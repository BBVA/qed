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

	// column family handlers
	cfHandles rocksdb.ColumnFamilyHandles

	// global options
	globalOpts *rocksdb.Options
	// per column family options
	cfOpts []*rocksdb.Options

	// read/write options
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

	cfNames := []string{
		storage.DefaultTable.String(),
		storage.IndexTable.String(),
		storage.HyperCacheTable.String(),
		storage.HistoryCacheTable.String(),
		storage.FSMStateTable.String(),
	}

	// global options
	globalOpts := rocksdb.NewDefaultOptions()
	globalOpts.SetCreateIfMissing(true)
	globalOpts.SetCreateIfMissingColumnFamilies(true)
	globalOpts.IncreaseParallelism(4)
	var stats *rocksdb.Statistics
	if opts.EnableStatistics {
		stats = rocksdb.NewStatistics()
		globalOpts.SetStatistics(stats)
	}

	// Per column family options
	cfOpts := []*rocksdb.Options{
		rocksdb.NewDefaultOptions(),
		getIndexTableOpts(),
		getHyperCacheTableOpts(),
		getHistoryCacheTableOpts(),
		getFsmStateTableOpts(),
	}

	db, cfHandles, err := rocksdb.OpenDBColumnFamilies(opts.Path, globalOpts, cfNames, cfOpts)
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
		cfHandles:      cfHandles,
		checkPointPath: checkPointPath,
		checkpoints:    make(map[uint64]string),
		globalOpts:     globalOpts,
		cfOpts:         cfOpts,
		wo:             rocksdb.NewDefaultWriteOptions(),
		ro:             rocksdb.NewDefaultReadOptions(),
	}

	if rms == nil && stats != nil {
		rms = newRocksDBMetrics(stats)
	}

	return store, nil
}

func getIndexTableOpts() *rocksdb.Options {
	// index table is append-only so we have to optimize for
	// read amplification

	// TODO change this!!!

	bbto := rocksdb.NewDefaultBlockBasedTableOptions()
	bbto.SetFilterPolicy(rocksdb.NewBloomFilterPolicy(10))
	opts := rocksdb.NewDefaultOptions()
	opts.SetBlockBasedTableFactory(bbto)
	opts.SetCompression(rocksdb.SnappyCompression)
	// in normal mode, by default, we try to minimize space amplification,
	// so we set:
	//
	// L0 size = 64MBytes * 2 (min_write_buffer_number_to_merge) * \
	//              8 (level0_file_num_compaction_trigger)
	//         = 1GBytes
	// L1 size close to L0, 1GBytes, max_bytes_for_level_base = 1GBytes,
	//   max_bytes_for_level_multiplier = 2
	// L2 size is 2G, L3 is 4G, L4 is 8G, L5 16G...
	//
	opts.SetWriteBufferSize(64 * 1024 * 1024)
	opts.SetMaxWriteBufferNumber(5)
	opts.SetMinWriteBufferNumberToMerge(2)
	opts.SetLevel0FileNumCompactionTrigger(8)
	// MaxBytesForLevelBase is the total size of L1, should be close to
	// the size of L0
	opts.SetMaxBytesForLevelBase(1 * 1024 * 1024 * 1024)
	opts.SetMaxBytesForLevelMultiplier(2)
	// files in L1 will have TargetFileSizeBase bytes
	opts.SetTargetFileSizeBase(64 * 1024 * 1024)
	opts.SetTargetFileSizeMultiplier(10)
	// io parallelism
	opts.SetMaxBackgroundCompactions(2)
	opts.SetMaxBackgroundFlushes(2)
	return opts
}

func getHyperCacheTableOpts() *rocksdb.Options {
	bbto := rocksdb.NewDefaultBlockBasedTableOptions()
	bbto.SetFilterPolicy(rocksdb.NewBloomFilterPolicy(10))
	opts := rocksdb.NewDefaultOptions()
	opts.SetBlockBasedTableFactory(bbto)
	opts.SetCompression(rocksdb.SnappyCompression)
	return opts
}

func getHistoryCacheTableOpts() *rocksdb.Options {
	bbto := rocksdb.NewDefaultBlockBasedTableOptions()
	bbto.SetFilterPolicy(rocksdb.NewBloomFilterPolicy(10))
	opts := rocksdb.NewDefaultOptions()
	opts.SetBlockBasedTableFactory(bbto)
	opts.SetCompression(rocksdb.SnappyCompression)
	return opts
}

func getFsmStateTableOpts() *rocksdb.Options {
	// FSM state contains only one key that is updated on every
	// add event operation. We should try to reduce write and
	// space amplification by keeping a lower number of levels.
	bbto := rocksdb.NewDefaultBlockBasedTableOptions()
	opts := rocksdb.NewDefaultOptions()
	opts.SetBlockBasedTableFactory(bbto)
	opts.SetCompression(rocksdb.SnappyCompression)
	// we try to reduce write and space amplification, so we:
	//   * set a low size for the in-memory write buffers
	//   * reduce the number of write buffers
	//   * activate merging before flushing
	//   * set parallelism to 1
	opts.SetWriteBufferSize(4 * 1024 * 1024)
	opts.SetMaxWriteBufferNumber(3)
	opts.SetMinWriteBufferNumberToMerge(2)
	opts.SetMaxBackgroundCompactions(1)
	opts.SetMaxBackgroundFlushes(1)
	return opts
}

