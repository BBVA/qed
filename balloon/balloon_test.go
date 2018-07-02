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
	"github.com/bbva/qed/balloon/proof"
	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/storage/cache"
	"github.com/bbva/qed/testutils/rand"
	"github.com/bbva/qed/testutils/storage"
)

func TestAdd(t *testing.T) {

	frozen, frozenCloseF := storage.NewBPlusStorage()
	leaves, leavesCloseF := storage.NewBPlusStorage()
	defer frozenCloseF()
	defer leavesCloseF()

	cache := cache.NewSimpleCache(0)
	hasher := new(hashing.XorHasher)

	hyperT := hyper.NewFakeTree(string(0x00), cache, leaves, hasher)
	historyT := history.NewFakeTree(string(0x00), frozen, hasher)
	balloon := NewHyperBalloon(hasher, historyT, hyperT)

	var testCases = []struct {
		event         string
		hyperDigest   []byte
		historyDigest []byte
		version       uint64
	}{
		{"test event 0", []byte{0x4a}, []byte{0x4a}, 0},
		{"test event 1", []byte{0x01}, []byte{0x01}, 1},
		{"test event 2", []byte{0x4b}, []byte{0x4a}, 2},
		{"test event 3", []byte{0x00}, []byte{0x00}, 3},
		{"test event 4", []byte{0x4a}, []byte{0x4a}, 4},
		{"test event 5", []byte{0x01}, []byte{0x00}, 5},
		{"test event 6", []byte{0x4b}, []byte{0x4d}, 6},
		{"test event 7", []byte{0x00}, []byte{0x07}, 7},
		{"test event 8", []byte{0x4a}, []byte{0x41}, 8},
		{"test event 9", []byte{0x01}, []byte{0x0b}, 9},
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

	t.Skip("TODO: Decide wether this snapshotting tests are required…")
	frozen, frozenCloseF := storage.NewBPlusStorage()
	leaves, leavesCloseF := storage.NewBPlusStorage()
	defer frozenCloseF()
	defer leavesCloseF()

	cache := cache.NewSimpleCache(0)
	hasher := new(hashing.XorHasher)

	hyperT := hyper.NewFakeTree(string(0x0), cache, leaves, hasher)
	historyT := history.NewFakeTree(string(0x0), frozen, hasher)
	balloon := NewHyperBalloon(hasher, historyT, hyperT)

	key := []byte{0x5a}
	var version uint64 = 0
	expectedHyperAuditPath := map[string][]byte{
		"50|3": {0x00},
		"40|4": {0x00},
		"5a|0": {0x00},
		"80|7": {0x00},
		"00|6": {0x00},
		"5c|2": {0x00},
		"58|1": {0x00},
		"5b|0": {0x00},
	}
	expectedHistoryAuditPath := map[string][]byte{
		"0|0": {0x5a},
	}

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

	if !compareAuditPaths(expectedHyperAuditPath, proof.HyperProof.AuditPath()) {
		t.Fatalf("Wrong hyper audit path: expected %v, actual %v", expectedHyperAuditPath, proof.HyperProof.AuditPath())
	}

	if !compareAuditPaths(expectedHistoryAuditPath, proof.HistoryProof.AuditPath()) {
		t.Fatalf("Wrong history audit path: expected %v, actual %v", expectedHistoryAuditPath, proof.HistoryProof.AuditPath())
	}

}

func compareAuditPaths(expected, actual proof.AuditPath) bool {
	if len(expected) != len(actual) {
		return false
	}

	for k, v := range expected {
		if !bytes.Equal(v, actual[k]) {
			return false
		}
	}
	return true
}

type FakeVerifiable struct {
	result bool
}

func NewFakeVerifiable(result bool) *FakeVerifiable {
	return &FakeVerifiable{result}
}

func (f FakeVerifiable) Verify(commitment, event, version []byte) bool {
	return f.result
}

func (f FakeVerifiable) AuditPath() proof.AuditPath {
	return make(map[string][]byte)
}

func TestVerify(t *testing.T) {

	testCases := []struct {
		exists         bool
		hyperOK        bool
		historyOK      bool
		currentVersion uint64
		queryVersion   uint64
		actualVersion  uint64
		expectedResult bool
	}{
		// Event exists, queryVersion <= actualVersion, and both trees verify it
		{true, true, true, uint64(0), uint64(0), uint64(0), true},
		// Event exists, queryVersion <= actualVersion, but HyperTree does not verify it
		{true, false, true, uint64(0), uint64(0), uint64(0), false},
		// Event exists, queryVersion <= actualVersion, but HistoryTree does not verify it
		{true, true, false, uint64(0), uint64(0), uint64(0), false},

		// Event exists, queryVersion > actualVersion, and both trees verify it
		{true, true, true, uint64(1), uint64(1), uint64(0), true},
		// Event exists, queryVersion > actualVersion, but HyperTree does not verify it
		{true, false, true, uint64(1), uint64(1), uint64(0), false},

		// Event does not exist, HyperTree verifies it
		{false, true, false, uint64(0), uint64(0), uint64(0), true},
		// Event does not exist, HyperTree does not verify it
		{false, false, false, uint64(0), uint64(0), uint64(0), false},
	}

	for i, c := range testCases {
		event := []byte("Yadda yadda")
		commitment := &Commitment{
			[]byte("Some hyperDigest"),
			[]byte("Some historyDigest"),
			c.actualVersion,
		}
		proof := NewMembershipProof(
			c.exists,
			NewFakeVerifiable(c.hyperOK),
			NewFakeVerifiable(c.historyOK),
			c.currentVersion,
			c.queryVersion,
			c.actualVersion,
			event,
			new(hashing.Sha256Hasher),
		)
		result := proof.Verify(commitment, event)

		if result != c.expectedResult {
			t.Fatalf("Unexpected result '%v' in test case '%d'", result, i)
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

	frozen, frozenCloseF := storage.NewBoltStorage(fmt.Sprintf("%s/frozen", path))
	leaves, leavesCloseF := storage.NewBoltStorage(fmt.Sprintf("%s/leaves", path))
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

	frozen, frozenCloseF := storage.NewBadgerStorage(fmt.Sprintf("%s/frozen", path))
	leaves, leavesCloseF := storage.NewBadgerStorage(fmt.Sprintf("%s/leaves", path))
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
