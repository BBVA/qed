/*
   Copyright 2018 Banco Bilbao Vizcaya Argentaria, S.A.

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
	"github.com/bbva/qed/balloon/navigator"
	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/storage"
)

const fifoKeySize = 10

type PassThroughCache struct {
	prefix byte
	store  storage.Store
}

func NewPassThroughCache(prefix byte, store storage.Store) *PassThroughCache {
	return &PassThroughCache{prefix, store}
}

func (c PassThroughCache) Get(pos navigator.Position) (hashing.Digest, bool) {
	pair, err := c.store.Get(c.prefix, pos.Bytes())
	if err != nil {
		return nil, false
	}
	return pair.Value, true
}

type Entry struct {
	key   [fifoKeySize]byte
	value hashing.Digest
}

type FIFOReadThroughCache struct {
	prefix        byte
	store         storage.Store
	cacheSize     uint8
	cached        map[[fifoKeySize]byte]uint8
	cachedEntries []*Entry
	cacheIndex    uint8
}

func NewFIFOReadThroughCache(prefix byte, store storage.Store, cacheSize uint8) *FIFOReadThroughCache {
	return &FIFOReadThroughCache{
		prefix:        prefix,
		store:         store,
		cacheSize:     cacheSize,
		cached:        make(map[[fifoKeySize]byte]uint8, 0),
		cachedEntries: make([]*Entry, cacheSize),
		cacheIndex:    0,
	}
}

func (c FIFOReadThroughCache) Get(pos navigator.Position) (hashing.Digest, bool) {
	var key [fifoKeySize]byte
	copy(key[:], pos.Bytes())
	index, ok := c.cached[key]
	if !ok {
		pair, err := c.store.Get(c.prefix, pos.Bytes())
		if err != nil {
			return nil, false
		}
		return pair.Value, true
	}
	return c.cachedEntries[index].value, ok
}

func (c *FIFOReadThroughCache) Put(pos navigator.Position, value hashing.Digest) {
	var key [fifoKeySize]byte
	copy(key[:], pos.Bytes())
	// evict entry
	evicted := c.cachedEntries[c.cacheIndex]
	if evicted != nil {
		delete(c.cached, evicted.key)
	}
	// insert new
	c.cached[key] = c.cacheIndex
	c.cachedEntries[c.cacheIndex] = &Entry{key, value}
	// new index
	c.cacheIndex = (c.cacheIndex + 1) % c.cacheSize
}

func (c *FIFOReadThroughCache) Fill(r storage.KVPairReader) (err error) {
	defer r.Close()
	for {
		entries := make([]*storage.KVPair, 100)
		n, err := r.Read(entries)
		if err != nil || n == 0 {
			break
		}
		for _, entry := range entries {
			if entry != nil {
				var key [fifoKeySize]byte
				copy(key[:], entry.Key)
				// evict entry
				evicted := c.cachedEntries[c.cacheIndex]
				if evicted != nil {
					delete(c.cached, evicted.key)
				}
				// insert new
				c.cached[key] = c.cacheIndex
				c.cachedEntries[c.cacheIndex] = &Entry{key, entry.Value}
				// new index
				c.cacheIndex = (c.cacheIndex + 1) % c.cacheSize
			}
		}
	}
	return nil
}

func (c FIFOReadThroughCache) Size() int {
	return len(c.cached)
}
