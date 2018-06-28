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

package hyper

import (
	"math"
	"sync"

	"github.com/bbva/qed/balloon/position"
	"github.com/bbva/qed/balloon/proof"
	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/log"
	"github.com/bbva/qed/metrics"
)

type Tree struct {
	id            []byte
	cache         Cache
	leaves        Store
	cacheLevel    uint64
	numBits       uint64
	hasher        hashing.Hasher
	interiorHash  hashing.InteriorHasher
	leafHash      hashing.LeafHasher
	defaultHashes [][]byte
	sync.RWMutex
}

// NewTree returns  a new Hyper Tree given all its dependencies
func NewTree(id string, cache Cache, leaves Store, hasher hashing.Hasher) *Tree {
	var m sync.RWMutex
	cacheLevels := uint64(math.Max(0.0, math.Floor(math.Log(float64(cache.Size()))/math.Log(2.0))))
	numBits := hasher.Len()

	tree := &Tree{
		[]byte(id),
		cache,
		leaves,
		numBits - cacheLevels,
		numBits,
		hasher,
		hashing.InteriorHasherF(hasher),
		hashing.LeafHasherF(hasher),
		make([][]byte, numBits),
		m,
	}

	// init default hashes cache
	tree.defaultHashes[0] = hasher.Do(tree.id, []byte{0x0})
	for i := uint64(1); i < numBits; i++ {
		tree.defaultHashes[i] = hasher.Do(tree.defaultHashes[i-1], tree.defaultHashes[i-1])
	}

	return tree
}

func (t Tree) Close() {
	t.Lock()
	defer t.Unlock()
	// t.cache.Close()
	t.leaves.Close()
}

func (t *Tree) Add(key []byte, value []byte) ([]byte, error) {
	t.Lock()
	defer t.Unlock()

	err := t.leaves.Add(key, value)
	if err != nil {
		return nil, err
	}
	return t.toCache(key, value, NewRootPosition(t.numBits, t.cacheLevel)), nil
}

func (t Tree) ProveMembership(key []byte, value []byte) (*proof.Proof, []byte, error) {
	t.Lock()
	defer t.Unlock()

	value, err := t.leaves.Get(key) // TODO check existence
	if err != nil {
		log.Debug(t.leaves)
		return nil, nil, err
	}

	ap := make(proof.AuditPath)
	rootPos := NewRootPosition(t.numBits, t.cacheLevel)
	t.auditPathFromCache(key, value, rootPos, ap)

	return proof.NewProof(rootPos, ap, t.hasher), value, nil
}

func (t *Tree) toCache(key, value []byte, pos position.Position) []byte {

	// if we are beyond the cache zone
	// we need to go to database to get
	// nodes
	if !pos.ShouldBeCached() {
		metrics.Hyper.Add("storage_reads", 1)
		first := pos.FirstLeaf()
		last := pos.LastLeaf()
		d := t.leaves.GetRange(first.Key(), last.Key())
		return t.fromStorage(d, value, pos)
	}

	// if not, the node hash is the hash of our left and right child
	direction := pos.Direction(key)
	var left, right, digest []byte

	switch {
	case direction == position.Left:
		left = t.toCache(key, value, pos.Left())
		right = t.fromCache(pos.Right())
	case direction == position.Right:
		left = t.fromCache(pos.Left())
		right = t.toCache(key, value, pos.Right())
	case direction == position.Halt:
		//this should never happen => TODO should return an error
		return nil
	}

	metrics.Hyper.Add("interior_hash", 1)
	posId := pos.Id()
	digest = t.interiorHash(posId, left, right)

	// we re-cache all the nodes on each update
	// if the node is whithin the cache area
	if pos.ShouldBeCached() {
		metrics.Hyper.Add("update", 1)
		t.cache.Put(posId, digest)
	}

	return digest
}

func (t *Tree) fromCache(pos position.Position) []byte {

	// get the value from the cache
	cachedHash, cached := t.cache.Get(pos.Id())
	if cached {
		metrics.Hyper.Add("cached_hash", 1)
		return cachedHash
	}

	// if there is no value in the cache,return a default hash
	metrics.Hyper.Add("default_hash", 1)
	return t.defaultHashes[pos.Height()]

}

func (t *Tree) fromStorage(d [][]byte, value []byte, pos position.Position) []byte {

	// if we are a leaf, return our hash
	if len(d) == 1 && pos.IsLeaf() {
		metrics.Hyper.Add("leaf", 1)
		metrics.Hyper.Add("leaf_hash", 1)
		return t.leafHash(d[0], value)
	}

	// if there are no more childs,
	// return a default hash if i'm not in root node
	if len(d) == 0 && pos.Height() != t.numBits {
		metrics.Hyper.Add("default_hash", 1)
		return t.defaultHashes[pos.Height()]
	}

	if len(d) > 0 && pos.IsLeaf() {
		panic("this should never happen (unsorted LeavesSlice or broken split?)")
	}

	rightChild := pos.Right()
	leftSlice, rightSlice := Split(d, rightChild.Key())

	left := t.fromStorage(leftSlice, value, pos.Left())
	right := t.fromStorage(rightSlice, value, rightChild)
	metrics.Hyper.Add("interior_hash", 1)
	return t.interiorHash(pos.Id(), left, right)
}

func (t *Tree) auditPathFromCache(key, value []byte, pos position.Position, ap proof.AuditPath) (err error) {
	// if we are beyond the cache zone
	// we need to go to database to get
	// nodes
	if !pos.ShouldBeCached() {
		first := pos.FirstLeaf()
		last := pos.LastLeaf()
		leaves := t.leaves.GetRange(first.Key(), last.Key())
		return t.auditPathFromStorage(leaves, key, value, pos, ap)
	}

	direction := pos.Direction(key)

	switch {
	case direction == position.Left:
		right := pos.Right()
		ap[right.StringId()] = t.fromCache(right)
		t.auditPathFromCache(key, value, pos.Left(), ap)
	case direction == position.Right:
		left := pos.Left()
		ap[left.StringId()] = t.fromCache(left)
		t.auditPathFromCache(key, value, pos.Right(), ap)
	case direction == position.Halt:
		panic("this should never happen")
	}

	return
}

func (t *Tree) auditPathFromStorage(d [][]byte, key, value []byte, pos position.Position, ap proof.AuditPath) (err error) {

	direction := pos.Direction(key)

	rightChild := pos.Right()
	leftSlice, rightSlice := Split(d, rightChild.Key())

	switch {
	case direction == position.Halt && pos.IsLeaf():
		ap[pos.StringId()] = t.fromStorage(d, value, pos)
	case direction == position.Left:
		right := pos.Right()
		ap[right.StringId()] = t.fromStorage(rightSlice, value, right)
		t.auditPathFromStorage(leftSlice, key, value, pos.Left(), ap)
	case direction == position.Right:
		left := pos.Left()
		ap[left.StringId()] = t.fromStorage(leftSlice, value, left)
		t.auditPathFromStorage(rightSlice, key, value, pos.Right(), ap)
	}

	return

}
