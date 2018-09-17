package common

import (
	"fmt"
	"os"

	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/storage/badger"
)

func OpenBadgerStore(path string) (*badger.BadgerStore, func()) {
	store := badger.NewBadgerStore(path)
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

type FakeCache struct {
	FixedDigest hashing.Digest
}

func NewFakeCache(fixedDigest hashing.Digest) *FakeCache {
	return &FakeCache{fixedDigest}
}

func (c FakeCache) Get(Position) (hashing.Digest, bool) {
	return hashing.Digest{0x0}, true
}
