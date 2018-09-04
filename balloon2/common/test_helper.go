package common

import (
	"fmt"
	"os"

	"github.com/bbva/qed/db/badger"
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
	FixedDigest Digest
}

func NewFakeCache(fixedDigest Digest) *FakeCache {
	return &FakeCache{fixedDigest}
}

func (c FakeCache) Get(Position) (Digest, bool) {
	return Digest{0x0}, true
}
