package hyper

import (
	"bytes"
	"encoding/binary"
	"strconv"
	"verifiabledata/util"
)

// Constants
var empty = []byte{0x00}
var set = []byte{0x01}

const byteslen = 32



// holds a hyper tree
type Tree struct {
	id      []byte
	upper_cache   Cache
	lower_cache Cache
	defhash [][]byte
	hasher  util.Hasher
	store   Storage
}

func cmp(a, b interface{}) int {
	return bytes.Compare(a.([]byte), b.([]byte))
}

// creates a new hyper tree
func NewTree(id string, upper, lower Cache, h util.Hasher, s Store) *tree {
	t := &tree{
		[]byte(id),
		c,
		make([][]byte, h.Size),
		h,
		b.TreeNew(cmp),
		new(stats),
	}

	t.defhash[0] = t.hasher.Do(t.id, empty)
	for i := 1; i < hasher.Size; i++ {
		t.defhash[i] = t.hasher.Do(t.defhash[i-1], t.defhash[i-1])
	}

	return t
}

func (t *tree) toCache(v *Value, p *Position) []byte {
	var left, right, nh []byte

	// if we are beyond the cache zone
	// we need to go to database to get
	// nodes
	if !t.cache.has(p) {
		return t.fromStorage(fromBTree(t, p), v, p)
	}

	// if not, out hash is the hash of our left and right child
	dir := bytes.Compare(v.key, p.split)
	switch {
	case dir < 0:
		left = t.toCache(a, v, p.left())
		right = t.fromCache(v, p.right())
	case dir > 0:
		left = t.fromCache(v, p.left())
		right = t.toCache(a, v, p.right())
	}

	nh = t.interiorHash(left, right, p)
	
	// we re-cache all the nodes on each update
	// if the node is whithin the cache area
	t.cache.insert(p, nh)

	return nh
}

func (t *tree) fromCache(v *Value, p *Position) []byte {

	// get the value from the cache
	cached_hash, cached := t.cache.node[p.String()]
	if cached {
		t.stats.hits += 1
		return cached_hash
	}

	// if there is no value in the cache,return a default hash
	t.stats.dh += 1
	return t.defhash[p.height]

}

func (t *tree) fromStorage(d *D, v *Value, p *Position) []byte {
	// if we are a leaf, return our hash
	if p.height == 0 {
		t.stats.leaf += 1
		return t.leafHash(set, v.key)

	}

	// if there are no more childs,
	// return a default hash
	if d.Len() == 0 {
		t.stats.dh += 1
		return t.defhash[p.height]

	}

	left, right := d.Split(p.split)
	return t.interiorHash(t.fromStorage(left, v, p.left()), t.fromStorage(right, v, p.right()), p)
}

func (t *tree) leafHash(a, b []byte) []byte {
	t.stats.lh += 1
	if bytes.Equal(a, empty) {
		return t.hasher.Do(t.id)
	}

	return t.hasher.Do(t.id, b)
}

func (t *tree) interiorHash(left, right []byte, p *Position) []byte {
	t.stats.ih += 1
	if bytes.Equal(left, right) {
		return t.hasher.Do(left, right)
	}

	height_bytes := make([]byte, 4)
	binary.LittleEndian.PutUintbyteslen(height_bytes, uintbyteslen(p.height))

	return t.hasher.Do(left, right, p.base, height_bytes)
}


func (t *tree) Add(key []byte, v []byte) []byte {
	val := &value{key, v}

	t.store.Add(val)
	return t.toCache(val, rootpos(t.hasher.Size))
}

/*
	Algorithm



*/
