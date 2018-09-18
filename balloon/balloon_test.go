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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bbva/qed/balloon/common"
	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/log"
	"github.com/bbva/qed/testutils/rand"
	storage_utils "github.com/bbva/qed/testutils/storage"
	"github.com/bbva/qed/util"
)

func TestAdd(t *testing.T) {

	log.SetLogger("TestAdd", log.SILENT)

	store, closeF := storage_utils.OpenBPlusTreeStore()
	defer closeF()

	balloon, err := NewBalloon(store, hashing.NewSha256Hasher)
	require.NoError(t, err)

	for i := uint64(0); i < 9; i++ {
		commitment, mutations, err := balloon.Add(rand.Bytes(128))
		store.Mutate(mutations)

		require.NoError(t, err)
		assert.Truef(t, len(mutations) > 0, "There should be some mutations in test %d", i)
		assert.Equalf(t, i, commitment.Version, "Wrong version in test %d", i)
		assert.NotNil(t, commitment.HyperDigest, "The HyperDigest shouldn't be nil in test %d", i)
		assert.NotNil(t, commitment.HistoryDigest, "The HistoryDigest shouldn't be nil in test %d", i)
	}

}

func TestQueryMembership(t *testing.T) {

	log.SetLogger("TestQueryMembership", log.SILENT)

	store, closeF := storage_utils.OpenBPlusTreeStore()
	defer closeF()

	balloon, err := NewBalloon(store, hashing.NewFakeXorHasher)
	require.NoError(t, err)

	testCases := []struct {
		key     []byte
		version uint64
	}{
		{[]byte{0x5a}, uint64(0)},
	}

	for i, c := range testCases {
		_, mutations, err := balloon.Add(c.key)
		require.NoErrorf(t, err, "Error adding event %d", i)
		store.Mutate(mutations)

		proof, err := balloon.QueryMembership(c.key, c.version)

		require.NoError(t, err)
		assert.True(t, proof.Exists, "The event should exist in test %d ", i)
		assert.Equalf(t, c.version, proof.QueryVersion, "The query version does not match in test %d : expected %d, actual %d", i, c.version, proof.QueryVersion)
		assert.Equalf(t, c.version, proof.ActualVersion, "The actual version does not match in test %d : expected %d, actual %d", i, c.version, proof.ActualVersion)
		assert.NotNil(t, proof.HyperProof, "The hyper proof should not be nil in test %d ", i)
		assert.NotNil(t, proof.HistoryProof, "The history proof should not be nil in test %d ", i)
	}

}

func TestMembershipProofVerify(t *testing.T) {

	log.SetLogger("TestMembershipProofVerify", log.SILENT)

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
			hashing.Digest("Some hyperDigest"),
			hashing.Digest("Some historyDigest"),
			c.actualVersion,
		}
		proof := NewMembershipProof(
			c.exists,
			common.NewFakeVerifiable(c.hyperOK),
			common.NewFakeVerifiable(c.historyOK),
			c.currentVersion,
			c.queryVersion,
			c.actualVersion,
			event,
			hashing.NewSha256Hasher(),
		)

		result := proof.Verify(event, commitment)

		require.Equalf(t, c.expectedResult, result, "Unexpected result '%v' in test case '%d'", result, i)
	}
}

func TestQueryConsistencyProof(t *testing.T) {

	log.SetLogger("TestQueryConsistencyProof", log.SILENT)

	testCases := []struct {
		start, end uint64
	}{
		{uint64(0), uint64(2)},
	}

	for i, c := range testCases {
		store, closeF := storage_utils.OpenBPlusTreeStore()
		defer closeF()
		balloon, err := NewBalloon(store, hashing.NewFakeXorHasher)
		require.NoError(t, err)

		for j := 0; j <= int(c.end); j++ {
			_, mutations, err := balloon.Add(util.Uint64AsBytes(uint64(j)))
			require.NoErrorf(t, err, "Error adding event %d", j)
			store.Mutate(mutations)
		}

		proof, err := balloon.QueryConsistency(c.start, c.end)

		require.NoError(t, err)
		assert.Equalf(t, c.start, proof.Start, "The query start does not match in test %d: expected %d, actual %d", i, c.start, proof.Start)
		assert.Equalf(t, c.end, proof.End, "The query end does not match in test %d: expected %d, actual %d", i, c.end, proof.End)
		assert.Truef(t, len(proof.AuditPath) > 0, "The lenght of the audith path should be >0 in test %d ", i)
	}
}

func TestConsistencyProofVerify(t *testing.T) {
	// Tests already done in history>proof_test.go
}

// TODO move this to integration
func TestAddQueryAndVerify(t *testing.T) {

	log.SetLogger("TestCacheWarmingUp", log.SILENT)

	store, closeF := storage_utils.OpenBadgerStore(t, "/var/tmp/ballon_test.db")
	defer closeF()

	// start balloon
	balloon, err := NewBalloon(store, hashing.NewSha256Hasher)
	require.NoError(t, err)

	event := []byte("Never knows best")

	// Add event
	commitment, mutations, err := balloon.Add(event)
	require.NoError(t, err)
	store.Mutate(mutations)

	// Query event
	proof, err := balloon.QueryMembership(event, commitment.Version)
	require.NoError(t, err)

	// Verify
	correct := proof.Verify(event, commitment)
	require.True(t, correct, "The proof should verify correctly")

}

func TestCacheWarmingUp(t *testing.T) {

	log.SetLogger("TestCacheWarmingUp", log.SILENT)

	store, closeF := storage_utils.OpenBadgerStore(t, "/var/tmp/ballon_test.db")
	defer closeF()

	// start balloon
	balloon, err := NewBalloon(store, hashing.NewSha256Hasher)
	require.NoError(t, err)

	// add 100 elements
	var lastCommitment *Commitment
	for i := uint64(0); i < 100; i++ {
		commitment, mutations, err := balloon.Add(util.Uint64AsBytes(i))
		require.NoError(t, err)
		lastCommitment = commitment
		store.Mutate(mutations)
	}

	// close balloon
	balloon.Close()
	balloon = nil

	// open balloon again
	balloon, err = NewBalloon(store, hashing.NewSha256Hasher)
	require.NoError(t, err)

	// query for all elements
	for i := uint64(0); i < 100; i++ {
		key := util.Uint64AsBytes(i)
		proof, err := balloon.QueryMembership(key, lastCommitment.Version)
		require.NoError(t, err)
		require.Truef(t, proof.Verify(key, lastCommitment), "The proof should verify correctly for element %d", i)
	}
}

func BenchmarkAddBadger(b *testing.B) {
	store, closeF := storage_utils.OpenBadgerStore(b, "/var/tmp/ballon_bench.db")
	defer closeF()

	balloon, err := NewBalloon(store, hashing.NewSha256Hasher)
	require.NoError(b, err)

	b.ResetTimer()
	b.N = 10000
	for i := 0; i < b.N; i++ {
		event := rand.Bytes(128)
		_, mutations, _ := balloon.Add(event)
		store.Mutate(mutations)
	}

}
