package hyper

import (
	"bytes"
	"encoding/binary"
	"verifiabledata/merkle"
	"verifiabledata/util"
)

var empty = []byte{0x00}
var set = []byte{0x01}

// Tree holds a hyper tree structure
type Tree struct {
	id            []byte
	hasher        *util.Hasher
	defaultHashes [][]byte
	cache         merkle.Cache
	store         Storage
	stats         *stats
	cacheArea     *area
}

func NewTree(id string, hasher *util.Hasher, cacheLevels int, cache merkle.Cache, store Storage) *Tree {
	tree := &Tree{
		[]byte(id),
		hasher,
		make([][]byte, hasher.Size),
		cache,
		store,
		new(stats),
		newArea(hasher.Size-cacheLevels, hasher.Size),
	}

	// init default hashes cache
	tree.defaultHashes[0] = tree.hasher.Do(tree.id, empty)
	for i := 1; i < int(hasher.Size); i++ {
		tree.defaultHashes[i] = tree.hasher.Do(tree.defaultHashes[i-1], tree.defaultHashes[i-1])
	}

	return tree
}

// Add inserts a new key-value pair into the tree and returns the
// root hash as a commitment.
func (t *Tree) Add(key []byte, value []byte) []byte {
	t.store.Add(key, value)
	return t.toCache(key, value, rootpos(t.hasher.Size))
}

// INTERNALS

// Area is the area of the tree designated by its min height and its max height
type area struct {
	minHeigth, maxHeigth int
}

// check if a position is whithing the caching area
func (a area) has(p *Position) bool {
	if p.height > a.minHeigth && p.height <= a.maxHeigth {
		return true
	}
	return false
}

// creates a new area structure, initialized with max and min boundaries
func newArea(min, max int) *area {
	return &area{
		min,
		max,
	}
}

func (t *Tree) toCache(key, value []byte, pos *Position) []byte {
	var left, right, nodeHash []byte

	// if we are beyond the cache zone
	// we need to go to database to get
	// nodes
	if !t.cacheArea.has(pos) {
		t.stats.disk += 1
		return t.fromStorage(t.store.GetRange(pos.base, pos.split), pos)
	}

	// if not, the node hash is the hash of our left and right child
	dir := bytes.Compare(key, pos.split)
	switch {
	case dir < 0:
		left = t.toCache(key, value, pos.left())
		right = t.fromCache(pos.right())
	case dir > 0:
		left = t.fromCache(pos.left())
		right = t.toCache(key, value, pos.right())
	}

	nodeHash = t.interiorHash(left, right, pos)

	// we re-cache all the nodes on each update
	// if the node is whithin the cache area
	if t.cacheArea.has(pos) {
		t.stats.update += 1
		t.cache.Put(pos.base, nodeHash)
	}

	return nodeHash
}

func (t *Tree) fromCache(pos *Position) []byte {

	// get the value from the cache
	cachedHash, cached := t.cache.Get(pos.base)
	if cached {
		t.stats.hits += 1
		return cachedHash
	}

	// if there is no value in the cache,return a default hash
	t.stats.dh += 1
	return t.defaultHashes[pos.height]

}

func (t *Tree) fromStorage(d LeavesSlice, pos *Position) []byte {
	// if we are a leaf, return our hash
	if pos.height == 0 {
		t.stats.leaf += 1
		return t.leafHash(set, pos.base)
	}

	// if there are no more childs,
	// return a default hash
	if len(d) == 0 {
		t.stats.dh += 1
		return t.defaultHashes[pos.height]
	}

	left, right := d.Split(pos.split)

	return t.interiorHash(t.fromStorage(left, pos.left()), t.fromStorage(right, pos.right()), pos)
}

func (t *Tree) leafHash(a, base []byte) []byte {
	t.stats.lh += 1
	if bytes.Equal(a, empty) {
		return t.hasher.Do(t.id)
	}

	return t.hasher.Do(t.id, base)
}

func (t *Tree) interiorHash(left, right []byte, pos *Position) []byte {
	t.stats.ih += 1
	if bytes.Equal(left, right) {
		return t.hasher.Do(left, right)
	}

	heightBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(heightBytes, uint32(pos.height))

	return t.hasher.Do(left, right, pos.base, heightBytes)
}
