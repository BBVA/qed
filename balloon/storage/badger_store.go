package storage

import (
	"bytes"
	"log"

	"github.com/dgraph-io/badger"
)

type BadgerStorage struct {
	db *badger.DB
}

func (s *BadgerStorage) Add(key []byte, value []byte) error {
	return s.db.Update(func(txn *badger.Txn) error {
		err := txn.Set(key, value)
		return err
	})
}

func (s *BadgerStorage) Get(key []byte) ([]byte, error) {
	var value []byte
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		v, err := item.Value()
		if err != nil {
			return err
		}
		value = make([]byte, len(v))
		copy(value, v)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return value, nil
}

func (s *BadgerStorage) GetRange(start, end []byte) LeavesSlice {
	var leaves LeavesSlice

	s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Seek(start); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()
			if bytes.Compare(k, end) > 0 {
				break
			}
			leaves = append(leaves, k)
		}
		return nil
	})

	return leaves
}

func NewBadgerStorage(path string) *BadgerStorage {
	opts := badger.DefaultOptions
	opts.Dir = path
	opts.ValueDir = path
	db, err := badger.Open(opts)
	if err != nil {
		log.Fatal(err)
	}
	return &BadgerStorage{db}
}
