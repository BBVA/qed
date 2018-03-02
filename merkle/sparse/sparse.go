// Copyright 2018 BBVA. All rights reserved.
// Use of this source code is governed by a Apache 2 License
// that can be found in the LICENSE file

package sparse

import (
	"math/big"
	"verifiabledata/util"
)

// Commitment is the digest of the hash function as proof of the event insertion
type Commitment []byte

// Tree implements a Sparse Merkle Tree as stated in the paper
// ....
type Tree struct {
	d     Store  // contains leaves
	xi    Store  // cache ξ contains the default hashes
	delta Store  // cache δ contains inserted nodes
	n     uint64 // the size of the Digest
}

var (
	Zero = big.NewInt(0)
	One  = big.NewInt(1)
	Two  = big.NewInt(2)
)

// NewTree returns an instance of an Sparse Merkle Tree as stated by
// http://.....
func NewTree(d, xi, delta Store, n uint64) *Tree {
	return &Tree{
		d, xi, delta, n,
	}
}

type Node struct {
	k []byte
	v uint64
	d uint64
}

// r ← SMT.Add(k, v): On input of a key k and an associated value v, the Add method
// inserts v into the leaf with index H(k), updates the relative information δ, and outputs
// the new root hash r.
func (t *Tree) Add(h []byte, v uint64) ([]byte, error) {

	node := &Node{
		h, v, 0,
	}
	if err := t.d.Add(node); err != nil {
		return nil, err
	}

	// start at root, recursively calculate hash
	root, _ := t.calculateTree(big.NewInt(0), big.NewInt(0).Lsh(big.NewInt(1), 256), t.n)

	return Commitment(root), nil
}

// calculateTree gets the Root node of the tree given the number
// of leaf descendants n, the current depth d, and the base mask b.
func (t *Tree) calculateTree(n, b *big.Int, d uint64) ([]byte, error) {

	// if the node is in the cache, get it
	if node, err := t.delta.Get(b.Bytes(), d); err == nil {
		return node.k, nil
	}

	// if the number of descendants is 0
	// return the default hash stored in xi
	if n.Cmp(Zero) == 0 {
		if node, err := t.xi.Get(b.Bytes(), d); err == nil {
			return node.k, nil
		}
	}

	// if we are at a leaf, return its hash
	// which is equivalent to b in our implementation
	if d == 0 && n.Cmp(One) == 0 {
		return b.Bytes(), nil
	}

	// Calculate the subtree recursively
	left_child, err := t.calculateTree(big.NewInt(0).Div(n, Two), b, d-1)
	if err != nil {
		return nil, err
	}

	// update b for right tranversal updating b
	right_child, err := t.calculateTree(big.NewInt(0).Div(n, Two), big.NewInt(0).SetBit(b, int(t.n-d), 1), d-1)
	if err != nil {
		return nil, err
	}
	return util.Hash(left_child, right_child), nil
}