func (s *RocksDBStore) Mutate(mutations []*storage.Mutation) error {
	batch := rocksdb.NewWriteBatch()
	defer batch.Destroy()
	for _, m := range mutations {
		batch.PutCF(s.cfHandles[m.Table], m.Key, m.Value)
	}
	err := s.db.Write(s.wo, batch)
	return err
}

func (s *RocksDBStore) Get(table storage.Table, key []byte) (*storage.KVPair, error) {
	result := new(storage.KVPair)
	result.Key = key
	v, err := s.db.GetBytesCF(s.ro, s.cfHandles[table], key)
	if err != nil {
		return nil, err
	}
	if v == nil {
		return nil, storage.ErrKeyNotFound
	}
	result.Value = v
	return result, nil
}

func (s *RocksDBStore) GetRange(table storage.Table, start, end []byte) (storage.KVRange, error) {
	result := make(storage.KVRange, 0)
	it := s.db.NewIteratorCF(s.ro, s.cfHandles[table])
	defer it.Close()
	for it.Seek(start); it.Valid(); it.Next() {
		keySlice := it.Key()
		key := make([]byte, keySlice.Size())
		copy(key, keySlice.Data())
		keySlice.Free()
		if bytes.Compare(key, end) > 0 {
			break
		}
		valueSlice := it.Value()
		value := make([]byte, valueSlice.Size())
		copy(value, valueSlice.Data())
		result = append(result, storage.KVPair{Key: key, Value: value})
		valueSlice.Free()
	}

	return result, nil
}

func (s *RocksDBStore) GetLast(table storage.Table) (*storage.KVPair, error) {
	it := s.db.NewIteratorCF(s.ro, s.cfHandles[table])
	defer it.Close()
	it.SeekForPrev([]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
	if it.Valid() {
		result := new(storage.KVPair)
		keySlice := it.Key()
		key := make([]byte, keySlice.Size())
		copy(key, keySlice.Data())
		keySlice.Free()
		result.Key = key
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
	it *rocksdb.Iterator
}

func NewRocksDBKVPairReader(cfHandle *rocksdb.ColumnFamilyHandle, db *rocksdb.DB) *RocksDBKVPairReader {
	opts := rocksdb.NewDefaultReadOptions()
	opts.SetFillCache(false)
	it := db.NewIteratorCF(opts, cfHandle)
	it.SeekToFirst()
	return &RocksDBKVPairReader{it}
}

func (r *RocksDBKVPairReader) Read(buffer []*storage.KVPair) (n int, err error) {
	for n = 0; r.it.Valid() && n < len(buffer); r.it.Next() {
		keySlice := r.it.Key()
		valueSlice := r.it.Value()
		key := make([]byte, keySlice.Size())
		value := make([]byte, valueSlice.Size())
		copy(key, keySlice.Data())
		copy(value, valueSlice.Data())
		keySlice.Free()
		valueSlice.Free()
		buffer[n] = &storage.KVPair{Key: key, Value: value}
		n++
	}
	return n, err
}

func (r *RocksDBKVPairReader) Close() {
	r.it.Close()
}

func (s *RocksDBStore) GetAll(table storage.Table) storage.KVPairReader {
	return NewRocksDBKVPairReader(s.cfHandles[table], s.db)
}

func (s *RocksDBStore) Close() error {

	for _, cf := range s.cfHandles {
		cf.Destroy()
	}

	if s.db != nil {
		s.db.Close()
	}

	if s.stats != nil {
		s.stats.Destroy()
	}
	if s.globalOpts != nil {
		s.globalOpts.Destroy()
	}

	for _, opt := range s.cfOpts {
		opt.Destroy()
	}

	if s.ro != nil {
		s.ro.Destroy()
	}
	if s.wo != nil {
		s.wo.Destroy()
	}

	return nil
}

// Snapshot takes a snapshot of the store, and returns and id
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

	tables := []storage.Table{
		storage.DefaultTable,
		storage.IndexTable,
		storage.HyperCacheTable,
		storage.HistoryCacheTable,
		storage.FSMStateTable,
	}
	for _, table := range tables {

		it := checkDB.NewIteratorCF(ro, s.cfHandles[table])
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
				Table: pb.Table(table),
				Key:   key,
				Value: value,
			}

			// write entries to disk
			if err := writeTo(entry, w); err != nil {
				return err
			}
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
		table := storage.Table(kv.GetTable())
		batch.PutCF(s.cfHandles[table], kv.Key, kv.Value)

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
