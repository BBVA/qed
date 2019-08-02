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

package rocksdb

// #include <stdlib.h>
// #include <rocksdb/c.h>
// #include <extended.h>
import "C"
import "unsafe"

// CompressionType specifies the block compression.
// DB contents are stored in a set of blocks, each of which holds a
// sequence of key,value pairs. Each block may be compressed before
// being stored in a file. The following enum describes which
// compression method (if any) is used to compress a block.
type CompressionType uint

// Compression types
const (
	NoCompression     = CompressionType(C.rocksdb_no_compression)
	SnappyCompression = CompressionType(C.rocksdb_snappy_compression)
)

// Options represent all of the available options when opening a database with Open.
type Options struct {
	c *C.rocksdb_options_t

	// Hold references for GC.
	env  *Env
	bbto *BlockBasedTableOptions
	cst  *C.rocksdb_slicetransform_t
}

// NewDefaultOptions creates the default Options.
func NewDefaultOptions() *Options {
	return &Options{c: C.rocksdb_options_create()}
}

// SetCreateIfMissing specifies whether the database
// should be created if it is missing.
// Default: false
func (o *Options) SetCreateIfMissing(value bool) {
	C.rocksdb_options_set_create_if_missing(o.c, boolToUchar(value))
}

// SetEnv sets the specified object to interact with the environment,
// e.g. to read/write files, schedule background work, etc.
// Default: DefaultEnv
func (o *Options) SetEnv(value *Env) {
	o.env = value
	C.rocksdb_options_set_env(o.c, value.c)
}

// IncreaseParallelism sets the level of parallelism.
//
// By default, RocksDB uses only one background thread for flush and
// compaction. Calling this function will set it up such that total of
// `totalThreads` is used. Good value for `totalThreads` is the number of
// cores. You almost definitely want to call this function if your system is
// bottlenecked by RocksDB.
func (o *Options) IncreaseParallelism(totalThreads int) {
	C.rocksdb_options_increase_parallelism(o.c, C.int(totalThreads))
}

// SetMaxWriteBufferNumber sets the maximum number of write buffers (memtables)
// that are built up in memory.
//
// The default is 2, so that when 1 write buffer is being flushed to
// storage, new writes can continue to the other write buffer.
// Default: 2
func (o *Options) SetMaxWriteBufferNumber(value int) {
	C.rocksdb_options_set_max_write_buffer_number(o.c, C.int(value))
}

// SetMinWriteBufferNumberToMerge sets the minimum number of write buffers
// that will be merged together before writing to storage.
//
// If set to 1, then all write buffers are flushed to L0 as individual files
// and this increases read amplification because a get request has to check
// in all of these files. Also, an in-memory merge may result in writing lesser
// data to storage if there are duplicate records in each of these
// individual write buffers.
// Default: 1
func (o *Options) SetMinWriteBufferNumberToMerge(value int) {
	C.rocksdb_options_set_min_write_buffer_number_to_merge(o.c, C.int(value))
}

// SetMaxOpenFiles sets the number of open files that can be used by the DB.
//
// You may need to increase this if your database has a large working set
// (budget one open file per 2MB of working set).
// Default: 1000
func (o *Options) SetMaxOpenFiles(value int) {
	C.rocksdb_options_set_max_open_files(o.c, C.int(value))
}

// SetMaxFileOpeningThreads sets the maximum number of file opening threads.
// If max_open_files is -1, DB will open all files on db.Open(). You can
// use this option to increase the number of threads used to open the files.
// Default: 16
func (o *Options) SetMaxFileOpeningThreads(value int) {
	C.rocksdb_options_set_max_file_opening_threads(o.c, C.int(value))
}

