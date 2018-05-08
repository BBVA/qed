// Copyright © 2018 Banco Bilbao Vizcaya Argentaria S.A.  All rights reserved.
// Use of this source code is governed by an Apache 2 License
// that can be found in the LICENSE file

/*
	Package hyper implements an hyper tree (HT), which is an sparse merkle tree (SMT) with
	optimizations for horizontal scalability.
	The work here is based in the paper
		Efficient Sparse Merkle Trees: Caching Strategies and Secure (Non-)Membership Proofs
		https://eprint.iacr.org/2016/683
*/
package hyper

import (
	"math"

	"verifiabledata/balloon/hashing"
	"verifiabledata/balloon/storage"
	"verifiabledata/log"
)

// An Hyper Tree is an optimized Sparse Merkle Tree with
// aggresive caching and optimized storage.
//
// It builds the caching and storage strategy with the assumption
// that the hash function used to compute the key of the values
// begin inserted satisfy the sparse condition stated in the paper.
//
// This makes possible to relay on an external database as the promise
// is that it will be very unlikely that a query to the database returns more
// than two elements. This approach minimizes the queries made to the database:
// one insert and one search per element being inserted, and one search per element
// being proved.
//
// The database used can be either a B+Tree based one for quick queries and ok
// insertion rate, or LSM with fast insterion and a key index to quickly scan keys.
//
// The API exported by the tree does not contain vissible state, just the operations
// that can be made to the tree, ensuring its integrity, and the strcutres returned by
// those public methods.
type Tree struct {
	id             []byte // tree-wide constant
	leafHasher     leafHasher
	interiorHasher interiorHasher
	defaultHashes  [][]byte
	cache          storage.Cache
	leaves         storage.Store
	stats          *stats
	cacheArea      *area
	digestLength   int
	ops            chan interface{} // serialize operations
}

// MembershipProof holds the audit information needed the verify
// membership
type MembershipProof struct {
	AuditPath   [][]byte
	ActualValue []byte
}

// NewTree returns  a new Hyper Tree given all its dependencies
func NewTree(id string, cache storage.Cache, leaves storage.Store, hasher hashing.Hasher) *Tree {

	cacheLevels := int(math.Max(0.0, math.Floor(math.Log(float64(cache.Size()))/math.Log(2.0))))
	digestLength := len(hasher([]byte("a test event"))) * 8

	tree := &Tree{
		[]byte(id),
		leafHasherF(hasher),
		interiorHasherF(hasher),
		make([][]byte, digestLength),
		cache,
		leaves,
		new(stats),
		newArea(digestLength-cacheLevels, digestLength),
		digestLength,
		nil,
	}

	// init default hashes cache
	tree.defaultHashes[0] = hasher(tree.id, Empty)
	for i := 1; i < int(digestLength); i++ {
		tree.defaultHashes[i] = hasher(tree.defaultHashes[i-1], tree.defaultHashes[i-1])
	}
	tree.ops = tree.operations()

	return tree
}

// Internally we use a channel API to serialize operations
// but external we use exported methods to be called
// by others.
// These methods returns a channel with an appropriate type
// for each operation to be consumed from when the data arrives.

// Queues an Add operation to the tree and returns a channel
// when the result []byte will be sent when ready
//
// The resulting []byte correspond to the hash of the root element
// of the tree.
//
// The algorithm will first store the key-value in the provided leaves store
// and then will proceed to compute the root hash.
//
// The algorithm need to compute at least a n+1 hashes per each inserted elements,
// when n is the size in bits of the key, but on average this number will
// be higher [....]
func (t Tree) Add(digest, index []byte) chan []byte {
	result := make(chan []byte, 0)
	t.ops <- &add{
		digest,
		index,
		result,
	}
	return result
}

// Queues a Prove operation to the tree and returns a channel
// when the result *MembershipProof will be sent when ready
func (t Tree) Prove(key []byte) chan *MembershipProof {
	result := make(chan *MembershipProof, 0)
	t.ops <- &proof{key, result}
	return result
}

// Queues a close operation to the tree and returns a channel
// were a true or false will be send when the operation is completed
func (t Tree) Close() chan bool {
	result := make(chan bool)
	t.ops <- &close{true, result}
	return result
}

// INTERNALS

type add struct {
	digest []byte
	index  []byte
	result chan []byte
}

