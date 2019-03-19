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

	"github.com/bbva/qed/storage"
)

const keySize = 34

type SimpleCache struct {
	cached map[[keySize]byte][]byte
}

func NewSimpleCache(initialSize uint64) *SimpleCache {
	return &SimpleCache{make(map[[keySize]byte][]byte, initialSize)}
}

func (c SimpleCache) Get(key []byte) ([]byte, bool) {
	var k [keySize]byte
	copy(k[:], key)
	value, ok := c.cached[k]
	return value, ok
}

func (c *SimpleCache) Put(key []byte, value []byte) {
	var k [keySize]byte
	copy(k[:], key)
	c.cached[k] = value
}

func (c *SimpleCache) Fill(r storage.KVPairReader) (err error) {
	defer r.Close()
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
			}
		}
	}
	return nil
}

func (c SimpleCache) Size() int {
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