// OptimizeForPointLookup optimize the DB for point lookups.
//
// Use this if you don't need to keep the data sorted, i.e. you'll never use
// an iterator, only Put() and Get() API calls
//
// If you use this with rocksdb >= 5.0.2, you must call `SetAllowConcurrentMemtableWrites(false)`
// to avoid an assertion error immediately on opening the db.
func (o *Options) OptimizeForPointLookup(blockCacheSizeMb uint64) {
	C.rocksdb_options_optimize_for_point_lookup(o.c, C.uint64_t(blockCacheSizeMb))
}

// SetAllowConcurrentMemtableWrites sets whether to allow concurrent memtable writes. Conccurent writes are
// not supported by all memtable factories (currently only SkipList memtables).
// As of rocksdb 5.0.2 you must call `SetAllowConcurrentMemtableWrites(false)`
// if you use `OptimizeForPointLookup`.
func (o *Options) SetAllowConcurrentMemtableWrites(allow bool) {
	C.rocksdb_options_set_allow_concurrent_memtable_write(o.c, boolToUchar(allow))
}

// SetMemtableVectorRep sets a MemTableRep which is backed by a vector.
//
// On iteration, the vector is sorted. This is useful for workloads where
// iteration is very rare and writes are generally not issued after reads begin.
func (o *Options) SetMemtableVectorRep() {
	C.rocksdb_options_set_memtable_vector_rep(o.c)
}

// SetHashSkipListRep sets a hash skip list as MemTableRep.
//
// It contains a fixed array of buckets, each
// pointing to a skiplist (null if the bucket is empty).
//
// bucketCount:             number of fixed array buckets
// skiplistHeight:          the max height of the skiplist
// skiplistBranchingFactor: probabilistic size ratio between adjacent
//                          link lists in the skiplist
func (o *Options) SetHashSkipListRep(bucketCount int, skiplistHeight, skiplistBranchingFactor int32) {
	C.rocksdb_options_set_hash_skip_list_rep(o.c, C.size_t(bucketCount), C.int32_t(skiplistHeight), C.int32_t(skiplistBranchingFactor))
}

// SetHashLinkListRep sets a hashed linked list as MemTableRep.
//
// It contains a fixed array of buckets, each pointing to a sorted single
// linked list (null if the bucket is empty).
//
// bucketCount: number of fixed array buckets
func (o *Options) SetHashLinkListRep(bucketCount int) {
	C.rocksdb_options_set_hash_link_list_rep(o.c, C.size_t(bucketCount))
}

// SetBlockBasedTableFactory sets the block based table factory.
//
// This is the default table type that we inherited from LevelDB, which was
// designed for storing data in hard disk or flash device.
//
// In block-based table, data is chucked into (almost) fix-sized blocks
// (default block size is 4k). Each block, in turn, keeps a bunch of entries.
//
// When storing data, we can compress and/or encode data efficiently within a
// block, which often resulted in a much smaller data size compared with the raw
// data size.
//
// As for the record retrieval, we'll first locate the block where target record
// may reside, then read the block to memory, and finally search that record within
// the block. Of course, to avoid frequent reads of the same block, we introduced
// the block cache to keep the loaded blocks in the memory.
func (o *Options) SetBlockBasedTableFactory(value *BlockBasedTableOptions) {
	o.bbto = value
	C.rocksdb_options_set_block_based_table_factory(o.c, value.c)
}

