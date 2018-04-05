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
	"encoding/binary"
	"math"
	"verifiabledata/balloon/hashing"
	"verifiabledata/balloon/storage"
)

// Constant Zero is the 0x0 byte, and it is used as a prefix to know
// if a node has a zero digest.
var Zero = []byte{0x0}

// Constant One is the 0x1 byte, and it is used as a prefix to know
// if a node has a non-zero digest.
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
type Tree struct {
	frozen storage.Store // already computed nodes, that will not change
	hasher hashing.Hasher
}

// NewTree returns a new history tree
func NewTree(frozen storage.Store, hasher hashing.Hasher) *Tree {
	return &Tree{
		frozen,
		hasher,
	}
}

// Returns the current layer or depth of the tree
func (t *Tree) getDepth(index uint64) uint64 {
	return uint64(math.Ceil(math.Log2(float64(index + 1))))
}

// Given an event the system appends it to the history tree as
// the i:th entry and then outputs a root hash as a commitment
// t.ps://eprint.iacr.org/2015/007.pdf
func (t *Tree) Add(eventDigest []byte, index []byte) ([]byte, error) {
	version := binary.LittleEndian.Uint64(index)
	// calculate commitment as C_n = A_n(0,d)
	depth := t.getDepth(version)
	rootPos := newPosition(0, depth)
	rootDigest, err := t.computeNodeHash(eventDigest, rootPos, version)
	if err != nil {
		return nil, err
	}
	return rootDigest, nil
}

func (t *Tree) computeNodeHash(eventDigest []byte, pos *Position, version uint64) ([]byte, error) {

	var digest []byte

	// try to unfroze first
	if version >= pos.Index+pow(2, pos.Layer)-1 {
		digest, err := t.frozen.Get(pos.GetBytes())
		if err == nil {
			return digest, nil
		}
	}

	switch {
	// we are at a leaf: A_v(i,0)
	case pos.Layer == 0 && version >= pos.Index:
		digest = t.hasher(Zero, eventDigest)
		break
	// A_v(i,r)
	case version < pos.Index+pow(2, pos.Layer-1):
		newPos := newPosition(pos.Index, pos.Layer-1)
		hash, err := t.computeNodeHash(eventDigest, newPos, version)
		if err != nil {
			return nil, err
		}
		digest = t.hasher(One, hash)
		break
	// A_v(i,r)
	case version >= pos.Index+pow(2, pos.Layer-1):
		newPos1 := newPosition(pos.Index, pos.Layer-1)
		hash1, err := t.computeNodeHash(eventDigest, newPos1, version)
		if err != nil {
			return nil, err
		}
		newPos2 := newPosition(pos.Index+pow(2, pos.Layer-1), pos.Layer-1)
		hash2, err := t.computeNodeHash(eventDigest, newPos2, version)
		if err != nil {
			return nil, err
		}
		digest = t.hasher(One, hash1, hash2)
		break
	}

	// froze the node with its new digest
	if version >= pos.Index+pow(2, pos.Layer)-1 {
		err := t.frozen.Add(pos.GetBytes(), digest)
		if err != nil {
			// if it was already frozen nothing happens
		}
	}

	return digest, nil
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
					digest, _ := t.Add(msg.digest, msg.index)
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

func NewAdd(digest, index []byte) (*Add, chan []byte) {
	result := make(chan []byte)
	return &Add{
		digest,
		index,
		result,
	}, result
}

type Stop struct {
	stop bool
	result chan bool
}

func NewStop() (*Stop, chan bool) {
	result := make(chan bool)
	return &Stop{true, result}, result
}
// Utility to calculate arbitraty pow and return an int64
func pow(x, y uint64) uint64 {
	return uint64(math.Pow(float64(x), float64(y)))
}
