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

package hyper

import (
	"bytes"
	"crypto/rand"
	"testing"

	"github.com/BBVA/qed/balloon/hashing"
	"github.com/BBVA/qed/balloon/storage"
	"github.com/BBVA/qed/balloon/storage/cache"
)

func TestAdd(t *testing.T) {
	store, closeF := openBPlusStorage()
	defer closeF()

	cache := cache.NewSimpleCache(0)
	hasher := hashing.XorHasher
	ht := NewTree(string(0x0), cache, store, hasher)

	var testCases = []struct {
		key        []byte
		commitment []byte
	}{
		{[]byte{0x00}, []byte{0x0}},
		{[]byte{0x01}, []byte{0x0}},
		{[]byte{0x2}, []byte{0x0}},
		{[]byte{0x3}, []byte{0x0}},
		{[]byte{0x4}, []byte{0x0}},
		{[]byte{0x5}, []byte{0x0}},
		{[]byte{0x6}, []byte{0x0}},
		{[]byte{0x7}, []byte{0x0}},
		{[]byte{0x8}, []byte{0x0}},
		{[]byte{0x9}, []byte{0x06}},
	}
	value := []byte{0x01}

	for i, e := range testCases {
		commitment := <-ht.Add(e.key, value)

		if bytes.Compare(commitment, e.commitment) != 0 {
			t.Fatalf("Expected commitment for test %d: %x, Actual: %x", i, e.commitment, commitment)
		}
	}

}

func TestProve(t *testing.T) {

	store, closeF := openBPlusStorage()
	defer closeF()

	cache := cache.NewSimpleCache(0)
	hasher := hashing.XorHasher

	ht := NewTree(string(0x0), cache, store, hasher)

	key := []byte{0x5a}
	value := []byte{0x01}

	<-ht.Add(key, value)

	expectedPath := [][]byte{
		[]byte{0x00},
		[]byte{0x00},
		[]byte{0x00},
		[]byte{0x00},
		[]byte{0x00},
		[]byte{0x00},
		[]byte{0x00},
		[]byte{0x00},
	}
	proof := <-ht.Prove(key)

	if !comparePaths(expectedPath, proof.AuditPath) {
		t.Fatalf("Invalid path: expected %v, actual %v", expectedPath, proof.AuditPath)
	}

}

func comparePaths(expected, actual [][]byte) bool {
	for i, e := range expected {
		if !bytes.Equal(e, actual[i]) {
			return false
		}
	}
	return true
}

func randomBytes(n int) []byte {
	bytes := make([]byte, n)
	_, err := rand.Read(bytes)
	if err != nil {
		panic(err)
	}

	return bytes
}

func BenchmarkAdd(b *testing.B) {
	store, closeF := openBadgerStorage("/var/tmp/hyper_tree_test.db") //openBoltStorage()
	defer closeF()

	cache := cache.NewSimpleCache(storage.SIZE25)
	hasher := hashing.Sha256Hasher
	ht := NewTree(string(0x0), cache, store, hasher)

	b.N = 1000000
	for i := 0; i < b.N; i++ {
		key := hashing.Sha256Hasher(randomBytes(32))
		value := randomBytes(1)
		store.Add(key, value)
		<-ht.Add(key, value)
	}
	b.Logf("stats = %+v\n", ht.stats)
}
