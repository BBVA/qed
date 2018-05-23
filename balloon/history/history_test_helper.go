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
	"github.com/bbva/qed/balloon/hashing"
	"github.com/bbva/qed/balloon/storage"
	"github.com/bbva/qed/log"
)

// fakeLeafHasherF is a test helper function that prints in debug level the
// hashing event.
func fakeLeafHasherF(hasher hashing.Hasher) leafHasher {
	return func(a, key []byte) []byte {
		digest := hasher(a, key)
		log.Debug("Hashing leaf: a-> %b key-> %b :=> %b\n", a, key, digest)
		return digest
	}
}

// fakeInteriorHasherF is a test helper function that prints in debug level
// the hashing event.
func fakeInteriorHasherF(hasher hashing.Hasher) interiorHasher {
	return func(a, left, right []byte) []byte {
		digest := hasher(a, left, right)
		log.Debug("Hashing interior: a-> %b left-> %b right-> %b :=> %b\n", a, left, right, digest)
		return digest
	}
}

// fakeLeafHasherCleanF is a test helper function that only use the key
// param for the salt.
func fakeLeafHasherCleanF(hasher hashing.Hasher) leafHasher {
	return func(a, key []byte) []byte {
		return hasher(key)
	}
}

// fakeInteriorHasherCleanF is a test helper function that only hasher with
// the left and right paramenters.
func fakeInteriorHasherCleanF(hasher hashing.Hasher) interiorHasher {
	return func(a, left, right []byte) []byte {
		return hasher(left, right)
	}
}

// NewFakeTree is a test helper public function that returns a history.Tree
// pointer
func NewFakeTree(frozen storage.Store, hasher hashing.Hasher) *Tree {

	tree := NewTree(frozen, hasher)
	tree.leafHasher = fakeLeafHasherF(hasher)
	tree.interiorHasher = fakeInteriorHasherF(hasher)

	return tree
}

// NewFakeCleanTree is a test helper public function.
func NewFakeCleanTree(frozen storage.Store, hasher hashing.Hasher) *Tree {

	tree := NewTree(frozen, hasher)
	tree.leafHasher = fakeLeafHasherCleanF(hasher)
	tree.interiorHasher = fakeInteriorHasherCleanF(hasher)

	return tree
}

// NewFakeProof is a test helper public function.
func NewFakeProof(auditPath []Node, index uint64, hasher hashing.Hasher) *Proof {

	proof := NewProof(auditPath, index, hasher)
	proof.leafHasher = fakeLeafHasherF(hasher)
	proof.interiorHasher = fakeInteriorHasherF(hasher)

	return proof
}