// SetPlainTableFactory sets a plain table factory with prefix-only seek.
//
// Plain table stores data in a sequence of key/value pairs. Compared with
// block-based table, which employs mostly binary search for entry lookup,
// the well designed hash-based index in plain table enables us to locate
// data magnitudes faster. No memory copy is needed. Plain table bypasses
// the concept of "block" and therefore avoids the overhead inherent in block-based
// table, like extra block lookup, block cache, etc.
//
// For this factory, you need to set prefix_extractor properly to make it
// work. Look-up will starts with prefix hash lookup for key prefix. Inside the
// hash bucket found, a binary search is executed for hash conflicts. Finally,
// a linear search is used.
//
// Limitations:
//
// - File size may not be greater than 2^31 - 1 (i.e., `2147483647`) bytes.
// - Data compression/Delta encoding is not supported, which may resulted in
//	 bigger file size compared with block-based table.
// - Backward (Iterator.Prev()) scan is not supported.
// - Non-prefix-based Seek() is not supported.
// - Table loading is slower since indexes are built on the fly by 2-pass table scanning.
// - Only support mmap mode.
//
// keyLen: 			plain table has optimization for fix-sized keys,
// 					which can be specified via keyLen. Alternatively, you can
// 					pass 0 if your keys have variable lengths.
// bloomBitsPerKey: the number of bits used for bloom filer per prefix. You
//                  may disable it by passing a zero.
// hashTableRatio:  the desired utilization of the hash table used for prefix
//                  hashing. hashTableRatio = number of prefixes / #buckets
//                  in the hash table
// indexSparseness: inside each prefix, need to build one index record for how
//                  many keys for binary search inside each hash bucket.
func (o *Options) SetPlainTableFactory(keyLen uint32, bloomBitsPerKey int, hashTableRatio float64, indexSparseness int) {
	C.rocksdb_options_set_plain_table_factory(o.c, C.uint32_t(keyLen), C.int(bloomBitsPerKey), C.double(hashTableRatio), C.size_t(indexSparseness))
}

// SetCreateIfMissingColumnFamilies specifies whether the column families
// should be created if they are missing.
func (o *Options) SetCreateIfMissingColumnFamilies(value bool) {
	C.rocksdb_options_set_create_missing_column_families(o.c, boolToUchar(value))
}

// SetPrefixExtractor sets the prefic extractor.
//
// If set, use the specified function to determine the
// prefixes for keys. These prefixes will be placed in the filter.
// Depending on the workload, this can reduce the number of read-IOP
// cost for scans when a prefix is passed via ReadOptions to
// db.NewIterator().
// Default: nil
func (o *Options) SetPrefixExtractor(st SliceTransform) {
	if nst, ok := st.(nativeSliceTransform); ok {
		o.cst = nst.c
	} else {
		idx := registerSliceTransform(st)
		o.cst = C.rocksdb_slicetransform_create_ext(C.uintptr_t(idx))
	}
	C.rocksdb_options_set_prefix_extractor(o.c, o.cst)
}

// SetCompression sets the compression algorithm.
// Default: SnappyCompression, which gives lightweight but fast
// compression.
func (o *Options) SetCompression(value CompressionType) {
	C.rocksdb_options_set_compression(o.c, C.int(value))
}

// SetNumLevels sets the number of levels for this database.
// Default: 7
func (o *Options) SetNumLevels(value int) {
	C.rocksdb_options_set_num_levels(o.c, C.int(value))
}

// SetLevel0FileNumCompactionTrigger sets the number of files
// to trigger level-0 compaction.
//
// A value <0 means that level-0 compaction will not be
// triggered by number of files at all.
// Default: 4
func (o *Options) SetLevel0FileNumCompactionTrigger(value int) {
	C.rocksdb_options_set_level0_file_num_compaction_trigger(o.c, C.int(value))
}

// SetLevel0SlowdownWritesTrigger sets the soft limit on number of level-0 files.
//
// We start slowing down writes at this point.
// A value <0 means that no writing slow down will be triggered by
// number of files in level-0.
// Default: 8
func (o *Options) SetLevel0SlowdownWritesTrigger(value int) {
	C.rocksdb_options_set_level0_slowdown_writes_trigger(o.c, C.int(value))
}

// SetLevel0StopWritesTrigger sets the maximum number of level-0 files.
// We stop writes at this point.
// Default: 12
func (o *Options) SetLevel0StopWritesTrigger(value int) {
	C.rocksdb_options_set_level0_stop_writes_trigger(o.c, C.int(value))
}

