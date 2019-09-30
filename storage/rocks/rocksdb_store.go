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
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/bbva/qed/metrics"
	"github.com/bbva/qed/rocksdb"
	"github.com/bbva/qed/storage"
	"github.com/bbva/qed/storage/pb"
	"github.com/bbva/qed/util"
)

type RocksDBStore struct {
	path string
	db   *rocksdb.DB

	stats *rocksdb.Statistics

	// backup handler and options
	backupEngine *rocksdb.BackupEngine
	backupOpts   *rocksdb.Options
	restoreOpts  *rocksdb.RestoreOptions

	// column family handlers
	cfHandles rocksdb.ColumnFamilyHandles

	// global options
	globalOpts *rocksdb.Options
	// per column family options
	cfOpts []*rocksdb.Options

	// block cache
	blockCache *rocksdb.Cache

	// read/write options
	ro *rocksdb.ReadOptions
	wo *rocksdb.WriteOptions

	// metrics
	metrics *rocksDBMetrics
}

type Options struct {
	Path             string
	EnableStatistics bool
	MaxTotalWalSize  uint64
	WALSizeLimitMB   uint64
	WALTtlSeconds    uint64
}

func DefaultOptions() *Options {
	return &Options{
		EnableStatistics: true,
		MaxTotalWalSize:  0,
		WALSizeLimitMB:   1 << 20,
		WALTtlSeconds:    0,
	}
}

func NewRocksDBStore(path string, ttl time.Duration) (*RocksDBStore, error) {
	opts := DefaultOptions()
	opts.Path = path
	opts.WALTtlSeconds = uint64(ttl.Seconds())
	return NewRocksDBStoreWithOpts(opts)
}

func NewRocksDBStoreWithOpts(opts *Options) (*RocksDBStore, error) {

	cfNames := []string{
		storage.DefaultTable.String(),
		storage.HyperTable.String(),
		storage.HyperCacheTable.String(),
		storage.HistoryTable.String(),
		storage.FSMStateTable.String(),
	}

	// env
	env := rocksdb.NewDefaultEnv()
	env.SetBackgroundThreads(5)
	env.SetHighPriorityBackgroundThreads(3)

	// global options
	globalOpts := rocksdb.NewDefaultOptions()
	globalOpts.SetCreateIfMissing(true)
	globalOpts.SetCreateIfMissingColumnFamilies(true)
	globalOpts.SetMaxTotalWalSize(opts.MaxTotalWalSize)
	globalOpts.SetWalSizeLimitMb(opts.WALSizeLimitMB)
	globalOpts.SetWALTtlSeconds(opts.WALTtlSeconds)
	//globalOpts.SetMaxOpenFiles(1000)
	globalOpts.SetEnv(env)
	// We build a LRU cache with a high pool ratio of 0.4 (40%). The lower pool
	// will cache data blocks and the higher will cache index and filters.
	blockCache := rocksdb.NewLRUCache(8*1024*1024*1024, 0.4) // 8GB
	var stats *rocksdb.Statistics
	if opts.EnableStatistics {
		stats = rocksdb.NewStatistics()
		globalOpts.SetStatistics(stats)
	}

	// Per column family options
	cfOpts := []*rocksdb.Options{
		rocksdb.NewDefaultOptions(),
		getHyperTableOpts(blockCache),
		getHyperTableOpts(blockCache), // hyperCacheOpts table options
		getHistoryTableOpts(blockCache),
		getFsmStateTableOpts(),
	}

	db, cfHandles, err := rocksdb.OpenDBColumnFamilies(opts.Path, globalOpts, cfNames, cfOpts)
	if err != nil {
		return nil, err
	}

	// Backup and restore stuff.
	backupPath := opts.Path + "/backups"
	err = os.MkdirAll(backupPath, 0755)
	if err != nil {
		return nil, err
	}

	backupOpts := rocksdb.NewDefaultOptions()
	be, err := rocksdb.OpenBackupEngine(backupOpts, backupPath)
	if err != nil {
		return nil, err
	}

	restoreOpts := rocksdb.NewRestoreOptions()

	store := &RocksDBStore{
		path:         opts.Path,
		db:           db,
		stats:        stats,
		cfHandles:    cfHandles,
		blockCache:   blockCache,
		backupEngine: be,
		backupOpts:   backupOpts,
		restoreOpts:  restoreOpts,
		globalOpts:   globalOpts,
		cfOpts:       cfOpts,
		wo:           rocksdb.NewDefaultWriteOptions(),
		ro:           rocksdb.NewDefaultReadOptions(),
	}

	if stats != nil {
		store.metrics = newRocksDBMetrics(store)
	}

	return store, nil
}

