// Copyright © 2018 Banco Bilbao Vizcaya Argentaria S.A.  All rights reserved.
// Use of this source code is governed by an Apache 2 License
// that can be found in the LICENSE file

package badger

import (
	"bytes"
	"log"
	"verifiabledata/balloon/storage"

	b "github.com/dgraph-io/badger"
)

type BadgerStorage struct {
	db *b.DB
}

func (s *BadgerStorage) Add(key []byte, value []byte) error {
	return s.db.Update(func(txn *b.Txn) error {
		err := txn.Set(key, value)
		return err
	})
}

func (s *BadgerStorage) Get(key []byte) ([]byte, error) {
	var value []byte
	err := s.db.View(func(txn *b.Txn) error {
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

func (s *BadgerStorage) GetRange(start, end []byte) storage.LeavesSlice {
	var leaves storage.LeavesSlice

	s.db.View(func(txn *b.Txn) error {
		opts := b.DefaultIteratorOptions
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

func (s *BadgerStorage) Close() error {
	return s.db.Close()
}

func NewBadgerStorage(path string) *BadgerStorage {
	opts := b.DefaultOptions
	opts.Dir = path
	opts.ValueDir = path
	db, err := b.Open(opts)
	if err != nil {
		log.Fatal(err)
	}
	return &BadgerStorage{db}
}
