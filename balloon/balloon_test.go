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
	"encoding/binary"
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bbva/qed/crypto/hashing"
	metrics_utils "github.com/bbva/qed/testutils/metrics"
	"github.com/bbva/qed/testutils/rand"
	storage_utils "github.com/bbva/qed/testutils/storage"
	"github.com/bbva/qed/util"
)

func TestAdd(t *testing.T) {

	store, closeF := storage_utils.OpenBPlusTreeStore()
	defer closeF()

	h := hashing.NewSha256Hasher()
	balloon, err := NewBalloon(store, hashing.NewSha256Hasher)
	require.NoError(t, err)

	for i := uint64(0); i < 9; i++ {
		eventHash := h.Do(rand.Bytes(128))
		snapshot, mutations, err := balloon.Add(eventHash)
		require.NoError(t, err)

		err = store.Mutate(mutations, nil)
		require.NoError(t, err)
		assert.Truef(t, len(mutations) > 0, "There should be some mutations in test %d", i)

		assert.Equalf(t, i, snapshot.Version, "Wrong version in test %d", i)
		assert.NotNil(t, snapshot.HyperDigest, "The HyperDigest shouldn't be nil in test %d", i)
		assert.NotNil(t, snapshot.HistoryDigest, "The HistoryDigest shouldn't be nil in test %d", i)
	}

}
func TestAddBulk(t *testing.T) {

	store, closeF := storage_utils.OpenBPlusTreeStore()
	defer closeF()

	h := hashing.NewSha256Hasher()
	balloon, err := NewBalloon(store, hashing.NewSha256Hasher)
	require.NoError(t, err)

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
		eventHashes = append(eventHashes, h.Do(event))
	}

	snapshotBulk, mutations, err := balloon.AddBulk(eventHashes)
	require.NoError(t, err)

	err = store.Mutate(mutations, nil)
	require.NoError(t, err)
	assert.Truef(t, len(mutations) > 0, "There should be some mutations")

	for i, snapshot := range snapshotBulk {
		assert.Equalf(t, uint64(i), snapshot.Version, "Wrong version in test %d", i)
		assert.NotNil(t, snapshot.HyperDigest, "The HyperDigest shouldn't be nil in test %d", i)
		assert.NotNil(t, snapshot.HistoryDigest, "The HistoryDigest shouldn't be nil in test %d", i)
	}
}

