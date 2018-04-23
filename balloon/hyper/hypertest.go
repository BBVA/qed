package hyper

import (
	"bytes"
	"fmt"
	"os"
	"verifiabledata/balloon/hashing"
	"verifiabledata/balloon/storage/badger"
	"verifiabledata/balloon/storage/bolt"
	"verifiabledata/balloon/storage/bplus"
)

func FakeLeafHasherF(hasher hashing.Hasher) LeafHasher {
	return func(id, value, base []byte) []byte {
		if bytes.Equal(value, Empty) {
			return hasher(Empty)
		}
		return hasher(base)
	}
}

func FakeInteriorHasherF(hasher hashing.Hasher) InteriorHasher {
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
		fmt.Printf("Unable to remove db file %s", err)
	}
}
