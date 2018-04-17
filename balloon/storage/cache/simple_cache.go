// Copyright © 2018 Banco Bilbao Vizcaya Argentaria S.A.  All rights reserved.
// Use of this source code is governed by an Apache 2 License
// that can be found in the LICENSE file

package cache

const keySize = 36
// SimpleCache is a cache that contains the hashes of the pre-computed nodes
type SimpleCache struct {
	nodes map[[keySize]byte][]byte // node map containing the cached hashes
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

// NewSimpleCache creates a new cache structure, already initialized
// with a specified size
func NewSimpleCache(size int) *SimpleCache {
	return &SimpleCache{make(map[[keySize]byte][]byte, size)}
}

