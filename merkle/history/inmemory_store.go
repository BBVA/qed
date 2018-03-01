package history

import (
	"fmt"
	"sync"
)

// Implements store interface, it uses sync.Map under the hood
type InmemoryStore struct {
	store *sync.Map
}

// Implements Add store interface to add Node to the store
func (m *InmemoryStore) Add(n *Node) error {
	_, loaded := m.store.LoadOrStore(*n.Pos, *n)
	if loaded {
		return fmt.Errorf("Node already in pos: %v", n.Pos)
	}
	return nil
}

// Implements Get store interface to get Node from the store
// given a Position
func (m *InmemoryStore) Get(p *Position) (*Node, error) {
	v, ok := m.store.Load(*p)
	if !ok {
		return nil, fmt.Errorf("Node with pos %v not found in storage", p)
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
