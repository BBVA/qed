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

package integration

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/bbva/qed/balloon"
	"github.com/bbva/qed/testutils/rand"
	"github.com/bbva/qed/testutils/storage"
	"github.com/stretchr/testify/assert"

	"github.com/bbva/qed/balloon/history"
	"github.com/bbva/qed/balloon/hyper"
	"github.com/bbva/qed/balloon/proof"
	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/storage/badger"
	"github.com/bbva/qed/storage/cache"
)

func TestAddAndVerify(t *testing.T) {

	hasher := new(hashing.Sha256Hasher)
	b, _, closeF := createBalloon("treeId", hasher)
	defer closeF()

	key := []byte("Never knows best")
	keyDigest := hasher.Do(key)

	commitment := <-b.Add(key)
	membershipProof := <-b.GenMembershipProof(key, commitment.Version)

	proof := balloon.NewMembershipProof(
		membershipProof.Exists,
		membershipProof.HyperProof,
		membershipProof.HistoryProof,
		membershipProof.CurrentVersion,
		membershipProof.QueryVersion,
		membershipProof.ActualVersion,
		keyDigest,
		hasher,
	)

	correct := proof.Verify(commitment, key)

	if !correct {
		t.Errorf("Proof is incorrect")
	}

}

func TestTamperAndVerify(t *testing.T) {

	hasher := new(hashing.Sha256Hasher)
	b, store, closeF := createBalloon("treeId", hasher)
	defer closeF()

	key := []byte("Never knows best")
	keyDigest := hasher.Do(key)

	commitment := <-b.Add(key)
	membershipProof := <-b.GenMembershipProof(key, commitment.Version)

	memProof := balloon.NewMembershipProof(
		membershipProof.Exists,
		membershipProof.HyperProof,
		membershipProof.HistoryProof,
		membershipProof.CurrentVersion,
		membershipProof.QueryVersion,
		membershipProof.ActualVersion,
		keyDigest,
		hasher,
	)

	correct := memProof.Verify(commitment, key)
	if !correct {
		t.Errorf("Proof is incorrect")
	}

	original, _ := store.Get(keyDigest)

	tamperVal := ^uint64(0) // max uint ftw!
	tpBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(tpBytes, tamperVal)

	err := store.Add(keyDigest, tpBytes)
	if err != nil {
		t.Fatal("store add returned not nil value")
	}

	tampered, _ := store.Get(keyDigest)
	if bytes.Compare(tpBytes, tampered) != 0 {
		t.Fatal("Tamper unsuccesfull")
	}
	if bytes.Compare(original, tampered) == 0 {
		t.Fatal("Tamper unsuccesfull")
	}

	tpMembershipProof := <-b.GenMembershipProof(key, commitment.Version)

	tpHyperProof := proof.NewProof(hyper.NewRootPosition(hasher.Len(), 0), tpMembershipProof.HyperProof.AuditPath(), hasher)

	if tpMembershipProof.HistoryProof != nil {
		t.Fatal("The history proof must be nil")
	}

	tpProof := balloon.NewMembershipProof(
		tpMembershipProof.Exists,
		tpHyperProof,
		nil,
		tpMembershipProof.CurrentVersion,
		tpMembershipProof.QueryVersion,
		tpMembershipProof.ActualVersion,
		keyDigest,
		hasher,
	)

	if tpProof.Verify(commitment, key) {
		t.Errorf("TamperProof unsuccessful")
	}

}

func TestDeleteAndVerify(t *testing.T) {

	hasher := new(hashing.Sha256Hasher)
	b, store, closeF := createBalloon("treeId", hasher)
	defer closeF()

	key := []byte("Never knows best")
	keyDigest := hasher.Do(key)

	commitment := <-b.Add(key)
	membershipProof := <-b.GenMembershipProof(key, commitment.Version)

	memProof := balloon.NewMembershipProof(
		membershipProof.Exists,
		membershipProof.HyperProof,
		membershipProof.HistoryProof,
		membershipProof.CurrentVersion,
		membershipProof.QueryVersion,
		membershipProof.ActualVersion,
		keyDigest,
		hasher,
	)

	correct := memProof.Verify(commitment, key)
	if !correct {
		t.Errorf("Proof is incorrect")
	}

	tamperVal := ^uint64(0) // max uint ftw!
	tpBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(tpBytes, tamperVal)

	err := store.Delete(keyDigest)
	if err != nil {
		t.Fatal("store.Delete returned not nil value")
	}

	tampered, _ := store.Get(keyDigest)
	if len(tampered) > 0 {
		t.Fatal("Tamper unsuccesfull")
	}

	tpMembershipProof := <-b.GenMembershipProof(key, commitment.Version)

	tpHyperProof := proof.NewProof(hyper.NewRootPosition(hasher.Len(), 0), tpMembershipProof.HyperProof.AuditPath(), hasher)

	if tpMembershipProof.HistoryProof != nil {
		t.Fatal("The history proof must be nil")
	}

	tpProof := balloon.NewMembershipProof(
		tpMembershipProof.Exists,
		tpHyperProof,
		nil,
		tpMembershipProof.CurrentVersion,
		tpMembershipProof.QueryVersion,
		tpMembershipProof.ActualVersion,
		keyDigest,
		hasher,
	)

	if tpProof.Verify(commitment, key) {
		t.Errorf("TamperProof unsuccessful")
	}

}

func TestGenIncrementalAndVerify(t *testing.T) {
	hasher := new(hashing.Sha256Hasher)
	b, store, closeF := createBalloon("treeId", hasher)
	defer closeF()

	c := make([]*balloon.Commitment, 0)
	for i := 0; i < 10; i++ {
		c[i] = <-b.Add(rand.Bytes(10))
	}

	start := 2
	end := 9
	proof := <-b.GenIncrementalProof(start, end)

	correct := proof.Verify(c[start], c[end])
	assert.True(t, correct, "Unable to verify incremental proof")
}

func createBalloon(id string, hasher hashing.Hasher) (*balloon.HyperBalloon, *badger.BadgerStorage, func()) {
	dir, err := ioutil.TempDir("/var/tmp/", "balloon.test")
	if err != nil {
		log.Fatal(err)
	}

	frozen, frozenCloseF := storage.NewBadgerStorage(fmt.Sprintf("%s/frozen", dir))
	leaves, leavesCloseF := storage.NewBadgerStorage(fmt.Sprintf("%s/leaves", dir))
	cache := cache.NewSimpleCache(1 << 20)

	hyperT := hyper.NewTree(id, cache, leaves, hasher)
	historyT := history.NewTree(id, frozen, hasher)
	balloon := balloon.NewHyperBalloon(hasher, historyT, hyperT)

	return balloon, leaves, func() {
		frozenCloseF()
		leavesCloseF()
		os.RemoveAll(dir)
	}
}
