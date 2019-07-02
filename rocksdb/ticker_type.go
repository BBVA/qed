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

// #include "extended.h"
import (
	"C"
)

// TickerType is the logical mapping of tickers defined in rocksdb::Tickers.
type TickerType uint32

const (
	// TickerBlockCacheMiss is the total number of block bache misses.
	TickerBlockCacheMiss = TickerType(C.BLOCK_CACHE_MISS)
	// TickerBlockCacheHit is the total number of block bache hits.
	TickerBlockCacheHit = TickerType(C.BLOCK_CACHE_HIT)
	// TickerBlockCacheAdd is the number of blocks added to block cache.
	TickerBlockCacheAdd = TickerType(C.BLOCK_CACHE_ADD)
	// TickerBlockCacheAddFailures is the number of failures when adding blocks to block cache.
	TickerBlockCacheAddFailures = TickerType(C.BLOCK_CACHE_ADD_FAILURES)
	// TickerBlockCacheIndexMiss is the number of cache misses when accessing index block from block cache.
	TickerBlockCacheIndexMiss = TickerType(C.BLOCK_CACHE_INDEX_MISS)
	// TickerBlockCacheIndexHit is the number cache hits when accessing index block from block cache.
	TickerBlockCacheIndexHit = TickerType(C.BLOCK_CACHE_INDEX_HIT)
	// TickerBlockCacheIndexAdd is the number of index blocks added to block cache.
	TickerBlockCacheIndexAdd = TickerType(C.BLOCK_CACHE_INDEX_ADD)
	// TickerBlockCacheIndexBytesInsert is the number of bytes of index blocks inserted into cache.
	TickerBlockCacheIndexBytesInsert = TickerType(C.BLOCK_CACHE_INDEX_BYTES_INSERT)
	// TickerBlockCacheIndexBytesEvict is the number of bytes of index block erased from cache.
	TickerBlockCacheIndexBytesEvict = TickerType(C.BLOCK_CACHE_INDEX_BYTES_EVICT)
	// TickerBlockCacheFilterMiss is the number of times cache misses when accessing filter block from block cache.
	TickerBlockCacheFilterMiss = TickerType(C.BLOCK_CACHE_FILTER_MISS)
	// TickerBlockCacheFilterHit is the number of times cache hits when accessing filter block from block cache.
	TickerBlockCacheFilterHit = TickerType(C.BLOCK_CACHE_FILTER_HIT)
	// TickerBlockCacheFilterAdd is the number of filter blocks added to block cache.
	TickerBlockCacheFilterAdd = TickerType(C.BLOCK_CACHE_FILTER_ADD)
	// TickerBlockCacheFilterBytesInsert is the number of bytes of bloom filter blocks inserted into cache.
	TickerBlockCacheFilterBytesInsert = TickerType(C.BLOCK_CACHE_FILTER_BYTES_INSERT)
	// TickerBlockCacheFilterBytesEvict is the number of bytes of bloom filter block erased from cache.
	TickerBlockCacheFilterBytesEvict = TickerType(C.BLOCK_CACHE_FILTER_BYTES_EVICT)
	// TickerBlockCacheDataMiss is the number of times cache misses when accessing data block from block cache.
	TickerBlockCacheDataMiss = TickerType(C.BLOCK_CACHE_DATA_MISS)
	// TickerBlockCacheDataHit is the number of times cache hits when accessing data block from block cache.
	TickerBlockCacheDataHit = TickerType(C.BLOCK_CACHE_DATA_HIT)
	// TickerBlockCacheDataAdd is the number of data blocks added to block cache.
	TickerBlockCacheDataAdd = TickerType(C.BLOCK_CACHE_DATA_ADD)
	// TickerBlockCacheDataBytesInsert is the number of bytes of data blocks inserted into cache.
	TickerBlockCacheDataBytesInsert = TickerType(C.BLOCK_CACHE_DATA_BYTES_INSERT)
	// TickerBlockCacheBytesRead is the number of bytes read from cache.
	TickerBlockCacheBytesRead = TickerType(C.BLOCK_CACHE_BYTES_READ)
	// TickerBlockCacheBytesWrite is the number of bytes written into cache.
	TickerBlockCacheBytesWrite = TickerType(C.BLOCK_CACHE_BYTES_WRITE)

	// TickerBloomFilterUseful is the number of times bloom filter has avoided file reads, i.e., negatives.
	TickerBloomFilterUseful = TickerType(C.BLOOM_FILTER_USEFUL)
	// TickerBloomFilterFullPositive is the number of times bloom FullFilter has not avoided the reads.
	TickerBloomFilterFullPositive = TickerType(C.BLOOM_FILTER_FULL_POSITIVE)
	// TickerBloomFilterFullTruePositive is the number of times bloom FullFilter has not avoided the reads and
	// data actually exist.
	TickerBloomFilterFullTruePositive = TickerType(C.BLOOM_FILTER_FULL_TRUE_POSITIVE)

	// TickerMemtableHit is the number of memtable hits.
	TickerMemtableHit = TickerType(C.MEMTABLE_HIT)
	// TickerMemtableMiss is the number of memtable misses.
	TickerMemtableMiss = TickerType(C.MEMTABLE_MISS)

	// TickerGetHitL0 is the number of Get() queries served by L0.
	TickerGetHitL0 = TickerType(C.GET_HIT_L0)
	// TickerGetHitL1 is the number of Get() queries served by L1.
	TickerGetHitL1 = TickerType(C.GET_HIT_L1)
	// TickerGetHitL2AndUp is the number of Get() queries served by L2 and up.
	TickerGetHitL2AndUp = TickerType(C.GET_HIT_L2_AND_UP)

	// TickerNumberKeysWritten is the number of keys written to the database via the Put and Write call's.
	TickerNumberKeysWritten = TickerType(C.NUMBER_KEYS_WRITTEN)
	// TickerNumberKeysRead is the number of Keys read.
	TickerNumberKeysRead = TickerType(C.NUMBER_KEYS_READ)
	// TickerNumberKeysUpdated is the number keys updated, if inplace update is enabled.
	TickerNumberKeysUpdated = TickerType(C.NUMBER_KEYS_UPDATED)
	// TickerBytesWritten is the number of uncompressed bytes issued by db.Put(),
	// db.Delete(), db.Merge(), and db.Write().
	TickerBytesWritten = TickerType(C.BYTES_WRITTEN)
	// TickerBytesRead is the number of uncompressed bytes read from db.Get().
	// It could be either from memtables, cache, or table files.
	// For the number of logical bytes read from db.MultiGet(),
	// please use NumberMultiGetBytesRead.
	TickerBytesRead = TickerType(C.BYTES_READ)
	// TickerStallMicros is the number of microseconds the writer has to wait for
	// compaction or flush to finish.
	TickerStallMicros = TickerType(C.STALL_MICROS)

	// TickerBlockCacheCompressedMiss is the number of misses in the compressed block cache.
	TickerBlockCacheCompressedMiss = TickerType(C.BLOCK_CACHE_COMPRESSED_MISS)
	// TickerBlockCacheCompressedHit is the number of hits in the compressed block cache.
	TickerBlockCacheCompressedHit = TickerType(C.BLOCK_CACHE_COMPRESSED_HIT)
	// TickerBlockCacheCompressedAdd is the number of blocks added to compressed block cache.
	TickerBlockCacheCompressedAdd = TickerType(C.BLOCK_CACHE_COMPRESSED_ADD)
	// TickerBlockCacheCompressedAddFailures is the number of failures when adding blocks to compressed block cache.
	TickerBlockCacheCompressedAddFailures = TickerType(C.BLOCK_CACHE_COMPRESSED_ADD_FAILURES)

	// TickerWALFileSynced is the number of times WAL sync is done.
	TickerWALFileSynced = TickerType(C.WAL_FILE_SYNCED)
	// TickerWALFileBytes is the number of bytes written to WAL.
	TickerWALFileBytes = TickerType(C.WAL_FILE_BYTES)

	// TickerCompactReadBytes is the number of bytes read during compaction.
	TickerCompactReadBytes = TickerType(C.COMPACT_READ_BYTES)
	// TickerCompactWriteBytes is the number of bytes written during compaction.
	TickerCompactWriteBytes = TickerType(C.COMPACT_WRITE_BYTES)
	// TickerFlushWriteBytes is the number of bytes written during flush.
	TickerFlushWriteBytes = TickerType(C.FLUSH_WRITE_BYTES)

	// TickerNumberBlockCompressed is the number of compressions executed.
	TickerNumberBlockCompressed = TickerType(C.NUMBER_BLOCK_COMPRESSED)
	// TickerNumberBlockDecompressed is the number of decompressions executed.
	TickerNumberBlockDecompressed = TickerType(C.NUMBER_BLOCK_DECOMPRESSED)
	// TickerNumberBlockNotCompressed is the number of blocks not compressed.
	TickerNumberBlockNotCompressed = TickerType(C.NUMBER_BLOCK_NOT_COMPRESSED)

	// TickerMergeOperationTotalTime is the number in # of all merge operations.
	TickerMergeOperationTotalTime = TickerType(C.MERGE_OPERATION_TOTAL_TIME)
	// TickerFilterOperationTotalTime is the number in # of all filter operations.
	TickerFilterOperationTotalTime = TickerType(C.FILTER_OPERATION_TOTAL_TIME)

	// Read amplification statistics.
	// Read amplification can be calculated using this formula
	// (READ_AMP_TOTAL_READ_BYTES / READ_AMP_ESTIMATE_USEFUL_BYTES)
	//
	// REQUIRES: ReadOptions::read_amp_bytes_per_bit to be enabled
	
	// TickerReadAmpEstimateUsefulBytes is the estimate of total bytes actually used.
	TickerReadAmpEstimateUsefulBytes = TickerType(C.READ_AMP_ESTIMATE_USEFUL_BYTES)
	// TickerReadAmpTotalReadBytes is the total size of loaded data blocks.
	TickerReadAmpTotalReadBytes = TickerType(C.READ_AMP_TOTAL_READ_BYTES)
)
