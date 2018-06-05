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

package history

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/storage/badger"
	"github.com/bbva/qed/storage/bplus"
	"github.com/bbva/qed/testutils/rand"
)

func TestAdd(t *testing.T) {
	var testCases = []struct {
		index      uint64
		commitment []byte
		event      []byte
	}{
		{0, []byte{0x4a}, []byte{0x4a}},
		{1, []byte{0x00}, []byte{0x4b}},
		{2, []byte{0x48}, []byte{0x48}},
		{3, []byte{0x01}, []byte{0x49}},
		{4, []byte{0x4e}, []byte{0x4e}},
		{5, []byte{0x01}, []byte{0x4f}},
		{6, []byte{0x4c}, []byte{0x4c}},
		{7, []byte{0x01}, []byte{0x4d}},
		{8, []byte{0x43}, []byte{0x42}},
		{9, []byte{0x00}, []byte{0x43}},
	}

	store, closeF := openBPlusStorage()
	defer closeF()

	ht := NewFakeTree(store, hashing.XorHasher)

	for i, e := range testCases {
		commitment := <-ht.Add(e.event, uInt64AsBytes(e.index))

		if !bytes.Equal(e.commitment, commitment) {
			t.Fatalf("Incorrect commitment for test %d: expected %x, actual %x", i, e.commitment, commitment)
		}
	}
}

func TestProveMembership(t *testing.T) {
	store, closeF := openBPlusStorage()
	defer closeF()

	var testCases = []struct {
		index      uint64
		commitment []byte
		event      []byte
	}{
		{0, []byte{0x4a}, []byte{0x4a}}, // 74
		{1, []byte{0x00}, []byte{0x4b}}, // 75
		{2, []byte{0x48}, []byte{0x48}}, // 72
		{3, []byte{0x01}, []byte{0x49}}, // 73
		{4, []byte{0x01}, []byte{0x50}}, // 80
		{5, []byte{0x01}, []byte{0x51}}, // 81
		{6, []byte{0x01}, []byte{0x52}}, // 82
	}

	ht := NewFakeTree(store, hashing.XorHasher)

	for _, e := range testCases {
		<-ht.Add(e.event, uInt64AsBytes(e.index))
	}

	expectedPath := [][]byte{
		{0x01},
		{0x00},
	}
	proof := <-ht.ProveMembership([]byte{0x5}, 6, 6)

	if !comparePaths(expectedPath, proof.Nodes) {
		t.Fatalf("Invalid path: expected %v, actual %v", expectedPath, proof.Nodes)
	}
}

func comparePaths(expected [][]byte, actual []Node) bool {
	if len(expected) != len(actual) {
		return false
	}

	for i, e := range expected {
		if !bytes.Equal(e, actual[i].Digest) {
			return false
		}
	}
	return true
}

func BenchmarkAdd(b *testing.B) {
	store, closeF := openBadgerStorage()
	defer closeF()
	ht := NewTree(store, hashing.Sha256Hasher)
	b.N = 100000
	for i := 0; i < b.N; i++ {
		key := rand.Bytes(64)
		<-ht.Add(key, uInt64AsBytes(uint64(i)))
	}
	b.Logf("stats = %+v\n", ht.stats)
}

func openBPlusStorage() (*bplus.BPlusTreeStorage, func()) {
	store := bplus.NewBPlusTreeStorage()
	return store, func() {
		store.Close()
	}
}

func openBadgerStorage() (*badger.BadgerStorage, func()) {
	store := badger.NewBadgerStorage("/var/tmp/history_store_test.db")
	return store, func() {
		fmt.Println("Cleaning...")
		store.Close()
		deleteFile("/var/tmp/history_store_test.db")
	}
}

func deleteFile(path string) {
	err := os.RemoveAll(path)
	if err != nil {
		fmt.Printf("Unable to remove db file %s", err)
	}
}
