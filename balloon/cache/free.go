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

package cache

import (
	"bytes"
	"runtime/debug"

	"github.com/bbva/qed/storage"
	"github.com/coocood/freecache"
)

type FreeCache struct {
	cached *freecache.Cache
}

// NewFreeCache funtion returns a new cache with a parametrized size.
func NewFreeCache(initialSize int) *FreeCache {
	cache := freecache.NewCache(initialSize)
	debug.SetGCPercent(20)
	return &FreeCache{cached: cache}
}

// Get function returns the value of a given key in cache, and a boolean showing if
// the key is or is not present.
func (c FreeCache) Get(key []byte) ([]byte, bool) {
	value, err := c.cached.Get(key)
	if err != nil {
		return nil, false
	}
	return value, true
}

// Put function adds a new key/value pair to the cache.
func (c *FreeCache) Put(key []byte, value []byte) {
	_ = c.cached.Set(key, value, 0)
}

// Fill function inserts a bulk of key/value elements into the cache.
func (c *FreeCache) Fill(r storage.KVPairReader) (err error) {
	defer r.Close()
	for {
		entries := make([]*storage.KVPair, 100)
		n, err := r.Read(entries)
		if err != nil || n == 0 {
			break
		}
		for _, entry := range entries {
			if entry != nil {
				_ = c.cached.Set(entry.Key, entry.Value, 0)
			}
		}
	}
	return nil
}

// Size function returns the number of items currently in the cache.
func (c FreeCache) Size() int {
	return int(c.cached.EntryCount())
}

// Equal function checks if every element from current cache (C) exists
// in the cache to compare (O). It does not check that every element from (O)
// exists in current cache (C).
func (c FreeCache) Equal(o *FreeCache) bool {
	it := c.cached.NewIterator()
	entry := it.Next()
	for entry != nil {
		v2, err := o.cached.Get(entry.Key)
		if err != nil {
			return false
		}
		if !bytes.Equal(entry.Value, v2) {
			return false
		}
		entry = it.Next()
	}
	return true
}
