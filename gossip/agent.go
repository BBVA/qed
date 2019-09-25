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

// Package gossip implements functionality to build gossip agents
// and control their life cycle: start/stop, join/leave the gossip network,
// send messages, ...
package gossip

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/bbva/qed/client"
	"github.com/bbva/qed/log"
	"github.com/bbva/qed/metrics"
	"github.com/hashicorp/memberlist"
	"github.com/prometheus/client_golang/prometheus"
)

// Agent exposes the necesary API to interact with
// the gossip network, the snapshot store, the
// QED log and the alerts store.
//
// The agent API enables QED users to implement
// and integrate its own tools and services with
// QED.
type Agent struct {
	// stateLock is used to protect critical information
	// from cocurrent access from the gossip
	// network
	stateLock sync.Mutex

	// parameters from command line
	// interface
	config Config

	// Self stores the peer information corresponding to
	// this agent instance. It is used to make routing
	// decissions.
	Self *Peer

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

	// processors enqueue tasks to be executed by
	// the tasks manager. They need to create
	// the context for each task to be able to execute.
	processors map[string]Processor

	// timeout signals when the default timeout has passed
	// to end an enqueue operation
	timeout *time.Ticker

	// A cached KV to be used by processors and tasks
	Cache Cache

	// In channel receives messages from the gossip
	// network to be processed by the agent
	In MessageBus

	// Out channel enqueue the messages to be forwarded to
	// other gossip agents
	Out MessageBus

	// quitCh channels signal the graceful shutdown
	// of the agent
	quitCh chan bool

	// Client to a running QED
	Qed *client.HTTPClient

	//Client to a notification service
	Notifier Notifier

	// Client to a snapshot store service
	SnapshotStore SnapshotStore

	//Client to a task manager service
	Tasks TasksManager

	// Logger
	log log.Logger
}

// NewAgentFromConfig creates new agent from a configuration object.
// It does not create external clients like QED, SnapshotStore or Notifier, nor
// a task manager.
func NewAgentFromConfig(conf *Config) (agent *Agent, err error) {
	options, err := configToOptions(conf)
	if err != nil {
		return nil, err
	}
	return NewAgent(options...)
}

// NewAgentFromConfigWithLogger creates new agent from a configuration object.
// It does not create external clients like QED, SnapshotStore or Notifier, nor
// a task manager.
func NewAgentFromConfigWithLogger(conf *Config, l log.Logger) (agent *Agent, err error) {
	options, err := configToOptions(conf)
	if err != nil {
		return nil, err
	}
	options = append(options, SetLogger(l))
	return NewAgent(options...)
}

// NewDefaultAgent returns a new agent with all the APIs initialized and
// with a cache of size bytes.
func NewDefaultAgent(conf *Config, qed *client.HTTPClient, s SnapshotStore, t TasksManager, n Notifier, l log.Logger) (*Agent, error) {
	options, err := configToOptions(conf)
	if err != nil {
		return nil, err
	}
	options = append(options, SetQEDClient(qed), SetSnapshotStore(s), SetTasksManager(t), SetNotifier(n), SetLogger(l))
	return NewAgent(options...)
}

// NewAgent returns a configured and started agent or error if
// it cannot be created.
// On return, the agent is already connected to the gossip network
// but it will not process any information.
// It will though enqueue request as soon as it is created. When those
// queues are full, messages will start to be dropped silently.
func NewAgent(options ...AgentOptionF) (*Agent, error) {
	agent := &Agent{
		quitCh:   make(chan bool),
		topology: NewTopology(),
		log:      log.L(),
	}

	// Run the options on the client
	for _, option := range options {
		if err := option(agent); err != nil {
			return nil, err
		}
	}

	// Set message buses
	agent.In = MessageBus{log: agent.log.Named("bus")}
	agent.Out = MessageBus{log: agent.log.Named("bus")}

	bindIP, bindPort, err := agent.config.AddrParts(agent.config.BindAddr)
	if err != nil {
		return nil, fmt.Errorf("Invalid bind address: %s", err)
	}

	var advertiseIP string
	var advertisePort int
	if agent.config.AdvertiseAddr != "" {
		advertiseIP, advertisePort, err = agent.config.AddrParts(agent.config.AdvertiseAddr)
		if err != nil {
			return nil, fmt.Errorf("Invalid advertise address: %s", err)
		}
	}

	agent.config.MemberlistConfig = memberlist.DefaultLocalConfig()
	agent.config.MemberlistConfig.BindAddr = bindIP
	agent.config.MemberlistConfig.BindPort = bindPort
	agent.config.MemberlistConfig.AdvertiseAddr = advertiseIP
	agent.config.MemberlistConfig.AdvertisePort = advertisePort
	agent.config.MemberlistConfig.Name = agent.config.NodeName
	agent.config.MemberlistConfig.Logger = agent.log.StdLogger(&log.StdLoggerOptions{InferLevels: true})

	// Configure delegates
	agent.config.MemberlistConfig.Delegate = newAgentDelegate(agent, agent.log)
	agent.config.MemberlistConfig.Events = &eventDelegate{agent, agent.log}

	agent.Self = NewPeer(agent.config.NodeName, advertiseIP, uint16(advertisePort), agent.config.Role)

	return agent, nil
}

