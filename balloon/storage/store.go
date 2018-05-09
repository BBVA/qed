package storage

import (
	"bytes"
	"sort"
)

type Store interface {
	Add(key []byte, value []byte) error
	GetRange(start, end []byte) LeavesSlice
	Get(key []byte) ([]byte, error)

	Close() error
}

type DeletableStore interface {
	Delete(key []byte) error

	Store
}

// LeavesSlice is intermediate data structure from database to memory
type LeavesSlice [][]byte

// Split splits the slice.
func (ls LeavesSlice) Split(s []byte) (left, right LeavesSlice) {
	// the smallest index i where d[i] >= s
	i := sort.Search(len(ls), func(i int) bool {
		return bytes.Compare(ls[i], s) >= 0
	})
	return ls[:i], ls[i:]
}
