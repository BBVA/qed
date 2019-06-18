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

// Package raftrocks provides access to RocksDB for Raft to store and retrieve
// log entries. It also provides key/value storage, and can be used as
// a LogStore and StableStore.
package raftrocks

import (
	"bytes"
	"errors"

	"github.com/bbva/qed/metrics"
	"github.com/bbva/qed/rocksdb"
	"github.com/bbva/qed/util"
	"github.com/hashicorp/go-msgpack/codec"
	"github.com/hashicorp/raft"
)

var (
	// ErrKeyNotFound is an error indicating a given key does not exist
	ErrKeyNotFound = errors.New("not found")
)

// table groups related key-value pairs under a
// consistent space.
type table uint32

const (
	defaultTable table = iota
	stableTable
	logTable
)

func (t table) String() string {
	var s string
	switch t {
	case defaultTable:
		s = "default"
	case stableTable:
		s = "stable"
	case logTable:
		s = "log"
	}
	return s
}

// RocksDBStore provides access to RocksDB for Raft to store and retrieve
// log entries. It also provides key/value storage, and can be used as
// a LogStore and StableStore.
type RocksDBStore struct {
	// db is the underlying handle to the db.
	db *rocksdb.DB

	stats *rocksdb.Statistics

	// The path to the RocksDB database directory.
	path string
	ro   *rocksdb.ReadOptions
	wo   *rocksdb.WriteOptions
	// column family handlers
	cfHandles rocksdb.ColumnFamilyHandles

	// global options
	globalOpts *rocksdb.Options
	// stable options
	stableBbto *rocksdb.BlockBasedTableOptions
	stableOpts *rocksdb.Options
	// log options
	logBbto *rocksdb.BlockBasedTableOptions
	logOpts *rocksdb.Options
	// block cache
	blockCache *rocksdb.Cache

	// metrics
	metrics *rocksDBMetrics
}

// Options contains all the configuration used to open the RocksDB instance.
type Options struct {
	// Path is the directory path to the RocksDB instance to use.
	Path string
	// TODO decide if we should use a diferent directory for the Rocks WAL

	// NoSync causes the database to skip fsync calls after each
	// write to the log. This is unsafe, so it should be used
	// with caution.
	NoSync bool

	EnableStatistics bool
}

// NewRocksDBStore takes a file path and returns a connected Raft backend.
func NewRocksDBStore(path string) (*RocksDBStore, error) {
	return New(Options{Path: path, NoSync: true})
}

