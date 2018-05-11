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

// Package bplus implements the storage engine interface for
// github.com/google/btree
package bplus

import (
	"bytes"
	"errors"

	"github.com/google/btree"

	"github.com/bbva/qed/balloon/storage"
)

type BPlusTreeStorage struct {
	store *btree.BTree
}

type KVPair struct {
	Key, Value []byte
}

func (p KVPair) Less(b btree.Item) bool {
	return bytes.Compare(p.Key, b.(KVPair).Key) < 0
}

func NewBPlusTreeStorage() *BPlusTreeStorage {
	return &BPlusTreeStorage{btree.New(2)}
}

func (s *BPlusTreeStorage) Add(key []byte, value []byte) error {
	s.store.ReplaceOrInsert(KVPair{key, value})
	return nil
}

func (s *BPlusTreeStorage) Get(key []byte) ([]byte, error) {
	item := s.store.Get(KVPair{key, nil})
	if item == nil {
		return make([]byte, 0), nil
	}
	return item.(KVPair).Value, nil
}

func (s *BPlusTreeStorage) GetRange(start, end []byte) storage.LeavesSlice {
	var leaves storage.LeavesSlice
	s.store.AscendGreaterOrEqual(KVPair{start, nil}, func(i btree.Item) bool {
		if bytes.Compare(i.(KVPair).Key, end) > 0 {
			return false
		}
		leaves = append(leaves, i.(KVPair).Key)
		return true
	})
	return leaves
}

func (s *BPlusTreeStorage) Delete([]byte) error { return errors.New("not implemented") }

func (s *BPlusTreeStorage) Close() error {
	s.store.Clear(false)
	return nil
}
