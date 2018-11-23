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
	"math/rand"
	"sync"

	"github.com/bbva/qed/gossip/member"
)

type PeerList struct {
	L []*member.Peer
}

type Filter func(m *member.Peer) bool

func (l *PeerList) Filter(f Filter) *PeerList {
	var b PeerList
	b.L = l.L[:0]
	for _, x := range l.L {
		if f(x) {
			b.L = append(b.L, x)
		}
	}

	return &b
}

func (l *PeerList) Exclude(list *PeerList) *PeerList {
	if list == nil {
		return l
	}
	return l.Filter(func(p *member.Peer) bool {
		for _, x := range list.L {
			if x.Name == p.Name {
				return false
			}
		}
		return true
	})
}

func (l PeerList) All() PeerList {
	return l
}

func (l *PeerList) Shuffle() *PeerList {
	rand.Shuffle(len(l.L), func(i, j int) {
		l.L[i], l.L[j] = l.L[j], l.L[i]
	})
	return l
}

func (l *PeerList) Update(m *member.Peer) error {
	var add bool = true
	for _, e := range l.L {
		if e.Name == m.Name {
			e = m
			add = false
		}
	}
	if add {
		l.L = append(l.L, m)
	}
	return nil
}

func (l *PeerList) Delete(m *member.Peer) error {
	for i, e := range l.L {
		if e.Name == m.Name {
			l.L[i] = l.L[len(l.L)-1]
			l.L[len(l.L)-1] = nil
			l.L = l.L[:len(l.L)-1]
		}
	}
	return nil
}

type Topology struct {
	m []PeerList
	sync.Mutex
}

func NewTopology() *Topology {
	m := make([]PeerList, member.Unknown)
	for i := member.Auditor; i < member.Unknown; i++ {
		m[i] = PeerList{
			L: make([]*member.Peer, 0),
		}
	}
	return &Topology{
		m: m,
	}
}

func (t *Topology) Update(p *member.Peer) error {
	t.Lock()
	defer t.Unlock()
	t.m[p.Meta.Role].Update(p)
	return nil
}

func (t *Topology) Delete(p *member.Peer) error {
	t.Lock()
	defer t.Unlock()
	t.m[p.Meta.Role].Delete(p)
	return nil
}

func (t *Topology) Get(kind member.Type) PeerList {
	t.Lock()
	defer t.Unlock()
	return t.m[kind]
}

func (t *Topology) Each(n int, p *PeerList) *PeerList {
	var b PeerList

	auditors := t.m[member.Auditor].Exclude(p).Shuffle()
	monitors := t.m[member.Monitor].Exclude(p).Shuffle()
	publishers := t.m[member.Publisher].Exclude(p).Shuffle()

	if len(auditors.L) > n {
		auditors.L = auditors.L[:n]
	}
	if len(monitors.L) > n {
		monitors.L = monitors.L[:n]
	}
	if len(publishers.L) > n {
		publishers.L = publishers.L[:n]
	}
	b.L = append(b.L, auditors.L...)
	b.L = append(b.L, auditors.L...)
	b.L = append(b.L, auditors.L...)

	return &b
}