// New uses the supplied options to open the RocksDB instance and prepare it for
// use as a raft backend.
func New(options Options) (*RocksDBStore, error) {

	// we need two column families, one for stable store and one for log store:
	// stable : used for storing key configurations.
	// log 	  : used for storing logs in a durable fashion.
	cfNames := []string{defaultTable.String(), stableTable.String(), logTable.String()}

	defaultOpts := rocksdb.NewDefaultOptions()

	// global options
	globalOpts := rocksdb.NewDefaultOptions()
	globalOpts.SetCreateIfMissing(true)
	globalOpts.SetCreateIfMissingColumnFamilies(true)
	blockCache := rocksdb.NewDefaultLRUCache(512 * 1024 * 1024)
	var stats *rocksdb.Statistics
	if options.EnableStatistics {
		stats = rocksdb.NewStatistics()
		globalOpts.SetStatistics(stats)
	}

	// stable store options
	stableBbto := rocksdb.NewDefaultBlockBasedTableOptions()
	stableOpts := rocksdb.NewDefaultOptions()
	stableOpts.SetBlockBasedTableFactory(stableBbto)

	// log store options
	logBbto := rocksdb.NewDefaultBlockBasedTableOptions()
	logBbto.SetBlockSize(32 * 1024)
	logBbto.SetCacheIndexAndFilterBlocks(true)
	logBbto.SetBlockCache(blockCache)
	logOpts := rocksdb.NewDefaultOptions()
	logOpts.SetUseFsync(!options.NoSync)
	// dio := directIOSupported(options.Path)
	// if dio {
	// 	logOpts.SetUseDirectIOForFlushAndCompaction(true)
	// }
	logOpts.SetCompression(rocksdb.NoCompression)

	// in normal mode, by default, we try to minimize write amplification,
	// so we set:
	//
	// L0 size = 256MBytes * 2 (min_write_buffer_number_to_merge) * \
	//              8 (level0_file_num_compaction_trigger)
	//         = 4GBytes
	// L1 size close to L0, 4GBytes, max_bytes_for_level_base = 4GBytes,
	//   max_bytes_for_level_multiplier = 2
	// L2 size is 8G, L3 is 16G, L4 is 32G, L5 64G...
	//
	// note this is the size of a shard, and the content of the store is expected
	// to be compacted by raft.
	//
	logOpts.SetMaxSubCompactions(2) // TODO what's this?
	logOpts.SetEnablePipelinedWrite(true)
	logOpts.SetWriteBufferSize(256 * 1024 * 1024)
	logOpts.SetMinWriteBufferNumberToMerge(2)
	logOpts.SetLevel0FileNumCompactionTrigger(8)
	logOpts.SetLevel0SlowdownWritesTrigger(17)
	logOpts.SetLevel0StopWritesTrigger(24)
	logOpts.SetMaxWriteBufferNumber(5)
	logOpts.SetNumLevels(7)
	// MaxBytesForLevelBase is the total size of L1, should be close to
	// the size of L0
	logOpts.SetMaxBytesForLevelBase(4 * 1024 * 1024 * 1024)
	logOpts.SetMaxBytesForLevelMultiplier(2)
	// files in L1 will have TargetFileSizeBase bytes
	logOpts.SetTargetFileSizeBase(256 * 1024 * 1024)
	logOpts.SetTargetFileSizeMultiplier(1)
	// IO parallelism
	logOpts.SetMaxBackgroundCompactions(2)
	logOpts.SetMaxBackgroundFlushes(2)

	cfOpts := []*rocksdb.Options{defaultOpts, stableOpts, logOpts}
	db, cfHandles, err := rocksdb.OpenDBColumnFamilies(options.Path, globalOpts, cfNames, cfOpts)
	if err != nil {
		return nil, err
	}

	// read/write options
	wo := rocksdb.NewDefaultWriteOptions()
	wo.SetSync(!options.NoSync)
	ro := rocksdb.NewDefaultReadOptions()
	ro.SetFillCache(false)

	store := &RocksDBStore{
		db:         db,
		stats:      stats,
		path:       options.Path,
		cfHandles:  cfHandles,
		stableBbto: stableBbto,
		stableOpts: stableOpts,
		logBbto:    logBbto,
		logOpts:    logOpts,
		blockCache: blockCache,
		globalOpts: globalOpts,
		ro:         ro,
		wo:         wo,
	}

	if stats != nil {
		store.metrics = newRocksDBMetrics(store)
	}

	return store, nil
}

// Close is used to gracefully close the DB connection.
func (s *RocksDBStore) Close() error {
	for _, cf := range s.cfHandles {
		cf.Destroy()
	}
	if s.db != nil {
		s.db.Close()
	}
	if s.stableBbto != nil {
		s.stableBbto.Destroy()
	}
	if s.stableOpts != nil {
		s.stableOpts.Destroy()
	}
	if s.blockCache != nil {
		s.blockCache.Destroy()
	}
	if s.logBbto != nil {
		s.logBbto.Destroy()
	}
	if s.logOpts != nil {
		s.logOpts.Destroy()
	}
	if s.stats != nil {
		s.stats.Destroy()
	}
	if s.wo != nil {
		s.wo.Destroy()
	}
	if s.ro != nil {
		s.ro.Destroy()
	}
	s.db = nil
	return nil
}

// FirstIndex returns the first known index from the Raft log.
func (s *RocksDBStore) FirstIndex() (uint64, error) {
	it := s.db.NewIteratorCF(rocksdb.NewDefaultReadOptions(), s.cfHandles[logTable])
	defer it.Close()
	it.SeekToFirst()
	if it.Valid() {
		slice := it.Key()
		defer slice.Free()
		key := make([]byte, slice.Size())
		copy(key, slice.Data())
		return util.BytesAsUint64(key), nil
	}
	return 0, nil
}

