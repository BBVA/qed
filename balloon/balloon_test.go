/*
   Copyright 2018-2019 Banco Bilbao Vizcaya Argentaria, S.A.

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
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/log"
	metrics_utils "github.com/bbva/qed/testutils/metrics"
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

	hasher := hashing.NewSha256Hasher()

	for i := uint64(0); i < 9; i++ {
		eventHash := hasher.Do(rand.Bytes(128))
		snapshot, mutations, err := balloon.Add(eventHash)
		require.NoError(t, err)

		err = store.Mutate(mutations)
		require.NoError(t, err)
		assert.Truef(t, len(mutations) > 0, "There should be some mutations in test %d", i)

		assert.Equalf(t, i, snapshot.Version, "Wrong version in test %d", i)
		assert.NotNil(t, snapshot.HyperDigest, "The HyperDigest shouldn't be nil in test %d", i)
		assert.NotNil(t, snapshot.HistoryDigest, "The HistoryDigest shouldn't be nil in test %d", i)
	}

}
func TestAddBulk(t *testing.T) {

	log.SetLogger("TestAddBulk", log.SILENT)

	store, closeF := storage_utils.OpenBPlusTreeStore()
	defer closeF()

	balloon, err := NewBalloon(store, hashing.NewSha256Hasher)
	require.NoError(t, err)

	hasher := hashing.NewSha256Hasher()

	events := [][]byte{
		[]byte("The year’s at the spring,"),
		[]byte("And day's at the morn;"),
		[]byte("Morning's at seven;"),
		[]byte("The hill-side’s dew-pearled;"),
		[]byte("The lark's on the wing;"),
		[]byte("The snail's on the thorn;"),
		[]byte("God's in his heaven—"),
		[]byte("All's right with the world!"),
	}

	var eventHashes []hashing.Digest
	for _, event := range events {
		eventHashes = append(eventHashes, hasher.Do(event))
	}

	snapshotBulk, mutations, err := balloon.AddBulk(eventHashes)
	require.NoError(t, err)

	err = store.Mutate(mutations)
	require.NoError(t, err)
	assert.Truef(t, len(mutations) > 0, "There should be some mutations")

	for i, snapshot := range snapshotBulk {
		assert.Equalf(t, uint64(i), snapshot.Version, "Wrong version in test %d", i)
		assert.NotNil(t, snapshot.HyperDigest, "The HyperDigest shouldn't be nil in test %d", i)
		assert.NotNil(t, snapshot.HistoryDigest, "The HistoryDigest shouldn't be nil in test %d", i)
	}
}

func TestQueryMembershipConsistency(t *testing.T) {

	log.SetLogger("TestQueryMembership", log.SILENT)

	testCases := []struct {
		key     []byte
		version uint64
		exists  bool
	}{
		{[]byte{0x5a}, uint64(0), true},
		{nil, uint64(42), false},
	}

	hasher := hashing.NewSha256Hasher()

	for i, c := range testCases {
		store, closeF := storage_utils.OpenBPlusTreeStore()

		balloon, err := NewBalloon(store, hashing.NewSha256Hasher)
		require.NoError(t, err)

		if c.key != nil {
			_, mutations, err := balloon.Add(hasher.Do(c.key))
			require.NoErrorf(t, err, "Error adding event %d", i)
			err = store.Mutate(mutations)
			require.NoError(t, err)
		}

		proof, err := balloon.QueryMembershipConsistency(c.key, c.version)

		require.NoError(t, err)
		assert.True(t, proof.Exists == c.exists, "The event should exist in test %d ", i)

		if c.exists {
			assert.Equalf(t, c.version, proof.QueryVersion, "The query version does not match in test %d : expected %d, actual %d", i, c.version, proof.QueryVersion)
			assert.Equalf(t, c.version, proof.ActualVersion, "The actual version does not match in test %d : expected %d, actual %d", i, c.version, proof.ActualVersion)
			assert.NotNil(t, proof.HyperProof, "The hyper proof should not be nil in test %d ", i)
			assert.NotNil(t, proof.HistoryProof, "The history proof should not be nil in test %d ", i)
		}

		closeF()
	}

}

func TestQueryMembership(t *testing.T) {

	log.SetLogger("TestQueryMembership", log.SILENT)

	testCases := []struct {
		key    []byte
		exists bool
	}{
		{[]byte{0x5a}, true},
		{nil, false},
	}

	hasher := hashing.NewSha256Hasher()

	for i, c := range testCases {
		store, closeF := storage_utils.OpenBPlusTreeStore()

		balloon, err := NewBalloon(store, hashing.NewSha256Hasher)
		require.NoError(t, err)

		if c.key != nil {
			_, mutations, err := balloon.Add(hasher.Do(c.key))
			require.NoErrorf(t, err, "Error adding event %d", i)
			err = store.Mutate(mutations)
			require.NoError(t, err)
		}

		proof, err := balloon.QueryMembership(c.key)

		require.NoError(t, err)
		assert.True(t, proof.Exists == c.exists, "The event should exist in test %d ", i)

		if c.exists {
			assert.NotNil(t, proof.HyperProof, "The hyper proof should not be nil in test %d ", i)
			assert.NotNil(t, proof.HistoryProof, "The history proof should not be nil in test %d ", i)
		}

		closeF()
	}
}

func TestQueryConsistencyProof(t *testing.T) {

	log.SetLogger("TestQueryConsistencyProof", log.SILENT)

	testCases := []struct {
		additions, start, end uint64
		ok                    bool
	}{
		{uint64(2), uint64(0), uint64(2), true},
		{uint64(0), uint64(30), uint64(600), false},
	}

	hasher := hashing.NewFakeXorHasher()

	for i, c := range testCases {
		store, closeF := storage_utils.OpenBPlusTreeStore()
		defer closeF()
		balloon, err := NewBalloon(store, hashing.NewFakeXorHasher)
		require.NoError(t, err)

		for j := 0; j <= int(c.additions); j++ {
			eventHash := hasher.Do(util.Uint64AsBytes(uint64(j)))
			_, mutations, err := balloon.Add(eventHash)
			require.NoErrorf(t, err, "Error adding event %d", j)
			err = store.Mutate(mutations)
			require.NoError(t, err)
		}

		proof, err := balloon.QueryConsistency(c.start, c.end)

		if c.ok {
			require.NoError(t, err)
			assert.Equalf(t, c.start, proof.Start, "The query start does not match in test %d: expected %d, actual %d", i, c.start, proof.Start)
			assert.Equalf(t, c.end, proof.End, "The query end does not match in test %d: expected %d, actual %d", i, c.end, proof.End)
			assert.Truef(t, len(proof.AuditPath) > 0, "The length of the audith path should be >0 in test %d ", i)
		} else {
			require.Error(t, err)
		}
	}
}

func TestConsistencyProofVerify(t *testing.T) {
	// Tests already done in history>proof_test.go
}

func TestAddQueryAndVerify(t *testing.T) {
	log.SetLogger("TestCacheWarmingUp", log.SILENT)

	store, closeF := storage_utils.OpenRocksDBStore(t, "/var/tmp/balloon.test.1")
	defer closeF()

	// start balloon
	b, err := NewBalloon(store, hashing.NewSha256Hasher)
	assert.NoError(t, err)

	event := []byte("Never knows best")
	hasher := hashing.NewSha256Hasher()

	// Add event
	snapshot, mutations, err := b.Add(hasher.Do(event))
	require.NoError(t, err)
	err = store.Mutate(mutations)
	require.NoError(t, err)

	// Query event
	proof, err := b.QueryMembershipConsistency(event, snapshot.Version)
	assert.NoError(t, err)

	// Verify
	assert.True(t, proof.Verify(event, snapshot), "The proof should verify correctly")
}

func TestCacheWarmingUp(t *testing.T) {

	log.SetLogger("TestCacheWarmingUp", log.SILENT)

	store, closeF := storage_utils.OpenRocksDBStore(t, "/var/tmp/ballon_test.db")
	defer closeF()

	// start balloon
	balloon, err := NewBalloon(store, hashing.NewSha256Hasher)
	require.NoError(t, err)

	hasher := hashing.NewSha256Hasher()

	// add 100 elements
	var lastSnapshot *Snapshot
	for i := uint64(0); i < 100; i++ {
		eventHash := hasher.Do(util.Uint64AsBytes(i))
		snapshot, mutations, err := balloon.Add(eventHash)
		require.NoError(t, err)
		lastSnapshot = snapshot
		err = store.Mutate(mutations)
		require.NoError(t, err)
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
		proof, err := balloon.QueryMembershipConsistency(key, lastSnapshot.Version)
		require.NoError(t, err)
		require.Truef(t, proof.Verify(key, lastSnapshot), "The proof should verify correctly for element %d", i)
	}
}

func TestGenIncrementalAndVerify(t *testing.T) {
	log.SetLogger("TestGenIncrementalAndVerify", log.SILENT)

	store, closeF := storage_utils.OpenRocksDBStore(t, "/var/tmp/balloon.test.3")
	defer closeF()

	b, err := NewBalloon(store, hashing.NewSha256Hasher)
	assert.NoError(t, err)

	size := 10
	s := make([]*Snapshot, size)
	for i := 0; i < size; i++ {
		event := hashing.Digest(fmt.Sprintf("Never knows %d best", i))
		snapshot, mutations, _ := b.Add(event)
		err = store.Mutate(mutations)
		require.NoError(t, err)
		s[i] = snapshot
	}

	start := uint64(1)
	end := uint64(7)
	proof, err := b.QueryConsistency(start, end)
	assert.NoError(t, err)

	correct := proof.Verify(s[start], s[end])
	assert.True(t, correct, "Unable to verify incremental proof")
}

func BenchmarkAddRocksDB(b *testing.B) {

	log.SetLogger("BenchmarkAddRocksDB", log.SILENT)

	store, closeF := storage_utils.OpenRocksDBStore(b, "/var/tmp/balloon_bench.db")
	defer closeF()

	balloon, err := NewBalloon(store, hashing.NewSha256Hasher)
	require.NoError(b, err)

	hasher := hashing.NewSha256Hasher()

	balloonMetrics := metrics_utils.CustomRegister(AddTotal)
	srvCloseF := metrics_utils.StartMetricsServer(balloonMetrics, store)
	defer srvCloseF()

	b.ResetTimer()
	b.N = 2000000
	for i := 0; i < b.N; i++ {
		event := rand.Bytes(128)
		_, mutations, err := balloon.Add(hasher.Do(event))
		require.NoError(b, err)
		require.NoError(b, store.Mutate(mutations))
		AddTotal.Inc()
	}

}

func BenchmarkAddBulkRocksDB(b *testing.B) {

	log.SetLogger("BenchmarkAddBulkRocksDB", log.SILENT)

	store, closeF := storage_utils.OpenRocksDBStore(b, "/var/tmp/balloon_bench.db")
	defer closeF()

	balloon, err := NewBalloon(store, hashing.NewSha256Hasher)
	require.NoError(b, err)

	hasher := hashing.NewSha256Hasher()

	balloonMetrics := metrics_utils.CustomRegister(AddTotal)
	srvCloseF := metrics_utils.StartMetricsServer(balloonMetrics, store)
	defer srvCloseF()

	b.ResetTimer()
	b.N = 2000000
	for i := 0; i < b.N; i++ {
		events := []hashing.Digest{hasher.Do(rand.Bytes(128))}
		_, mutations, err := balloon.AddBulk(events)
		require.NoError(b, err)
		require.NoError(b, store.Mutate(mutations))
		AddTotal.Inc()
	}

}

func BenchmarkQueryRocksDB(b *testing.B) {
	var events [][]byte
	log.SetLogger("BenchmarkQueryRocksDB", log.SILENT)

	store, closeF := storage_utils.OpenRocksDBStore(b, "/var/tmp/ballon_bench.db")
	defer closeF()

	balloon, err := NewBalloon(store, hashing.NewSha256Hasher)
	require.NoError(b, err)

	hasher := hashing.NewSha256Hasher()

	b.N = 1000000
	for i := 0; i < b.N; i++ {
		event := rand.Bytes(128)
		eventHash := hasher.Do(event)
		events = append(events, event)
		_, mutations, _ := balloon.Add(eventHash)
		_ = store.Mutate(mutations)
	}

	b.ResetTimer()
	for i, e := range events {
		_, err := balloon.QueryMembershipConsistency(e, uint64(i))
		require.NoError(b, err)
	}

}

func BenchmarkQueryRocksDBParallel(b *testing.B) {
	var events [][]byte
	log.SetLogger("BenchmarkQueryRocksDB", log.SILENT)

	store, closeF := storage_utils.OpenRocksDBStore(b, "/var/tmp/ballon_bench.db")
	defer closeF()

	balloon, err := NewBalloon(store, hashing.NewSha256Hasher)
	require.NoError(b, err)

	hasher := hashing.NewSha256Hasher()

	b.N = 1000000
	for i := 0; i < b.N; i++ {
		event := rand.Bytes(128)
		eventHash := hasher.Do(event)
		events = append(events, event)
		_, mutations, _ := balloon.Add(eventHash)
		_ = store.Mutate(mutations)
	}

	b.ResetTimer()
	n := int64(-1)
	b.SetParallelism(100)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			i := atomic.AddInt64(&n, 1)
			event := events[i]
			_, err := balloon.QueryMembershipConsistency(event, uint64(i))
			require.NoError(b, err)
		}
	})

}
