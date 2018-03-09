// Copyright 2018 BBVA. All rights reserved.
// Use of this source code is governed by a Apache 2 License
// that can be found in the LICENSE file

/*
	Package sparse implements a sparse merkle tree as stated in the paper
	    https://eprint.iacr.org/2016/683.

*/
package sparse

import (
	"bytes"
	"encoding/binary"
	"math/big"
	"verifiabledata/util"
)

// Tree implements a Sparse Merkle Tree as stated in the paper
// https://eprint.iacr.org/2016/683
type Tree struct {
	id      []byte   // tree id
	leaves  Store    // D strcuture, contains leaf nodes
	cache   cache    // cache δ (delta) contains inserted nodes
	defhash [][]byte // ξ (xi) contains the default hashes

	hash *util.Hasher // the hash function and it's properties
}

// constants
var (
	Empty, Set, Zero, One []byte
)

func init() {
	// Empty is an empty key.
	Empty = []byte{0x0}
	// Set is a set key.
	Set = []byte{0x1}
	// Numbers to compare agains
	Zero = []byte{0x00}
	One = []byte{0x01}
}

// NewTree returns an instance of a Sparse Merkle Tree
func NewTree(id string, leaves Store, c cache, hash *util.Hasher) *Tree {
	var t Tree

	t.id = hash.Do([]byte(id))
	t.leaves = leaves
	t.cache = c
	t.hash = hash

	t.defhash = make([][]byte, t.hash.Size)
	t.defhash[0] = hash.Do(t.id, Empty)
	for i := 1; i < t.hash.Size; i++ {
		t.defhash[i] = hash.Do(t.defhash[i-1], t.defhash[i-1])
	}

	return &t
}

// (Dk, Dv) pair is a node of k, v later stored in a storage D whicih implements the Storage intrerface
type Node struct {
	k []byte
	v uint64
}

// (r ← Add(k, v)). On input of a key-value pair (k, v), the Add algorithm
// inserts (k, v) into the non-authenticated data structure D (overwriting any previous pair
// with the same key), gives the oracle an opportunity to perform initial adjustments to the
// relative information δ, and uses Recursion 5 to refresh δ and output the new root hash r.
//
func (t *Tree) Add(k []byte, v uint64) ([]byte, error) {

	value := make([]byte, 8)
	binary.LittleEndian.PutUint64(value, v)

	node := &Node{
		k, v,
	}
	if err := t.leaves.Add(node); err != nil {
		return nil, err
	}

	return t.update(k, value, make([]byte, 32), t.hash.Size), nil

}

// Traversing the tree down to the leaf with index h_k ← H(k) and
// value v, the update routine derives the new root hash as follows:
//
func (t *Tree) update(key, value, base []byte, depth int) []byte {
	var split []byte
	var left, right []byte

	if depth == 0 {
		return t.leafHash(value, base)
	}

	// left traversals use base but
	// right traversals update base setting to 1 the
	// bit j = size - depth
	split = make([]byte, len(base))
	copy(split, base)
	j := t.hash.Size - depth
	bitSet(split, j)

	// calcualte the descendants = 2^h
	n := big.NewInt(0).Exp(big.NewInt(2), big.NewInt(int64(depth)), nil).Bytes()

	// if key <= split
	// do a left traversal
	if bytes.Compare(key, split) < 0 {
		left = t.update(key, value, base, depth-1)
		right = t.rootHash(n, split, depth-1)
	} else {
		left = t.rootHash(n, base, depth-1)
		right = t.update(key, value, split, depth-1)
	}
	ih := t.interiorHash(left, right, base, depth)
	t.cache.Update(left, right, base, ih, t.defhash, depth)
	return ih
}

// Recursion 2 (rh).Starting from the base b ← 0 N , depth d ← N,
// and data structure D containing key-value// pairs (Dk , Dv), the root hash is
// derived as follows:
//
func (t *Tree) rootHash(ndesc, base []byte, depth int) []byte {
	var split []byte

	// calcualte the descendants = 2^h
	n := big.NewInt(0).Exp(big.NewInt(2), big.NewInt(int64(depth)), nil).Bytes()

	switch {
	case t.cache.Exists(depth, base):
		return t.cache.Get(depth, base)
	case bytes.Equal(n, Zero):
		return t.defhash[depth]
	case depth == 0 && bytes.Equal(n, One):
		return t.leafHash(Set, base)
	case !bytes.Equal(n, One) && depth == 0:
		panic("this should never happen!!")
	default:
		// left traversals use base but
		// right traversals update base setting to 1 the
		// bit j = size - depth
		split = make([]byte, len(base))
		copy(split, base)
		j := t.hash.Size - depth

		bitSet(split, j)

		left := t.rootHash(n, base, depth-1)
		right := t.rootHash(n, split, depth-1)
		return t.interiorHash(left, right, base, depth)
	}
}

func (t *Tree) leafHash(a, base []byte) []byte {
	if bytes.Equal(a, Empty) {
		return t.hash.Do(t.id)
	}
	return t.hash.Do(t.id, base)
}

func (t *Tree) interiorHash(left, right, base []byte, depth int) []byte {
	if bytes.Equal(left, right) {
		return t.hash.Do(left, right)
	}

	depth_bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(depth_bytes, uint32(depth))

	return t.hash.Do(left, right, base, depth_bytes)
}

func bitIsSet(bits []byte, i uint64) bool { return bits[i/8]&(1<<uint(7-i%8)) != 0 }
func bitSet(bits []byte, i int)           { bits[i/8] |= 1 << uint(7-i%8) }
