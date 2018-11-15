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
	"github.com/hashicorp/go-msgpack/codec"
	"github.com/hashicorp/memberlist"
)

type NodeMeta struct {
	Role NodeType
}

type Node struct {
	config     *Config
	meta       *NodeMeta
	memberlist *memberlist.Memberlist
	broadcasts *memberlist.TransmitLimitedQueue

	topology     *Topology
	topologyLock sync.RWMutex

	stateLock sync.Mutex
	state     NodeState
}

// NodeState is the state of the Node instance.
type NodeState int

const (
	NodeAlive NodeState = iota
	NodeLeaving
	NodeLeft
	NodeShutdown
)

func (s NodeState) String() string {
	switch s {
	case NodeAlive:
		return "alive"
	case NodeLeaving:
		return "leaving"
	case NodeLeft:
		return "left"
	case NodeShutdown:
		return "shutdown"
	default:
		return "unknown"
	}
}

type Topology struct {
	members map[NodeType]map[string]*Member
	sync.Mutex
}

func NewTopology() *Topology {
	members := make(map[NodeType]map[string]*Member)
	for i := 0; i < int(MaxType); i++ {
		members[NodeType(i)] = make(map[string]*Member)
	}
	return &Topology{
		members: members,
	}
}

func (t *Topology) Update(m *Member) error {
	t.Lock()
	defer t.Unlock()
	t.members[m.Role][m.Name] = m
	return nil
}

func (t *Topology) Delete(m *Member) error {
	t.Lock()
	defer t.Unlock()
	delete(t.members[m.Role], m.Name)
	return nil
}

func (t *Topology) Get(kind NodeType) []*Member {
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
	Role   NodeType
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

type DelegateBuilder func(*Node) memberlist.Delegate

func Create(conf *Config, delegate DelegateBuilder) (node *Node, err error) {

	meta := &NodeMeta{
		Role: conf.Role,
	}

	node = &Node{
		config:   conf,
		meta:     meta,
		topology: NewTopology(),
		state:    NodeAlive,
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
	conf.MemberlistConfig.Delegate = delegate(node)
	conf.MemberlistConfig.Events = &eventDelegate{node}

	node.memberlist, err = memberlist.Create(conf.MemberlistConfig)
	if err != nil {
		return nil, err
	}

	// Print local member info
	localNode := node.memberlist.LocalNode()
	log.Infof("Local member %s:%d", localNode.Addr, localNode.Port)

	// Set broadcast queue
	node.broadcasts = &memberlist.TransmitLimitedQueue{
		NumNodes: func() int {
			return node.memberlist.NumMembers()
		},
		RetransmitMult: 2,
	}

	return node, nil
}

// Join asks the Node instance to join.
func (n *Node) Join(addrs []string) (int, error) {

	if n.State() != NodeAlive {
		return 0, fmt.Errorf("Node can't join after Leave or Shutdown")
	}

	if len(addrs) > 0 {
		log.Debugf("Trying to join the cluster using members: %v", addrs)
		return n.memberlist.Join(addrs)
	}
	return 0, nil
}

func (n *Node) Leave() error {

	// Check the current state
	n.stateLock.Lock()
	if n.state == NodeLeft {
		n.stateLock.Unlock()
		return nil
	} else if n.state == NodeLeaving {
		n.stateLock.Unlock()
		return fmt.Errorf("Leave already in progress")
	} else if n.state == NodeShutdown {
		n.stateLock.Unlock()
		return fmt.Errorf("Leave called after Shutdown")
	}
	n.state = NodeLeaving
	n.stateLock.Unlock()

	// Attempt the memberlist leave
	err := n.memberlist.Leave(n.config.BroadcastTimeout)
	if err != nil {
		return err
	}

	// Wait for the leave to propagate through the cluster. The broadcast
	// timeout is how long we wait for the message to go out from our own
	// queue, but this wait is for that message to propagate through the
	// cluster. In particular, we want to stay up long enough to service
	// any probes from other nodes before they learn about us leaving.
	time.Sleep(n.config.LeavePropagateDelay)

	// Transition to Left only if we not already shutdown
	n.stateLock.Lock()
	if n.state != NodeShutdown {
		n.state = NodeLeft
	}
	n.stateLock.Unlock()
	return nil

}

// Shutdown forcefully shuts down the Node instance, stopping all network
// activity and background maintenance associated with the instance.
//
// This is not a graceful shutdown, and should be preceded by a call
// to Leave. Otherwise, other nodes in the cluster will detect this node's
// exit as a node failure.
//
// It is safe to call this method multiple times.
func (n *Node) Shutdown() error {

	n.stateLock.Lock()
	defer n.stateLock.Unlock()

	if n.state == NodeShutdown {
		return nil
	}

	if n.state != NodeLeft {
		log.Info("node: Shutdown without a Leave")
	}

	n.state = NodeShutdown
	err := n.memberlist.Shutdown()
	if err != nil {
		return err
	}

	return nil
}

func (n *Node) Memberlist() *memberlist.Memberlist {
	return n.memberlist
}

func (n *Node) State() NodeState {
	n.stateLock.Lock()
	defer n.stateLock.Unlock()
	return n.state
}

func (n *Node) getMember(peer *memberlist.Node) *Member {
	meta, err := n.decodeMetadata(peer.Meta)
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

func (n *Node) handleNodeJoin(peer *memberlist.Node) {
	member := n.getMember(peer)
	member.Status = StatusAlive
	n.topology.Update(member)
	log.Debugf("%s member joined: %s %s:%d", member.Role, member.Name, member.Addr, member.Port)
}

func (n *Node) handleNodeLeave(peer *memberlist.Node) {
	member := n.getMember(peer)
	n.topology.Delete(member)
	log.Debugf("%s member left: %s %s:%d", member.Role, member.Name, member.Addr, member.Port)
}

func (n *Node) handleNodeUpdate(peer *memberlist.Node) {
	// ignore
}

func (n *Node) encodeMetadata() ([]byte, error) {
	var buf bytes.Buffer
	encoder := codec.NewEncoder(&buf, &codec.MsgpackHandle{})
	if err := encoder.Encode(n.meta); err != nil {
		log.Errorf("Failed to encode node metadata: %v", err)
		return nil, err
	}
	return buf.Bytes(), nil
}

func (n *Node) decodeMetadata(buf []byte) (*NodeMeta, error) {
	meta := &NodeMeta{}
	reader := bytes.NewReader(buf)
	decoder := codec.NewDecoder(reader, &codec.MsgpackHandle{})
	if err := decoder.Decode(meta); err != nil {
		log.Errorf("Failed to decode node metadata: %v", err)
		return nil, err
	}
	return meta, nil
}
