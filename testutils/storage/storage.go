package storage

import (
	"fmt"
	"os"

	bd "github.com/bbva/qed/db/badger"
	bp "github.com/bbva/qed/db/bplus"
	"github.com/bbva/qed/storage/badger"
	"github.com/bbva/qed/storage/bolt"
	"github.com/bbva/qed/storage/bplus"
)

func NewBPlusTreeStore() (*bp.BPlusTreeStore, func()) {
	store := bp.NewBPlusTreeStore()
	return store, func() {
		store.Close()
	}
}

func NewBadgerStore(path string) (*bd.BadgerStore, func()) {
	store := bd.NewBadgerStore(path)
	return store, func() {
		store.Close()
		deleteFile(path)
	}
}

func NewBPlusStorage() (*bplus.BPlusTreeStorage, func()) {
	store := bplus.NewBPlusTreeStorage()
	return store, func() {
		store.Close()
	}
}

func NewBoltStorage(path string) (*bolt.BoltStorage, func()) {
	store := bolt.NewBoltStorage(path, "test")
	return store, func() {
		store.Close()
		deleteFile(path)
	}
}

func NewBadgerStorage(path string) (*badger.BadgerStorage, func()) {
	store := badger.NewBadgerStorage(path)
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
