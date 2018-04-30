// Copyright © 2018 Banco Bilbao Vizcaya Argentaria S.A.  All rights reserved.
// Use of this source code is governed by an Apache 2 License
// that can be found in the LICENSE file

package balloon

import (
	"testing"

	"verifiabledata/balloon/hashing"
	"verifiabledata/balloon/history"
	"verifiabledata/balloon/hyper"
	"verifiabledata/balloon/storage/cache"
)

type FakeVerifiable struct {
	result bool
}

func NewFakeVerifiable(result bool) *FakeVerifiable {
	return &FakeVerifiable{result}
}

func (f *FakeVerifiable) Verify(commitment, event []byte, version uint) bool {
	return f.result
}

func TestVerify(t *testing.T) {

	testCases := []struct {
		exists         bool
		hyperOK        bool
		historyOK      bool
		queryVersion   uint
		actualVersion  uint
		expectedResult bool
	}{
		// Event exists, queryVersion <= actualVersion, and both trees verify it
		{true, true, true, uint(0), uint(0), true},
		// Event exists, queryVersion <= actualVersion, but HyperTree does not verify it
		{true, false, true, uint(0), uint(0), false},
		// Event exists, queryVersion <= actualVersion, but HistoryTree does not verify it
		{true, true, false, uint(0), uint(0), false},

		// Event exists, queryVersion > actualVersion, and both trees verify it
		{true, true, true, uint(1), uint(0), true},
		// Event exists, queryVersion > actualVersion, but HyperTree does not verify it
		{true, false, true, uint(1), uint(0), false},

		// Event does not exist, HyperTree verifies it
		{false, true, false, uint(0), uint(0), true},
		// Event does not exist, HyperTree does not verify it
		{false, false, false, uint(0), uint(0), false},
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

func createBalloon(id string, hasher hashing.Hasher) (*HyperBalloon, func()) {
	frozen, frozenCloseF := openBPlusStorage()
	leaves, leavesCloseF := openBPlusStorage()
	cache := cache.NewSimpleCache(0)

	//TODO: this should not be part of the test

	hyperT := hyper.NewFakeTree(string(0x0), cache, leaves, hasher)
	historyT := history.NewFakeTree(frozen, hasher)
	balloon := NewHyperBalloon(hasher, historyT, hyperT)

	return balloon, func() {
		frozenCloseF()
		leavesCloseF()
	}
}

func TestAddAndVerify(t *testing.T) {
	id := string(0x0)
	hasher := hashing.Sha256Hasher

	balloon, closeF := createBalloon(id, hasher)
	defer closeF()

	key := []byte("Never knows best")
	// keyDigest := hasher(key)

	commitment := <-balloon.Add(key)
	membershipProof := <-balloon.GenMembershipProof(key, commitment.Version)

	proof := ToBalloonProof(id, membershipProof, hasher)
	correct := proof.Verify(commitment, key)

	if !correct {
		t.Errorf("Proof is incorrect")
	}

}
