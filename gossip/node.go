package gossip

import (
	"bytes"
	"fmt"
	"net"
	"sync"

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
		RetransmitMult: 0,
	}

	return node, nil
}

// Join asks the Node instance to join.
func (n *Node) Join(addrs []string) (int, error) {
	if len(addrs) > 0 {
		log.Debugf("Trying to join the cluster using members: %v", addrs)
		return n.memberlist.Join(addrs)
	}
	return 0, nil
}

func (n *Node) Shutdown() error {
	err := n.memberlist.Shutdown()
	if err != nil {
		return err
	}
	return nil
}

func (n *Node) handleNodeJoin(peer *memberlist.Node) {
	meta, err := n.decodeMetadata(peer.Meta)
	if err != nil {
		panic(err)
	}
	member := &Member{
		Name:   peer.Name,
		Addr:   net.IP(peer.Addr),
		Port:   peer.Port,
		Role:   meta.Role,
		Status: StatusAlive,
	}
	n.topology.Update(member)
	log.Debugf("%s member joined: %s %s:%d", member.Role, member.Name, member.Addr, member.Port)
}

func (n *Node) handleNodeLeave(peer *memberlist.Node) {

}

func (n *Node) handleNodeUpdate(peer *memberlist.Node) {

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
