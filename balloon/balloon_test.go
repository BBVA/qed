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

package balloon

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/bbva/qed/balloon/history"
	"github.com/bbva/qed/balloon/hyper"
	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/storage/badger"
	"github.com/bbva/qed/storage/bolt"
	"github.com/bbva/qed/storage/bplus"
	"github.com/bbva/qed/storage/cache"
	"github.com/bbva/qed/testutils/rand"
)

func TestAdd(t *testing.T) {

	frozen, frozenCloseF := openBPlusStorage()
	leaves, leavesCloseF := openBPlusStorage()
	defer frozenCloseF()
	defer leavesCloseF()

	cache := cache.NewSimpleCache(0)
	hasher := new(hashing.XorHasher)

	hyperT := hyper.NewFakeTree(string(0x0), cache, leaves, hasher)
	historyT := history.NewFakeTree(string(0x0), frozen, hasher)
	balloon := NewHyperBalloon(hasher, historyT, hyperT)

	var testCases = []struct {
		event         string
		hyperDigest   []byte
		historyDigest []byte
		version       uint64
	}{
		{"test event 0", []byte{0x0}, []byte{0x4a}, 0},
		{"test event 1", []byte{0x0}, []byte{0x01}, 1},
		{"test event 2", []byte{0x0}, []byte{0x4a}, 2},
		{"test event 3", []byte{0x0}, []byte{0x0}, 3},
		{"test event 4", []byte{0x0}, []byte{0x4a}, 4},
		{"test event 5", []byte{0x0}, []byte{0x0}, 5},
		{"test event 6", []byte{0x0}, []byte{0x4d}, 6},
		{"test event 7", []byte{0x0}, []byte{0x7}, 7},
		{"test event 8", []byte{0x0}, []byte{0x41}, 8},
		{"test event 9", []byte{0x0}, []byte{0x0b}, 9},
	}

	for i, e := range testCases {

		commitment := <-balloon.Add([]byte(e.event))

		if commitment.Version != e.version {
			t.Fatalf("Wrong version for test %d: expected %d, actual %d", i, e.version, commitment.Version)
		}

		if bytes.Compare(commitment.HyperDigest, e.hyperDigest) != 0 {
			t.Fatalf("Wrong index digest for test %d: expected: %x, Actual: %x", i, e.hyperDigest, commitment.HyperDigest)
		}

		if bytes.Compare(commitment.HistoryDigest, e.historyDigest) != 0 {
			t.Fatalf("Wrong history digest for test %d: expected: %x, Actual: %x", i, e.historyDigest, commitment.HistoryDigest)
		}
	}

}

//https://play.golang.org/p/nP241T7HXBj
// test event 0 : 4a [1001010] - 00 [0]
// test event 1 : 4b [1001011] - 01 [1]
// test event 2 : 48 [1001000] - 02 [10]
// test event 3 : 49 [1001001] - 03 [11]
// test event 4 : 4e [1001110] - 04 [100]
// test event 5 : 4f [1001111] - 05 [101]
// test event 6 : 4c [1001100] - 06 [110]
// test event 7 : 4d [1001101] - 07 [111]
// test event 8 : 42 [1000010] - 08 [1000]
// test event 9 : 43 [1000011] - 09 [1001]

func deleteFilesInDir(path string) {
	os.RemoveAll(fmt.Sprintf("%s/leaves.db", path))
	os.RemoveAll(fmt.Sprintf("%s/frozen.db", path))
}

func BenchmarkAddBolt(b *testing.B) {
	path := "/var/tmp/bench_balloon_add"
	os.MkdirAll(path, os.FileMode(0755))

	frozen, frozenCloseF := openBoltStorage(fmt.Sprintf("%s/frozen", path))
	leaves, leavesCloseF := openBoltStorage(fmt.Sprintf("%s/leaves", path))
	defer frozenCloseF()
	defer leavesCloseF()
	defer deleteFilesInDir(path)

	cache := cache.NewSimpleCache(1 << 25)
	hasher := new(hashing.Sha256Hasher)

	hyperT := hyper.NewTree(string(0x0), cache, leaves, hasher)
	historyT := history.NewTree(string(0x0), frozen, hasher)
	balloon := NewHyperBalloon(hasher, historyT, hyperT)

	b.ResetTimer()
	b.N = 10000
	for i := 0; i < b.N; i++ {
		event := rand.Bytes(128)
		r := balloon.Add(event)
		<-r
	}

}

func BenchmarkAddBadger(b *testing.B) {
	path := "/var/tmp/bench_balloon_add"
	os.MkdirAll(path, os.FileMode(0755))

	frozen, frozenCloseF := openBadgerStorage(fmt.Sprintf("%s/frozen", path))
	leaves, leavesCloseF := openBadgerStorage(fmt.Sprintf("%s/leaves", path))
	defer frozenCloseF()
	defer leavesCloseF()
	defer deleteFilesInDir(path)

	cache := cache.NewSimpleCache(1 << 25)
	hasher := new(hashing.Sha256Hasher)

	hyperT := hyper.NewTree(string(0x0), cache, leaves, hasher)
	historyT := history.NewTree(string(0x0), frozen, hasher)
	balloon := NewHyperBalloon(hasher, historyT, hyperT)

	b.ResetTimer()
	b.N = 10000
	for i := 0; i < b.N; i++ {
		event := rand.Bytes(128)
		r := balloon.Add(event)
		<-r
	}

}

func openBPlusStorage() (*bplus.BPlusTreeStorage, func()) {
	store := bplus.NewBPlusTreeStorage()
	return store, func() {
		store.Close()
	}
}

func openBoltStorage(path string) (*bolt.BoltStorage, func()) {
	store := bolt.NewBoltStorage(path, "test")
	return store, func() {
		store.Close()
		deleteFile(path)
	}
}

func openBadgerStorage(path string) (*badger.BadgerStorage, func()) {
	store := badger.NewBadgerStorage(path)
	return store, func() {
		store.Close()
		deleteFile(path)
	}
}

func deleteFile(path string) {
	err := os.RemoveAll(path)
	if err != nil {
		fmt.Printf("Unable to remove db file %s", err)
	}
}
