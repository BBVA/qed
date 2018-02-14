// Copyright 2018 BBVA. All rights reserved.
// Use of this source code is governed by a Apache 2 License
// that can be found in the LICENSE file

package memory

import (
	"fmt"
	"sync"
	"verifiabledata/tree"
	"verifiabledata/store"
)

type MemoryStore struct {
	store *sync.Map
}

func (m *MemoryStore) Add(n tree.Node) error {
	m.store.Store(n.Pos,n)
	return nil
}

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

func NewMemoryStore() store.Store {
	return &MemoryStore{
		new(sync.Map),
	}
}