// SetMaxBytesForLevelBase sets the maximum total data size for a level.
//
// It is the max total for level-1.
// Maximum number of bytes for level L can be calculated as
// (max_bytes_for_level_base) * (max_bytes_for_level_multiplier ^ (L-1))
//
// For example, if max_bytes_for_level_base is 20MB, and if
// max_bytes_for_level_multiplier is 10, total data size for level-1
// will be 20MB, total file size for level-2 will be 200MB,
// and total file size for level-3 will be 2GB.
// Default: 10MB
func (o *Options) SetMaxBytesForLevelBase(value uint64) {
	C.rocksdb_options_set_max_bytes_for_level_base(o.c, C.uint64_t(value))
}

// SetMaxBytesForLevelMultiplier sets the max Bytes for level multiplier.
// Default: 10
func (o *Options) SetMaxBytesForLevelMultiplier(value float64) {
	C.rocksdb_options_set_max_bytes_for_level_multiplier(o.c, C.double(value))
}

// SetTargetFileSizeBase sets the target file size for compaction.
//
// Target file size is per-file size for level-1.
// Target file size for level L can be calculated by
// target_file_size_base * (target_file_size_multiplier ^ (L-1))
//
// For example, if target_file_size_base is 2MB and
// target_file_size_multiplier is 10, then each file on level-1 will
// be 2MB, and each file on level 2 will be 20MB,
// and each file on level-3 will be 200MB.
// Default: 2MB
func (o *Options) SetTargetFileSizeBase(value uint64) {
	C.rocksdb_options_set_target_file_size_base(o.c, C.uint64_t(value))
}

// SetTargetFileSizeMultiplier sets the target file size multiplier for compaction.
// Default: 1
func (o *Options) SetTargetFileSizeMultiplier(value int) {
	C.rocksdb_options_set_target_file_size_multiplier(o.c, C.int(value))
}

// SetWriteBufferSize sets the amount of data to build up in memory
// (backed by an unsorted log on disk) before converting to a sorted on-disk file.
//
// Larger values increase performance, especially during bulk loads.
// Up to max_write_buffer_number write buffers may be held in memory
// at the same time, so you may wish to adjust this parameter to control
// memory usage.
// Also, a larger write buffer will result in a longer recovery time
// the next time the database is opened.
// Default: 4MB
func (o *Options) SetWriteBufferSize(value int) {
	C.rocksdb_options_set_write_buffer_size(o.c, C.size_t(value))
}

// SetDbWriteBufferSize sets the amount of data to build up
// in memtables across all column families before writing to disk.
//
// This is distinct from write_buffer_size, which enforces a limit
// for a single memtable.
//
// This feature is disabled by default. Specify a non-zero value
// to enable it.
//
// Default: 0 (disabled)
func (o *Options) SetDbWriteBufferSize(value int) {
	C.rocksdb_options_set_db_write_buffer_size(o.c, C.size_t(value))
}

// SetMaxSubCompactions sets the maximum number of threads that will
// concurrently perform a compaction job by breaking it into multiple,
// smaller ones that are run simultaneously.
// Default: 1 (i.e. no subcompactions)
func (o *Options) SetMaxSubCompactions(value int) {
	C.rocksdb_options_set_max_subcompactions(o.c, C.uint(value))
}

// SetEnablePipelinedWrite improves concurrent write throughput in
// case WAL is enabled. By default, a single write thread queue is
// maintained for concurrent writers. The thread gets to the head
// of the queue becomes write batch group leader and responsible
// for writing to WAL and memtable for the batch group.
// One observation is that WAL writes and memtable writes are sequential
// and by making them run in parallel we can increase throughput.
// For one single writer WAL writes and memtable writes has to run
// sequentially. With concurrent writers, once the previous writer
// finish WAL write, the next writer waiting in the write queue can
// start to write WAL while the previous writer still have memtable
// write ongoing.
func (o *Options) SetEnablePipelinedWrite(value bool) {
	C.rocksdb_options_set_enable_pipelined_write(o.c, boolToUchar(value))
}

