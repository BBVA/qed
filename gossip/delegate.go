package gossip

import (
	"log"

	"github.com/hashicorp/memberlist"
)

// NodeTypeDelegate is an AliveDelegate...
type eventDelegate struct {
	node *Node
}

func (e *eventDelegate) NotifyJoin(n *memberlist.Node) {
	e.node.handleNodeJoin(n)
}

func (e *eventDelegate) NotifyLeave(n *memberlist.Node) {
	e.node.handleNodeLeave(n)
}

func (e *eventDelegate) NotifyUpdate(n *memberlist.Node) {
	e.node.handleNodeUpdate(n)
}

func NewFakeDelegate() DelegateBuilder {
	return func(n *Node) memberlist.Delegate {
		return &fakeDelegate{n}
	}
}

type fakeDelegate struct {
	node *Node
}

// NodeMeta is used to retrieve meta-data about the current node
// when broadcasting an alive message. It's length is limited to
// the given byte size. This metadata is available in the Node structure.
func (d *fakeDelegate) NodeMeta(limit int) []byte {
	meta, err := d.node.encodeMetadata()
	if err != nil {
		log.Fatalf("Unable to encode node metadata: %v", err)
	}
	return meta
}

// NotifyMsg is called when a user-data message is received.
// Care should be taken that this method does not block, since doing
// so would block the entire UDP packet receive loop. Additionally, the byte
// slice may be modified after the call returns, so it should be copied if needed
func (d *fakeDelegate) NotifyMsg([]byte) {}

// GetBroadcasts is called when user data messages can be broadcast.
// It can return a list of buffers to send. Each buffer should assume an
// overhead as provided with a limit on the total byte size allowed.
// The total byte size of the resulting data to send must not exceed
// the limit. Care should be taken that this method does not block,
// since doing so would block the entire UDP packet receive loop.
func (d *fakeDelegate) GetBroadcasts(overhead, limit int) [][]byte {
	return d.node.broadcasts.GetBroadcasts(overhead, limit)
}

// LocalState is used for a TCP Push/Pull. This is sent to
// the remote side in addition to the membership information. Any
// data can be sent here. See MergeRemoteState as well. The `join`
// boolean indicates this is for a join instead of a push/pull.
func (d *fakeDelegate) LocalState(join bool) []byte {
	return []byte{}
}

// MergeRemoteState is invoked after a TCP Push/Pull. This is the
// state received from the remote side and is the result of the
// remote side's LocalState call. The 'join'
// boolean indicates this is for a join instead of a push/pull.
func (d *fakeDelegate) MergeRemoteState(buf []byte, join bool) {}
