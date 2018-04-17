// Copyright © 2018 Banco Bilbao Vizcaya Argentaria S.A.  All rights reserved.
// Use of this source code is governed by an Apache 2 License
// that can be found in the LICENSE file

package history

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"os"
	"testing"
	"verifiabledata/balloon/hashing"
	"verifiabledata/balloon/storage/badger"
	"verifiabledata/balloon/storage/bplus"
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
	}

	store, closeF := openBPlusStorage()
	defer closeF()

	ht := NewTree(store, fakeLeafHasherF(hashing.XorHasher), fakeInteriorHasherF(hashing.XorHasher))

	for _, e := range testCases {
		t.Logf("Testing event: %b", e.event)
		commitment := <-ht.Add(e.event, uInt64AsBytes(e.index))

		if !bytes.Equal(e.commitment, commitment) {
			t.Fatalf("Incorrect commitment: expected %b, actual %b", e.commitment, commitment)
		}
	}
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
	store, closeF := openBadgerStorage()
	defer closeF()
	ht := NewTree(store, LeafHasherF(hashing.Sha256Hasher), InteriorHasherF(hashing.Sha256Hasher))
	b.N = 100000
	for i := 0; i < b.N; i++ {
		key := randomBytes(64)
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
	store := badger.NewBadgerStorage("/tmp/history_store_test.db")
	return store, func() {
		fmt.Println("Cleaning...")
		store.Close()
		deleteFile("/tmp/history_store_test.db")
	}
}

func deleteFile(path string) {
	err := os.RemoveAll(path)
	if err != nil {
		fmt.Printf("Unable to remove db file %s", err)
	}
}
