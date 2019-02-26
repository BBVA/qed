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

	"time"

	b "github.com/dgraph-io/badger"
	bo "github.com/dgraph-io/badger/options"
	"github.com/dgraph-io/badger/protos"
	"github.com/dgraph-io/badger/y"

	"github.com/bbva/qed/log"
	"github.com/bbva/qed/storage"
)

type BadgerStore struct {
	db                  *b.DB
	vlogTicker          *time.Ticker // runs every 1m, check size of vlog and run GC conditionally.
	mandatoryVlogTicker *time.Ticker // runs every 10m, we always run vlog GC.
}

// Options contains all the configuration used to open the Badger db
type Options struct {
	// Path is the directory path to the Badger db to use.
	Path string

	// BadgerOptions contains any specific Badger options you might
	// want to specify.
	BadgerOptions *b.Options

	// NoSync causes the database to skip fsync calls after each
	// write to the log. This is unsafe, so it should be used
	// with caution.
	NoSync bool

	// ValueLogGC enables a periodic goroutine that does a garbage
	// collection of the value log while the underlying Badger is online.
	ValueLogGC bool

	// GCInterval is the interval between conditionally running the garbage
	// collection process, based on the size of the vlog. By default, runs every 1m.
	GCInterval time.Duration

	// GCInterval is the interval between mandatory running the garbage
	// collection process. By default, runs every 10m.
	MandatoryGCInterval time.Duration

	// GCThreshold sets threshold in bytes for the vlog size to be included in the
	// garbage collection cycle. By default, 1GB.
	GCThreshold int64
}

func NewBadgerStore(path string) (*BadgerStore, error) {
	return NewBadgerStoreOpts(&Options{Path: path})
}

func NewBadgerStoreOpts(opts *Options) (*BadgerStore, error) {

	var bOpts b.Options
	if bOpts = b.DefaultOptions; opts.BadgerOptions != nil {
		bOpts = *opts.BadgerOptions
	}

	bOpts.TableLoadingMode = bo.MemoryMap
	bOpts.ValueLogLoadingMode = bo.FileIO
	bOpts.Dir = opts.Path
	bOpts.ValueDir = opts.Path
	bOpts.SyncWrites = false
	bOpts.ValueThreshold = 1 << 11 // LSM mode

	db, err := b.Open(bOpts)
	if err != nil {
		return nil, err
	}

	store := &BadgerStore{db: db}
	// Start GC routine
	if opts.ValueLogGC {

		var gcInterval time.Duration
		var mandatoryGCInterval time.Duration
		var threshold int64

		if gcInterval = 1 * time.Minute; opts.GCInterval != 0 {
			gcInterval = opts.GCInterval
		}
		if mandatoryGCInterval = 10 * time.Minute; opts.MandatoryGCInterval != 0 {
			mandatoryGCInterval = opts.MandatoryGCInterval
		}
		if threshold = int64(1 << 30); opts.GCThreshold != 0 {
			threshold = opts.GCThreshold
		}

		store.vlogTicker = time.NewTicker(gcInterval)
		store.mandatoryVlogTicker = time.NewTicker(mandatoryGCInterval)
		go store.runVlogGC(db, threshold)
	}

	return store, nil
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
		it.Seek([]byte{prefix, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
		if it.ValidForPrefix([]byte{prefix}) {
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
	if s.vlogTicker != nil {
		s.vlogTicker.Stop()
	}
	if s.mandatoryVlogTicker != nil {
		s.mandatoryVlogTicker.Stop()
	}
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
func (s *BadgerStore) Backup(w io.Writer, until uint64) error {
	err := s.db.View(func(txn *b.Txn) error {
		opts := b.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			if item.Version() > until {
				// Ignore versions great than given timestamp
				break
			}
			val, err := item.Value()
			if err != nil {
				log.Infof("Key [%x]. Error while fetching value [%v]\n", item.Key(), err)
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

func (s *BadgerStore) GetLastVersion() (uint64, error) {
	var version uint64
	err := s.db.View(func(txn *b.Txn) error {
		opts := b.DefaultIteratorOptions
		opts.PrefetchValues = false
		opts.Reverse = true
		it := txn.NewIterator(opts)
		defer it.Close()
		// we are using a reversed iterator so we need to seek for
		// the last possible key
		it.Rewind()
		if it.Valid() {
			item := it.Item()
			version = item.Version()
		}
		return nil
	})
	return version, err
}

func (b *BadgerStore) runVlogGC(db *b.DB, threshold int64) {
	// Get initial size on start.
	_, lastVlogSize := db.Size()

	runGC := func() {
		var err error
		for err == nil {
			// If a GC is successful, immediately run it again.
			log.Debug("VlogGC task: running...")
			err = db.RunValueLogGC(0.7)
		}
		log.Debug("VlogGC task: done.")
		_, lastVlogSize = db.Size()
	}

	for {
		select {
		case <-b.vlogTicker.C:
			_, currentVlogSize := db.Size()
			if currentVlogSize < lastVlogSize+threshold {
				continue
			}
			runGC()
		case <-b.mandatoryVlogTicker.C:
			runGC()
		}
	}
}
