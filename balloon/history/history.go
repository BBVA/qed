// Copyright 2018 BBVA. All rights reserved.
// Use of this source code is governed by a Apache 2 License
// that can be found in the LICENSE file

/*
	Package history implements a history tree structure as described in the paper
	    Balloon: A Forward-Secure Append-Only Persistent Authenticated Data Structure
	    https://eprint.iacr.org/2015/007
*/
package history

import (
	"fmt"
	"math"
	"verifiabledata/util"
)

// Constant Zero is the 0x0 byte, and it is used as a prefix to know
// if a node has a zero digest.
var Zero = []byte{0x0}

// Constant One is the 0x1 byte, and it is used as a prefix to know
// if a node has a non-zero digest.
var One = []byte{0x1}

// Commitment holds the Digest as proof of the event insertion, and the verstion of
// the tree after the insertion, which is equivalent to the Position.Index
type Commitment struct {
	Version uint64
	Digest  []byte
}

// Position holds the index and the layer of a node in a tree
type Position struct {
	Index uint64
	Layer uint64
}

func (p *Position) String() string {
	return fmt.Sprintf("(i %d, l %d)", p.Index, p.Layer)
}

// A node holds its digest and its position
type Node struct {
	Pos    *Position
	Digest []byte
}

func (n *Node) String() string {
	return fmt.Sprintf("(P %s,  D %x)", n.Pos, n.Digest)
}

// A History tree is a tree structure with a version metadata.
// As described in the pag. 6-7 of the paper:
// http://tamperevident.cs.rice.edu/papers/paper-treehist.pdf
//
// For example, a tree with 5 leaf hashes x0, x1, x2, x3, x4
//    version 4
//
//    layer 3                 ____________ h(0,3)_____________
//                           |                               |
//    layer 2       _______h(0,2)_______                 __h(4,2)__
//                  |                  |                 |        |
//    layer 1   __h(0,1)__         __h(2,1)__         __h(4,1)__  ▢
//              |        |         |        |         |        |
//    layer 0  x0(0,0) x1(1,0) x2(2,0)   x3(3,0)    x4(4,0)    ▢
//
//
// Each a is a pair of index and layer (a position). The index starts in 0.
// The size of the tree is the last index number plus 1.
// The next element of the tree can be calculated with the
// following formula:
//
// How to calculate the commitment:
//	C_n = A_n(0,d) where n is the version of the tree (the last index) and
//				   d is the depth of the tree
//  For example, using the previous tree the commitment to calculate will be
//
//	C_4 = h(0,3) = A_4(0,3)
//
//	Where A_v(i,r) is calculated as follows:
//  A_v(i,0) = H(0||X_i)   if v >= i
//	A_v(i,r) = H(1||A_v(i,r-1)||▢)                      if v < i + pow(2,r-1)
//	           H(1||A_v(i,r-1)||A_v(i+pow(2,r-1),r-1))  if v >= i + pow(2,r-1)
//	A_v(i,r) = FH(i,r)  whenever v >= i + pow(2,r) - 1
//
// The depth of the tree is the maxium layer level, and can be calculated
// with the following formula:
//
//	layer = ceil(log(index))
//
//
type Tree struct {
	frozen Store // already computed nodes, that will not change
	events Store // layer 0 storage
	size   uint64
	hash   *util.Hasher
}

// Returns a new history tree
func NewTree(frozen, events Store, hash *util.Hasher) *Tree {
	return &Tree{
		frozen, events, 0, hash,
	}
}

// Returns the current layer or depth of the tree
func (t *Tree) getDepth(index uint64) uint64 {
	if index == 0 {
		return 0
	}
	return uint64(math.Ceil(math.Log2(float64(index))))
}

// Recursively traverses the tree computing the root node
// using the algorithm documented above.
func (t *Tree) getNode(i, r, v uint64) (*Node, error) {
	var node *Node
	pos := newpos(i, r)
	// try to unfroze first
	if v >= i+pow(2, r)-1 {
		node, err := t.frozen.Get(pos)
		if err == nil {
			return node, nil
		}
	}

	switch {
	case r == 0 && v >= i:
		a, err := t.events.Get(pos)
		if err != nil {
			return nil, err
		}
		digest := t.hash.Do(Zero, a.Digest)
		node = &Node{pos, digest}
		break

	case v < i+pow(2, r-1):
		a, err := t.getNode(i, r-1, v)
		if err != nil {
			return nil, err
		}
		digest := t.hash.Do(One, a.Digest, Zero)
		node = &Node{pos, digest}
		break

	case v >= i+pow(2, r-1):
		A_v1, err := t.getNode(i, r-1, v)
		if err != nil {
			return nil, err
		}
		A_v2, err := t.getNode(i+pow(2, r-1), r-1, v)
		if err != nil {
			return nil, err
		}
		digest := t.hash.Do(One, A_v1.Digest, A_v2.Digest)
		node = &Node{pos, digest}
		break
	}

	// froze the node with its new digest
	if v >= i+pow(2, r)-1 {
		err := t.frozen.Add(node)
		if err != nil {
			// if it was already frozen nothing happens
		}
	}

	return node, nil
}

// Given an event the system appends it to the history tree as
// the i:th entry and then outputs a commitment
// t.ps://eprint.iacr.org/2015/007.pdf
func (t *Tree) Add(key []byte, value []byte) ([]byte, error) {

}

func (t *Tree) Add(data []byte) (*Commitment, *Node, error) {

	node := &Node{
		Pos:    &Position{t.size, 0},
		Digest: t.hash.Do(data),
	}

	// add event to storage
	if err := t.events.Add(node); err != nil {
		return nil, nil, err
	}

	// increase tree size
	t.size += 1

	// calculate commitment as C_n = A_n(0,d)
	d := t.getDepth(t.size)
	v := t.size - 1
	rootNode, err := t.getNode(0, d, v)
	if err != nil {
		// TODO: rollback inclusion in storage if we cannot calculate a commitment
		return nil, nil, err
	}
	C_n := &Commitment{
		Version: v,
		Digest:  rootNode.Digest,
	}
	return C_n, node, nil

}

// Utility to calculate arbitraty pow and return an int64
func pow(x, y uint64) uint64 {
	return uint64(math.Pow(float64(x), float64(y)))
}

// Utility to allocate a new Position
func newpos(i, l uint64) *Position {
	return &Position{i, l}
}
