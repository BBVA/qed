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

package badger

import (
	"bytes"
	"encoding/binary"
	"io"
	"log"

	b "github.com/dgraph-io/badger"
	bo "github.com/dgraph-io/badger/options"
	"github.com/dgraph-io/badger/protos"
	"github.com/dgraph-io/badger/y"

	"github.com/bbva/qed/storage"
)

type BadgerStore struct {
	db *b.DB
}

func NewBadgerStore(path string) (*BadgerStore, error) {
	opts := b.DefaultOptions
	opts.TableLoadingMode = bo.MemoryMap
	opts.ValueLogLoadingMode = bo.FileIO
	opts.Dir = path
	opts.ValueDir = path
	opts.SyncWrites = false

	return NewBadgerStoreOpts(path, opts)
}

func NewBadgerStoreOpts(path string, opts b.Options) (*BadgerStore, error) {
	db, err := b.Open(opts)
	if err != nil {
		return nil, err
	}
	return &BadgerStore{db}, nil
}

func (s BadgerStore) Mutate(mutations []*storage.Mutation) error {
	return s.db.Update(func(txn *b.Txn) error {
		for _, m := range mutations {
			key := append([]byte{m.Prefix}, m.Key...)
			err := txn.Set(key, m.Value)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (s BadgerStore) GetRange(prefix byte, start, end []byte) (storage.KVRange, error) {
	result := make(storage.KVRange, 0)
	startKey := append([]byte{prefix}, start...)
	endKey := append([]byte{prefix}, end...)
	err := s.db.View(func(txn *b.Txn) error {
		opts := b.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Seek(startKey); it.Valid(); it.Next() {
			item := it.Item()
			var key []byte
			key = item.KeyCopy(key)
			if bytes.Compare(key, endKey) > 0 {
				break
			}
			var value []byte
			value, err := item.ValueCopy(value)
			if err != nil {
				return err
			}
			result = append(result, storage.KVPair{key[1:], value})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s BadgerStore) Get(prefix byte, key []byte) (*storage.KVPair, error) {
	result := new(storage.KVPair)
	result.Key = key
	err := s.db.View(func(txn *b.Txn) error {
		k := append([]byte{prefix}, key...)
		item, err := txn.Get(k)
		if err != nil {
			return err
		}
		value, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		result.Value = value
		return nil
	})
	switch err {
	case nil:
		return result, nil
	case b.ErrKeyNotFound:
		return nil, storage.ErrKeyNotFound
	default:
		return nil, err
	}
}

func (s BadgerStore) GetLast(prefix byte) (*storage.KVPair, error) {
	result := new(storage.KVPair)
	err := s.db.View(func(txn *b.Txn) error {
		var err error
		opts := b.DefaultIteratorOptions
		opts.PrefetchValues = false
		opts.Reverse = true
		it := txn.NewIterator(opts)
		defer it.Close()
		// we are using a reversed iterator so we need to seek for
		// the last possible key for history prefix
		it.Seek([]byte{prefix, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
		if it.Valid() {
			item := it.Item()
			key := item.KeyCopy(nil)
			result.Key = key[1:]
			result.Value, err = item.ValueCopy(nil)
		} else {
			err = b.ErrKeyNotFound
		}
		return err
	})
	switch err {
	case nil:
		return result, nil
	case b.ErrKeyNotFound:
		return nil, storage.ErrKeyNotFound
	default:
		return nil, err
	}
}

type BadgerKVPairReader struct {
	prefix byte
	txn    *b.Txn
	it     *b.Iterator
}

func NewBadgerKVPairReader(prefix byte, txn *b.Txn) *BadgerKVPairReader {
	opts := b.DefaultIteratorOptions
	opts.PrefetchSize = 10
	it := txn.NewIterator(opts)
	it.Seek([]byte{prefix})
	return &BadgerKVPairReader{prefix, txn, it}
}

func (r *BadgerKVPairReader) Read(buffer []*storage.KVPair) (n int, err error) {
	for n = 0; r.it.ValidForPrefix([]byte{r.prefix}) && n < len(buffer); r.it.Next() {
		item := r.it.Item()
		var key, value []byte
		key = item.KeyCopy(key)
		value, err := item.ValueCopy(value)
		if err != nil {
			break
		}
		buffer[n] = &storage.KVPair{key[1:], value}
		n++
	}

	// TODO should i close the iterator and transaction?
	return n, err
}

func (r *BadgerKVPairReader) Close() {
	r.it.Close()
	r.txn.Discard()
}

func (s BadgerStore) GetAll(prefix byte) storage.KVPairReader {
	return NewBadgerKVPairReader(prefix, s.db.NewTransaction(false))
}

func (s BadgerStore) Close() error {
	return s.db.Close()
}

func (s BadgerStore) Delete(prefix byte, key []byte) error {
	return s.db.Update(func(txn *b.Txn) error {
		k := append([]byte{prefix}, key...)
		return txn.Delete(k)
	})
}

// Borrowed from github.com/dgraph-io/badger/backup.go
func writeTo(entry *protos.KVPair, w io.Writer) error {
	if err := binary.Write(w, binary.LittleEndian, uint64(entry.Size())); err != nil {
		return err
	}
	buf, err := entry.Marshal()
	if err != nil {
		return err
	}
	_, err = w.Write(buf)
	return err
}

// Backup dumps a protobuf-encoded list of all entries in the database into the
// given writer, that are newer than the specified version.
//
// Borrowed from github.com/dgraph-io/badger/backup.go
func (s *BadgerStore) Backup(w io.Writer, since uint64) error {
	err := s.db.View(func(txn *b.Txn) error {
		opts := b.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			if item.Version() < since {
				// Ignore versions less than given timestamp
				continue
			}
			val, err := item.Value()
			if err != nil {
				log.Printf("Key [%x]. Error while fetching value [%v]\n", item.Key(), err)
				continue
			}

			entry := &protos.KVPair{
				Key:       y.Copy(item.Key()),
				Value:     y.Copy(val),
				UserMeta:  []byte{item.UserMeta()},
				Version:   item.Version(),
				ExpiresAt: item.ExpiresAt(),
			}

			// Write entries to disk
			if err := writeTo(entry, w); err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

func (s *BadgerStore) Load(r io.Reader) error {
	return s.db.Load(r)
}
