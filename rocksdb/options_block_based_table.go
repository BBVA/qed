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

// BlockBasedTableOptions represents block-based table options.
type BlockBasedTableOptions struct {
	c *C.rocksdb_block_based_table_options_t

	// Hold references for GC.
	cache *Cache

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
	o.fp = nil
}

// SetCacheIndexAndFilterBlocks is indicating if we'd put index/filter blocks to the block cache.
// If not specified, each "table reader" object will pre-load index/filter
// block during table initialization.
// Default: false
func (o *BlockBasedTableOptions) SetCacheIndexAndFilterBlocks(value bool) {
	C.rocksdb_block_based_options_set_cache_index_and_filter_blocks(o.c, boolToUchar(value))
}

// SetBlockSize sets the approximate size of user data packed per block.
// Note that the block size specified here corresponds to uncompressed data.
// The actual size of the unit read from disk may be smaller if
// compression is enabled. This parameter can be changed dynamically.
// Default: 4K
func (o *BlockBasedTableOptions) SetBlockSize(blockSize int) {
	C.rocksdb_block_based_options_set_block_size(o.c, C.size_t(blockSize))
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
