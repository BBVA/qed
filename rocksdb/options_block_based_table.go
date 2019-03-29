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

// #include <rocksdb/c.h>
import "C"

// IndexType specifies the index type that will be used for this table.
type IndexType uint

const (
	// KBinarySearchIndexType is a space efficient index block that is optimized for
	// binary-search-based index.
	KBinarySearchIndexType IndexType = iota
	// KHashSearchIndexType is the hash index, if enabled, will do the hash lookup when
	// `Options.prefix_extractor` is provided.
	KHashSearchIndexType
	// KTwoLevelIndexSearchIndexType is a two-level index implementation. Both
	// levels are binary search indexes.
	KTwoLevelIndexSearchIndexType
)

// BlockBasedTableOptions represents block-based table options.
type BlockBasedTableOptions struct {
	c *C.rocksdb_block_based_table_options_t

	// Hold references for GC.
	cache     *Cache
	cacheComp *Cache

	// We keep these so we can free their memory in Destroy.
	fp *C.rocksdb_filterpolicy_t
}

// NewDefaultBlockBasedTableOptions creates a default BlockBasedTableOptions object.
func NewDefaultBlockBasedTableOptions() *BlockBasedTableOptions {
	return &BlockBasedTableOptions{c: C.rocksdb_block_based_options_create()}
}

// Destroy deallocates the BlockBasedTableOptions object.
func (o *BlockBasedTableOptions) Destroy() {
	//C.rocksdb_filterpolicy_destroy(o.fp)
	C.rocksdb_block_based_options_destroy(o.c)
	o.c = nil
	o.cache = nil
	o.cacheComp = nil
	o.fp = nil
}

// SetCacheIndexAndFilterBlocks is indicating if we'd put index/filter blocks to the block cache.
// If not specified, each "table reader" object will pre-load index/filter
// block during table initialization.
// Default: false
func (o *BlockBasedTableOptions) SetCacheIndexAndFilterBlocks(value bool) {
	C.rocksdb_block_based_options_set_cache_index_and_filter_blocks(o.c, boolToUchar(value))
}

// SetPinL0FilterAndIndexBlocksInCache sets cache_index_and_filter_blocks.
// If is true and the below is true (hash_index_allow_collision), then
// filter and index blocks are stored in the cache, but a reference is
// held in the "table reader" object so the blocks are pinned and only
// evicted from cache when the table reader is freed.
// Default: false
func (o *BlockBasedTableOptions) SetPinL0FilterAndIndexBlocksInCache(value bool) {
	C.rocksdb_block_based_options_set_pin_l0_filter_and_index_blocks_in_cache(o.c, boolToUchar(value))
}

// SetPinTopLevelIndexAndFilterInCache pins top-level indexes.
// Default: false
func (o *BlockBasedTableOptions) SetPinTopLevelIndexAndFilterInCache(value bool) {
	C.rocksdb_block_based_options_set_pin_top_level_index_and_filter(o.c, boolToUchar(value))
}

// SetCacheIndexAndFilterBlocksWithHighPriority priority to high for index and filter blocks
// in block cache. It only affect LRUCache so far, and need to use together with
// high_pri_pool_ratio when calling NewLRUCache(). If the feature is enabled, LRU-list in LRU
// cache will be split into two parts, one for high-pri blocks and one for low-pri blocks.
// Data blocks will be inserted to the head of low-pri pool. Index and filter blocks will be
// inserted to the head of high-pri pool. If the total usage in the high-pri pool exceed
// capacity * high_pri_pool_ratio, the block at the tail of high-pri pool will overflow to the
// head of low-pri pool, after which it will compete against data blocks to stay in cache.
// Eviction will start from the tail of low-pri pool.
func (o *BlockBasedTableOptions) SetCacheIndexAndFilterBlocksWithHighPriority(value bool) {
	C.rocksdb_block_based_options_set_cache_index_and_filter_blocks_with_high_priority(o.c, boolToUchar(value))
}

// SetHashIndexAllowCollision when enabled, prefix hash index for block-based table
// will not store prefix and allow hash collision, reducing memory consumption.
// Default false
func (o *BlockBasedTableOptions) SetHashIndexAllowCollision(value bool) {
	C.rocksdb_block_based_options_set_hash_index_allow_collision(o.c, boolToUchar(value))
}

