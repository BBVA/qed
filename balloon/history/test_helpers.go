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

/* +build !release */

package history

import (
	"encoding/binary"

	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/log"
)

// fakeLeafHasherF is a test helper function that prints in debug level the
// hashing event.
func fakeLeafHasherF(hasher hashing.Hasher) hashing.LeafHasher {
	return func(a, key []byte) []byte {
		digest := hasher.Do(a, key)
		log.Debug("Hashing leaf: a-> %b key-> %b :=> %b\n", a, key, digest)
		return digest
	}
}

// fakeInteriorHasherF is a test helper function that prints in debug level
// the hashing event.
func fakeInteriorHasherF(hasher hashing.Hasher) hashing.InteriorHasher {
	return func(a, left, right []byte) []byte {
		digest := hasher.Do(a, left, right)
		log.Debug("Hashing interior: a-> %b left-> %b right-> %b :=> %b\n", a, left, right, digest)
		return digest
	}
}

// fakeLeafHasherCleanF is a test helper function that only use the key
// param for the salt.
func fakeLeafHasherCleanF(hasher hashing.Hasher) hashing.LeafHasher {
	return func(a, key []byte) []byte {
		return hasher.Do(key)
	}
}

// fakeInteriorHasherCleanF is a test helper function that only hasher with
// the left and right paramenters.
func fakeInteriorHasherCleanF(hasher hashing.Hasher) hashing.InteriorHasher {
	return func(a, left, right []byte) []byte {
		return hasher.Do(left, right)
	}
}

// NewFakeTree is a test helper public function that returns a history.Tree
// pointer
func NewFakeTree(id string, frozen Store, hasher hashing.Hasher) *Tree {

	tree := NewTree(id, frozen, hasher)
	tree.leafHash = fakeLeafHasherF(hasher)
	tree.interiorHash = fakeInteriorHasherF(hasher)

	return tree
}

// NewFakeCleanTree is a test helper public function.
func NewFakeCleanTree(id string, frozen Store, hasher hashing.Hasher) *Tree {

	tree := NewTree(id, frozen, hasher)
	tree.leafHash = fakeLeafHasherCleanF(hasher)
	tree.interiorHash = fakeInteriorHasherCleanF(hasher)

	return tree
}

// NewFakeProof is a test helper public function.
/*
func NewFakeProof(auditPath [][]byte, index uint64, hasher hashing.Hasher) *Proof {

	proof := NewProof(auditPath, index, hasher)
	proof.leafHash = fakeLeafHasherF(hasher)
	proof.interiorHash = fakeInteriorHasherF(hasher)

	return proof
}
*/

func uint64AsBytes(index uint64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, index)
	return b
}

func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}
