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
	"math"
	"verifiabledata/store"
	"verifiabledata/tree"
	"verifiabledata/util"
)

// Constant Zero is the 0x0 byte, and it is used as a prefix to know
// if a node has a zero digest.
var Zero = []byte{0x0}

// Constant One is the 0x1 byte, and it is used as a prefix to know
// if a node has a non-zero digestg.
var One = []byte{0x1}

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
type HistoryTree struct {
	frozen store.Store // already computed nodes, that will not change
	events store.Store // layer 0 storage
	size   uint64
}

// Returns a new history tree
func NewHistoryTree(frozen, events store.Store) *HistoryTree {
	return &HistoryTree{
		frozen, events, 0,
	}
}

// Returns the current layer or depth of the tree
func (ht *HistoryTree) currentLayer(index uint64) uint64 {
	if index == 0 {
		return 0
	}
	return uint64(math.Ceil(math.Log2(float64(index))))
}

// Returns the current position of the tree
func (ht *HistoryTree) currentPosition() *tree.Position {
	return &tree.Position{ht.size, ht.currentLayer(ht.size)}
}

// Returns the current version of the tree
func (ht *HistoryTree) currentVersion() uint64 {
	return ht.size - 1
}

// Recursively traverses the tree computing the root node
func (ht *HistoryTree) getNode(pos *tree.Position, version uint64) (*tree.Node, error) {
	var node *tree.Node

	// 	whenever v >= i + pow(2,r) - 1
	// 		A_v(i,r) = FH(i,r)
	if version >= util.Pow(2, pos.Layer)-1 {
		node, err := ht.frozen.Get(pos)
		if err == nil {
			return node, nil
		}
	}

	switch {
	// if v >= i
	// 	A_v(i,0) = H(0||X_i)
	case pos.Layer == 0 && version >= pos.Index:
		X_i, err := ht.events.Get(pos)
		if err != nil {
			return nil, err
		}
		digest := util.Hash(Zero, X_i.Digest)
		node = &tree.Node{X_i.Pos, digest}
		break

	// if v < i + pow(2,r-1)
	// 	A_v(i,r) = H(1||A_v(i,r-1)||▢)
	case version < pos.Index+util.Pow(2, pos.Layer-1):
		X_i, err := ht.getNode(pos.SetLayer(pos.Layer-1), version)
		if err != nil {
			return nil, err
		}
		digest := util.Hash(One, X_i.Digest, Zero)
		node = &tree.Node{X_i.Pos, digest}
		break

	// if v >= i + pow(2,r-1)
	// 	H(1||A_v(i,r-1)||A_v(i+pow(2,r-1),r-1))
	case version >= pos.Index+util.Pow(2, pos.Layer-1):
		new_pos := pos.SetLayer(pos.Layer - 1)
		A_v1, err := ht.getNode(new_pos, version)
		if err != nil {
			return nil, err
		}
		A_v2, err := ht.getNode(new_pos.SetIndex(pos.Index+util.Pow(2, pos.Layer-1)), version)
		if err != nil {
			return nil, err
		}
		digest := util.Hash(One, A_v1.Digest, A_v2.Digest)
		node = &tree.Node{pos, digest}
		break
	}

	// froze the node
	if version >= util.Pow(2, pos.Layer)-1 {
		err := ht.frozen.Add(node)
		if err != nil {
			// if it was already frozen nothing happens
		}
	}

	return node, nil
}

// Given an event e the system appends it to the history tree H as
// the i:th event and then outputs a commitment
// https://eprint.iacr.org/2015/007.pdf
func (ht *HistoryTree) Add(data []byte) (*tree.Node, error) {

	// tree version
	version := ht.size

	node := &tree.Node{
		ht.currentPosition().SetLayer(0),
		util.Hash(data),
	}

	// add event to storage
	if err := ht.events.Add(node); err != nil {
		return nil, err
	}

	// increase tree size
	ht.size += 1

	// calculate commitment as C_n = A_n(0,d)
	d := ht.currentLayer(ht.size)
	A_n := &tree.Position{0, d}

	C_n, err := ht.getNode(A_n, version)
	if err != nil {
		// TODO: rollback inclusion in storage if we cannot calculate a commitment
		return nil, err
	}

	return C_n, nil

}
