package hyper

import (
	"bytes"
	"os"

	"verifiabledata/balloon/hashing"
	"verifiabledata/balloon/storage"
	"verifiabledata/balloon/storage/badger"
	"verifiabledata/balloon/storage/bolt"
	"verifiabledata/balloon/storage/bplus"
	"verifiabledata/log"
)

func fakeLeafHasherF(hasher hashing.Hasher) leafHasher {
	return func(id, value, base []byte) []byte {
		if bytes.Equal(value, Empty) {
			return hasher(Empty)
		}
		return hasher(base)
	}
}

func fakeInteriorHasherF(hasher hashing.Hasher) interiorHasher {
	return func(left, right, base, height []byte) []byte {
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
