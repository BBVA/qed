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
	"math/rand"
	"net"

	"github.com/hashicorp/memberlist"
	"github.com/pkg/errors"
)

// Status is the state of the Agent instance.
type Status int32

const (
	AgentStatusAlive Status = iota
	AgentStatusLeaving
	AgentStatusLeft
	AgentStatusShutdown
	AgentStatusFailed
)

func (s Status) String() string {
	switch s {
	case AgentStatusAlive:
		return "alive"
	case AgentStatusLeaving:
		return "leaving"
	case AgentStatusLeft:
		return "left"
	case AgentStatusShutdown:
		return "shutdown"
	default:
		return "failed"
	}
}

// Member is a single member of the gossip cluster.
type Peer struct {
	Name   string
	Addr   net.IP
	Port   uint16
	Meta   Meta
	Status Status
}

// Returns a memberlist node from a peer
// datra
func (p Peer) Node() *memberlist.Node {
	return &memberlist.Node{
		Name: p.Name,
		Addr: p.Addr,
		Port: p.Port,
	}
}

//Returns a new peer from the parameters configuration
func NewPeer(name, addr string, port uint16, role string) *Peer {
	meta := Meta{
		Role: role,
	}

	return &Peer{
		Name: name,
		Addr: net.ParseIP(addr),
		Port: port,
		Meta: meta,
	}
}

// Builds a new peer from the memberlist.Node data
func ParsePeer(node *memberlist.Node) (*Peer, error) {
	var meta Meta
	err := meta.Decode(node.Meta)
	if err != nil {
		return nil, errors.Wrap(err, "Error parsing peer: unable to decode meta")
	}
	return &Peer{
		Name: node.Name,
		Addr: node.Addr,
		Port: node.Port,
		Meta: meta,
	}, nil
}

// Implements a list of peers
// which is able to filter, merge and
// take elements from the head
type PeerList struct {
	L []*Peer
}

func NewPeerList() *PeerList {
	return &PeerList{
		L: make([]*Peer, 0),
	}
}

// A filter function returns if a peer must
// me selected or not
type Filter func(m *Peer) bool

// Returns a filtered peer list, containg only
// the peers the filter selected
func (l *PeerList) Filter(f Filter) *PeerList {
	var b PeerList
	b.L = make([]*Peer, 0)
	for _, x := range l.L {
		if f(x) {
			b.L = append(b.L, x)
		}
	}

	return &b
}

// Appends a peer list to the current list
func (l *PeerList) Append(m *PeerList) {
	if m == nil {
		return
	}
	l.L = append(l.L, m.L...)
}

// Returnsa new list with n peers included
// starting in the head of the list.
func (l *PeerList) Take(n int) *PeerList {
	if n > len(l.L) {
		return nil
	}

	return &PeerList{
		L: l.L[:n],
	}
}

// Returns a list with all the peers from the exclusion
// list removed
func (l *PeerList) Exclude(ex *PeerList) *PeerList {
	if ex == nil {
		return l
	}
	return l.Filter(func(p *Peer) bool {
		for _, x := range ex.L {
			if x.Name == p.Name {
				return false
			}
		}
		return true
	})
}

// Returns the list randomly shuffled
func (l *PeerList) Shuffle() *PeerList {
	rand.Shuffle(len(l.L), func(i, j int) {
		l.L[i], l.L[j] = l.L[j], l.L[i]
	})
	return l
}

// Updates a peer data by its name
func (l *PeerList) Update(m *Peer) {
	for i, e := range l.L {
		if e.Name == m.Name {
			l.L[i] = m
			return
		}
	}
	l.L = append(l.L, m)
}

// Deletes a peer from the list by its name
func (l *PeerList) Delete(m *Peer) {
	for i, e := range l.L {
		if e.Name == m.Name {
			copy(l.L[i:], l.L[i+1:])
			l.L[len(l.L)-1] = nil
			l.L = l.L[:len(l.L)-1]
			return
		}
	}
}

func (l PeerList) Size() int {
	return len(l.L)
}
