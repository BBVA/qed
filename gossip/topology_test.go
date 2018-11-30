package gossip

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/bbva/qed/gossip/member"
	"github.com/stretchr/testify/require"
)

func setupPeerList(size int) *PeerList {
	peers := make([]*member.Peer, 0)
	for i := 0; i < size; i++ {
		name := fmt.Sprintf("name%d", i)
		port, _ := strconv.Atoi(fmt.Sprintf("900%d", i))
		role := member.Type(i % int(member.Unknown))
		peer := member.NewPeer(name, "127.0.0.1", uint16(port), role)
		peers = append(peers, peer)
	}
	return &PeerList{peers}
}

func TestFilterPeerList(t *testing.T) {
	list := setupPeerList(10)

	// filter auditor types
	filtered := list.Filter(func(m *member.Peer) bool {
		return m.Meta.Role == member.Auditor
	})

	require.Truef(t, list.Size() > filtered.Size(), "The filtered list should have less elements")
	for _, e := range filtered.L {
		require.Truef(t, member.Auditor == e.Meta.Role, "The role cannot be different to Auditor")
	}
}

func TestExcludePeerList(t *testing.T) {
	list := setupPeerList(10)

	// exclude auditors
	filtered := list.Filter(func(m *member.Peer) bool {
		return m.Meta.Role == member.Auditor
	})
	included := list.Exclude(filtered)

	require.Truef(t, list.Size() > included.Size(), "The included list should have less elements")
	for _, e := range included.L {
		require.Truef(t, member.Auditor != e.Meta.Role, "The role cannot be Auditor")
	}
}

func TestExcludeNonIncludedPeerList(t *testing.T) {
	list := setupPeerList(10)

	// exclude unknown
	var uknown PeerList
	uknown.L = append(uknown.L, member.NewPeer("uknown", "127.0.0.1", 10000, member.Unknown))
	included := list.Exclude(&uknown)

	require.Truef(t, list.Size() == included.Size(), "The included list should have the same size")
	for _, e := range included.L {
		require.Truef(t, member.Unknown != e.Meta.Role, "The role cannot be Unknown")
	}
}

func TestAllPeerList(t *testing.T) {
	list := setupPeerList(10)

	all := list.All()

	require.Equalf(t, list, &all, "The lists should be equal")
}

func TestSizePeerList(t *testing.T) {
	list := setupPeerList(10)

	require.Equalf(t, 10, list.Size(), "The size should match")
}

func TestUpdatePeerList(t *testing.T) {
	var list PeerList

	list.Update(member.NewPeer("auditor1", "127.0.0.1", 9001, member.Auditor))
	require.Equalf(t, 1, list.Size(), "The size should have been incremented by 1")

	list.Update(member.NewPeer("auditor2", "127.0.0.1", 9002, member.Auditor))
	require.Equalf(t, 2, list.Size(), "The size should have been incremented by 2")

	// update the previous one
	list.Update(member.NewPeer("auditor2", "127.0.0.1", 9002, member.Auditor))
	require.Equalf(t, 2, list.Size(), "The size should have been incremented by 2")

	// update the previous one changing status
	p := member.NewPeer("auditor2", "127.0.0.1", 9002, member.Auditor)
	p.Status = member.Leaving
	list.Update(p)
	require.Equalf(t, 2, list.Size(), "The size should have been incremented by 2")
	require.Equalf(t, member.Leaving, list.L[1].Status, "The status should have been updated")
}

func TestDeletePeerList(t *testing.T) {
	list := setupPeerList(10)
	list2 := setupPeerList(10)

	// filter auditor types
	auditors := list.Filter(func(m *member.Peer) bool {
		return m.Meta.Role == member.Auditor
	})

	// delete auditors
	for _, e := range auditors.L {
		list2.Delete(e)
	}

	require.Truef(t, 10 > list2.Size(), "The new list should have less elements")
	for _, e := range list2.L {
		require.Truef(t, member.Auditor != e.Meta.Role, "The role cannot be Auditor")
	}
}

func TestDeleteNotIncludedPeerList(t *testing.T) {
	list := setupPeerList(10)

	list.Delete(member.NewPeer("unknown", "127.0.0.1", 10000, member.Unknown))

	require.Truef(t, 10 == list.Size(), "The new list should have the same size")

}
