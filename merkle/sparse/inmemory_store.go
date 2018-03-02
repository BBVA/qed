// Copyright 2018 BBVA. All rights reserved.
// Use of this source code is governed by a Apache 2 License
// that can be found in the LICENSE file
package sparse

import (
	"fmt"
	"sync"
)

// Implements store interface, it uses sync.Map under the hood
type InmemoryStore struct {
	store *sync.Map
}

type pos struct {
	k []byte
	d uint64
}

func newpos(n *Node) pos {
	return pos{
		n.k,
		n.d,
	}
}

// Implements Add store interface to add Node to the store
func (m *InmemoryStore) Add(n *Node) error {
	_, loaded := m.store.LoadOrStore(newpos(n), *n)
	if loaded {
		return fmt.Errorf("Node already in pos: %v", n)
	}
	return nil
}

// Implements Get store interface to get Node from the store
// given a Position
func (m *InmemoryStore) Get(b []byte, d uint64) (*Node, error) {
	v, ok := m.store.Load(pos{b, d})
	if !ok {
		return nil, fmt.Errorf("Node with pos %v not found in storage", b)
	}
	node, ok := v.(Node)
	if ok {
		return &node, nil
	}
	return nil, fmt.Errorf("Error getting node from storage")
}

// Returns a new instance of a Store
func NewInmemoryStore() *InmemoryStore {
	return &InmemoryStore{
		new(sync.Map),
	}
}
