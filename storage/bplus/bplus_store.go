package bplus

import (
	"bytes"

	"github.com/bbva/qed/storage"
	"github.com/google/btree"
)

type BPlusTreeStore struct {
	db *btree.BTree
}

func NewBPlusTreeStore() *BPlusTreeStore {
	return &BPlusTreeStore{btree.New(2)}
}

type KVItem struct {
	Key, Value []byte
}

func (p KVItem) Less(b btree.Item) bool {
	return bytes.Compare(p.Key, b.(KVItem).Key) < 0
}

func (s *BPlusTreeStore) Mutate(mutations []storage.Mutation) error {
	for _, m := range mutations {
		key := append([]byte{m.Prefix}, m.Key...)
		s.db.ReplaceOrInsert(KVItem{key, m.Value})
	}
	return nil
}

func (s BPlusTreeStore) GetRange(prefix byte, start, end []byte) (storage.KVRange, error) {
	result := make(storage.KVRange, 0)
	startKey := append([]byte{prefix}, start...)
	endKey := append([]byte{prefix}, end...)
	s.db.AscendGreaterOrEqual(KVItem{startKey, nil}, func(i btree.Item) bool {
		key := i.(KVItem).Key
		if bytes.Compare(key, endKey) > 0 {
			return false
		}
		result = append(result, storage.KVPair{key[1:], i.(KVItem).Value})
		return true
	})
	return result, nil
}

func (s BPlusTreeStore) Get(prefix byte, key []byte) (*storage.KVPair, error) {
	result := new(storage.KVPair)
	result.Key = key
	k := append([]byte{prefix}, key...)
	item := s.db.Get(KVItem{k, nil})
	if item != nil {
		result.Value = item.(KVItem).Value
		return result, nil
	} else {
		return nil, storage.ErrKeyNotFound
	}
}

func (s BPlusTreeStore) GetLast(prefix byte) (*storage.KVPair, error) {
	result := new(storage.KVPair)
	s.db.DescendGreaterThan(KVItem{[]byte{prefix}, nil}, func(i btree.Item) bool {
		item := i.(KVItem)
		result.Key = item.Key[1:]
		result.Value = item.Value
		return false
	})
	if result.Key == nil {
		return nil, storage.ErrKeyNotFound
	}
	return result, nil
}

func (s BPlusTreeStore) GetAll(prefix byte) storage.KVPairReader {
	return NewBPlusKVPairReader(prefix, s.db)
}

type BPlusKVPairReader struct {
	prefix  byte
	db      *btree.BTree
	lastKey []byte
}

func NewBPlusKVPairReader(prefix byte, db *btree.BTree) *BPlusKVPairReader {
	return &BPlusKVPairReader{
		prefix:  prefix,
		db:      db,
		lastKey: []byte{prefix},
	}
}

func (r *BPlusKVPairReader) Read(buffer []*storage.KVPair) (n int, err error) {
	n = 0
	r.db.AscendGreaterOrEqual(KVItem{r.lastKey, nil}, func(i btree.Item) bool {
		if n >= len(buffer) {
			return false
		}
		key := i.(KVItem).Key
		if bytes.Compare(key, r.lastKey) != 0 {
			buffer[n] = &storage.KVPair{key[1:], i.(KVItem).Value}
			n++
		}
		r.lastKey = key
		return true
	})
	return n, nil
}

func (r *BPlusKVPairReader) Close() {
	r.db = nil
}

func (s BPlusTreeStore) Close() error {
	s.db.Clear(false)
	return nil
}