type proof struct {
	key    []byte
	result chan *MembershipProof
}

type close struct {
	stop   bool
	result chan bool
}

// Run listens in channel operations to execute in the tree
func (t *Tree) operations() chan interface{} {
	operations := make(chan interface{}, 0)
	go func() {
		for {
			select {
			case op := <-operations:
				switch msg := op.(type) {
				case *close:
					t.leaves.Close()
					msg.result <- true
					return
				case *add:
					digest, err := t.add(msg.digest, msg.index)
					if err != nil {
						log.Error("Operations error: ", err)
					}
					msg.result <- digest
				case *proof:
					proof, err := t.auditPath(msg.key)
					if err != nil {
						log.Error("Operations error: ", err)
					}
					msg.result <- proof
				default:
					log.Error("Hyper tree Run() message not implemented!!")
				}

			}
		}
	}()
	return operations
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

func (t *Tree) auditPath(key []byte) (*MembershipProof, error) {
	value, err := t.leaves.Get(key) // TODO check existence
	if err != nil {
		log.Debug(t.leaves)
		return nil, err
	}
	path := t.calcAuditPathFromCache(key, rootPosition(t.digestLength))

	return &MembershipProof{
		path,
		value,
	}, nil
}

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
		d := t.leaves.GetRange(pos.base, pos.end())
		return t.fromStorage(d, value, pos)
	}

	// if not, the node hash is the hash of our left and right child
	isleft := !bitIsSet(key, t.digestLength-pos.height)
	if isleft {
		left = t.toCache(key, value, pos.left())
		right = t.fromCache(pos.right())
	} else {
		left = t.fromCache(pos.left())
		right = t.toCache(key, value, pos.right())
	}

	t.stats.ih += 1
	nodeHash = t.interiorHasher(left, right, pos.base, pos.heightBytes())

	// we re-cache all the nodes on each update
	// if the node is whithin the cache area
	if t.cacheArea.has(pos) {
		t.stats.update += 1
		t.cache.Put(pos.Key(), nodeHash)
	}

	return nodeHash
}

func (t *Tree) fromCache(pos *Position) []byte {

	// get the value from the cache
	cachedHash, cached := t.cache.Get(pos.Key())
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
	if len(d) == 1 && pos.height == 0 {
		t.stats.leaf += 1
		t.stats.lh += 1
		return t.leafHasher(t.id, value, pos.base)
	}

	// if there are no more childs,
	// return a default hash if i'm not in root node
	if len(d) == 0 && pos.height != pos.n {
		t.stats.dh += 1
		return t.defaultHashes[pos.height]
	}

	if len(d) > 0 && pos.height == 0 {
		panic("this should never happen (unsorted LeavesSlice or broken split?)")
	}

	leftSlice, rightSlice := d.Split(pos.split)

	left := t.fromStorage(leftSlice, value, pos.left())
	right := t.fromStorage(rightSlice, value, pos.right())
	t.stats.ih += 1
	return t.interiorHasher(left, right, pos.base, pos.heightBytes())
}

func (t *Tree) calcAuditPathFromCache(key []byte, pos *Position) [][]byte {
	// if we are beyond the cache zone
	// we need to go to database to get
	// nodes
	if !t.cacheArea.has(pos) {
		leaves := t.leaves.GetRange(pos.base, pos.end())
		return t.calcAuditPathFromStorage(leaves, key, pos)
	}

	if !bitIsSet(key, t.digestLength-pos.height) { // if k_j == 0
		return append(
			t.calcAuditPathFromCache(key, pos.left()),
			t.fromCache(pos.right()))
	}
	return append(
		t.calcAuditPathFromCache(key, pos.right()),
		t.fromCache(pos.left()))

}

func (t *Tree) calcAuditPathFromStorage(d storage.LeavesSlice, key []byte, pos *Position) [][]byte {
	if pos.height == 0 {
		return nil
	}
	left, right := d.Split(pos.split)

	if !bitIsSet(key, t.digestLength-pos.height) { // if k_j == 0
		return append(
			t.calcAuditPathFromStorage(left, key, pos.left()),
			t.fromStorage(right, key, pos.right()))
	}
	return append(
		t.calcAuditPathFromStorage(right, key, pos.right()),
		t.fromStorage(left, key, pos.left()))
}
