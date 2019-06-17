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
	"fmt"

	"github.com/bbva/qed/storage"
	"github.com/bbva/qed/testutils/rand"
	"github.com/bbva/qed/util"
)

type FakeCache struct {
	FixedValue []byte
	cached     map[[keySize]byte][]byte
}

func NewFakeCache(fixedValue []byte) *FakeCache {
	return &FakeCache{
		FixedValue: fixedValue,
		cached:     make(map[[keySize]byte][]byte),
	}
}

func (c FakeCache) Get(key []byte) ([]byte, bool) {
	var k [keySize]byte
	copy(k[:], key)
	value, ok := c.cached[k]
	if !ok {
		return []byte{0x0}, true
	}
	return value, ok
}

func (c *FakeCache) Put(key []byte, value []byte) {
	var k [keySize]byte
	copy(k[:], key)
	c.cached[k] = value
}

func (c *FakeCache) Fill(r storage.KVPairReader) error {
	return fmt.Errorf("Operation not supported")
}

func (c FakeCache) Size() int {
	return len(c.cached)
}

func (c FakeCache) Equal(o *FakeCache) bool {
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

type FakeKVPairReader struct {
	Remaining uint64
	index     uint64
}

func NewFakeKVPairReader(numElems uint64) *FakeKVPairReader {
	return &FakeKVPairReader{numElems, 0}
}

func (r *FakeKVPairReader) Read(buffer []*storage.KVPair) (n int, err error) {
	for n = 0; r.Remaining > 0 && n < len(buffer); n++ {
		buffer[n] = &storage.KVPair{
			Key:   util.Uint64AsBytes(r.index),
			Value: rand.Bytes(8),
		}
		r.Remaining--
		r.index++
	}
	return n, nil
}
func (r *FakeKVPairReader) Close() {
	r.Remaining = 0
}
