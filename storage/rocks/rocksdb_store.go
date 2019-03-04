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
package rocks

import (
	"bytes"

	"github.com/bbva/qed/rocksdb"
	"github.com/bbva/qed/storage"
)

type RocksDBStore struct {
	db *rocksdb.DB
}

type rocksdbOpts struct {
	Path string
}

func NewRocksDBStore(path string) (*RocksDBStore, error) {
	return NewRocksDBStoreOpts(&rocksdbOpts{Path: path})
}

func NewRocksDBStoreOpts(opts *rocksdbOpts) (*RocksDBStore, error) {
	rocksdbOpts := rocksdb.NewDefaultOptions()
	rocksdbOpts.SetCreateIfMissing(true)
	rocksdbOpts.IncreaseParallelism(4)
	rocksdbOpts.SetMaxWriteBufferNumber(5)
	rocksdbOpts.SetMinWriteBufferNumberToMerge(2)

	blockOpts := rocksdb.NewDefaultBlockBasedTableOptions()
	blockOpts.SetFilterPolicy(rocksdb.NewBloomFilterPolicy(10))
	rocksdbOpts.SetBlockBasedTableFactory(blockOpts)

	db, err := rocksdb.OpenDB(opts.Path, rocksdbOpts)
	if err != nil {
		return nil, err
	}

	store := &RocksDBStore{db: db}
	return store, nil
}

func (s RocksDBStore) Mutate(mutations []*storage.Mutation) error {
	batch := rocksdb.NewWriteBatch()
	defer batch.Destroy()
	for _, m := range mutations {
		key := append([]byte{m.Prefix}, m.Key...)
		batch.Put(key, m.Value)
	}
	err := s.db.Write(rocksdb.NewDefaultWriteOptions(), batch)
	return err
}

func (s RocksDBStore) Get(prefix byte, key []byte) (*storage.KVPair, error) {
	result := new(storage.KVPair)
	result.Key = key
	k := append([]byte{prefix}, key...)
	v, err := s.db.GetBytes(rocksdb.NewDefaultReadOptions(), k)
	if err != nil {
		return nil, err
	}
	if v == nil {
		return nil, storage.ErrKeyNotFound
	}
	result.Value = v
	return result, nil
}

func (s RocksDBStore) GetRange(prefix byte, start, end []byte) (storage.KVRange, error) {
	result := make(storage.KVRange, 0)
	startKey := append([]byte{prefix}, start...)
	endKey := append([]byte{prefix}, end...)
	it := s.db.NewIterator(rocksdb.NewDefaultReadOptions())
	defer it.Close()
	for it.Seek(startKey); it.Valid(); it.Next() {
		keySlice := it.Key()
		key := make([]byte, keySlice.Size())
		copy(key, keySlice.Data())
		if bytes.Compare(key, endKey) > 0 {
			break
		}
		valueSlice := it.Value()
		value := make([]byte, valueSlice.Size())
		copy(value, valueSlice.Data())
		result = append(result, storage.KVPair{key[1:], value})
	}

	return result, nil
}

func (s RocksDBStore) GetLast(prefix byte) (*storage.KVPair, error) {
	it := s.db.NewIterator(rocksdb.NewDefaultReadOptions())
	defer it.Close()
	it.SeekForPrev([]byte{prefix, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
	if it.ValidForPrefix([]byte{prefix}) {
		result := new(storage.KVPair)
		keySlice := it.Key()
		key := make([]byte, keySlice.Size())
		copy(key, keySlice.Data())
		result.Key = key[1:]
		valueSlice := it.Value()
		value := make([]byte, valueSlice.Size())
		copy(value, valueSlice.Data())
		result.Value = value
		return result, nil
	}
	return nil, storage.ErrKeyNotFound
}

type RocksDBKVPairReader struct {
	prefix byte
	it     *rocksdb.Iterator
}

func NewRocksDBKVPairReader(prefix byte, db *rocksdb.DB) *RocksDBKVPairReader {
	opts := rocksdb.NewDefaultReadOptions()
	opts.SetFillCache(false)
	it := db.NewIterator(opts)
	it.Seek([]byte{prefix})
	return &RocksDBKVPairReader{prefix, it}
}

func (r *RocksDBKVPairReader) Read(buffer []*storage.KVPair) (n int, err error) {
	for n = 0; r.it.ValidForPrefix([]byte{r.prefix}) && n < len(buffer); r.it.Next() {
		keySlice := r.it.Key()
		valueSlice := r.it.Value()
		key := make([]byte, keySlice.Size())
		value := make([]byte, valueSlice.Size())
		copy(key, keySlice.Data())
		copy(value, valueSlice.Data())
		buffer[n] = &storage.KVPair{Key: key[1:], Value: value}
		n++
	}
	return n, err
}

func (r *RocksDBKVPairReader) Close() {
	r.it.Close()
}

func (s RocksDBStore) GetAll(prefix byte) storage.KVPairReader {
	return NewRocksDBKVPairReader(prefix, s.db)
}

func (s RocksDBStore) Close() error {
	s.db.Close()
	return nil
}