func TestQueryMembershipConsistency(t *testing.T) {

	testCases := []struct {
		key     []byte
		version uint64
		exists  bool
	}{
		{[]byte{0x5a}, uint64(0), true},
		{nil, uint64(42), false},
	}

	hasherF := hashing.NewSha256Hasher
	h := hasherF()

	for i, c := range testCases {
		store, closeF := storage_utils.OpenBPlusTreeStore()

		balloon, err := NewBalloon(store, hasherF)
		require.NoError(t, err)

		if c.key != nil {
			_, mutations, err := balloon.Add(h.Do(c.key))
			require.NoErrorf(t, err, "Error adding event %d", i)
			err = store.Mutate(mutations, nil)
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

	testCases := []struct {
		key    []byte
		exists bool
	}{
		{[]byte{0x5a}, true},
		{nil, false},
	}

	hasherF := hashing.NewSha256Hasher
	h := hasherF()

	for i, c := range testCases {
		store, closeF := storage_utils.OpenBPlusTreeStore()

		balloon, err := NewBalloon(store, hasherF)
		require.NoError(t, err)

		if c.key != nil {
			_, mutations, err := balloon.Add(h.Do(c.key))
			require.NoErrorf(t, err, "Error adding event %d", i)
			err = store.Mutate(mutations, nil)
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

	testCases := []struct {
		additions, start, end uint64
		ok                    bool
	}{
		{uint64(2), uint64(0), uint64(2), true},
		{uint64(0), uint64(30), uint64(600), false},
	}

	hasherF := hashing.NewSha256Hasher
	h := hasherF()

	for i, c := range testCases {
		store, closeF := storage_utils.OpenBPlusTreeStore()
		defer closeF()
		balloon, err := NewBalloon(store, hasherF)
		require.NoError(t, err)

		for j := 0; j <= int(c.additions); j++ {
			eventHash := h.Do(util.Uint64AsBytes(uint64(j)))
			_, mutations, err := balloon.Add(eventHash)
			require.NoErrorf(t, err, "Error adding event %d", j)
			err = store.Mutate(mutations, nil)
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
	t.Skip()
	// Tests already done in history>proof_test.go
}

func TestAddQueryAndVerify(t *testing.T) {

	store, closeF := storage_utils.OpenRocksDBStore(t, "/var/tmp/balloon.test.1")
	defer closeF()

	// start balloon
	h := hashing.NewSha256Hasher()
	b, err := NewBalloon(store, hashing.NewSha256Hasher)
	require.NoError(t, err)

	event := []byte("Never knows best")

	// Add event
	snapshot, mutations, err := b.Add(h.Do(event))
	require.NoError(t, err)
	require.NoError(t, store.Mutate(mutations, nil))

	// Query event
	proof, err := b.QueryMembershipConsistency(event, snapshot.Version)
	require.NoError(t, err)

	// Verify
	require.True(t, proof.Verify(event, snapshot), "The proof should verify correctly")
}

func TestCacheWarmingUp(t *testing.T) {

	store, closeF := storage_utils.OpenRocksDBStore(t, "/var/tmp/ballon_test.db")
	defer closeF()

	// start balloon
	raftBalloonHasherF := hashing.NewSha256Hasher
	h := raftBalloonHasherF()
	balloon, err := NewBalloon(store, raftBalloonHasherF)
	require.NoError(t, err)

	// add 100 elements
	var lastSnapshot *Snapshot
	for i := uint64(0); i < 100; i++ {
		eventHash := h.Do(util.Uint64AsBytes(i))
		snapshot, mutations, err := balloon.Add(eventHash)
		require.NoError(t, err)
		lastSnapshot = snapshot
		err = store.Mutate(mutations, nil)
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

	store, closeF := storage_utils.OpenRocksDBStore(t, "/var/tmp/balloon.test.3")
	defer closeF()

	h := hashing.NewSha256Hasher()
	b, err := NewBalloon(store, hashing.NewSha256Hasher)
	assert.NoError(t, err)

	size := 10
	s := make([]*Snapshot, size)
	for i := 0; i < size; i++ {
		event := h.Do([]byte(fmt.Sprintf("Never knows %d best", i)))
		snapshot, mutations, _ := b.Add(event)
		err = store.Mutate(mutations, nil)
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

func TestAddBulkAndQuery(t *testing.T) {

	store, closeF := storage_utils.OpenRocksDBStore(t, "/var/tmp/balloon.test.6")
	defer closeF()

	h := hashing.NewSha256Hasher()
	balloon, err := NewBalloon(store, hashing.NewSha256Hasher)
	require.NoError(t, err)

	size := 1000
	// insert
	bulk := make([]hashing.Digest, 0)
	for i := 0; i < size; i++ {
		eventDigest := h.Do([]byte(fmt.Sprintf("Never knows %d best", i)))
		bulk = append(bulk, eventDigest)
		if i%20 == 0 || i == size-1 {
			_, mutations, err := balloon.AddBulk(bulk)
			require.NoError(t, err)
			require.NoError(t, store.Mutate(mutations, nil))
			bulk = make([]hashing.Digest, 0)
		}
	}

	// query
	for i := 0; i < size; i++ {
		event := []byte(fmt.Sprintf("Never knows %d best", i))
		proof, err := balloon.QueryMembership(event)
		require.NoError(t, err)
		require.Equalf(t, h.Do(event), proof.KeyDigest, "index %d", i)
		require.Equalf(t, uint64(i), proof.ActualVersion, "index %d", i)
	}

}

func TestAddAndQuery(t *testing.T) {

	store, closeF := storage_utils.OpenRocksDBStore(t, "/var/tmp/balloon.test.7")
	defer closeF()

	h := hashing.NewSha256Hasher()
	balloon, err := NewBalloon(store, hashing.NewSha256Hasher)
	require.NoError(t, err)

	size := 1000
	// insert
	for i := 0; i < size; i++ {
		eventDigest := h.Do([]byte(fmt.Sprintf("Never knows %d best", i)))
		_, mutations, err := balloon.Add(eventDigest)
		require.NoError(t, err)
		require.NoError(t, store.Mutate(mutations, nil))
	}

	// query
	for i := 0; i < size; i++ {
		event := []byte(fmt.Sprintf("Never knows %d best", i))
		proof, err := balloon.QueryMembership(event)
		require.NoError(t, err)
		require.True(t, proof.Exists, "index %d", i)
		require.Equalf(t, h.Do(event), proof.KeyDigest, "index %d", i)
		require.Equalf(t, uint64(i), proof.ActualVersion, "index %d", i)
	}

}

func TestAddAndQueryConsistency(t *testing.T) {

	store, closeF := storage_utils.OpenRocksDBStore(t, "/var/tmp/balloon.test.7")
	defer closeF()

	h := hashing.NewSha256Hasher()
	balloon, err := NewBalloon(store, hashing.NewSha256Hasher)
	require.NoError(t, err)

	size := 10
	snapshots := make([]*Snapshot, size)
	// insert
	for i := 0; i < size; i++ {
		eventDigest := h.Do([]byte(fmt.Sprintf("Never knows %d best", i)))
		snapshot, mutations, err := balloon.Add(eventDigest)
		require.NoError(t, err)
		require.NoError(t, store.Mutate(mutations, nil))
		snapshots[i] = snapshot
	}

	// query and verify
	for i := 0; i < size; i++ {
		for j := i; j < size; j++ {
			proof, err := balloon.QueryConsistency(uint64(i), uint64(j))
			require.NoErrorf(t, err, "start %d, end %d", i, j)
			require.True(t, len(proof.AuditPath) > 0, "start %d, end %d", i, j)
			require.Equalf(t, uint64(i), proof.Start, "start %d, end %d", i, j)
			require.Equalf(t, uint64(j), proof.End, "start %d, end %d", i, j)
			require.Truef(t, proof.Verify(snapshots[i], snapshots[j]), "start %d, end %d", i, j)
		}
	}

}

func BenchmarkAddRocksDB(b *testing.B) {

	store, closeF := storage_utils.OpenRocksDBStore(b, "/var/tmp/balloon_bench.db")
	defer closeF()

	h := hashing.NewSha256Hasher()
	balloon, err := NewBalloon(store, hashing.NewSha256Hasher)
	require.NoError(b, err)

	balloonMetrics := metrics_utils.CustomRegister(AddTotal)
	srvCloseF := metrics_utils.StartMetricsServer(balloonMetrics, store)
	defer srvCloseF()

	b.ResetTimer()
	b.N = 2000000
	for i := 0; i < b.N; i++ {
		event := rand.Bytes(128)
		_, mutations, err := balloon.Add(h.Do(event))
		require.NoError(b, err)
		require.NoError(b, store.Mutate(mutations, nil))
		AddTotal.Inc()
	}
}

func BenchmarkAddBulkRocksDB(b *testing.B) {

	store, closeF := storage_utils.OpenRocksDBStore(b, "/var/tmp/balloon_bench.db")
	defer closeF()

	h := hashing.NewSha256Hasher()
	balloon, err := NewBalloon(store, hashing.NewSha256Hasher)
	bulkSize := uint64(20)
	eventDigests := make([]hashing.Digest, bulkSize)
	require.NoError(b, err)

	balloonMetrics := metrics_utils.CustomRegister(AddTotal)
	srvCloseF := metrics_utils.StartMetricsServer(balloonMetrics, store)
	defer srvCloseF()

	b.ResetTimer()
	b.N = 1000000
	for i := uint64(0); i < uint64(b.N); i++ {
		idx := i % bulkSize

		index := make([]byte, 8)
		binary.LittleEndian.PutUint64(index, i)
		event := append(rand.Bytes(32), index...)
		eventDigests[idx] = h.Do(event)

		if idx == bulkSize-1 {
			_, mutations, err := balloon.AddBulk(eventDigests)
			require.NoError(b, err)
			require.NoError(b, store.Mutate(mutations, nil))
			AddTotal.Add(float64(bulkSize))
		}
	}
}
func BenchmarkQueryRocksDB(b *testing.B) {

	var events [][]byte

	store, closeF := storage_utils.OpenRocksDBStore(b, "/var/tmp/ballon_bench.db")
	defer closeF()

	h := hashing.NewSha256Hasher()
	balloon, err := NewBalloon(store, hashing.NewSha256Hasher)
	require.NoError(b, err)

	b.N = 1000000
	for i := 0; i < b.N; i++ {
		event := rand.Bytes(128)
		eventHash := h.Do(event)
		events = append(events, event)
		_, mutations, _ := balloon.Add(eventHash)
		_ = store.Mutate(mutations, nil)
	}

	b.ResetTimer()
	for i, e := range events {
		_, err := balloon.QueryMembershipConsistency(e, uint64(i))
		require.NoError(b, err)
	}
}

func BenchmarkQueryRocksDBParallel(b *testing.B) {

	var events [][]byte

	store, closeF := storage_utils.OpenRocksDBStore(b, "/var/tmp/ballon_bench.db")
	defer closeF()

	h := hashing.NewSha256Hasher()
	balloon, err := NewBalloon(store, hashing.NewSha256Hasher)
	require.NoError(b, err)

	b.N = 1000000
	for i := 0; i < b.N; i++ {
		event := rand.Bytes(128)
		eventHash := h.Do(event)
		events = append(events, event)
		_, mutations, _ := balloon.Add(eventHash)
		_ = store.Mutate(mutations, nil)
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
