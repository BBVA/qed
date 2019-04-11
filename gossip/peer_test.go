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
	"testing"

	"github.com/stretchr/testify/require"
)

var roles = []string{
	"auditor",
	"monitor",
	"publisher",
}

func setupPeerList(size int) *PeerList {
	peers := make([]*Peer, 0)
	for i := 0; i < size; i++ {
		name := fmt.Sprintf("name%d", i)
		port := uint16(9000 + i)
		role := roles[i%len(roles)]
		peer := NewPeer(name, "127.0.0.1", port, role)
		peers = append(peers, peer)
	}
	return &PeerList{peers}
}

func TestFilterPeerList(t *testing.T) {
	list := setupPeerList(10)

	filtered := list.Filter(func(m *Peer) bool {
		return m.Meta.Role == "auditor"
	})

	require.Truef(t, list.Size() > filtered.Size(), "The filtered list should have less elements")
	for _, e := range filtered.L {
		require.Truef(t, "auditor" == e.Meta.Role, "The role cannot be different to Auditor")
	}
}

func TestExcludePeerList(t *testing.T) {
	list := setupPeerList(10)

	// exclude auditors
	filtered := list.Filter(func(m *Peer) bool {
		return m.Meta.Role == "auditor"
	})
	included := list.Exclude(filtered)

	require.Truef(t, list.Size() > included.Size(), "The included list should have less elements")
	for _, e := range included.L {
		require.Truef(t, "auditor" != e.Meta.Role, "The role cannot be Auditor")
	}
}

func TestExcludeNonIncludedPeerList(t *testing.T) {
	list := setupPeerList(10)

	// exclude unknown
	var uknown PeerList
	uknown.L = append(uknown.L, NewPeer("uknown", "127.0.0.1", 10000, "unknown"))
	included := list.Exclude(&uknown)

	require.Truef(t, list.Size() == included.Size(), "The included list should have the same size")
	for _, e := range included.L {
		require.Truef(t, "unknown" != e.Meta.Role, "The role cannot be Unknown")
	}
}

func TestSizePeerList(t *testing.T) {
	list := setupPeerList(10)

	require.Equalf(t, 10, list.Size(), "The size should match")
}

func TestUpdatePeerList(t *testing.T) {
	var list PeerList

	list.Update(NewPeer("auditor1", "127.0.0.1", 9001, "auditor"))
	require.Equalf(t, 1, list.Size(), "The size should have been incremented by 1")

	list.Update(NewPeer("auditor2", "127.0.0.1", 9002, "auditor"))
	require.Equalf(t, 2, list.Size(), "The size should have been incremented by 2")

	// update the previous one
	list.Update(NewPeer("auditor2", "127.0.0.1", 9002, "auditor"))
	require.Equalf(t, 2, list.Size(), "The size should have been incremented by 2")

	// update the previous one changing status
	p := NewPeer("auditor2", "127.0.0.1", 9002, "auditor")
	p.Status = AgentStatusLeaving
	list.Update(p)
	require.Equalf(t, 2, list.Size(), "The size should have been incremented by 2")
	require.Equalf(t, AgentStatusLeaving, list.L[1].Status, "The status should have been updated")
}

func TestDeletePeerList(t *testing.T) {
	list := setupPeerList(10)
	list2 := setupPeerList(10)

	// filter auditor types
	auditors := list.Filter(func(m *Peer) bool {
		return m.Meta.Role == "auditor"
	})

	// delete auditors
	for _, e := range auditors.L {
		list2.Delete(e)
	}

	require.Truef(t, 10 > list2.Size(), "The new list should have less elements")
	for _, e := range list2.L {
		require.Truef(t, "auditor" != e.Meta.Role, "The role cannot be Auditor")
	}
}

func TestDeleteNotIncludedPeerList(t *testing.T) {
	list := setupPeerList(10)

	list.Delete(NewPeer("unknown", "127.0.0.1", 10000, "unknown"))

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