// LastIndex returns the last known index from the Raft log.
func (s *RocksDBStore) LastIndex() (uint64, error) {
	it := s.db.NewIteratorCF(rocksdb.NewDefaultReadOptions(), s.cfHandles[logTable])
	defer it.Close()
	it.SeekToLast()
	if it.Valid() {
		slice := it.Key()
		defer slice.Free()
		key := make([]byte, slice.Size())
		copy(key, slice.Data())
		return util.BytesAsUint64(key), nil
	}
	return 0, nil
}

// GetLog gets a log entry at a given index.
func (s *RocksDBStore) GetLog(index uint64, log *raft.Log) error {
	val, err := s.db.GetBytesCF(s.ro, s.cfHandles[logTable], util.Uint64AsBytes(index))
	if err != nil {
		return err
	}
	if val == nil {
		return raft.ErrLogNotFound
	}
	return decodeMsgPack(val, log)
}

// StoreLog stores a single raft log.
func (s *RocksDBStore) StoreLog(log *raft.Log) error {
	val, err := encodeMsgPack(log)
	if err != nil {
		return err
	}
	return s.db.PutCF(s.wo, s.cfHandles[logTable], util.Uint64AsBytes(log.Index), val.Bytes())
}

// StoreLogs stores a set of raft logs.
func (s *RocksDBStore) StoreLogs(logs []*raft.Log) error {
	batch := rocksdb.NewWriteBatch()
	for _, log := range logs {
		key := util.Uint64AsBytes(log.Index)
		val, err := encodeMsgPack(log)
		if err != nil {
			return err
		}
		batch.PutCF(s.cfHandles[logTable], key, val.Bytes())
	}
	return s.db.Write(s.wo, batch)
}

// DeleteRange deletes logs within a given range inclusively.
func (s *RocksDBStore) DeleteRange(min, max uint64) error {
	batch := rocksdb.NewWriteBatch()
	batch.DeleteRangeCF(s.cfHandles[logTable], util.Uint64AsBytes(min), util.Uint64AsBytes(max+1))
	return s.db.Write(s.wo, batch)
}

// Set is used to set a key/value set outside of the raft log.
func (s *RocksDBStore) Set(key []byte, val []byte) error {
	if err := s.db.PutCF(s.wo, s.cfHandles[stableTable], key, val); err != nil {
		return err
	}
	return nil
}

// Get is used to retrieve a value from the k/v store by key
func (s *RocksDBStore) Get(key []byte) ([]byte, error) {
	val, err := s.db.GetBytesCF(s.ro, s.cfHandles[stableTable], key)
	if err != nil {
		return nil, err
	}
	if val == nil {
		return nil, ErrKeyNotFound
	}
	return val, nil
}

// SetUint64 is like Set, but handles uint64 values
func (s *RocksDBStore) SetUint64(key []byte, val uint64) error {
	return s.Set(key, util.Uint64AsBytes(val))
}

// GetUint64 is like Get, but handles uint64 values
func (s *RocksDBStore) GetUint64(key []byte) (uint64, error) {
	val, err := s.Get(key)
	if err != nil {
		return 0, err
	}
	return util.BytesAsUint64(val), nil
}

func (s *RocksDBStore) RegisterMetrics(registry metrics.Registry) {
	if registry != nil {
		registry.MustRegister(s.metrics.collectors()...)
	}
}

// Decode reverses the encode operation on a byte slice input
func decodeMsgPack(buf []byte, out interface{}) error {
	r := bytes.NewBuffer(buf)
	hd := codec.MsgpackHandle{}
	dec := codec.NewDecoder(r, &hd)
	return dec.Decode(out)
}

// Encode writes an encoded object to a new bytes buffer
func encodeMsgPack(in interface{}) (*bytes.Buffer, error) {
	buf := bytes.NewBuffer(nil)
	hd := codec.MsgpackHandle{}
	enc := codec.NewEncoder(buf, &hd)
	err := enc.Encode(in)
	return buf, err
}
