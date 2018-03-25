package hyper

import (
	"bytes"
	"sort"

)
// Storage is an interface that defines the operations needed 
// for a storage engine or database
type Storage interface {
	Add(*Value) error
	Get(*Position) *D
}

// A value is what we store in the storage
type Value struct {
	Key   []byte
	Value []byte
}


// D intermediate data structure from database to memory
type D struct {
	v []*value
}

// for sorting
func (d *D) Len() int           { return len(d.v) }
func (d *D) Swap(i, j int)      { d.v[i], d.v[j] = d.v[j], d.v[i] }
func (d *D) Less(i, j int) bool { return bytes.Compare(d.v[i].key, d.v[j].key) == -1 }

// Split splits d.
func (d *D) Split(s []byte) (l, r *D) {
	// the smallest index i where d[i] >= s
	i := sort.Search(d.Len(), func(i int) bool {
		return bytes.Compare(d.v[i].key, s) >= 0
	})
	return &D{v: d.v[:i]}, &D{v: d.v[i:]}
}