// SetUseFsync enable/disable fsync.
//
// If true, then every store to stable storage will issue a fsync.
// If false, then every store to stable storage will issue a fdatasync.
// This parameter should be set to true while storing data to
// filesystem like ext3 that can lose files after a reboot.
// Default: false
func (o *Options) SetUseFsync(value bool) {
	C.rocksdb_options_set_use_fsync(o.c, C.int(btoi(value)))
}

// SetUseDirectReads enable/disable direct I/O mode (O_DIRECT) for reads.
// Default: false
func (o *Options) SetUseDirectReads(value bool) {
	C.rocksdb_options_set_use_direct_reads(o.c, boolToUchar(value))
}

// SetUseDirectIOForFlushAndCompaction enable/disable direct I/O mode (O_DIRECT)
// for both reads and writes in background flush and compactions.
// When true, new_table_reader_for_compaction_inputs is forced to true.
// Default: false
func (o *Options) SetUseDirectIOForFlushAndCompaction(value bool) {
	C.rocksdb_options_set_use_direct_io_for_flush_and_compaction(o.c, boolToUchar(value))
}

// SetMaxTotalWalSize sets the maximum total wal size in bytes.
// Once write-ahead logs exceed this size, we will start forcing the flush of
// column families whose memtables are backed by the oldest live WAL file
// (i.e. the ones that are causing all the space amplification). If set to 0
// (default), we will dynamically choose the WAL size limit to be
// [sum of all write_buffer_size * max_write_buffer_number] * 4
// Default: 0
func (o *Options) SetMaxTotalWalSize(value uint64) {
	C.rocksdb_options_set_max_total_wal_size(o.c, C.uint64_t(value))
}

// SetDBLogDir specifies the absolute info LOG dir.
//
// If it is empty, the log files will be in the same dir as data.
// If it is non empty, the log files will be in the specified dir,
// and the db data dir's absolute path will be used as the log file
// name's prefix.
// Default: empty
func (o *Options) SetDBLogDir(value string) {
	cValue := C.CString(value)
	defer C.free(unsafe.Pointer(cValue))
	C.rocksdb_options_set_db_log_dir(o.c, cValue)
}

// SetWalDir specifies the absolute dir path for write-ahead logs (WAL).
//
// If it is empty, the log files will be in the same dir as data.
// If it is non empty, the log files will be in the specified dir,
// When destroying the db, all log files and the dir are deleted.
// Default: empty
func (o *Options) SetWalDir(value string) {
	cValue := C.CString(value)
	defer C.free(unsafe.Pointer(cValue))
	C.rocksdb_options_set_wal_dir(o.c, cValue)
}

// SetMaxBackgroundCompactions sets the maximum number of
// concurrent background jobs, submitted to
// the default LOW priority thread pool
// Default: 1
func (o *Options) SetMaxBackgroundCompactions(value int) {
	C.rocksdb_options_set_max_background_compactions(o.c, C.int(value))
}

// SetMaxBackgroundFlushes sets the maximum number of
// concurrent background memtable flush jobs, submitted to
// the HIGH priority thread pool.
//
// By default, all background jobs (major compaction and memtable flush) go
// to the LOW priority pool. If this option is set to a positive number,
// memtable flush jobs will be submitted to the HIGH priority pool.
// It is important when the same Env is shared by multiple db instances.
// Without a separate pool, long running major compaction jobs could
// potentially block memtable flush jobs of other db instances, leading to
// unnecessary Put stalls.
// Default: 0
func (o *Options) SetMaxBackgroundFlushes(value int) {
	C.rocksdb_options_set_max_background_flushes(o.c, C.int(value))
}

// SetMaxLogFileSize sets the maximal size of the info log file.
//
// If the log file is larger than `max_log_file_size`, a new info log
// file will be created.
// If max_log_file_size == 0, all logs will be written to one log file.
// Default: 0
func (o *Options) SetMaxLogFileSize(value int) {
	C.rocksdb_options_set_max_log_file_size(o.c, C.size_t(value))
}

