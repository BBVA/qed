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

package hyper

import (
	"os"

	"github.com/bbva/qed/balloon/hashing"
	"github.com/bbva/qed/balloon/hyper/storage"
	"github.com/bbva/qed/log"
	"github.com/bbva/qed/storage/badger"
	"github.com/bbva/qed/storage/bolt"
	"github.com/bbva/qed/storage/bplus"
)

func fakeLeafHasherF(hasher hashing.Hasher) hashing.LeafHasher {
	return func(id, base []byte) []byte {
		return hasher(id, base)
	}
}

func fakeInteriorHasherF(hasher hashing.Hasher) hashing.InteriorHasher {
	return func(id, left, right []byte) []byte {
		return hasher(left, right)
	}
}

func openBPlusStorage() (*bplus.BPlusTreeStorage, func()) {
	store := bplus.NewBPlusTreeStorage()
	return store, func() {
		store.Close()
	}
}

func openBadgerStorage(path string) (*badger.BadgerStorage, func()) {
	store := badger.NewBadgerStorage(path)
	return store, func() {
		store.Close()
		deleteFile(path)
	}
}

func openBoltStorage(path string) (*bolt.BoltStorage, func()) {
	store := bolt.NewBoltStorage(path, "test")
	return store, func() {
		store.Close()
		deleteFile(path)
	}
}

func deleteFile(path string) {
	err := os.RemoveAll(path)
	if err != nil {
		log.Debugf("Unable to remove db file %s", err)
	}
}

func NewFakeTree(id string, cache storage.Cache, leaves storage.Store, hasher hashing.Hasher) *Tree {

	tree := NewTree(id, cache, leaves, hasher)
	tree.leafHasher = fakeLeafHasherF(hasher)
	tree.interiorHasher = fakeInteriorHasherF(hasher)

	return tree
}

func NewFakeProof(id string, auditPath [][]byte, hasher hashing.Hasher) *Proof {

	proof := NewProof(id, auditPath, hasher)
	proof.leafHasher = fakeLeafHasherF(hasher)
	proof.interiorHasher = fakeInteriorHasherF(hasher)

	return proof
}
