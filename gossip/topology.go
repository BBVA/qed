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
	"sync"
)

// Hold the gossip network information as this node sees it.
// This information can be used to route messages to other nodes.
type Topology struct {
	m map[string]*PeerList
	sync.Mutex
}

// Returns a new empty topology
func NewTopology() *Topology {
	m := make(map[string]*PeerList)
	return &Topology{
		m: m,
	}
}

// Updates the topology with the peer
// information
func (t *Topology) Update(p *Peer) error {
	t.Lock()
	defer t.Unlock()
	l, ok := t.m[p.Meta.Role]
	if !ok {
		t.m[p.Meta.Role] = NewPeerList()
		l = t.m[p.Meta.Role]
	}

	l.Update(p)
	return nil
}

// Deletes a peer from the topology
func (t *Topology) Delete(p *Peer) error {
	t.Lock()
	defer t.Unlock()
	l := t.m[p.Meta.Role]
	l.Delete(p)

	return nil
}

// Returns a list of peers of a given kind
func (t *Topology) Get(kind string) *PeerList {
	t.Lock()
	defer t.Unlock()
	return t.m[kind]
}

// Returns a peer list of each kind with n elements on each kind,
// Each list is built excluding all the nodes in the list l, shuffling the result,
// and taking the n elements from the head of the list.
func (t *Topology) Each(n int, l *PeerList) *PeerList {
	var p PeerList

	for _, list := range t.m {
		p.Append(list.Exclude(l).Shuffle().Take(n))
	}
	return &p
}
