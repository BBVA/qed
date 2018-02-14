// Copyright 2018 BBVA. All rights reserved.
// Use of this source code is governed by a Apache 2 License
// that can be found in the LICENSE file


// Package memory implements a Store compatible interface 
// using sync.Map as a backend.
package memory

import (
	"fmt"
	"sync"
	"verifiabledata/tree"
	"verifiabledata/store"
)


// Contains unexported fields
type MemoryStore struct {
	store *sync.Map
}

// Implements Add store interface to add tree.Node to the store
func (m *MemoryStore) Add(n tree.Node) error {
	m.store.Store(n.Pos,n)
	return nil
}

// Implements Get store interface to get tree.Node from the store
// given a tree.Position
func (m *MemoryStore) Get(p tree.Position) (*tree.Node, error) {
	v, ok := m.store.Load(p)
	if ! ok {
		return nil, fmt.Errorf("Key not found")
	}
	node, ok := v.(tree.Node)
	if ok {
		return &node, nil
	}
	return nil, fmt.Errorf("Error getting node from storage")
}


// Returns a new instance of a MemoryStore
func NewMemoryStore() store.Store {
	return &MemoryStore{
		new(sync.Map),
	}
}
