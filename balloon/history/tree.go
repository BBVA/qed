/*
   Copyright 2018 Banco Bilbao Vizcaya Argentaria, S.A.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

// Package history implements a history tree structure as described in the paper
//     Balloon: A Forward-Secure Append-Only Persistent Authenticated Data
//     Structure
//     https://eprint.iacr.org/2015/007
package history

import (
	"bytes"
	"encoding/binary"
	"errors"
	"math"
	"sync"

	"github.com/bbva/qed/balloon/position"
	"github.com/bbva/qed/balloon/proof"
	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/metrics"
)

// A History tree is a tree structure with a version metadata.
// As described in the pag. 6-7 of the paper:
// http://tamperevident.cs.rice.edu/papers/paper-treehist.pdf
//
// For example, a tree with 5 leaf hashes x0, x1, x2, x3, x4
//    version 4
//
//    layer 3                 ____________ h(0,3)______________
//                           |                                |
//    layer 2       _______h(0,2)_______                 ___h(4,2)___
//                  |                  |                 |          |
//    layer 1   __h(0,1)__         __h(2,1)__         __h(4,1)__    ▢
//              |        |         |      º  |        |        |
//    layer 0  x0(0,0) x1(1,0) x2(2,0)   x3(3,0)    x4(4,0)    ▢
//
//
// Each a is a pair of index and layer (a position). The index starts in 0.
// The size of the tree is the last index number plus 1.
// The next element of the tree can be calculated with the
// following formula:
//
// How to calculate the commitment:
// C_n = A_n(0,d) where n is the version of the tree (the last index) and
// d is the depth of the tree
//
// For example, using the previous tree the commitment to calculate will be
//  C_4 = h(0,3) = A_4(0,3)
//
// Where A_v(i,r) is calculated as follows:
//  A_v(i,0) = H(0||X_i)   if v >= i
//  A_v(i,r) = H(1||A_v(i,r-1)||▢)                      if v <  i + pow(2,r-1)
//             H(1||A_v(i,r-1)||A_v(i+pow(2,r-1),r-1))  if v >= i + pow(2,r-1)
//  A_v(i,r) = FH(i,r)                                 w/e v >= i + pow(2,r)-1
//
// The depth of the tree is the maxium layer level, and can be calculated
// with the following formula:
//  layer = ceil(log(index))
//

type Tree struct {
	treeId       []byte
	frozen       Store
	hasher       hashing.Hasher
	interiorHash hashing.InteriorHasher
	leafHash     hashing.LeafHasher
	sync.RWMutex
}

// NewTree returns a new history Tree struct.
func NewTree(id string, frozen Store, hasher hashing.Hasher) *Tree {
	var m sync.RWMutex

	t := &Tree{
		[]byte(id),
		frozen,
		hasher,
		hashing.InteriorHasherF(hasher),
		hashing.LeafHasherF(hasher),
		m,
	}

	return t
}

func (t Tree) Close() {
	t.Lock()
	defer t.Unlock()
	t.frozen.Close()
}

func (t *Tree) Add(eventDigest, index []byte) ([]byte, error) {
	t.Lock()
	defer t.Unlock()

	version := binary.LittleEndian.Uint64(index)

	// calculate commitment as C_n = A_n(0,d)
	rootDigest, err := t.computeHash(eventDigest, NewRootPosition(version), version)
	if err != nil {
		return nil, err
	}
	return rootDigest, nil
}

func (t Tree) ProveMembership(key []byte, index, version uint64) (*proof.Proof, error) {
	t.Lock()
	defer t.Unlock()

	if index < 0 || index > version {
		return nil, errors.New("invalid index, has to be: 0 <= index <= version")
	}

	ap := make(proof.AuditPath)
	pos := NewRootPosition(version)
	err := t.auditPath(key, index, version, pos, ap)
	if err != nil {
		return nil, err
	}

	return proof.NewProof(pos, ap, t.hasher), nil
}

func (t Tree) ProveIncremental(startKey, endKey []byte, startVersion, endVersion uint64) (*IncrementalProof, error) {
	t.Lock()
	defer t.Unlock()

	if startVersion < 0 || startVersion > endVersion {
		return nil, errors.New("invalid startVersion, has to be: 0 <= endVersion <= endVersion")
	}

	ap := make(proof.AuditPath)
	endPosition := NewRootPosition(endVersion)

	err := t.incAuditPath(startKey, endKey, startVersion, endVersion, endPosition, ap)
	if err != nil {
		return nil, err
	}
	return NewIncrementalProof(startVersion, endVersion, ap, t.interiorHash, t.leafHash), nil
}

func (t Tree) getDepth(index uint64) uint64 {
	return uint64(math.Ceil(math.Log2(float64(index + 1))))
}

// uInt64AsBytes returns the []byte representation of a unit64
func uInt64AsBytes(i uint64) []byte {
	valuebytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(valuebytes, i)
	return valuebytes
}

func (t *Tree) computeHash(eventDigest []byte, pos position.Position, version uint64) ([]byte, error) {

	var digest []byte

	// try to unfreeze it first
	if pos.ShouldBeCached() {
		metrics.History.Add("unfreezing", 1)
		digest, err := t.frozen.Get(pos.Id())
		if err == nil && len(digest) != 0 {
			return digest, nil
		}
	}

	direction := pos.Direction(uInt64AsBytes(version))

	switch {
	case direction == position.Halt && pos.IsLeaf():
		metrics.History.Add("leaf_hashes", 1)
		digest = t.leafHash(pos.Id(), eventDigest)
	case direction == position.Left:
		hash, err := t.computeHash(eventDigest, pos.Left(), version)
		if err != nil {
			return nil, err
		}
		metrics.History.Add("leaf_hashes", 1)
		digest = t.leafHash(pos.Id(), hash)
	case direction == position.Right:
		hash1, err := t.computeHash(eventDigest, pos.Left(), version)
		if err != nil {
			return nil, err
		}
		hash2, err := t.computeHash(eventDigest, pos.Right(), version)
		if err != nil {
			return nil, err
		}
		metrics.History.Add("internal_hashes", 1)
		digest = t.interiorHash(pos.Id(), hash1, hash2)
	}

	// freeze the node with its new digest
	if pos.ShouldBeCached() {
		metrics.History.Add("freezing", 1)
		err := t.frozen.Add(pos.Id(), digest)
		if err != nil {
			// if it was already frozen nothing happens
		}
	}

	return digest, nil
}

func (t Tree) auditPath(key []byte, targetIndex, version uint64, pos position.Position, ap proof.AuditPath) (err error) {

	direction := pos.Direction(uInt64AsBytes(targetIndex))

	switch {
	case direction == position.Halt && pos.IsLeaf():
		err = t.appendHashToPath(key, pos, version, ap)
		return
	case direction == position.Left:
		right := pos.Right()
		if bytes.Compare(right.Key(), uInt64AsBytes(version)) <= 0 {
			t.appendHashToPath(key, right, version, ap)
		}
		return t.auditPath(key, targetIndex, version, pos.Left(), ap)
	case direction == position.Right:
		t.appendHashToPath(key, pos.Left(), version, ap)
		return t.auditPath(key, targetIndex, version, pos.Right(), ap)
	default:
		panic("WTF")
		return
	}

}

func (t Tree) appendHashToPath(key []byte, pos position.Position, version uint64, ap proof.AuditPath) (err error) {
	hash, err := t.computeHash(key, pos, version)
	if err != nil {
		return err
	}
	ap[pos.StringId()] = hash
	return
}

func (t Tree) incAuditPath(startKey, endKey []byte, startIndex, endIndex uint64, pos position.Position, ap proof.AuditPath) (err error) {
	startDirection := pos.Direction(uInt64AsBytes(startIndex))
	endDirection := pos.Direction(uInt64AsBytes(endIndex))

	switch {
	case startDirection == endDirection && startDirection == position.Halt && pos.IsLeaf():
		t.appendHashToPath(endKey, pos, endIndex, ap)
	case startDirection == endDirection && startDirection == position.Left:
		err = t.incAuditPath(startKey, endKey, startIndex, endIndex, pos.Left(), ap)
	case startDirection == endDirection && startDirection == position.Right:
		t.appendHashToPath(endKey, pos.Left(), endIndex, ap)
		err = t.incAuditPath(startKey, endKey, startIndex, endIndex, pos.Right(), ap)
	case startDirection == position.Left:
		err = t.auditPath(startKey, startIndex, endIndex, pos.Left(), ap) // debería ser el last key, no endIndex
		err = t.auditPath(endKey, endIndex, endIndex, pos.Right(), ap)
	}
	return err
}
