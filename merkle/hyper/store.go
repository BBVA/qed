package hyper

import (
	"bytes"
	"sort"
)

// Storage is an interface that defines the operations needed
// for a storage engine or database
type Storage interface {
	Add(key []byte, value []byte) error
	Get(*Position) D
}

// D intermediate data structure from database to memory
type D [][]byte

// Split splits d.
func (d D) Split(s []byte) (left, right D) {
	// the smallest index i where d[i] >= s
	i := sort.Search(len(d), func(i int) bool {
		return bytes.Compare(d[i], s) >= 0
	})
	return d[:i], d[i:]
}
