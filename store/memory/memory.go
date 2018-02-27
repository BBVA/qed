// Copyright 2018 BBVA. All rights reserved.
// Use of this source code is governed by a Apache 2 License
// that can be found in the LICENSE file

// Package memory implements a Store compatible interface
// using sync.Map as a backend.
package memory

import (
	"bytes"
	"fmt"
	"sort"
	"sync"
	"verifiabledata/store"
	"verifiabledata/tree"
)

// Implements store interface, it uses sync.Map under the hood
type Store struct {
	store *sync.Map
}

// Implements Add store interface to add tree.Node to the store
func (m *Store) Add(n *tree.Node) (*tree.Node, error) {
	_, loaded := m.store.LoadOrStore(*n.Pos, *n)
	if loaded {
		return nil, fmt.Errorf("Node already in pos: %v", n.Pos)
	}
	return n, nil
}

// Implements Get store interface to get tree.Node from the store
// given a tree.Position
func (m *Store) Get(n *tree.Node) (*tree.Node, error) {
	v, ok := m.store.Load(*n.Pos)
	if !ok {
		return nil, fmt.Errorf("Node with pos %v not found in storage", n.Pos)
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

// Implements store interface, it used []tree.Node under the hood
// also implements sort.Interface and uses binary search to do
// sorted inserts using sort.Search to do so
type SortedStore struct {
	sync.RWMutex
	data []tree.Node
}

// Implements Add store interface to add tree.Node to the store
func (s *SortedStore) Add(n *tree.Node) (*tree.Node, error) {
	s.Lock()
	defer s.Unlock()
	i := sort.Search(len(s.data), func(i int) bool {
		return bytes.Compare(s.data[i].Digest, n.Digest) >= 0
	})
	// Insert as per https://github.com/golang/go/wiki/SliceTricks
	s.data = append(s.data, tree.Node{})
	copy(s.data[i+1:], s.data[i:])
	s.data[i] = *n
	return n, nil
}

// Implements Get store interface to get tree.Node from the store
// given a tree.Position
func (s *SortedStore) Get(n *tree.Node) (*tree.Node, error) {
	s.Lock()
	defer s.Unlock()
	i := sort.Search(len(s.data), func(i int) bool {
		return bytes.Equal(s.data[i].Digest, n.Digest) 
	})
	if i > 0 {
		return &s.data[i], nil
	}
	return nil, fmt.Errorf("SortedStore: Node not found")
}

// Returns a new instance of a SortedStore with 1000 nodes capacity
func NewSortedStore() store.Store {
	var s SortedStore
	s.data = make([]tree.Node, 1000)
	return &s
}
