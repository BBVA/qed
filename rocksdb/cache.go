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

// #include "rocksdb/c.h"
// #include "extended.h"
import "C"

// Cache is a cache used to store data read from data in memory.
type Cache struct {
	c *C.rocksdb_cache_t
}

// NewDefaultLRUCache create a new LRU cache with a fixed size capacity.
// num_shard_bits = -1 means it is automatically determined: every shard
// will be at least 512KB and number of shard bits will not exceed 6.
// strict_capacity_limit = false
// high_pri_pool_ration = 0.0
func NewDefaultLRUCache(capacity int) *Cache {
	return &Cache{
		c: C.rocksdb_cache_create_lru(C.size_t(capacity)),
	}
}

// NewLRUCache creates a new LRU cache with a fixed size capacity
// and high priority pool ration. The cache is sharded
// to 2^num_shard_bits shards, by hash of the key. The total capacity
// is divided and evenly assigned to each shard. If strict_capacity_limit
// is set, insert to the cache will fail when cache is full. User can also
// set percentage of the cache reserves for high priority entries via
// high_pri_pool_pct.
// num_shard_bits = -1 means it is automatically determined: every shard
// will be at least 512KB and number of shard bits will not exceed 6.
func NewLRUCache(capacity int, highPriorityPoolRatio float64) *Cache {
	return &Cache{
		c: C.rocksdb_cache_create_lru_with_ratio(
			C.size_t(capacity),
			C.double(highPriorityPoolRatio),
		),
	}
}

// GetUsage returns the Cache memory usage.
func (c *Cache) GetUsage() int {
	return int(C.rocksdb_cache_get_usage(c.c))
}

// GetPinnedUsage returns the Cache pinned memory usage.
func (c *Cache) GetPinnedUsage() int {
	return int(C.rocksdb_cache_get_pinned_usage(c.c))
}

// Destroy deallocates the Cache object.
func (c *Cache) Destroy() {
	C.rocksdb_cache_destroy(c.c)
	c.c = nil
}