// The hyper table has the more varied behavior. It receives
// a mixed workload of point lookups and write/updates.
// The values are higher than the ones inserted in other tables (~1KB).
func getHyperTableOpts(blockCache *rocksdb.Cache) *rocksdb.Options {

	// Keys in this table are positions in the hyper tree and
	// values are batches of at most 31 hashes of 32B.
	// We try to keep hot keys at the lowest levels of the LSM
	// tree. Hot keys corresponds with those batches at the
	// highest levels of the hyper tree, which are more frequently
	// touched on every operation.

	bbto := rocksdb.NewDefaultBlockBasedTableOptions()
	bbto.SetFilterPolicy(rocksdb.NewFullBloomFilterPolicy(10))
	// In order to have a fine-grained control over the memory usage
	// we cache SST's index and filters in the block cache. The alternative
	// would be to leave RocksDB keep those files memory mapped, but
	// the only way to control memory usage would be through the property
	// max_open_files.
	bbto.SetCacheIndexAndFilterBlocks(true)
	// To avoid filter and index eviction from block cache we pin
	// those from L0 and move them to the high priority pool.
	bbto.SetPinL0FilterAndIndexBlocksInCache(true)
	bbto.SetCacheIndexAndFilterBlocksWithHighPriority(true)
	// activate partition filters
	bbto.SetPartitionFilters(true)
	bbto.SetPinTopLevelIndexAndFilterInCache(true)
	bbto.SetIndexType(rocksdb.KTwoLevelIndexSearchIndexType)
	bbto.SetBlockCache(blockCache)
	// increase block size to 16KB
	bbto.SetBlockSize(16 * 1024)

	opts := rocksdb.NewDefaultOptions()
	opts.SetBlockBasedTableFactory(bbto)
	opts.SetCompression(rocksdb.SnappyCompression)

	// We use level style compaction with high concurrency.
	// Memtable size is 128MB and the total number of level 0
	// files is 8. This means compaction is triggered when L0
	// grows to 1GB. L1 size is 1GB and every level is 8 times
	// larger than the previous one. L2 is 8GB, L3 is 64GB,
	// L4 is 512GB, L5 is 8TB (note that given a ~1KB uncompressed
	// key-value pair, 8TB can contain up to around 8 billion)

	// L0 size = 64MB * 1 (min_write_buffer_number_to_merge) * \
	// 				8 (level0_file_num_compaction_trigger)
	// 		   = 512MB
	// L1 size = 64MB (target_file_base) * 8 (target_file_size_multiplier)
	//		   = 512MB = max_bytes_for_level_base
	// L2 size = 64MB (target_file_base) * 8^2 (target_file_size_multiplier)
	// 		   = 4GB = 512 (max_bytes_for_level_base) * 8 (max_bytes_for_level_multiplier)
	// L2 size = 64MB (target_file_base) * 8^3 (target_file_size_multiplier)
	// 		   = 32GB = 512 (max_bytes_for_level_base) * 8^2 (max_bytes_for_level_multiplier)
	// ...
	opts.SetMaxSubCompactions(2)
	opts.SetWriteBufferSize(64 * 1024 * 1024) // 64MB
	opts.SetMaxWriteBufferNumber(3)
	opts.SetMinWriteBufferNumberToMerge(2)
	opts.SetLevel0FileNumCompactionTrigger(8)
	opts.SetLevel0SlowdownWritesTrigger(24)
	opts.SetLevel0StopWritesTrigger(33)
	opts.SetTargetFileSizeBase(64 * 1024 * 1024) // 64MB
	opts.SetTargetFileSizeMultiplier(8)
	opts.SetMaxBytesForLevelBase(512 * 1024 * 1024) // 512MB
	opts.SetMaxBytesForLevelMultiplier(8)
	opts.SetNumLevels(7)

	// io parallelism
	opts.SetMaxBackgroundCompactions(8)
	opts.SetMaxBackgroundFlushes(1)
	return opts
}

