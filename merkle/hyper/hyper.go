package hyper

import (
	"bytes"
	"encoding/binary"
	"strconv"
	"verifiabledata/util"

	"github.com/cznic/b"
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

// A position identifies a unique node in the tree by its base, split and height
type position struct {
	base   []byte // the left-most leaf in this node subtree
	split  []byte // the left-most leaf in the right branch of this node subtree
	height int    // height in the tree of this node
	n      int    // number of bits in the hash key
}

// returns a string representation of the position
func (p position) String() string {
	return string(p.base[:32]) + strconv.Itoa(p.height)
	// return fmt.Sprintf("%x-%d", p.base, p.height)
}

// returns a new position pointing to the left child
func (p position) left() *position {
	var np position
	np.base = p.base
	np.height = p.height - 1
	np.n = p.n

	np.split = make([]byte, 32)
	copy(np.split, np.base)

	bitSet(np.split, p.n-p.height)

	return &np
}

// returns a new position pointing to the right child
func (p position) right() *position {
	var np position
	np.base = p.split
	np.height = p.height - 1
	np.n = p.n

	np.split = make([]byte, 32)
	copy(np.split, np.base)

	bitSet(np.split, p.n-p.height)

	return &np
}

// creates the tree root position
func rootpos(n int) *position {
	var p position
	p.base = make([]byte, 32)
	p.split = make([]byte, 32)
	p.height = n
	p.n = n

	bitSet(p.split, 0)

	return &p
}

type stats struct {
	hits   uint64
	disk   uint64
	dh     uint64
	update uint64
	leaf   uint64
	lh     uint64
	ih     uint64
	lend   float64
}

// a cache contains the hashes of the pre computed nodes
type cache struct {
	n         int               // number of bits in the hash key
	node      map[string][]byte // node map containing the cached hashes
	minHeight int               // min height of the cache
}

// creates a new cache structure, already initialized with
func newcache(n int) *cache {
	return &cache{
		n,
		make(map[string][]byte, 2097153), // 2^(layers+1) - 1 nodes in the cache
		n - 27,
	}
}

// holds a hyper tree
type tree struct {
	cache   *cache
	defhash [][]byte
	id      []byte
	hasher  *util.Hasher
	store   *b.Tree
	stats   *stats // cache statistics
}

func cmp(a, b interface{}) int {
	return bytes.Compare(a.([]byte), b.([]byte))
}

// creates a new hyper tree
func newtree(id string) *tree {
	hasher := util.Hash256()
	t := &tree{
		newcache(hasher.Size),
		make([][]byte, hasher.Size),
		[]byte(id),
		hasher,
		b.TreeNew(cmp),
		new(stats),
	}

	t.defhash[0] = t.hasher.Do(t.id, empty)
	for i := 1; i < hasher.Size; i++ {
		t.defhash[i] = t.hasher.Do(t.defhash[i-1], t.defhash[i-1])
	}

	return t
}

func (t *tree) fromBTree(p *position) *D {
	var d D
	var err error
	var k, v interface{}

	t.stats.disk += 1

	d.v = make([]*value, 0)

	iter, _ := t.store.Seek(p.base)
	defer iter.Close()

	defer func() {
		a := float64(len(d.v))
		if  a > t.stats.lend { t.stats.lend = a }
		// t.stats.lend = (float64(len(d.v)) + t.stats.lend) / float64(t.stats.disk)
	}()

	n := 0
	for {
		k, v, err = iter.Next()
		if err != nil {
			return &d
		}
		if bytes.Compare(k.([]byte), p.split) < 0 {
			d.v = append(d.v, &value{k.([]byte), v.([]byte)})
		} else {
			return &d
		}
		n += 1
	}

	return &d
}

func (t *tree) toCache(v *value, p *position) []byte {
	var left, right, nh []byte

	// if we are beyond the cache zone
	// we need to go to database to get
	// nodes
	if p.height < t.cache.minHeight {
		return t.fromStorage(t.fromBTree(p), v, p)
	}

	// if not, out hash is the hash of our left and right child
	dir := bytes.Compare(v.key, p.split)
	switch {
	case dir < 0:
		left = t.toCache(v, p.left())
		right = t.fromCache(v, p.right())
	case dir > 0:
		left = t.fromCache(v, p.left())
		right = t.toCache(v, p.right())
	}

	nh = t.interiorHash(left, right, p)
	// we re-cache all the nodes on each update
	// if the node is whithin the cache area
	if p.height >= t.cache.minHeight {
		t.stats.update += 1
		t.cache.node[p.String()] = nh
	}

	return nh
}

func (t *tree) fromCache(v *value, p *position) []byte {

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

func (t *tree) fromStorage(d *D, v *value, p *position) []byte {
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

func (t *tree) interiorHash(left, right []byte, p *position) []byte {
	t.stats.ih += 1
	if bytes.Equal(left, right) {
		return t.hasher.Do(left, right)
	}

	height_bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(height_bytes, uint32(p.height))

	return t.hasher.Do(left, right, p.base, height_bytes)
}

func bitSet(bits []byte, i int)   { bits[i/8] |= 1 << uint(7-i%8) }
func bitUnset(bits []byte, i int) { bits[i/8] &= 0 << uint(7-i%8) }

func (t *tree) Add(key []byte, v []byte) []byte {
	val := &value{key, v}

	t.store.Set(key, v)
	return t.toCache(val, rootpos(t.hasher.Size))
}

/*
	Algorithm



*/
