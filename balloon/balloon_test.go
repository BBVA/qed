package balloon

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"os"
	"testing"

	"qed/balloon/hashing"
	"qed/balloon/history"
	"qed/balloon/hyper"
	"qed/balloon/storage"
	"qed/balloon/storage/badger"
	"qed/balloon/storage/bolt"
	"qed/balloon/storage/bplus"
	"qed/balloon/storage/cache"
)

func TestAdd(t *testing.T) {

	frozen, frozenCloseF := openBPlusStorage()
	leaves, leavesCloseF := openBPlusStorage()
	defer frozenCloseF()
	defer leavesCloseF()

	cache := cache.NewSimpleCache(0)
	hasher := hashing.XorHasher

	hyperT := hyper.NewFakeTree(string(0x0), cache, leaves, hasher)
	historyT := history.NewFakeTree(frozen, hasher)
	balloon := NewHyperBalloon(hasher, historyT, hyperT)

	var testCases = []struct {
		event         string
		hyperDigest   []byte
		historyDigest []byte
		version       uint64
	}{
		{"test event 0", []byte{0x4a}, []byte{0x4a}, 0},
		{"test event 1", []byte{0x1}, []byte{0x00}, 1},
		{"test event 2", []byte{0x49}, []byte{0x48}, 2},
		{"test event 3", []byte{0x0}, []byte{0x01}, 3},
		{"test event 4", []byte{0x4e}, []byte{0x4e}, 4},
		{"test event 5", []byte{0x1}, []byte{0x01}, 5},
		{"test event 6", []byte{0x4d}, []byte{0x4c}, 6},
		{"test event 7", []byte{0x0}, []byte{0x01}, 7},
		{"test event 8", []byte{0x42}, []byte{0x43}, 8},
		{"test event 9", []byte{0x1}, []byte{0x00}, 9},
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

func TestGenMembershipProof(t *testing.T) {

	frozen, frozenCloseF := openBPlusStorage()
	leaves, leavesCloseF := openBPlusStorage()
	defer frozenCloseF()
	defer leavesCloseF()

	cache := cache.NewSimpleCache(0)
	hasher := hashing.XorHasher

	hyperT := hyper.NewTree(string(0x0), cache, leaves, hasher)
	historyT := history.NewFakeTree(frozen, hasher)
	balloon := NewHyperBalloon(hasher, historyT, hyperT)

	key := []byte{0x5a}
	var version uint64 = 0
	expectedHyperProof := [][]byte{
		[]byte{0x00},
		[]byte{0x00},
		[]byte{0x00},
		[]byte{0x00},
		[]byte{0x00},
		[]byte{0x00},
		[]byte{0x00},
		[]byte{0x00},
	}
	expectedHistoryProof := [][]byte{}

	<-balloon.Add(key)

	proof := <-balloon.GenMembershipProof(key, version)

	if !proof.Exists {
		t.Fatalf("Wrong proof: the event should exist")
	}

	if proof.QueryVersion != version {
		t.Fatalf("The query version does not match: expected %d, actual %d", version, proof.QueryVersion)
	}

	if proof.ActualVersion != version {
		t.Fatalf("The actual version does not match: expected %d, actual %d", version, proof.ActualVersion)
	}

	if !compareHyperProofs(expectedHyperProof, proof.HyperProof) {
		t.Fatalf("Wrong hyper proof: expected %v, actual %v", expectedHyperProof, proof.HyperProof)
	}

	if !compareHistoryProofs(expectedHistoryProof, proof.HistoryProof) {
		t.Fatalf("Wrong history proof: expected %v, actual %v", expectedHistoryProof, proof.HistoryProof)
	}

}

func compareHyperProofs(expected, actual [][]byte) bool {
	for i, e := range expected {
		if !bytes.Equal(e, actual[i]) {
			return false
		}
	}
	return true
}

func compareHistoryProofs(expected [][]byte, actual []history.Node) bool {
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

func randomBytes(n int) []byte {
	bytes := make([]byte, n)
	_, err := rand.Read(bytes)
	if err != nil {
		panic(err)
	}

	return bytes
}

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

	cache := cache.NewSimpleCache(storage.SIZE25)
	hasher := hashing.Sha256Hasher

	hyperT := hyper.NewTree(string(0x0), cache, leaves, hasher)
	historyT := history.NewTree(frozen, hasher)
	balloon := NewHyperBalloon(hasher, historyT, hyperT)

	b.ResetTimer()
	b.N = 10000
	for i := 0; i < b.N; i++ {
		event := randomBytes(128)
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

	cache := cache.NewSimpleCache(storage.SIZE25)
	hasher := hashing.Sha256Hasher

	hyperT := hyper.NewTree(string(0x0), cache, leaves, hasher)
	historyT := history.NewTree(frozen, hasher)
	balloon := NewHyperBalloon(hasher, historyT, hyperT)

	b.ResetTimer()
	b.N = 10000
	for i := 0; i < b.N; i++ {
		event := randomBytes(128)
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
