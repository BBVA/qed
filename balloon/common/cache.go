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

package common

import (
	"bytes"

	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/log"
	"github.com/bbva/qed/storage"
)

type Cache interface {
	Get(pos Position) (hashing.Digest, bool)
}

type ModifiableCache interface {
	Put(pos Position, value hashing.Digest)
	Fill(r storage.KVPairReader) error
	Size() int
	Cache
}
type PassThroughCache struct {
	prefix byte
	store  storage.Store
}

func NewPassThroughCache(prefix byte, store storage.Store) *PassThroughCache {
	return &PassThroughCache{prefix, store}
}

func (c PassThroughCache) Get(pos Position) (hashing.Digest, bool) {
	pair, err := c.store.Get(c.prefix, pos.Bytes())
	if err != nil {
		return nil, false
	}
	return pair.Value, true
}

const keySize = 34

type SimpleCache struct {
	cached map[[keySize]byte]hashing.Digest
}

func NewSimpleCache(initialSize uint64) *SimpleCache {
	return &SimpleCache{make(map[[keySize]byte]hashing.Digest, initialSize)}
}

func (c SimpleCache) Get(pos Position) (hashing.Digest, bool) {
	var key [keySize]byte
	copy(key[:], pos.Bytes())
	digest, ok := c.cached[key]
	return digest, ok
}

func (c *SimpleCache) Put(pos Position, value hashing.Digest) {
	var key [keySize]byte
	copy(key[:], pos.Bytes())
	c.cached[key] = value
}

func (c *SimpleCache) Fill(r storage.KVPairReader) (err error) {
	defer r.Close()
	log.Info("Warming up hyper cache...")
	cached := 0
	for {
		entries := make([]*storage.KVPair, 100)
		n, err := r.Read(entries)
		if err != nil || n == 0 {
			break
		}
		for _, entry := range entries {
			if entry != nil {
				var key [keySize]byte
				copy(key[:], entry.Key)
				c.cached[key] = entry.Value
				cached++
			}
		}
	}
	log.Infof("Warming up done, elements cached: %d", cached)
	return nil
}

func (c *SimpleCache) Size() int {
	return len(c.cached)
}

func (c SimpleCache) Equal(o *SimpleCache) bool {
	for k, v1 := range c.cached {
		v2, ok := o.cached[k]
		if !ok {
			return false
		}
		if !bytes.Equal(v1, v2) {
			return false
		}
	}
	return true
}
