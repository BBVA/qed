/*
   Copyright 2018-2019 Banco Bilbao Vizcaya Argentaria, S.A.

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

package bplus

import (
	"bytes"
	"io"

	"github.com/bbva/qed/metrics"

	"github.com/bbva/qed/storage"
	"github.com/google/btree"
)

type BPlusTreeStore struct {
	db *btree.BTree
}

func NewBPlusTreeStore() *BPlusTreeStore {
	return &BPlusTreeStore{btree.New(2)}
}

func (s *BPlusTreeStore) Mutate(mutations []*storage.Mutation) error {
	for _, m := range mutations {
		key := append([]byte{m.Table.Prefix()}, m.Key...)
		s.db.ReplaceOrInsert(KVItem{key, m.Value})
	}
	return nil
}

func (s BPlusTreeStore) GetRange(table storage.Table, start, end []byte) (storage.KVRange, error) {
	result := make(storage.KVRange, 0)
	startKey := append([]byte{table.Prefix()}, start...)
	endKey := append([]byte{table.Prefix()}, end...)
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

func (s BPlusTreeStore) Get(table storage.Table, key []byte) (*storage.KVPair, error) {
	result := new(storage.KVPair)
	result.Key = key
	k := append([]byte{table.Prefix()}, key...)
	item := s.db.Get(KVItem{k, nil})
	if item != nil {
		result.Value = item.(KVItem).Value
		return result, nil
	} else {
		return nil, storage.ErrKeyNotFound
	}
}

func (s BPlusTreeStore) GetLast(table storage.Table) (*storage.KVPair, error) {
	result := new(storage.KVPair)
	s.db.DescendGreaterThan(KVItem{[]byte{table.Prefix()}, nil}, func(i btree.Item) bool {
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

func (s BPlusTreeStore) GetAll(table storage.Table) storage.KVPairReader {
	return NewBPlusKVPairReader(table, s.db)
}

func (s BPlusTreeStore) Close() error {
	s.db.Clear(false)
	return nil
}

func (s BPlusTreeStore) Dump(w io.Writer, id uint64) error {
	panic("Not implemented")
}

func (s BPlusTreeStore) Load(r io.Reader) error {
	panic("Not implemented")
}

func (s BPlusTreeStore) Snapshot() (uint64, error) {
	panic("Not implemented")
}

func (s BPlusTreeStore) Backup(metatada string) error {
	panic("Not implemented")
}

func (s *BPlusTreeStore) GetBackupsInfo() []storage.BackupInfo {
	panic("Not implemented")
}

func (s BPlusTreeStore) RestoreFromBackup(backupID uint32, dbDir, walDir string) error {
	panic("Not implemented")
}

func (s BPlusTreeStore) RegisterMetrics(registry metrics.Registry) {
	panic("Not implemented")
}

type KVItem struct {
	Key, Value []byte
}

func (p KVItem) Less(b btree.Item) bool {
	return bytes.Compare(p.Key, b.(KVItem).Key) < 0
}

type BPlusKVPairReader struct {
	prefix  byte
	db      *btree.BTree
	lastKey []byte
}

func NewBPlusKVPairReader(table storage.Table, db *btree.BTree) *BPlusKVPairReader {
	return &BPlusKVPairReader{
		prefix:  table.Prefix(),
		db:      db,
		lastKey: []byte{table.Prefix()},
	}
}

func (r *BPlusKVPairReader) Read(buffer []*storage.KVPair) (n int, err error) {
	n = 0
	r.db.AscendGreaterOrEqual(KVItem{r.lastKey, nil}, func(i btree.Item) bool {
		if n >= len(buffer) {
			return false
		}
		key := i.(KVItem).Key

		if bytes.Compare(key[:1], r.lastKey[:1]) == 0 && bytes.Compare(key, r.lastKey) != 0 {
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
