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

func setupTopology(size int) *Topology {
	topology := NewTopology()
	for i := 0; i < size; i++ {
		name := fmt.Sprintf("name%d", i)
		port := uint16(9000 + i)
		role := roles[i%len(roles)]
		peer := NewPeer(name, "127.0.0.1", port, role)
		_ = topology.Update(peer)
	}
	return topology
}

func TestUpdateAndDeleteTopology(t *testing.T) {
	topology := NewTopology()

	peer := NewPeer("auditor", "127.0.0.1", 9000, "auditor")
	_ = topology.Update(peer)

	auditors := topology.Get("auditor")
	require.Truef(t, 1 == auditors.Size(), "The topology must include one auditor")

	_ = topology.Delete(peer)

	auditors = topology.Get("auditor")
	require.Truef(t, 0 == auditors.Size(), "The topology must include zero auditor")

}

func TestEachWithoutExclusionsTopology(t *testing.T) {
	topology := setupTopology(10)

	each := topology.Each(1, nil)

	require.Truef(t, 3 == each.Size(), "It must include only 3 elements")

	auditors := each.Filter(func(m *Peer) bool {
		return m.Meta.Role == "auditor"
	})
	monitors := each.Filter(func(m *Peer) bool {
		return m.Meta.Role == "monitor"
	})
	publishers := each.Filter(func(m *Peer) bool {
		return m.Meta.Role == "publisher"
	})

	require.Truef(t, 1 == auditors.Size(), "It must include only one auditor")
	require.Truef(t, 1 == monitors.Size(), "It must include only one monitor")
	require.Truef(t, 1 == publishers.Size(), "It must include only one publisher")
}

func TestEachWithExclusionsTopology(t *testing.T) {
	topology := setupTopology(10)

	excluded := topology.Get("auditor")
	each := topology.Each(1, excluded)

	require.Truef(t, 2 == each.Size(), "It must include only 2 elements")

	auditors := each.Filter(func(m *Peer) bool {
		return m.Meta.Role == "auditor"
	})
	monitors := each.Filter(func(m *Peer) bool {
		return m.Meta.Role == "monitor"
	})
	publishers := each.Filter(func(m *Peer) bool {
		return m.Meta.Role == "publisher"
	})

	require.Truef(t, 0 == auditors.Size(), "It must not include any auditor")
	require.Truef(t, 1 == monitors.Size(), "It must include only one monitor")
	require.Truef(t, 1 == publishers.Size(), "It must include only one publisher")
}
