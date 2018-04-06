package hyper

import (
	"bytes"
	"encoding/binary"
	"verifiabledata/balloon/hashing"
	"verifiabledata/balloon/storage"
)

// Constant Empty is a constant for empty leaves
var Empty = []byte{0x00}

// Constant Set is a constant for non-empty leaves
var Set = []byte{0x01}

// Tree holds a hyper tree structure
type Tree struct {
	id            []byte // tree-wide constant
	hasher        hashing.Hasher
	defaultHashes [][]byte
	cache         storage.Cache
	leaves         storage.Store
	stats         *stats
	cacheArea     *area
	digestLength  int
}

// Add inserts a new key-value pair into the tree and returns the
// root hash as a commitment.
func (t *Tree) add(key []byte, value []byte) ([]byte, error) {
	err := t.leaves.Add(key, value)
	if err != nil {
		return nil, err
	}
	return t.toCache(key, value, rootPosition(t.digestLength)), nil
}

func NewTree(id string, cache storage.Cache, leaves storage.Store, hasher hashing.Hasher) *Tree {

	digestLength := len(hasher([]byte("una cosa mala")))*8
	cacheLevels := 30

	tree := &Tree{
		[]byte(id),
		hasher,
		make([][]byte, digestLength),
		cache,
		leaves,
		new(stats),
		newArea(digestLength-cacheLevels, digestLength),
		digestLength,
	}

	// init default hashes cache
	tree.defaultHashes[0] = tree.hasher(tree.id, Empty)
	for i := 1; i < int(digestLength); i++ {
		tree.defaultHashes[i] = tree.hasher(tree.defaultHashes[i-1], tree.defaultHashes[i-1])
	}

	return tree
}

// Run listens in channel operations to execute in the tree
func (t *Tree) Run(operations chan interface{}) {
	go func() {
		for {
			select {
			case op := <-operations:
				switch msg := op.(type) {
				case *Stop:
					if msg.stop {
						msg.result <- true
						return
					}
				case *Add:
					digest, _ := t.add(msg.digest, msg.index)
					msg.result <- digest
				default:
					panic("Hyper tree Run() message not implemented!!")
				}

			}
		}
	}()
}

// These are the operations the tree supports and
// together form the channel based API

type Add struct {
	digest []byte
	index  []byte
	result chan []byte
}

func NewAdd(digest, index []byte, result chan []byte) *Add {
	return &Add{
		digest,
		index,
		result,
	}
}

type Stop struct {
	stop bool
	result chan bool
}

func NewStop(result chan bool) *Stop {
	return &Stop{true, result}
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
		return t.fromStorage(t.leaves.GetRange(pos.base, pos.split), value, pos)
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

func (t *Tree) fromStorage(d storage.LeavesSlice, value []byte, pos *Position) []byte {
	// if we are a leaf, return our hash
	if pos.height == 0 {
		t.stats.leaf += 1
		return t.leafHash(value, pos.base)
	}

	// if there are no more childs,
	// return a default hash
	if len(d) == 0 {
		t.stats.dh += 1
		return t.defaultHashes[pos.height]
	}

	leftSlice, rightSlice := d.Split(pos.split)

	left := t.fromStorage(leftSlice, value, pos.left())
	right := t.fromStorage(rightSlice, value, pos.right())
	return t.interiorHash(left, right, pos)
}

func (t *Tree) leafHash(a, base []byte) []byte {
	t.stats.lh += 1
	if bytes.Equal(a, Empty) {
		return t.hasher(t.id)
	}

	return t.hasher(t.id, base)
}

func (t *Tree) interiorHash(left, right []byte, pos *Position) []byte {
	t.stats.ih += 1
	if bytes.Equal(left, right) {
		return t.hasher(left, right)
	}

	heightBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(heightBytes, uint32(pos.height))

	return t.hasher(left, right, pos.base, heightBytes)
}
