/*
   Copyright 2018-2019 Banco Bilbao Vizcaya Argentaria, S.A.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/
package gossip

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/bbva/qed/client"
	"github.com/bbva/qed/gossip/member"
	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/log"
	"github.com/bbva/qed/metrics"
	"github.com/bbva/qed/protocol"
	"github.com/coocood/freecache"
	"github.com/hashicorp/memberlist"
	"github.com/prometheus/client_golang/prometheus"
)

// hashedBatch contains a the received
// batch from the gossip network, and also
// a digest of its contents to univocally
// identify it.
type hashedBatch struct {
	batch  *protocol.BatchSnapshots
	digest hashing.Digest
}

// Agent exposes the necesary API to interact with
// the gossip network, the snapshot store, the
// QED log and the alerts store to the processors.
type Agent struct {
	// stateLock is used to protect critical information
	// from cocurrent access from the gossip
	// network
	stateLock sync.Mutex

	// config stores parameters from command line
	// interface
	config *Config

	// Self stores the peer information corresponding to
	// this agent instance. It is used to make routing
	// decissions.
	self *member.Peer

	// metricsServer exposes an HTTP service with
	// all the its metrics and also its processors
	// metrics.
	metrics *metrics.Server

	// gossip gives access to the
	// memberlist API to interact with the network
	// and its members
	gossip *memberlist.Memberlist

	// broadcasts gives access to the
	// broadcast gossip API
	broadcasts *memberlist.TransmitLimitedQueue

	// Topology holds the network topology
	// as this agent instance sees it.
	topology *Topology

	// taskManager is in charge of tasks execution.
	tasks *TasksManager

	// processed cache contains a reference to the last
	// messages sent to processors. This is used to
	// deduplicate messages coming from the network.
	// which has been already processed.
	processed *freecache.Cache

	// processors enqueue tasks to be executed by
	// the tasks manager. They need to create
	// the context for each task to be able to execute.
	processors map[string]Processor

	// In channel receives hashed batches from the gossip
	// network to be processed by the agent
	inCh chan *hashedBatch

	// Out channel enqueue the batches to be forwarded to
	// other gossip agents
	outCh chan *protocol.BatchSnapshots

	// alerts channel enqueue the alerts messages to be sent
	// to the alerting service
	alertsCh chan string

	// quitCh channels signal the graceful shutdown
	// of the agent
	quitCh chan bool

	// timeout signals when the default timeout has passed
	// to end an enqueue operation
	timeout *time.Ticker

	// qeq client
	qed *client.HTTPClient

	// store client
	snapshotStore SnapshotStore
}

// NewAgent returns a configured and started agent or error if
// it cannot be created.
// On return, the agent is already connected to the gossip network
// but it will not process any information until start is called.
// It will though enqueue request as soon as it is created. When those
// queues are full, messages will start to be dropped silently.
func NewAgent(conf *Config, processors map[string]Processor, m *metrics.Server) (agent *Agent, err error) {
	log.Infof("New agent %s\n", conf.NodeName)

	transport := http.DefaultTransport.(*http.Transport)
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: false}
	httpClient := http.DefaultClient
	httpClient.Transport = transport
	qed, err := client.NewHTTPClient(
		client.SetHttpClient(httpClient),
		client.SetURLs(conf.QEDUrls[0], conf.QEDUrls[1:]...),
		client.SetAPIKey(conf.APIKey),
	)
	if err != nil {
		return nil, err
	}

	agent = &Agent{
		config:        conf,
		metrics:       m,
		topology:      NewTopology(),
		processors:    processors,
		processed:     freecache.NewCache(1 << 20),
		tasks:         NewTasksManager(200*time.Millisecond, 10, 100*time.Millisecond),
		inCh:          make(chan *hashedBatch, 1<<16),
		outCh:         make(chan *protocol.BatchSnapshots, 1<<16),
		alertsCh:      make(chan string, 100),
		timeout:       time.NewTicker(conf.TimeoutQueues),
		quitCh:        make(chan bool),
		qed:           qed,
		snapshotStore: NewRestSnapshotStore(conf.SnapshotStoreUrls),
	}

	bindIP, bindPort, err := conf.AddrParts(conf.BindAddr)
	if err != nil {
		return nil, fmt.Errorf("Invalid bind address: %s", err)
	}

	var advertiseIP string
	var advertisePort int
	if conf.AdvertiseAddr != "" {
		advertiseIP, advertisePort, err = conf.AddrParts(conf.AdvertiseAddr)
		if err != nil {
			return nil, fmt.Errorf("Invalid advertise address: %s", err)
		}
	}

	conf.MemberlistConfig = memberlist.DefaultLocalConfig()
	conf.MemberlistConfig.BindAddr = bindIP
	conf.MemberlistConfig.BindPort = bindPort
	conf.MemberlistConfig.AdvertiseAddr = advertiseIP
	conf.MemberlistConfig.AdvertisePort = advertisePort
	conf.MemberlistConfig.Name = conf.NodeName
	conf.MemberlistConfig.Logger = log.GetLogger()

	// Configure delegates
	conf.MemberlistConfig.Delegate = newAgentDelegate(agent)
	conf.MemberlistConfig.Events = &eventDelegate{agent}

	agent.self = member.NewPeer(conf.NodeName, advertiseIP, uint16(advertisePort), conf.Role)

	agent.gossip, err = memberlist.Create(conf.MemberlistConfig)
	if err != nil {
		return nil, err
	}

	// Print local member info
	agent.self = member.ParsePeer(agent.gossip.LocalNode())
	log.Infof("Local member %+v", agent.self)

	// Set broadcast queue
	agent.broadcasts = &memberlist.TransmitLimitedQueue{
		NumNodes: func() int {
			return agent.gossip.NumMembers()
		},
		RetransmitMult: 2,
	}

	// register metrics of each processor
	for _, p := range processors {
		agent.RegisterMetrics(p.Metrics())
	}

	return agent, nil
}

// Enables the processing engines of the
// agent
func (a *Agent) Start() {
	a.metrics.Start()
	a.processor()
	a.alerter()
}

// Register a new processor into the agent, to add some tasks per batch
// to be executed by the task manager.
func (a *Agent) RegisterProcessor(name string, p Processor) {
	a.stateLock.Lock()
	defer a.stateLock.Unlock()
	a.processors[name] = p
	a.RegisterMetrics(p.Metrics())
}

// Deregister a processor per name. It will not fail if the
// processor does not exist.
func (a *Agent) DeregisterProcessor(name string) {
	a.stateLock.Lock()
	defer a.stateLock.Unlock()
	delete(a.processors, name)
}

// Enqueue a hashed batch in the input channel
// or returns a timeout error
func (a *Agent) HashedBatch(h *hashedBatch) error {
	for {
		select {
		case <-a.timeout.C:
			return ChTimedOut
		case a.inCh <- h:
			return nil
		}
	}
}

// Enqueue a task into the task manager
// or return a timeout error
func (a *Agent) Task(t Task) error {
	return a.tasks.Add(t)
}

// Returns the agent qued client already
// configured
func (a *Agent) Qued() *client.HTTPClient {
	return a.qed
}

// returns a snapshot store instance already
// configured
func (a *Agent) SnapshotStore() SnapshotStore {
	return a.snapshotStore
}

// Register a slice of collectors in the agent metrics server
func (a *Agent) RegisterMetrics(cs []prometheus.Collector) {
	for _, c := range cs {
		a.metrics.Register(c)
	}
}

// Alert enqueue a message into the alerts
// queue to be sent by the agent to the alerting
// service
func (a *Agent) Alert(msg string) error {
	for {
		select {
		case <-a.timeout.C:
			return ChTimedOut
		case a.alertsCh <- msg:
			return nil
		}
	}
}

// alerters posts using the defult http client an alert message
// to a random url selected from the configuration list of urls.
//
// The connection timeout is 200 ms and the response read
// timeout is another 200 ms
//
// The alerter will end its goroutines when the agent quit channel
// is closed.
func (a *Agent) alerter() {

	client := &http.Client{
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				// timeout calling the server
				conn, err := net.DialTimeout(netw, addr, 200*time.Millisecond)
				if err != nil {
					return nil, err
				}
				// timeout reading from the connection
				conn.SetDeadline(time.Now().Add(200 * time.Millisecond))
				return conn, nil
			},
		}}

	go func() {
		for {
			select {
			case msg := <-a.alertsCh:
				n := len(a.config.AlertsServiceUrls)
				server := a.config.AlertsServiceUrls[0]
				if n > 1 {
					server = a.config.AlertsServiceUrls[rand.Intn(n)]
				}

				resp, err := client.Post(server, "application/json", bytes.NewBufferString(msg))
				if err != nil {
					log.Infof("Agent had an error sending the alert %v because %v ", msg, err)
					continue
				}
				defer resp.Body.Close()
				_, err = io.Copy(ioutil.Discard, resp.Body)
				if err != nil {
					log.Infof("Agent had the error %v when reading the response from the alert %v ", err, msg)
				}
			case <-a.quitCh:
				return
			}
		}
	}()
}

// Deduplicates received hashed batches,
// launch processors to generate the needed tasks
// and enqueue the batch to be forwarded by the gossip
// network.
//
// Processors are launched on separated goroutines and
// there is no limit on how many processor are enqueing tasks
// at any given moment.
//
// The processor will end when the general quitCh from the agent
// is closed.
func (a *Agent) processor() {
	go func() {
		for {
			select {
			case hashedBatch := <-a.inCh:
				// batches are idenfified by its digest. If a batch is already
				// processed, we do not process it again, and we do not retransmit it
				// again
				_, err := a.processed.Get(hashedBatch.digest)
				if err == nil {
					continue
				}
				a.processed.Set(hashedBatch.digest, []byte{0x0}, 0)

				// each batch is sent to all the agent processors which
				// will generate the tasks to be executed.
				ctx := context.WithValue(context.Background(), "batch", hashedBatch.batch)
				for _, p := range a.processors {
					p.Process(a, ctx)
				}

				// we do not wait for the processors to finish to enqueue
				// the batch into the out queue
				for {
					select {
					case <-a.timeout.C:
						log.Infof("Agent timed out enqueuing batch in out channel")
					case a.outCh <- hashedBatch.batch:
					}
				}
			case b := <-a.outCh:
				// as soon as we have a batch ready for retransmission, we try to send
				// it after applying all the routing contraints
				go a.Send(b)
			case <-a.quitCh:
				return
			}
		}
	}()
}

// Sends a batch using the gossip network reliable transport
// to  other nodes based on the routing policy applied
func (a *Agent) Send(batch *protocol.BatchSnapshots) {

	if batch.TTL <= 0 {
		return
	}

	batch.TTL -= 1
	from := batch.From
	batch.From = a.self
	msg, _ := batch.Encode()
	for _, dst := range a.route(from) {
		log.Debugf("Sending batch to %+v\n", dst.Name)
		a.gossip.SendReliable(dst, msg)
	}
}

// Returns the list of nodes to which a batch can be sent
// given the source of the communication and the internal
// agent topology.
func (a *Agent) route(src *member.Peer) []*memberlist.Node {
	var excluded PeerList

	dst := make([]*memberlist.Node, 0)

	excluded.L = append(excluded.L, src)
	excluded.L = append(excluded.L, a.self)

	peers := a.topology.Each(1, &excluded)
	for _, p := range peers.L {
		dst = append(dst, p.Node())
	}
	return dst
}

// Join asks the Agent instance to join
// the nodes with the give addrs addresses.
func (a *Agent) Join(addrs []string) (int, error) {
	if a.State() != member.Alive {
		return 0, fmt.Errorf("Agent can't join after Leave or Shutdown")
	}

	if len(addrs) > 0 {
		return a.gossip.Join(addrs)
	}

	return 0, nil
}

// Leave ask the agent to leave the gossip
// network gracefully, communicating to others
// this agent want to leave
func (a *Agent) Leave() error {

	// Check the current state
	a.stateLock.Lock()
	switch a.self.Status {
	case member.Left:
		a.stateLock.Unlock()
		return nil
	case member.Leaving:
		a.stateLock.Unlock()
		return fmt.Errorf("Leave already in progress")
	case member.Shutdown:
		a.stateLock.Unlock()
		return fmt.Errorf("Leave called after Shutdown")
	default:
		a.self.Status = member.Leaving
		a.stateLock.Unlock()
	}

	// Attempt the memberlist leave
	err := a.gossip.Leave(a.config.BroadcastTimeout)
	if err != nil {
		return err
	}

	// Wait for the leave to propagate through the cluster. The broadcast
	// timeout is how long we wait for the message to go out from our own
	// queue, but this wait is for that message to propagate through the
	// cluster. In particular, we want to stay up long enough to service
	// any probes from other agents before they learn about us leaving.
	time.Sleep(a.config.LeavePropagateDelay)

	// Transition to Left only if we not already shutdown
	a.stateLock.Lock()
	if a.self.Status != member.Shutdown {
		a.self.Status = member.Left
	}
	a.stateLock.Unlock()
	return nil

}

// Shutdown forcefully shuts down the Agent instance, stopping all network
// activity and background maintenance associated with the instance.
//
// This is not a graceful shutdown, and should be preceded by a call
// to Leave. Otherwise, other agents in the cluster will detect this agent's
// exit as a agent failure.
//
// It is safe to call this method multiple times.
func (a *Agent) Shutdown() error {
	log.Infof("Shutting down agent %s", a.config.NodeName)
	a.stateLock.Lock()
	defer a.stateLock.Unlock()

	a.metrics.Shutdown()

	if a.self.Status == member.Shutdown {
		return nil
	}

	if a.self.Status != member.Left {
		log.Info("agent: Shutdown without a Leave")
	}

	a.self.Status = member.Shutdown
	err := a.gossip.Shutdown()
	if err != nil {
		return err
	}
	close(a.quitCh)
	return nil
}

// Returns the memberlist object to manage the gossip api
// directly
func (a *Agent) Memberlist() *memberlist.Memberlist {
	return a.gossip
}

// Returns the broadcast facility to manage broadcasts messages
// directly
func (a *Agent) Broadcasts() *memberlist.TransmitLimitedQueue {
	return a.broadcasts
}

// Returns this agent IP address and port
func (a *Agent) GetAddrPort() (net.IP, uint16) {
	n := a.gossip.LocalNode()
	return n.Addr, n.Port
}

// Returns this agent status. This can be used to
// check if we should stop doing something based
// on the state of the agent in the gossip network.
func (a *Agent) State() member.Status {
	a.stateLock.Lock()
	defer a.stateLock.Unlock()
	return a.self.Status
}
