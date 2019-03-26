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
package rocks

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"github.com/bbva/qed/rocksdb"
	"github.com/bbva/qed/storage"
	"github.com/bbva/qed/storage/pb"
)

type RocksDBStore struct {
	db *rocksdb.DB

	stats *rocksdb.Statistics

	// checkpoints are stored in a path on the same
	// folder as the database, so rocksdb uses hardlinks instead
	// of copies
	checkPointPath string

	// each checkpoint is created in a subdirectory
	// inside checkPointPath folder
	checkpoints map[uint64]string

	ro *rocksdb.ReadOptions
	wo *rocksdb.WriteOptions
}

type Options struct {
	Path             string
	EnableStatistics bool
}

func NewRocksDBStore(path string) (*RocksDBStore, error) {
	return NewRocksDBStoreOpts(&Options{Path: path, EnableStatistics: true})
}

func NewRocksDBStoreOpts(opts *Options) (*RocksDBStore, error) {
	rocksdbOpts := rocksdb.NewDefaultOptions()
	rocksdbOpts.SetCreateIfMissing(true)
	rocksdbOpts.IncreaseParallelism(4)
	rocksdbOpts.SetMaxWriteBufferNumber(5)
	rocksdbOpts.SetMinWriteBufferNumberToMerge(2)

	var stats *rocksdb.Statistics
	if opts.EnableStatistics {
		stats = rocksdb.NewStatistics()
		rocksdbOpts.SetStatistics(stats)
	}

	blockOpts := rocksdb.NewDefaultBlockBasedTableOptions()
	blockOpts.SetFilterPolicy(rocksdb.NewBloomFilterPolicy(10))
	rocksdbOpts.SetBlockBasedTableFactory(blockOpts)

	db, err := rocksdb.OpenDB(opts.Path, rocksdbOpts)
	if err != nil {
		return nil, err
	}
	checkPointPath := opts.Path + "/checkpoints"
	err = os.MkdirAll(checkPointPath, 0755)
	if err != nil {
		return nil, err
	}

	store := &RocksDBStore{
		db:             db,
		stats:          stats,
		checkPointPath: checkPointPath,
		checkpoints:    make(map[uint64]string),
		wo:             rocksdb.NewDefaultWriteOptions(),
		ro:             rocksdb.NewDefaultReadOptions(),
	}

	if rms == nil && stats != nil {
		rms = newRocksDBMetrics(stats)
	}

	return store, nil
}

func (s RocksDBStore) Mutate(mutations []*storage.Mutation) error {
	batch := rocksdb.NewWriteBatch()
	defer batch.Destroy()
	for _, m := range mutations {
		key := append([]byte{m.Prefix}, m.Key...)
		batch.Put(key, m.Value)
	}
	err := s.db.Write(s.wo, batch)
	return err
}

func (s RocksDBStore) Get(prefix byte, key []byte) (*storage.KVPair, error) {
	result := new(storage.KVPair)
	result.Key = key
	k := append([]byte{prefix}, key...)
	v, err := s.db.GetBytes(s.ro, k)
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
	it := s.db.NewIterator(s.ro)
	defer it.Close()
	for it.Seek(startKey); it.Valid(); it.Next() {
		keySlice := it.Key()
		key := make([]byte, keySlice.Size())
		copy(key, keySlice.Data())
		keySlice.Free()
		if bytes.Compare(key, endKey) > 0 {
			break
		}
		valueSlice := it.Value()
		value := make([]byte, valueSlice.Size())
		copy(value, valueSlice.Data())
		result = append(result, storage.KVPair{Key: key[1:], Value: value})
		valueSlice.Free()
	}

	return result, nil
}

