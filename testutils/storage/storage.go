package storage

import (
	"fmt"
	"os"

	"github.com/bbva/qed/storage/badger"
	bd "github.com/bbva/qed/storage/badger"
	"github.com/bbva/qed/storage/bplus"
	bp "github.com/bbva/qed/storage/bplus"
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

func NewBPlusStorage() (*bplus.BPlusTreeStore, func()) {
	store := bplus.NewBPlusTreeStore()
	return store, func() {
		store.Close()
	}
}

func NewBadgerStorage(path string) (*badger.BadgerStore, func()) {
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