// SetBlockSize sets the approximate size of user data packed per block.
// Note that the block size specified here corresponds to uncompressed data.
// The actual size of the unit read from disk may be smaller if
// compression is enabled. This parameter can be changed dynamically.
// Default: 4K
func (o *BlockBasedTableOptions) SetBlockSize(blockSize int) {
	C.rocksdb_block_based_options_set_block_size(o.c, C.size_t(blockSize))
}

// SetPartitionFilters enables partition index filters.
// With partitioning, the index/filter of a SST file is partitioned into smaller
// blocks with an additional top-level index on them. When reading an index/filter,
// only top-level index is loaded into memory. The partitioned index/filter then uses
// the top-level index to load on demand into the block cache the partitions that are
// required to perform the index/filter query. The top-level index, which has much smaller
// memory footprint, can be stored in heap or block cache depending on the
// SetCacheIndexAndFilterBlocks setting.
// Default: false
func (o *BlockBasedTableOptions) SetPartitionFilters(value bool) {
	C.rocksdb_block_based_options_set_partition_filters(o.c, boolToUchar(value))
}

// SetMetadataBlockSize sets the approximate size of the blocks for index partitions.
// Default: 4K
func (o *BlockBasedTableOptions) SetMetadataBlockSize(blockSize uint64) {
	C.rocksdb_block_based_options_set_metadata_block_size(o.c, C.uint64_t(blockSize))
}

// SetBlockSizeDeviation sets the block size deviation.
// This is used to close a block before it reaches the configured
// 'block_size'. If the percentage of free space in the current block is less
// than this specified number and adding a new record to the block will
// exceed the configured block size, then this block will be closed and the
// new record will be written to the next block.
// Default: 10
func (o *BlockBasedTableOptions) SetBlockSizeDeviation(blockSizeDeviation int) {
	C.rocksdb_block_based_options_set_block_size_deviation(o.c, C.int(blockSizeDeviation))
}

// SetFilterPolicy sets the filter policy to reduce disk reads.
// Many applications will benefit from passing the result of
// NewBloomFilterPolicy() here.
// Default: nil
func (o *BlockBasedTableOptions) SetFilterPolicy(fp *FilterPolicy) {
	C.rocksdb_block_based_options_set_filter_policy(o.c, fp.policy)
	o.fp = fp.policy
}

// SetNoBlockCache specify whether block cache should be used or not.
// Default: false
func (o *BlockBasedTableOptions) SetNoBlockCache(value bool) {
	C.rocksdb_block_based_options_set_no_block_cache(o.c, boolToUchar(value))
}

// SetBlockCache sets the control over blocks (user data is stored in a set of blocks, and
// a block is the unit of reading from disk).
//
// If set, use the specified cache for blocks.
// If nil, rocksdb will automatically create and use an 8MB internal cache.
// Default: nil
func (o *BlockBasedTableOptions) SetBlockCache(cache *Cache) {
	o.cache = cache
	C.rocksdb_block_based_options_set_block_cache(o.c, cache.c)
}

// SetBlockCacheCompressed sets the cache for compressed blocks.
// If nil, rocksdb will not use a compressed block cache.
// Default: nil
func (o *BlockBasedTableOptions) SetBlockCacheCompressed(cache *Cache) {
	o.cacheComp = cache
	C.rocksdb_block_based_options_set_block_cache_compressed(o.c, cache.c)
}

// SetWholeKeyFiltering specify if whole keys in the filter (not just prefixes)
// should be placed.
// This must generally be true for gets opts be efficient.
// Default: true
func (o *BlockBasedTableOptions) SetWholeKeyFiltering(value bool) {
	C.rocksdb_block_based_options_set_whole_key_filtering(o.c, boolToUchar(value))
}

// SetIndexType sets the index type used for this table.
// kBinarySearch:
// A space efficient index block that is optimized for
// binary-search-based index.
//
// kHashSearch:
// The hash index, if enabled, will do the hash lookup when
// `Options.prefix_extractor` is provided.
//
// kTwoLevelIndexSearch:
// A two-level index implementation. Both levels are binary search indexes.
// Default: kBinarySearch
func (o *BlockBasedTableOptions) SetIndexType(value IndexType) {
	C.rocksdb_block_based_options_set_index_type(o.c, C.int(value))
}