func (s RocksDBStore) GetLast(prefix byte) (*storage.KVPair, error) {
	it := s.db.NewIterator(s.ro)
	defer it.Close()
	it.SeekForPrev([]byte{prefix, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
	if it.ValidForPrefix([]byte{prefix}) {
		result := new(storage.KVPair)
		keySlice := it.Key()
		key := make([]byte, keySlice.Size())
		copy(key, keySlice.Data())
		keySlice.Free()
		result.Key = key[1:]
		valueSlice := it.Value()
		value := make([]byte, valueSlice.Size())
		copy(value, valueSlice.Data())
		valueSlice.Free()
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
		keySlice.Free()
		valueSlice.Free()
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
	s.ro.Destroy()
	s.wo.Destroy()
	return nil
}

func (s RocksDBStore) Delete(prefix byte, key []byte) error {
	k := append([]byte{prefix}, key...)
	return s.db.Delete(rocksdb.NewDefaultWriteOptions(), k)
}

// Take a snapshot of the store, and returns and id
// to be used in the back up process. The state of the
// snapshot is stored in the store instance.
func (s *RocksDBStore) Snapshot() (uint64, error) {
	// create temp directory
	id := uint64(len(s.checkpoints) + 1)
	checkDir := fmt.Sprintf("%s/rocksdb-checkpoint-%d", s.checkPointPath, id)
	os.RemoveAll(checkDir)

	// create checkpoint
	checkpoint, err := s.db.NewCheckpoint()
	if err != nil {
		return 0, err
	}
	defer checkpoint.Destroy()
	checkpoint.CreateCheckpoint(checkDir, 0)

	s.checkpoints[id] = checkDir
	return id, nil
}

// Backup dumps a protobuf-encoded list of all entries in the database into the
// given writer, that are newer than the specified version.
func (s *RocksDBStore) Backup(w io.Writer, id uint64) error {

	checkDir := s.checkpoints[id]

	// open db for read-only
	opts := rocksdb.NewDefaultOptions()
	checkDB, err := rocksdb.OpenDBForReadOnly(checkDir, opts, true)
	if err != nil {
		return err
	}

	// open a new iterator and dump every key
	ro := rocksdb.NewDefaultReadOptions()
	ro.SetFillCache(false)
	it := checkDB.NewIterator(ro)
	defer it.Close()

	for it.SeekToFirst(); it.Valid(); it.Next() {
		keySlice := it.Key()
		valueSlice := it.Value()
		keyData := keySlice.Data()
		valueData := valueSlice.Data()
		key := append(keyData[:0:0], keyData...) // See https://github.com/go101/go101/wiki
		value := append(valueData[:0:0], valueData...)
		keySlice.Free()
		valueSlice.Free()

		entry := &pb.KVPair{
			Key:   key,
			Value: value,
		}

		// write entries to disk
		if err := writeTo(entry, w); err != nil {
			return err
		}
	}

	// remove checkpoint from list
	// order must be maintained,
	delete(s.checkpoints, id)

	// clean up only after we succesfully backup
	os.RemoveAll(checkDir)

	return nil
}

// Load reads a protobuf-encoded list of all entries from a reader and writes
// them to the database. This can be used to restore the database from a backup
// made by calling DB.Backup().
//
// DB.Load() should be called on a database that is not running any other
// concurrent transactions while it is running.
func (s *RocksDBStore) Load(r io.Reader) error {

	br := bufio.NewReaderSize(r, 16<<10)
	unmarshalBuf := make([]byte, 1<<10)
	batch := rocksdb.NewWriteBatch()
	wo := rocksdb.NewDefaultWriteOptions()
	wo.SetDisableWAL(true)

	for {
		var data uint64
		err := binary.Read(br, binary.LittleEndian, &data)
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		if cap(unmarshalBuf) < int(data) {
			unmarshalBuf = make([]byte, data)
		}

		kv := &pb.KVPair{}
		if _, err = io.ReadFull(br, unmarshalBuf[:data]); err != nil {
			return err
		}
		if err = kv.Unmarshal(unmarshalBuf[:data]); err != nil {
			return err
		}
		batch.Put(kv.Key, kv.Value)

		if batch.Count() == 1000 {
			s.db.Write(wo, batch)
			batch.Clear()
			continue
		}
	}

	if batch.Count() > 0 {
		return s.db.Write(wo, batch)
	}

	return nil
}

func writeTo(entry *pb.KVPair, w io.Writer) error {
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