// The history table is insert-only without updates so we have
// to optimize for an IO-bound and write-once workload.
func getHistoryTableOpts(blockCache *rocksdb.Cache) *rocksdb.Options {
	// This table performs both Get() and total order iterations.

	bbto := rocksdb.NewDefaultBlockBasedTableOptions()
	bbto.SetFilterPolicy(rocksdb.NewBloomFilterPolicy(10)) // TODO consider full filters instead of block filters
	// In order to have a fine-grained control over the memory usage
	// we cache SST's index and filters in the block cache. The alternative
	// would be to leave RocksDB keep those files memory mapped, but
	// the only way to control memory usage would be through the property
	// max_open_files.
	bbto.SetCacheIndexAndFilterBlocks(true)
	bbto.SetBlockCache(blockCache)
	// increase block size to 16KB
	bbto.SetBlockSize(16 * 1024)

	opts := rocksdb.NewDefaultOptions()
	opts.SetBlockBasedTableFactory(bbto)
	opts.SetCompression(rocksdb.SnappyCompression)

	// We use level style compaction with high concurrency.
	// Memtable size is 64MB and the total number of level 0
	// files is 8. This means compaction is triggered when L0
	// grows to 512MB. L1 size is 512MB and every level is 8 times
	// larger than the previous one. L2 is 4GB, L3 is 32GB,
	// L4 is 256GB, L5 is 2TB (note that given a 42B key-value
	// pair, 2TB can contain up to around 51 billion)

	// L0 size = 64MB * 1 (min_write_buffer_number_to_merge) * \
	// 				8 (level0_file_num_compaction_trigger)
	// 		   = 512MB
	// L1 size = 64MB (target_file_base) * 8 (target_file_size_multiplier)
	//		   = 512MB = max_bytes_for_level_base
	// L2 size = 64MB (target_file_base) * 8^2 (target_file_size_multiplier)
	// 		   = 4GB = 512 (max_bytes_for_level_base) * 8 (max_bytes_for_level_multiplier)
	// L2 size = 64MB (target_file_base) * 8^3 (target_file_size_multiplier)
	// 		   = 32GB = 512 (max_bytes_for_level_base) * 8^2 (max_bytes_for_level_multiplier)
	// ...
	opts.SetWriteBufferSize(64 * 1024 * 1024) // 64MB
	opts.SetMaxWriteBufferNumber(3)
	opts.SetMinWriteBufferNumberToMerge(1)
	opts.SetLevel0FileNumCompactionTrigger(8)
	opts.SetLevel0SlowdownWritesTrigger(17)
	opts.SetLevel0StopWritesTrigger(24)
	opts.SetTargetFileSizeBase(64 * 1024 * 1024) // 64MB
	opts.SetTargetFileSizeMultiplier(8)
	opts.SetMaxBytesForLevelBase(512 * 1024 * 1024) // 512MB
	opts.SetMaxBytesForLevelMultiplier(8)
	opts.SetNumLevels(5)

	// io parallelism
	opts.SetMaxBackgroundCompactions(8)
	opts.SetMaxBackgroundFlushes(1)
	return opts
}

// The FSM state table receives an update-only workload
// (only one point lookup when recovering), so we
// try to optimize for an IO-bound workload and multiple updates
// on the same key.
func getFsmStateTableOpts() *rocksdb.Options {

	// FSM state contains only one key that is updated on every
	// add event operation. We should try to reduce write and
	// space amplification by keeping a lower number of levels.

	bbto := rocksdb.NewDefaultBlockBasedTableOptions()
	// In order to have a fine-grained control over the memory usage
	// we cache SST's index and filters in the block cache. The alternative
	// would be to leave RocksDB keep those files memory mapped, but
	// the only way to control memory usage would be through the property
	// max_open_files.
	bbto.SetCacheIndexAndFilterBlocks(true)
	// decrease block size to 1KB
	bbto.SetBlockSize(1024)

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

	// io parallelism
	opts.SetMaxBackgroundCompactions(1)
	opts.SetMaxBackgroundFlushes(1)
	return opts
}

