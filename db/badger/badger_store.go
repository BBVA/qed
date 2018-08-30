package badger

import (
	"bytes"
	"io"
	"log"

	"github.com/bbva/qed/db"
	b "github.com/dgraph-io/badger"
	bo "github.com/dgraph-io/badger/options"
)

type BadgerStore struct {
	db *b.DB
}

func NewBadgerStore(path string) *BadgerStore {
	opts := b.DefaultOptions
	opts.TableLoadingMode = bo.MemoryMap
	opts.ValueLogLoadingMode = bo.FileIO
	opts.Dir = path
	opts.ValueDir = path
	opts.SyncWrites = false
	db, err := b.Open(opts)
	if err != nil {
		log.Fatal(err)
	}
	return &BadgerStore{db}
}

func (s BadgerStore) Mutate(mutations ...db.Mutation) error {
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

func (s BadgerStore) GetRange(prefix byte, start, end []byte) (db.KVRange, error) {
	result := make(db.KVRange, 0)
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
			result = append(result, db.KVPair{key[1:], value})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s BadgerStore) Get(prefix byte, key []byte) (*db.KVPair, error) {
	result := new(db.KVPair)
	result.Key = key
	err := s.db.View(func(txn *b.Txn) error {
		k := append([]byte{prefix}, key...)
		item, err := txn.Get(k)
		if err != nil {
			return err
		}
		var value []byte
		value, err = item.ValueCopy(value)
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
		return nil, db.ErrKeyNotFound
	default:
		return nil, err
	}
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
func (s *BadgerStore) Backup(w io.Writer, since uint64) error {
	_, err := s.db.Backup(w, since)
	return err
}

func (s *BadgerStore) Load(r io.Reader) error {
	return s.db.Load(r)
}
