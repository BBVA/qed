// Copyright 2018 BBVA. All rights reserved.
// Use of this source code is governed by a Apache 2 License
// that can be found in the LICENSE file

package history

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"testing"
	"verifiabledata/balloon/hashing"
	"verifiabledata/balloon/storage"
	"verifiabledata/util"
)

func TestAdd(t *testing.T) {
	var testCases = []struct {
		index      uint64
		commitment string
		event      string
	}{
		{0, "b4aa0a376986b4ab072ed536d41a4df65de5d46da15ff8756bc7657da01d2f52", "Hello World1"},
		{1, "8c93bb614746a51c1200f01a0ba5217686bf576b8bb0b095523ea38c740c567e", "Hello World2"},
		{2, "ef88d8ccbf2d620e83066e16be4e2f31db0a10713d5970da2a15dc57b64d760a", "Hello World3"},
		{3, "4cc083de4b21da14ca6d216293037b23c363580592b35eea724d5426b1dbd0ee", "Hello World4"},
		{4, "52b0222c6c43792823cfe719548a6f1b6ff01a5b2c8c08b3d9480d7b76a96d0f", "Hello World5"},
	}

	store, closeF := openStorage()
	defer closeF()

	ht := NewTree(store, hashing.Sha256Hasher)

	for _, e := range testCases {
		t.Log("Testing event: ", e.event)
		commitment := <- ht.Add([]byte(e.event), util.UInt64AsBytes(e.index))

		c := hex.EncodeToString(commitment)

		if e.commitment != c {
			t.Fatal("Incorrect commitment: ", e.commitment, " != ", c)
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
	store, closeF := openStorage()
	defer closeF()
	ht := NewTree(store, hashing.Sha256Hasher)
	b.N = 100000
	for i := 0; i < b.N; i++ {
		key := randomBytes(64)
		<- ht.Add(key, util.UInt64AsBytes(uint64(i)))
	}
	b.Logf("stats = %+v\n", ht.stats)
}

func openStorage() (*storage.BadgerStorage, func()) {
	store := storage.NewBadgerStorage("/tmp/badger_store_test.db")
	return store, func() {
		fmt.Println("Cleaning...")
		store.Close()
		deleteFile("/tmp/badger_store_test.db")
	}
}

func deleteFile(path string) {
	err := os.RemoveAll(path)
	if err != nil {
		fmt.Printf("Unable to remove db file %s", err)
	}
}
