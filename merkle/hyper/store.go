package hyper

import (
	"bytes"
	"sort"
)

// D is our data structure to authenticate
type D struct {
	v []*value
}

// for sorting
func (d *D) Len() int           { return len(d.v) }
func (d *D) Swap(i, j int)      { d.v[i], d.v[j] = d.v[j], d.v[i] }
func (d *D) Less(i, j int) bool { return bytes.Compare(d.v[i].key, d.v[j].key) == -1 }

// Split splits d.
func (d *D) Split(s []byte) (l, r D) {
	// the smallest index i where d[i] >= s
	i := sort.Search(d.Len(), func(i int) bool {
		return bytes.Compare(d.v[i].key, s) >= 0
	})
	return D{v: d.v[:i]}, D{v: d.v[i:]}
}

func (d *D) Insert(v *value) {
	var at int
	// the smallest index i where d.v[i] >= s
	d.v = append(d.v, v)
	at = sort.Search(d.Len(), func(i int) bool {
		return bytes.Compare(d.v[i].key, v.key) >= 0
	})

	copy(d.v[at+1:], d.v[at:])
	d.v[at] = v
}

func newD() D {
	var d D
	d.v = make([]*value, 0, 10000000)
	return d
}
