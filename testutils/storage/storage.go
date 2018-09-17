package storage

import (
	"fmt"
	"os"

	"github.com/stretchr/testify/require"

	bd "github.com/bbva/qed/storage/badger"
	bp "github.com/bbva/qed/storage/bplus"
)

func OpenBPlusTreeStore() (*bp.BPlusTreeStore, func()) {
	store := bp.NewBPlusTreeStore()
	return store, func() {
		store.Close()
	}
}

func OpenBadgerStore(t require.TestingT, path string) (*bd.BadgerStore, func()) {
	store, err := bd.NewBadgerStore(path)
	if err != nil {
		t.Errorf("Error opening badger store: %v", err)
		t.FailNow()
	}
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