// SetLogFileTimeToRoll sets the time for the info log file to roll (in seconds).
//
// If specified with non-zero value, log file will be rolled
// if it has been active longer than `log_file_time_to_roll`.
// Default: 0 (disabled)
func (o *Options) SetLogFileTimeToRoll(value int) {
	C.rocksdb_options_set_log_file_time_to_roll(o.c, C.size_t(value))
}

// SetKeepLogFileNum sets the maximal info log files to be kept.
// Default: 1000
func (o *Options) SetKeepLogFileNum(value int) {
	C.rocksdb_options_set_keep_log_file_num(o.c, C.size_t(value))
}

// SetWALTtlSeconds sets the WAL ttl in seconds.
//
// The following two options affect how archived logs will be deleted.
// 1. If both set to 0, logs will be deleted asap and will not get into
//    the archive.
// 2. If wal_ttl_seconds is 0 and wal_size_limit_mb is not 0,
//    WAL files will be checked every 10 min and if total size is greater
//    then wal_size_limit_mb, they will be deleted starting with the
//    earliest until size_limit is met. All empty files will be deleted.
// 3. If wal_ttl_seconds is not 0 and wall_size_limit_mb is 0, then
//    WAL files will be checked every wal_ttl_seconds / 2 and those that
//    are older than wal_ttl_seconds will be deleted.
// 4. If both are not 0, WAL files will be checked every 10 min and both
//    checks will be performed with ttl being first.
// Default: 0
func (o *Options) SetWALTtlSeconds(value uint64) {
	C.rocksdb_options_set_WAL_ttl_seconds(o.c, C.uint64_t(value))
}

// SetWalSizeLimitMb sets the WAL size limit in MB.
//
// If total size of WAL files is greater then wal_size_limit_mb,
// they will be deleted starting with the earliest until size_limit is met
// Default: 0
func (o *Options) SetWalSizeLimitMb(value uint64) {
	C.rocksdb_options_set_WAL_size_limit_MB(o.c, C.uint64_t(value))
}

// SetAllowMmapReads enables/disables mmap reads for reading sst tables.
// Default: false
func (o *Options) SetAllowMmapReads(value bool) {
	C.rocksdb_options_set_allow_mmap_reads(o.c, boolToUchar(value))
}

// SetAllowMmapWrites enables/disables mmap writes for writing sst tables.
// Default: false
func (o *Options) SetAllowMmapWrites(value bool) {
	C.rocksdb_options_set_allow_mmap_writes(o.c, boolToUchar(value))
}

// SetAtomicFlush enables/disables atomic flushes.
// If true, RocksDB supports flushing multiple column families and committing
// their results atomically to MANIFEST. Note that it is not
// necessary to set atomic_flush to true if WAL is always enabled since WAL
// allows the database to be restored to the last persistent state in WAL.
// This option is useful when there are column families with writes NOT
// protected by WAL.
// For manual flush, application has to specify which column families to
// flush atomically in db.Flush.
// For auto-triggered flush, RocksDB atomically flushes ALL column families.
//
// Currently, any WAL-enabled writes after atomic flush may be replayed
// independently if the process crashes later and tries to recover.
func (o *Options) SetAtomicFlush(value bool) {
	C.rocksdb_options_set_atomic_flush(o.c, boolToUchar(value))
}

// SetStatistics sets a statistics object to pass to the DB.
func (o *Options) SetStatistics(s *Statistics) {
	C.rocksdb_options_set_statistics(o.c, s.c)
}

// Destroy deallocates the Options object.
func (o *Options) Destroy() {
	C.rocksdb_options_destroy(o.c)
	if o.env != nil {
		o.env.Destroy()
	}
	if o.bbto != nil {
		o.bbto.Destroy()
	}
	if o.cst != nil {
		C.rocksdb_slicetransform_destroy(o.cst)
	}
	o.c = nil
	o.env = nil
	o.bbto = nil
}
