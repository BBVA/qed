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

// Package cache implements the cache used by the hyper tree to allow
// in-memory storage for the top N levels.
package cache

// keySize == len( H(x) ) + pos.heightBytes
const keySize = 36

// SimpleCache is a cache that contains the hashes of the pre-computed nodes
type SimpleCache struct {
	nodes map[[keySize]byte][]byte // node map containing the cached hashes
	size  uint64
}

func (c *SimpleCache) Put(key []byte, value []byte) error {
	var aux [keySize]byte
	copy(aux[:], key)
	c.nodes[aux] = value
	return nil
}

func (c *SimpleCache) Get(key []byte) ([]byte, bool) {
	var aux [keySize]byte
	copy(aux[:], key)
	result, ok := c.nodes[aux]
	return result, ok
}

func (c *SimpleCache) Exists(key []byte) bool {
	var aux [keySize]byte
	copy(aux[:], key)
	_, ok := c.nodes[aux]
	return ok
}

func (c *SimpleCache) Size() uint64 {
	return c.size
}

// NewSimpleCache creates a new cache structure, already initialized
// with a specified size
func NewSimpleCache(size uint64) *SimpleCache {
	return &SimpleCache{make(map[[keySize]byte][]byte, size), size}
}
