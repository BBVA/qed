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
	stats  *stats
	ops    chan interface{} // serialize operations
}

// NewTree returns a new history tree
func NewTree(frozen storage.Store, hasher hashing.Hasher) *Tree {
	t := &Tree{
		frozen,
		hasher,
		new(stats),
		nil,
	}
	// start tree goroutine to handle
	// tree operations
	t.ops = t.operations()

	return t
}

// Given an event the system appends it to the history tree as
// the i:th entry and then outputs a root hash as a commitment
// t.ps://eprint.iacr.org/2015/007.pdf
func (t *Tree) add(eventDigest []byte, index []byte) ([]byte, error) {
	version := binary.LittleEndian.Uint64(index)
	// calculate commitment as C_n = A_n(0,d)
	depth := t.getDepth(version)
	rootDigest, err := t.rootHash(eventDigest, 0, depth, version)
	if err != nil {
		return nil, err
	}
	return rootDigest, nil
}

// Returns the current layer or depth of the tree
func (t *Tree) getDepth(index uint64) uint64 {
	return uint64(math.Ceil(math.Log2(float64(index + 1))))
}

func uInt64AsBytes(i uint64) []byte {
	valuebytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(valuebytes, i)
	return valuebytes
}

func frozenKey(index, layer uint64) []byte {
	return append(uInt64AsBytes(index), uInt64AsBytes(layer)...)
}

func (t *Tree) rootHash(eventDigest []byte, index, layer uint64, version uint64) ([]byte, error) {

	var digest []byte

	// try to unfroze first
	if version >= index+pow(2, layer)-1 {
		t.stats.unfreezing++
		digest, err := t.frozen.Get(frozenKey(index, layer))
		if err == nil {
			t.stats.unfreezingHits++
			return digest, nil
		}
	}

	switch {
	// we are at a leaf: A_v(i,0)
	case layer == 0 && version >= index:
		digest = t.hasher(Zero, eventDigest)
		t.stats.leafHashes++
		break
	// A_v(i,r)
	case version < index+pow(2, layer-1):
		hash, err := t.rootHash(eventDigest, index, layer-1, version)
		if err != nil {
			return nil, err
		}
		digest = t.hasher(One, hash)
		t.stats.internalHashes++
		break
	// A_v(i,r)
	case version >= index+pow(2, layer-1):
		hash1, err := t.rootHash(eventDigest, index, layer-1, version)
		if err != nil {
			return nil, err
		}
		hash2, err := t.rootHash(eventDigest, index+pow(2, layer-1), layer-1, version)
		if err != nil {
			return nil, err
		}
		digest = t.hasher(One, hash1, hash2)
		t.stats.internalHashes++
		break
	}

	// froze the node with its new digest
	if version >= index+pow(2, layer)-1 {
		t.stats.freezing++
		err := t.frozen.Add(frozenKey(index, layer), digest)
		if err != nil {
			// if it was already frozen nothing happens
		}
	}

	return digest, nil
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
					t.frozen.Close()
					msg.result <- true
					return
				case *add:
					digest, _ := t.add(msg.digest, msg.index)
					msg.result <- digest
				default:
					panic("Hyper tree Run() message not implemented!!")
				}

			}
		}
	}()
	return operations
}

// Internally we use a channel API to serialize operations
// but external we use exported methods to be called
// by others.
// These methods returns a channel with an appropriate type
// for each operation to be consumed from when the data arrives.

type add struct {
	digest []byte
	index  []byte
	result chan []byte
}

// Queues an Add operation to the tree and returns a channel
// when the result []byte will be sent when ready
func (t Tree) Add(digest, index []byte) chan []byte {
	result := make(chan []byte, 0)
	t.ops <- &add{
		digest,
		index,
		result,
	}
	return result
}

type close struct {
	stop   bool
	result chan bool
}

// Queues a Stop operation to the tree and returns a channel
// were a true or false will be send when the operation is completed
func (t Tree) Close() chan bool {
	result := make(chan bool)
	t.ops <- &close{true, result}
	return result
}

// Utility to calculate arbitraty pow and return an int64
func pow(x, y uint64) uint64 {
	return uint64(math.Pow(float64(x), float64(y)))
}
