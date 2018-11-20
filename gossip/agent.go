/*
   Copyright 2018 Banco Bilbao Vizcaya Argentaria, n.A.
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
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/bbva/qed/log"
	"github.com/bbva/qed/protocol"
	"github.com/hashicorp/go-msgpack/codec"
	"github.com/hashicorp/memberlist"
)

type AgentMeta struct {
	Role AgentType
}

func (m *AgentMeta) Encode() ([]byte, error) {
	var buf bytes.Buffer
	encoder := codec.NewEncoder(&buf, &codec.MsgpackHandle{})
	if err := encoder.Encode(m); err != nil {
		log.Errorf("Failed to encode agent metadata: %v", err)
		return nil, err
	}
	return buf.Bytes(), nil
}

type Agent struct {
	config     *Config
	meta       *AgentMeta
	memberlist *memberlist.Memberlist
	broadcasts *memberlist.TransmitLimitedQueue

	topology     *Topology
	topologyLock sync.RWMutex

	stateLock sync.Mutex
	state     AgentState
}

// AgentState is the state of the Agent instance.
type AgentState int

const (
	AgentAlive AgentState = iota
	AgentLeaving
	AgentLeft
	AgentShutdown
)

func (s AgentState) String() string {
	switch s {
	case AgentAlive:
		return "alive"
	case AgentLeaving:
		return "leaving"
	case AgentLeft:
		return "left"
	case AgentShutdown:
		return "shutdown"
	default:
		return "unknown"
	}
}

type Topology struct {
	members map[AgentType]map[string]*Member
	sync.Mutex
}

func NewTopology() *Topology {
	members := make(map[AgentType]map[string]*Member)
	for i := 0; i < int(MaxType); i++ {
		members[AgentType(i)] = make(map[string]*Member)
	}
	return &Topology{
		members: members,
	}
}

func (t *Topology) Update(m *Member) error {
	t.Lock()
	defer t.Unlock()
	log.Debugf("Updating topology with member: %+v", m)
	t.members[m.Role][m.Name] = m
	return nil
}

func (t *Topology) Delete(m *Member) error {
	t.Lock()
	defer t.Unlock()
	delete(t.members[m.Role], m.Name)
	return nil
}

func (t *Topology) Get(kind AgentType) []*Member {
	t.Lock()
	defer t.Unlock()
	members := make([]*Member, 0)
	for _, member := range t.members[kind] {
		members = append(members, member)
	}
	return members
}

// Member is a single member of the gossip cluster.
type Member struct {
	Name   string
	Addr   net.IP
	Port   uint16
	Role   AgentType
	Status MemberStatus
}

// MemberStatus is the state that a member is in.
type MemberStatus int

const (
	StatusNone MemberStatus = iota
	StatusAlive
	StatusLeaving
	StatusLeft
	StatusFailed
)

func (s MemberStatus) String() string {
	switch s {
	case StatusNone:
		return "none"
	case StatusAlive:
		return "alive"
	case StatusLeaving:
		return "leaving"
	case StatusLeft:
		return "left"
	case StatusFailed:
		return "failed"
	default:
		panic(fmt.Sprintf("unknown MemberStatus: %d", s))
	}
}

type MessageHandler interface {
	HandleMsg([]byte)
}

type MessageHandlerBuilder func(*Agent) MessageHandler

type NopMessageHanler struct {
}

func (h *NopMessageHanler) HandleMsg([]byte) {}

func NewNopMessageHandler(*Agent) MessageHandler {
	return &NopMessageHanler{}
}

func Create(conf *Config, handler MessageHandlerBuilder) (agent *Agent, err error) {

	meta := &AgentMeta{
		Role: conf.Role,
	}

	agent = &Agent{
		config:   conf,
		meta:     meta,
		topology: NewTopology(),
		state:    AgentAlive,
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
	conf.MemberlistConfig.Delegate = newAgentDelegate(agent, handler(agent))
	conf.MemberlistConfig.Events = &eventDelegate{agent}

	agent.memberlist, err = memberlist.Create(conf.MemberlistConfig)
	if err != nil {
		return nil, err
	}

	// Print local member info
	localNode := agent.memberlist.LocalNode()
	log.Infof("Local member %s:%d", localNode.Addr, localNode.Port)

	// Set broadcast queue
	agent.broadcasts = &memberlist.TransmitLimitedQueue{
		NumNodes: func() int {
			return agent.memberlist.NumMembers()
		},
		RetransmitMult: 2,
	}

	return agent, nil
}

// Join asks the Agent instance to join.
func (a *Agent) Join(addrs []string) (int, error) {

	if a.State() != AgentAlive {
		return 0, fmt.Errorf("Agent can't join after Leave or Shutdown")
	}

	if len(addrs) > 0 {
		log.Debugf("Trying to join the cluster using members: %v", addrs)
		return a.memberlist.Join(addrs)
	}
	return 0, nil
}

func (a *Agent) Leave() error {

	// Check the current state
	a.stateLock.Lock()
	if a.state == AgentLeft {
		a.stateLock.Unlock()
		return nil
	} else if a.state == AgentLeaving {
		a.stateLock.Unlock()
		return fmt.Errorf("Leave already in progress")
	} else if a.state == AgentShutdown {
		a.stateLock.Unlock()
		return fmt.Errorf("Leave called after Shutdown")
	}
	a.state = AgentLeaving
	a.stateLock.Unlock()

	// Attempt the memberlist leave
	err := a.memberlist.Leave(a.config.BroadcastTimeout)
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
	if a.state != AgentShutdown {
		a.state = AgentLeft
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

	a.stateLock.Lock()
	defer a.stateLock.Unlock()

	if a.state == AgentShutdown {
		return nil
	}

	if a.state != AgentLeft {
		log.Info("agent: Shutdown without a Leave")
	}

	a.state = AgentShutdown
	err := a.memberlist.Shutdown()
	if err != nil {
		return err
	}

	return nil
}

func (a *Agent) Memberlist() *memberlist.Memberlist {
	return a.memberlist
}

func (a *Agent) Metadata() *AgentMeta {
	return a.meta
}

func (a *Agent) Broadcasts() *memberlist.TransmitLimitedQueue {
	return a.broadcasts
}

func (a *Agent) GetAddrPort() (net.IP, uint16) {
	n := a.memberlist.LocalNode()
	return n.Addr, n.Port
}

func (a *Agent) State() AgentState {
	a.stateLock.Lock()
	defer a.stateLock.Unlock()
	return a.state
}

func (a *Agent) getMember(peer *memberlist.Node) *Member {
	meta, err := a.decodeMetadata(peer.Meta)
	if err != nil {
		panic(err)
	}
	return &Member{
		Name: peer.Name,
		Addr: net.IP(peer.Addr),
		Port: peer.Port,
		Role: meta.Role,
	}
}

func (a *Agent) handleNodeJoin(peer *memberlist.Node) {
	member := a.getMember(peer)
	member.Status = StatusAlive
	a.topology.Update(member)
	log.Debugf("%s member joined: %s %s:%d", member.Role, member.Name, member.Addr, member.Port)
}

func (a *Agent) handleNodeLeave(peer *memberlist.Node) {
	member := a.getMember(peer)
	a.topology.Delete(member)
	log.Debugf("%s member left: %s %s:%d", member.Role, member.Name, member.Addr, member.Port)
}

func (a *Agent) handleNodeUpdate(peer *memberlist.Node) {
	// ignore
}

func (a *Agent) encodeMetadata() ([]byte, error) {
	var buf bytes.Buffer
	encoder := codec.NewEncoder(&buf, &codec.MsgpackHandle{})
	if err := encoder.Encode(a.meta); err != nil {
		log.Errorf("Failed to encode agent metadata: %v", err)
		return nil, err
	}
	return buf.Bytes(), nil
}

func (a *Agent) decodeMetadata(buf []byte) (*AgentMeta, error) {
	meta := &AgentMeta{}
	reader := bytes.NewReader(buf)
	decoder := codec.NewDecoder(reader, &codec.MsgpackHandle{})
	if err := decoder.Decode(meta); err != nil {
		log.Errorf("Failed to decode agent metadata: %v", err)
		return nil, err
	}
	return meta, nil
}

func memberToNode(members []*Member) []*memberlist.Node {
	list := make([]*memberlist.Node, 0)
	for _, m := range members {
		list = append(list, &memberlist.Node{Addr: m.Addr, Port: m.Port})
	}
	return list
}

func (a *Agent) GetPeers(max int, agentType AgentType, excluded *protocol.Source) []*memberlist.Node {

	fullList := a.topology.Get(agentType)

	var included []*Member
	if excluded != nil && agentType.String() == excluded.Role {
		included = excludePeers(fullList, excluded)
	} else {
		included = fullList
	}

	if len(included) <= max {
		return memberToNode(included)
	}

	var filteredList []*Member
	for i := 0; i < max; i++ {
		filteredList = append(filteredList, included[i])
	}

	return memberToNode(filteredList)
}

func excludePeers(peers []*Member, excluded *protocol.Source) []*Member {
	result := make([]*Member, 0)
	for _, p := range peers {
		if bytes.Equal(p.Addr, excluded.Addr) && p.Port == excluded.Port {
			continue
		}
		result = append(result, p)
	}
	return result
}