// Enables the processing engines of the
// agent
func (a *Agent) Start() {
	var err error

	if a.metrics != nil {
		a.log.Infof("Starting agent metrics server")
		go func() {
			if err := a.metrics.Start(); err != http.ErrServerClosed {
				a.log.Fatalf("Can't start metrics HTTP server: %s", err)
			}
		}()
	}

	if a.Tasks != nil {
		a.log.Infof("Starting task mamanger loop")
		a.Tasks.Start()
	}

	if a.Notifier != nil {
		a.log.Infof("Starting notifier mamanger loop")
		a.Notifier.Start()
	}

	a.log.Infof("Starting memberlist gossip netwotk")
	a.gossip, err = memberlist.Create(a.config.MemberlistConfig)
	if err != nil {
		a.log.Infof("Error creating the memberlist network; %v", err)
		return
	}

	// Print local member info
	a.Self, err = ParsePeer(a.gossip.LocalNode())
	if err != nil {
		a.log.Fatalf("Cannot parse peer: %v", err)
	}
	a.log.Infof("Local member %+v", a.Self)

	// Set broadcast queue
	a.broadcasts = &memberlist.TransmitLimitedQueue{
		NumNodes: func() int {
			return a.gossip.NumMembers()
		},
		RetransmitMult: 2,
	}

	if len(a.config.StartJoin) > 0 {
		a.log.Infof("Trying to joing gossip network with peers %v", a.config.StartJoin)
		n, err := a.Join(a.config.StartJoin)
		if n == 0 || err != nil {
			a.log.Fatalf("Unable to join gossip network because %v", err)
		}
		a.log.Infof("Joined gossip network with %d peers", n)
	}

	a.log.Infof("Starting agent sender loop")
	a.sender()
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

// Register a slice of collectors in the agent metrics server
func (a *Agent) RegisterMetrics(cs []prometheus.Collector) {
	a.metrics.MustRegister(cs...)
}

// Registers the agent in all output channels to send
// all the messages in the bus to other peers.
//
// Sender will create MaxSenders goroutines to send
// messages for all channels
func (a *Agent) sender() {
	var wg sync.WaitGroup
	var counter int
	for i := 0; i < MAXMESSAGEID; i++ {
		ch := make(chan *Message, 255)
		a.Out.pool[i] = append(a.Out.pool[i], ch)

		go func(ch chan *Message) {
			for {
				select {
				case msg := <-ch:
					// as soon as we have a batch ready for retransmission, we try to send
					// it after applying all the routing contraints
					wg.Add(1)
					go func() {
						defer wg.Done()
						a.log.Debugf("Agent sender loop: sending msg!")
						a.Send(msg)
					}()
					if counter >= a.config.MaxSenders {
						wg.Wait()
						counter = 0
					}
					counter++
				case <-a.quitCh:
					return
				}
			}
		}(ch)
	}
}

// Sends a batch using the gossip network reliable transport
// to  other nodes based on the routing policy applied
func (a *Agent) Send(msg *Message) {
	// if ttl is 0, the message dies here
	if msg.TTL == 0 {
		return
	}

	msg.TTL--
	wire, err := msg.Encode()
	if err != nil {
		a.log.Infof("Agent Send unable to encode message to gossip it")
		return
	}
	msg.From = a.Self
	for _, dst := range a.route(msg.From) {
		a.log.Debugf("Sending batch to %+v\n", dst.Name)
		_ = a.gossip.SendReliable(dst, wire)
	}
}

// Returns the list of nodes to which a batch can be sent
// given the source of the communication and the internal
// agent topology.
func (a *Agent) route(src *Peer) []*memberlist.Node {
	var excluded PeerList

	dst := make([]*memberlist.Node, 0)

	excluded.L = append(excluded.L, src)
	excluded.L = append(excluded.L, a.Self)

	peers := a.topology.Each(1, &excluded)
	for _, p := range peers.L {
		dst = append(dst, p.Node())
	}
	return dst
}

// Join asks the Agent instance to join
// the nodes with the give addrs addresses.
func (a *Agent) Join(addrs []string) (int, error) {
	if a.State() != AgentStatusAlive {
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
	switch a.Self.Status {
	case AgentStatusLeft:
		a.stateLock.Unlock()
		return nil
	case AgentStatusLeaving:
		a.stateLock.Unlock()
		return fmt.Errorf("Leave already in progress")
	case AgentStatusShutdown:
		a.stateLock.Unlock()
		return fmt.Errorf("Leave called after Shutdown")
	default:
		a.Self.Status = AgentStatusLeaving
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

	// Transition to AgentStatusLeft only if we not already shutdown
	a.stateLock.Lock()
	if a.Self.Status != AgentStatusShutdown {
		a.Self.Status = AgentStatusLeft
	}
	a.stateLock.Unlock()
	return nil

}

// AgentStatusShutdown forcefully shuts down the Agent instance, stopping all network
// activity and background maintenance associated with the instance.
//
// This is not a graceful shutdown, and should be preceded by a call
// to Leave. Otherwise, other agents in the cluster will detect this agent's
// exit as a agent failure.
//
// It is safe to call this method multiple times.
func (a *Agent) Shutdown() error {
	a.log.Infof("Shutting down agent %s", a.config.NodeName)
	a.stateLock.Lock()
	defer a.stateLock.Unlock()

	if a.Self.Status == AgentStatusShutdown {
		return nil
	}

	if a.Self.Status != AgentStatusLeft {
		a.log.Info("agent: Shutdown without a Leave")
	}

	a.Self.Status = AgentStatusShutdown
	err := a.gossip.Shutdown()
	if err != nil {
		return err
	}
	close(a.quitCh)

	if a.metrics != nil {
		a.metrics.Shutdown()
	}

	if a.Tasks != nil {
		a.Tasks.Stop()
	}

	if a.Notifier != nil {
		a.Notifier.Stop()
	}

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
func (a *Agent) State() Status {
	a.stateLock.Lock()
	defer a.stateLock.Unlock()
	return a.Self.Status
}
