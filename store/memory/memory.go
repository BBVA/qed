// Copyright 2018 BBVA. All rights reserved.
// Use of this source code is governed by a Apache 2 License
// that can be found in the LICENSE file

// Package memory implements a Store compatible interface
// using sync.Map as a backend.
package memory

import (
	"fmt"
	"sync"
	"verifiabledata/store"
	"verifiabledata/tree"
)

// Contains unexported fields
type Store struct {
	store *sync.Map
}

// Implements Add store interface to add tree.Node to the store
func (m *Store) Add(n *tree.Node) error {
	_, loaded := m.store.LoadOrStore(*n.Pos, *n)
	if loaded {
		return fmt.Errorf("Node already in pos: %v", n.Pos)
	}
	return nil
}

// Implements Get store interface to get tree.Node from the store
// given a tree.Position
func (m *Store) Get(p *tree.Position) (*tree.Node, error) {
	v, ok := m.store.Load(*p)
	if !ok {
		return nil, fmt.Errorf("Node with pos %v not found in storage", p)
	}
	node, ok := v.(tree.Node)
	if ok {
		return &node, nil
	}
	return nil, fmt.Errorf("Error getting node from storage")
}

// Returns a new instance of a Store
func NewStore() store.Store {
	return &Store{
		new(sync.Map),
	}
}
