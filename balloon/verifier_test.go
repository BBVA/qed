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
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/bbva/qed/balloon/hashing"
	"github.com/bbva/qed/balloon/history"
	"github.com/bbva/qed/balloon/hyper"
	"github.com/bbva/qed/balloon/storage"
	"github.com/bbva/qed/balloon/storage/cache"
	"github.com/bbva/qed/log"
)

type FakeVerifiable struct {
	result bool
}

func NewFakeVerifiable(result bool) *FakeVerifiable {
	return &FakeVerifiable{result}
}

func (f *FakeVerifiable) Verify(commitment, event []byte, version uint64) bool {
	return f.result
}

func TestVerify(t *testing.T) {

	testCases := []struct {
		exists         bool
		hyperOK        bool
		historyOK      bool
		queryVersion   uint64
		actualVersion  uint64
		expectedResult bool
	}{
		// Event exists, queryVersion <= actualVersion, and both trees verify it
		{true, true, true, uint64(0), uint64(0), true},
		// Event exists, queryVersion <= actualVersion, but HyperTree does not verify it
		{true, false, true, uint64(0), uint64(0), false},
		// Event exists, queryVersion <= actualVersion, but HistoryTree does not verify it
		{true, true, false, uint64(0), uint64(0), false},

		// Event exists, queryVersion > actualVersion, and both trees verify it
		{true, true, true, uint64(1), uint64(0), true},
		// Event exists, queryVersion > actualVersion, but HyperTree does not verify it
		{true, false, true, uint64(1), uint64(0), false},

		// Event does not exist, HyperTree verifies it
		{false, true, false, uint64(0), uint64(0), true},
		// Event does not exist, HyperTree does not verify it
		{false, false, false, uint64(0), uint64(0), false},
	}

	for i, c := range testCases {
		event := []byte("Yadda yadda")
		commitment := &Commitment{
			[]byte("Some hyperDigest"),
			[]byte("Some historyDigest"),
			c.actualVersion,
		}
		proof := NewProof(
			c.exists,
			NewFakeVerifiable(c.hyperOK),
			NewFakeVerifiable(c.historyOK),
			c.queryVersion,
			c.actualVersion,
			hashing.XorHasher,
		)
		result := proof.Verify(commitment, event)

		if result != c.expectedResult {
			t.Fatalf("Unexpected result '%v' in test case '%d'", result, i)
		}
	}
}

func createBalloon(id string, hasher hashing.Hasher) (*HyperBalloon, storage.DeletableStore, func()) {
	dir, err := ioutil.TempDir("/var/tmp/", "balloon.test")
	if err != nil {
		log.Fatal(err)
	}

	frozen, frozenCloseF := openBadgerStorage(fmt.Sprintf("%s/frozen", dir))
	leaves, leavesCloseF := openBadgerStorage(fmt.Sprintf("%s/leaves", dir))
	cache := cache.NewSimpleCache(storage.SIZE20)

	hyperT := hyper.NewTree(id, cache, leaves, hasher)
	historyT := history.NewTree(frozen, hasher)
	balloon := NewHyperBalloon(hasher, historyT, hyperT)

	return balloon, leaves, func() {
		frozenCloseF()
		leavesCloseF()
		os.RemoveAll(dir)
	}
}

func TestAddAndVerify(t *testing.T) {

	id := string(0x0)
	hasher := hashing.Sha256Hasher

	balloon, _, closeF := createBalloon(id, hasher)
	defer closeF()

	key := []byte("Never knows best")
	// keyDigest := hasher(key)

	commitment := <-balloon.Add(key)
	membershipProof := <-balloon.GenMembershipProof(key, commitment.Version)

	historyProof := history.NewProof(membershipProof.HistoryProof, commitment.Version, hasher)
	hyperProof := hyper.NewProof(id, membershipProof.HyperProof, hasher)

	proof := NewProof(
		membershipProof.Exists,
		hyperProof,
		historyProof,
		membershipProof.QueryVersion,
		membershipProof.ActualVersion,
		hasher,
	)

	correct := proof.Verify(commitment, key)

	if !correct {
		t.Errorf("Proof is incorrect")
	}

}

func TestTamperAndVerify(t *testing.T) {

	t.Skip("WIP")
	log.SetLogger("test", "info")

	id := string(0x0)
	hasher := hashing.Sha256Hasher

	balloon, store, closeF := createBalloon(id, hasher)
	defer closeF()

	key := []byte("Never knows best")

	commitment := <-balloon.Add(key)
	membershipProof := <-balloon.GenMembershipProof(key, commitment.Version)
	// log.Info(commitment.Version)

	historyProof := history.NewProof(membershipProof.HistoryProof, commitment.Version, hasher)
	hyperProof := hyper.NewProof(id, membershipProof.HyperProof, hasher)

	proof := NewProof(
		membershipProof.Exists,
		hyperProof,
		historyProof,
		membershipProof.QueryVersion,
		membershipProof.ActualVersion,
		hasher,
	)

	correct := proof.Verify(commitment, key)
	if !correct {
		t.Errorf("Proof is incorrect")
	}

	original, _ := store.Get(key)
	log.Info(original)

	tamperVal := ^uint64(0) // max uint ftw!
	tpBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(tpBytes, tamperVal)

	err := store.Add(key, tpBytes)
	if err != nil {
		t.Fatal("store add returned not nil value")
	}
	tampered, _ := store.Get(key)
	log.Info(tampered)
	if bytes.Compare(tpBytes, tampered) != 0 {
		t.Fatal("Tamper unsuccesfull")
	}

	if proof.Verify(commitment, key) {
		t.Errorf("TamperProof unsucessful")
	}

}

func TestDeleteAndVerify(t *testing.T) {

	t.Skip("WIP")
	log.SetLogger("test", "info")

	id := string(0x0)
	hasher := hashing.Sha256Hasher

	balloon, store, closeF := createBalloon(id, hasher)
	defer closeF()

	key := []byte("Never knows best")
	// keyDigest := hasher(key)

	<-balloon.Add([]byte("Starting balloon"))
	<-balloon.Add([]byte("Starting balloon"))
	<-balloon.Add([]byte("Starting balloon"))
	<-balloon.Add([]byte("Starting balloon"))
	<-balloon.Add([]byte("Starting balloon"))
	<-balloon.Add([]byte("Starting balloon"))
	commitment := <-balloon.Add(key)
	membershipProof := <-balloon.GenMembershipProof(key, commitment.Version)

	historyProof := history.NewProof(membershipProof.HistoryProof, commitment.Version, hasher)
	hyperProof := hyper.NewProof(id, membershipProof.HyperProof, hasher)

	proof := NewProof(
		membershipProof.Exists,
		hyperProof,
		historyProof,
		membershipProof.QueryVersion,
		membershipProof.ActualVersion,
		hasher,
	)

	if !proof.Verify(commitment, key) {
		t.Errorf("Proof is incorrect")
	}

	err := store.Delete(key)
	if err != nil {
		t.Fatal("store.Delete returned not nil value")
	}

	tampered, _ := store.Get(key)
	log.Info(tampered)
	if tampered != nil {
		t.Fatal("Tamper unsuccesfull")
	}

	if proof.Verify(commitment, key) {
		t.Errorf("TamperProof unsuccessfull")
	}

}
