package hyper

import (
	"bytes"
	"strconv"
	"verifiabledata/util"
)

// Constants
var empty = []byte{0x00}
var set = []byte{0x01}

// A value is what we store in a tree's leaf
type value struct {
	key   []byte
	value []byte
}

// creates a new value
func newvalue(k, v []byte) *value {
	return &value{k, v}
}

// A position identifies a unique node in the tree by its base, split and depth
type position struct {
	base  []byte // the left-most leaf in this node subtree
	split []byte // the left-most leaf in the right branch of this node subtree
	depth int    // depth in the tree of this node
	n     int    // number of bits in the hash key
}

// returns a string representation of the position
func (p position) String() string {
	return string(p.base[:32]) + strconv.Itoa(p.depth)
	// fmt.Sprintf("%x-%d", p.base, p.depth)
}

// returns a new position pointing to the left child
func (p position) left() *position {
	var np position
	np.base = p.base
	np.depth = p.depth - 1
	np.n = p.n

	np.split = make([]byte, 32)
	copy(np.split, np.base)

	bitSet(np.split, p.n-p.depth)

	return &np
}

// returns a new position pointing to the right child
func (p position) right() *position {
	var np position
	np.base = p.split
	np.depth = p.depth - 1
	np.n = p.n

	np.split = make([]byte, 32)
	copy(np.split, np.base)

	bitSet(np.split, p.n-p.depth)

	return &np
}

// creates the tree root position
func rootpos(n int) *position {
	var p position
	p.base = make([]byte, 32)
	p.split = make([]byte, 32)
	p.depth = n
	p.n = n

	bitSet(p.split, 0)

	return &p
}

// a cache contains the hashes of the pre computed nodes
type cache struct {
	n        int               // number of bits in the hash key
	node     map[string][]byte // node map containing the cached hashes
	depth    int               // current depth of the cache
	maxDepth int               // max depth of the cache
	hits     uint64
	miss     uint64
	dh       uint64
}

// creates a new cache structure, already initialized with
func newcache(n int) *cache {
	return &cache{
		n,
		make(map[string][]byte),
		n,
		n - 10,
		0,
		0,
		0,
	}
}

// pushes down a node in the cache
// which is completly wrong as the cached value
// would not be correct for a node below
func (c *cache) push(val []byte, p *position) {
	if p.depth <= c.maxDepth {
		return
	}
	c.node[p.left().String()] = val
	c.node[p.right().String()] = val
	n := p.depth - 1
	if n < c.depth {
		c.depth = n
	}
}

// holds a hyper tree
type tree struct {
	cache   *cache
	defhash [][]byte
	id      []byte
	hasher  *util.Hasher
}

// creates a new hyper tree
func newtree(id string) *tree {
	hasher := util.Hash256()
	t := &tree{
		newcache(hasher.Size),
		make([][]byte, hasher.Size),
		[]byte(id),
		hasher,
	}

	t.defhash[0] = t.hasher.Do(t.id, empty)
	for i := 1; i < hasher.Size; i++ {
		t.defhash[i] = t.hasher.Do(t.defhash[i-1], t.defhash[i-1])
	}

	return t
}

func (t *tree) toCache(v *value, p *position) []byte {
	var nh []byte

	// if we are a leaf, return our hash
	if p.depth == 0 {
		return t.leafHash(set, v.key)
	}

	// out hash is the hash of our childs, in a left traversal
	// the right branch comes from cache and viceversa
	dir := bytes.Compare(v.key, p.split)
	if dir < 0 {
		nh = t.hasher.Do(t.toCache(v, p.left()), t.fromCache(v, p.right()))
	} else {
		nh = t.hasher.Do(t.fromCache(v, p.left()), t.toCache(v, p.right()))
	}

	// if we are already in cache, we delete ourselves
	// because we are in the current path of insertion
	// if we are not cached, we cache ourselves now
	// for the future queries when we are not in the update
	// path
	_, cached := t.cache.node[p.String()]
	switch {
	case cached:
		delete(t.cache.node, p.String())
	case !cached && p.depth > t.cache.maxDepth:
		t.cache.node[p.String()] = nh
		if p.depth < t.cache.depth {
			t.cache.depth = p.depth
		}
	}

	return nh
}

func (t *tree) fromCache(v *value, p *position) []byte {

	// get the value from the cache
	cached_val, cached := t.cache.node[p.String()]
	if cached {
		t.cache.hits += 1
		return cached_val
	}
	
	// if there is no value in the cache, and the node is
	// below the current cache depth, return a default hash
	if p.depth < t.cache.depth {
		t.cache.dh += 1
		return t.defhash[p.depth]
	}
	
	// if cache depth is at maxDepth and
	// we doesn't have a value, we need to go to the 
	// storage to retrieve the tree node
	if p.depth <= t.cache.maxDepth {
		t.cache.miss += 1
		// go to database and iterate with the results
		// fmt.Println("Go to database at depth ", p.depth, " because max is ", t.cache.maxDepth)
		return t.defhash[p.depth]
	}
	
	// if the node is a leaf, return a default hash
	if p.depth == 0 {
		t.cache.dh += 1
		return t.defhash[p.depth]
	}
	
	// the hash if this node is the hash of its childs
	return t.hasher.Do(t.fromCache(v, p.left()), t.fromCache(v, p.right()))

}

func (t *tree) leafHash(a, b []byte) []byte {

	if bytes.Equal(a, empty) {
		return t.hasher.Do(t.id)
	}

	return t.hasher.Do(t.id, b)
}

func bitSet(bits []byte, i int)   { bits[i/8] |= 1 << uint(7-i%8) }
func bitUnset(bits []byte, i int) { bits[i/8] &= 0 << uint(7-i%8) }

/*
	Algorithm



*/
