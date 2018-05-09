package bplus

import (
	"bytes"
	"errors"

	"github.com/google/btree"

	"qed/balloon/storage"
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
