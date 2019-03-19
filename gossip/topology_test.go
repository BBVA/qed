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

func setupTopology(size int) *Topology {
	topology := NewTopology()
	for i := 0; i < size; i++ {
		name := fmt.Sprintf("name%d", i)
		port, _ := strconv.Atoi(fmt.Sprintf("900%d", i))
		role := member.Type(i % int(member.Unknown))
		peer := member.NewPeer(name, "127.0.0.1", uint16(port), role)
		topology.Update(peer)
	}
	return topology
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

func TestShufflePeerList(t *testing.T) {
	list := setupPeerList(10)

	shuffled := list.Shuffle()

	require.Truef(t, 10 == shuffled.Size(), "The new list should have the same size")
	for _, e := range list.L {
		require.Containsf(t, shuffled.L, e, "The element should remain in the list")
	}
}

func TestUpdateAndDeleteTopology(t *testing.T) {
	topology := NewTopology()

	peer := member.NewPeer("auditor", "127.0.0.1", 9000, member.Auditor)
	topology.Update(peer)

	auditors := topology.Get(member.Auditor)
	require.Truef(t, 1 == auditors.Size(), "The topology must include one auditor")

	topology.Delete(peer)

	auditors = topology.Get(member.Auditor)
	require.Truef(t, 0 == auditors.Size(), "The topology must include zero auditor")

}

func TestEachWithoutExclusionsTopology(t *testing.T) {
	topology := setupTopology(10)

	each := topology.Each(1, nil)

	require.Truef(t, 3 == each.Size(), "It must include only 3 elements")

	auditors := each.Filter(func(m *member.Peer) bool {
		return m.Meta.Role == member.Auditor
	})
	monitors := each.Filter(func(m *member.Peer) bool {
		return m.Meta.Role == member.Monitor
	})
	publishers := each.Filter(func(m *member.Peer) bool {
		return m.Meta.Role == member.Publisher
	})

	require.Truef(t, 1 == auditors.Size(), "It must include only one auditor")
	require.Truef(t, 1 == monitors.Size(), "It must include only one monitor")
	require.Truef(t, 1 == publishers.Size(), "It must include only one publisher")
}

func TestEachWithExclusionsTopology(t *testing.T) {
	topology := setupTopology(10)

	excluded := topology.Get(member.Auditor)
	each := topology.Each(1, &excluded)

	require.Truef(t, 2 == each.Size(), "It must include only 2 elements")

	auditors := each.Filter(func(m *member.Peer) bool {
		return m.Meta.Role == member.Auditor
	})
	monitors := each.Filter(func(m *member.Peer) bool {
		return m.Meta.Role == member.Monitor
	})
	publishers := each.Filter(func(m *member.Peer) bool {
		return m.Meta.Role == member.Publisher
	})

	require.Truef(t, 0 == auditors.Size(), "It must not include any auditor")
	require.Truef(t, 1 == monitors.Size(), "It must include only one monitor")
	require.Truef(t, 1 == publishers.Size(), "It must include only one publisher")
}
