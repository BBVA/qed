package storage

import (
	"bytes"
	"fmt"

	"github.com/cznic/b"
)

type BPlusTreeStorage struct {
	store *b.Tree
}

func NewBPlusTreeStorage() *BPlusTreeStorage {
	return &BPlusTreeStorage{b.TreeNew(cmp)}
}

func cmp(a, b interface{}) int {
	return bytes.Compare(a.([]byte), b.([]byte))
}

func (s *BPlusTreeStorage) Add(key []byte, value []byte) error {
	s.store.Set(key, value)
	return nil
}

func (s *BPlusTreeStorage) Get(key []byte) ([]byte, error) {
	value, ok := s.store.Get(key)
	if ok == false {
		return nil, fmt.Errorf("Unknown key %d", key)
	}
	return value.([]byte), nil
}

func (s *BPlusTreeStorage) GetRange(start, end []byte) LeavesSlice {
	var leaves LeavesSlice
	var err error
	var k interface{}

	iter, _ := s.store.Seek(start)
	defer iter.Close()

	n := 0
	for {
		k, _, err = iter.Next()
		if err != nil {
			return leaves
		}
		if bytes.Compare(k.([]byte), end) < 0 {
			leaves = append(leaves, k.([]byte))
		} else {
			return leaves
		}
		n++
	}

}

func (s *BPlusTreeStorage) Close() error {
	s.store.Close()
	return nil
}