func (s *RocksDBStore) Mutate(mutations []*storage.Mutation, metadata []byte) error {
	batch := rocksdb.NewWriteBatch()
	defer batch.Destroy()
	// IMPORTANT: This line must go before the PutCF. For some reason, if we set it after,
	// we cannot retrieve the metadata later with a writebatch handler.
	batch.PutLogData(metadata, len(metadata))
	for _, m := range mutations {
		batch.PutCF(s.cfHandles[m.Table], m.Key, m.Value)
	}
	return s.db.Write(s.wo, batch)
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

	if s.backupOpts != nil {
		s.backupOpts.Destroy()
	}

	if s.restoreOpts != nil {
		s.restoreOpts.Destroy()
	}

	if s.blockCache != nil {
		s.blockCache.Destroy()
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

// Backup uses the backupEngine to create backups with metadata. The backup directory has been
// set up previously.
func (s *RocksDBStore) Backup(metadata string) error {
	err := s.backupEngine.CreateNewBackupWithMetadata(s.db, metadata)
	if err != nil {
		return err
	}
	return nil
}

// Restore from latest backup gets the latest backup from the backup engine and restores
// it to the given paths.This can be used to restore the database from a backup
// made by calling DB.Backup().
func (s *RocksDBStore) RestoreFromLatestBackup(dbDir, walDir string) error {
	err := s.backupEngine.RestoreDBFromLatestBackup(dbDir, walDir, s.restoreOpts)
	if err != nil {
		return err
	}
	return nil
}

// Restore from backup looks for the given backupID in the backup engine, gets and restores
// it to the given paths. This can be used to restore the database from a backup
// made by calling DB.Backup().
func (s *RocksDBStore) RestoreFromBackup(backupID uint32, dbDir, walDir string) error {
	err := s.backupEngine.RestoreDBFromBackup(backupID, dbDir, walDir, s.restoreOpts)
	if err != nil {
		return err
	}
	return nil
}

// GetBackupsInfo function extract a list of backups from a backup engine, and iterate over them
// to parse its information.
func (s *RocksDBStore) GetBackupsInfo() []*storage.BackupInfo {
	bi := s.backupEngine.GetInfo()
	defer bi.Destroy()
	if bi == nil {
		return nil
	}

	backupsInfo := make([]*storage.BackupInfo, bi.GetCount())
	for i := 0; i < bi.GetCount(); i++ {
		info := &storage.BackupInfo{}
		info.ID = bi.GetBackupID(i)
		info.Timestamp = bi.GetTimestamp(i)
		info.Size = bi.GetSize(i)
		info.NumFiles = bi.GetNumFiles(i)
		info.Metadata = bi.GetAppMetadata(i)
		backupsInfo[i] = info
	}

	return backupsInfo
}

// DeleteBackup uses the backupEngine to delete the backup identified by backupID.
func (s *RocksDBStore) DeleteBackup(backupID uint32) error {
	err := s.backupEngine.DeleteBackup(backupID)
	if err != nil {
		return err
	}
	return nil
}

// FetchSnapshot fetches all WAL transactions from the first available
// seq_num to the last one specified in the lastSeqNum parameter, and dumps
// them to the given writer.
func (s *RocksDBStore) FetchSnapshot(w io.WriteCloser, since, until uint64, valid storage.ValidateF) error {

	extractor := rocksdb.NewLogDataExtractor(s.path)
	defer func() {
		extractor.Destroy()
	}()

	it, err := s.db.GetUpdatesSince(since) // we start on the first available seq_num
	if err != nil {
		return err
	}
	defer func() {
		it.Close()
		w.Close()
	}()

	for ; it.Valid(); it.Next() {
		batch, seqNum := it.GetBatch()
		defer batch.Destroy()
		if seqNum <= since {
			continue
		}
		if seqNum > until {
			break
		}

		ok, err := valid(batch.GetLogData(extractor))
		if err != nil {
			return err
		}

		if !ok {
			continue
		}

		data := batch.Data()
		size := util.Uint64AsBytes(uint64(len(data)))
		_, err = w.Write(append(size, data...))
		if err != nil {
			return err
		}
	}

	return nil
}

func readChunk(r io.Reader) ([]byte, error) {

	sizeBuff := make([]byte, 8)
	n, err := r.Read(sizeBuff)
	if err != nil || n != 8 {
		return nil, err
	}

	size := util.BytesAsUint64(sizeBuff)
	chunk := make([]byte, size)
	n, err = r.Read(chunk)
	if err != nil {
		return nil, err
	}
	if uint64(n) != size {
		return nil, fmt.Errorf("Corrupted chunk")
	}
	return chunk, nil

}

// LoadSnapshot reads a list of serialized batches from a reader,
// rehydrates them and write to the database if they fulfill the validation
// condition specified as parameter.
// This method should be called on a database that is not running
// any other concurrent transactions while it is running.
func (s *RocksDBStore) LoadSnapshot(r io.ReadCloser) error {

	wo := rocksdb.NewDefaultWriteOptions()

	for {
		buff, err := readChunk(r)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		batch := rocksdb.WriteBatchFrom(buff)
		defer batch.Destroy()

		err = s.db.Write(wo, batch)
		if err != nil {
			return nil
		}
	}

	return nil

}

// LastWALSequenceNumber returns the sequence number of the
// last transaction applied to the WAL. This sequence
// number can be used as upper limit when fetching transactions
// from the WAL.
func (s *RocksDBStore) LastWALSequenceNumber() uint64 {
	return s.db.GetLatestSequenceNumber()
}

func (s *RocksDBStore) RegisterMetrics(registry metrics.Registry) {
	if registry != nil {
		registry.MustRegister(s.metrics.collectors()...)
	}
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
