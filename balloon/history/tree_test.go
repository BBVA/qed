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

package history

import (
	"encoding/binary"
	"fmt"
	"os"
	"testing"

	"github.com/bbva/qed/balloon/proof"

	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/metrics"
	"github.com/bbva/qed/storage/badger"
	"github.com/bbva/qed/storage/bplus"
	"github.com/bbva/qed/testutils/rand"
	assert "github.com/stretchr/testify/require"
)

func TestAdd(t *testing.T) {

	testCases := []struct {
		eventDigest      []byte
		expectedRootHash []byte
	}{
		{[]byte{0x0}, []byte{0x0}},
		{[]byte{0x1}, []byte{0x1}},
		{[]byte{0x2}, []byte{0x3}},
		{[]byte{0x3}, []byte{0x0}},
		{[]byte{0x4}, []byte{0x4}},
		{[]byte{0x5}, []byte{0x1}},
		{[]byte{0x6}, []byte{0x7}},
		{[]byte{0x7}, []byte{0x0}},
		{[]byte{0x8}, []byte{0x8}},
		{[]byte{0x9}, []byte{0x1}},
	}

	frozen, close := openBPlusStorage()
	defer close()

	hasher := new(hashing.XorHasher)
	tree := NewTree("treeId", frozen, hasher)
	tree.leafHash = fakeLeafHasherCleanF(hasher)
	tree.interiorHash = fakeInteriorHasherCleanF(hasher)
	// Note that we are using fake hashing functions and the index
	// as the value of the event's digest to make predictable hashes

	for i, c := range testCases {
		index := uint64(i)
		rh, err := tree.Add(c.eventDigest, uint64AsBytes(index))
		assert.NoError(t, err, "Error while adding to the tree")
		assert.Equalf(t, c.expectedRootHash, rh, "Incorrect root hash for index %d", i)
	}

}

func TestProveMembership(t *testing.T) {

	testCases := []struct {
		eventDigest []byte
		auditPath   proof.AuditPath
	}{
		{
			[]byte{0x0},
			proof.AuditPath{"0|0": []uint8{0x0}}, // TODO this should be empty!!!
		},
		{
			[]byte{0x1},
			proof.AuditPath{"0|0": []uint8{0x0}, "1|0": []uint8{0x1}},
		},
		{
			[]byte{0x2},
			proof.AuditPath{"0|1": []uint8{0x1}, "2|0": []uint8{0x2}},
		},
		{
			[]byte{0x3},
			proof.AuditPath{"0|1": []uint8{0x1}, "2|0": []uint8{0x2}, "3|0": []uint8{0x3}},
		},
		{
			[]byte{0x4},
			proof.AuditPath{"0|2": []uint8{0x0}, "4|0": []uint8{0x4}},
		},
		{
			[]byte{0x5},
			proof.AuditPath{"0|2": []uint8{0x0}, "4|0": []uint8{0x4}, "5|0": []uint8{0x5}},
		},
		{
			[]byte{0x6},
			proof.AuditPath{"0|2": []uint8{0x0}, "4|1": []uint8{0x1}, "6|0": []uint8{0x6}},
		},
		{
			[]byte{0x7},
			proof.AuditPath{"0|2": []uint8{0x0}, "4|1": []uint8{0x1}, "6|0": []uint8{0x6}, "7|0": []uint8{0x7}},
		},
		{
			[]byte{0x8},
			proof.AuditPath{"0|3": []uint8{0x0}, "8|0": []uint8{0x8}},
		},
		{
			[]byte{0x9},
			proof.AuditPath{"0|3": []uint8{0x0}, "8|0": []uint8{0x8}, "9|0": []uint8{0x9}},
		},
	}

	frozen, close := openBPlusStorage()
	defer close()

	hasher := new(hashing.XorHasher)
	tree := NewTree("treeId", frozen, hasher)
	tree.leafHash = fakeLeafHasherCleanF(hasher)
	tree.interiorHash = fakeInteriorHasherCleanF(hasher)
	// Note that we are using fake hashing functions and the index
	// as the value of the event's digest to make predictable hashes

	for i, c := range testCases {
		index := uint64(i)
		_, err := tree.Add(c.eventDigest, uint64AsBytes(index))
		assert.NoError(t, err, "Error while adding to the tree")

		pf, err := tree.ProveMembership(c.eventDigest, index, index)
		assert.NoError(t, err, "Error proving membership")
		assert.Equalf(t, c.auditPath, pf.AuditPath(), "Incorrect audit path for index %d", i)
	}
}

func TestProveMembershipWithInvalidTargetVersion(t *testing.T) {
	frozen, close := openBPlusStorage()
	defer close()

	hasher := new(hashing.XorHasher)
	tree := NewTree("treeId", frozen, hasher)

	_, err := tree.ProveMembership([]byte{0x0}, 1, 0)
	assert.Error(t, err, "An error should occur")
}

func TestProveMembershipNonConsecutive(t *testing.T) {
	frozen, close := openBPlusStorage()
	defer close()

	hasher := new(hashing.XorHasher)
	tree := NewTree("treeId", frozen, hasher)
	tree.leafHash = fakeLeafHasherCleanF(hasher)
	tree.interiorHash = fakeInteriorHasherCleanF(hasher)
	// Note that we are using fake hashing functions and the index
	// as the value of the event's digest to make predictable hashes

	// add nine events
	for i := uint64(0); i < 9; i++ {
		eventDigest := uint64AsBytes(i)
		index := uint64AsBytes(i)
		_, err := tree.Add(eventDigest, index)
		assert.NoError(t, err, "Error while adding to the tree")
	}

	// query for membership with event 0 and version 8
	pf, err := tree.ProveMembership([]byte{0x0}, 0, 8)
	assert.NoError(t, err, "Error proving membership")
	expectedAuditPath := proof.AuditPath{"0|0": []uint8{0x0}, "1|0": []uint8{0x1}, "2|1": []uint8{0x1}, "4|2": []uint8{0x0}, "8|3": []uint8{0x8}}
	assert.Equal(t, expectedAuditPath, pf.AuditPath(), "Invalid audit path")
}

func TestProveIncremental(t *testing.T) {

	testCases := []struct {
		eventDigest []byte
		auditPath   proof.AuditPath
	}{
		{
			[]byte{0x0},
			proof.AuditPath{"0|0": []uint8{0x0}}, // TODO this should be empty!!!
		},
		{
			[]byte{0x1},
			proof.AuditPath{"0|0": []uint8{0x0}, "1|0": []uint8{0x1}},
		},
		{
			[]byte{0x2},
			proof.AuditPath{"0|0": []uint8{0x0}, "1|0": []uint8{0x1}, "2|0": []uint8{0x2}},
		},
		{
			[]byte{0x3},
			proof.AuditPath{"0|1": []uint8{0x1}, "2|0": []uint8{0x2}, "3|0": []uint8{0x3}},
		},
		{
			[]byte{0x4},
			proof.AuditPath{"0|1": []uint8{0x1}, "2|0": []uint8{0x2}, "3|0": []uint8{0x3}, "4|0": []uint8{0x4}},
		},
		{
			[]byte{0x5},
			proof.AuditPath{"0|2": []uint8{0x0}, "4|0": []uint8{0x4}, "5|0": []uint8{0x5}},
		},
		{
			[]byte{0x6},
			proof.AuditPath{"0|2": []uint8{0x0}, "4|0": []uint8{0x4}, "5|0": []uint8{0x5}, "6|0": []uint8{0x6}},
		},
		{
			[]byte{0x7},
			proof.AuditPath{"0|2": []uint8{0x0}, "4|1": []uint8{0x1}, "6|0": []uint8{0x6}, "7|0": []uint8{0x7}},
		},
		{
			[]byte{0x8},
			proof.AuditPath{"0|2": []uint8{0x0}, "4|1": []uint8{0x1}, "6|0": []uint8{0x6}, "7|0": []uint8{0x7}, "8|0": []uint8{0x8}},
		},
		{
			[]byte{0x9},
			proof.AuditPath{"0|3": []uint8{0x0}, "8|0": []uint8{0x8}, "9|0": []uint8{0x9}},
		},
	}

	frozen, close := openBPlusStorage()
	defer close()

	hasher := new(hashing.XorHasher)
	tree := NewTree("treeId", frozen, hasher)
	tree.leafHash = fakeLeafHasherCleanF(hasher)
	tree.interiorHash = fakeInteriorHasherCleanF(hasher)
	// Note that we are using fake hashing functions and the index
	// as the value of the event's digest to make predictable hashes

	for i, c := range testCases {
		index := uint64(i)
		_, err := tree.Add(c.eventDigest, uint64AsBytes(index))
		assert.NoError(t, err, "Error while adding to the tree")

		pf, err := tree.ProveIncremental(testCases[max(0, i-1)].eventDigest, c.eventDigest, uint64(max(0, i-1)), index)
		assert.NoError(t, err, "Error while querying for incremental proof in test case: %d", i)
		assert.Equal(t, c.auditPath, pf.AuditPath(), "Invalid audit path in test case: %d", i)
	}

}

func TestProveIncrementalWithInvalidTargetVersion(t *testing.T) {
	frozen, close := openBPlusStorage()
	defer close()

	hasher := new(hashing.XorHasher)
	tree := NewTree("treeId", frozen, hasher)

	_, err := tree.ProveIncremental([]byte{0x1}, []byte{0x0}, 1, 0)
	assert.Error(t, err, "An error should occur")
}

func TestProveIncrementalNonConsecutive(t *testing.T) { // TODO add more test cases!!
	frozen, close := openBPlusStorage()
	defer close()

	hasher := new(hashing.XorHasher)
	tree := NewTree("treeId", frozen, hasher)
	tree.leafHash = fakeLeafHasherCleanF(hasher)
	tree.interiorHash = fakeInteriorHasherCleanF(hasher)
	// Note that we are using fake hashing functions and the index
	// as the value of the event's digest to make predictable hashes

	// add nine events
	for i := uint64(0); i < 9; i++ {
		eventDigest := uint64AsBytes(i)
		index := uint64AsBytes(i)
		_, err := tree.Add(eventDigest, index)
		assert.NoError(t, err, "Error while adding to the tree")
	}

	// query for consistency with event 2 and version 8
	pf, err := tree.ProveIncremental(uint64AsBytes(2), uint64AsBytes(8), 2, 8)
	assert.NoError(t, err, "Error while querying for incremental proof")
	expectedAuditPath := proof.AuditPath{
		"0|1": []uint8{0x1}, "2|0": []uint8{0x2}, "3|0": []uint8{0x3},
		"4|2": []uint8{0x0}, "8|0": []uint8{0x8},
	}
	assert.Equal(t, expectedAuditPath, pf.AuditPath(), "Invalid audit path")
}

func TestProveIncrementalSameVersions(t *testing.T) {
	frozen, close := openBPlusStorage()
	defer close()

	hasher := new(hashing.XorHasher)
	tree := NewTree("treeId", frozen, hasher)
	tree.leafHash = fakeLeafHasherCleanF(hasher)
	tree.interiorHash = fakeInteriorHasherCleanF(hasher)
	// Note that we are using fake hashing functions and the index
	// as the value of the event's digest to make predictable hashes

	// add nine events
	for i := uint64(0); i < 9; i++ {
		eventDigest := uint64AsBytes(i)
		index := uint64AsBytes(i)
		_, err := tree.Add(eventDigest, index)
		assert.NoError(t, err, "Error while adding to the tree")
	}

	// query for consistency with event 8 and version 8
	pf, err := tree.ProveIncremental(uint64AsBytes(8), uint64AsBytes(8), 8, 8)
	assert.NoError(t, err, "Error while querying for incremental proof")
	expectedAuditPath := proof.AuditPath{"0|3": []uint8{0x0}, "8|0": []uint8{0x8}}
	assert.Equal(t, expectedAuditPath, pf.AuditPath(), "Invalid audit path")
}

func BenchmarkAdd(b *testing.B) {
	store, closeF := openBadgerStorage()
	defer closeF()
	ht := NewTree("treeId", store, hashing.NewSha256Hasher())
	b.N = 100000
	for i := 0; i < b.N; i++ {
		key := rand.Bytes(64)
		index := make([]byte, 8)
		binary.LittleEndian.PutUint64(index, uint64(i))
		ht.Add(key, index)
	}
	b.Logf("stats = %+v\n", metrics.History)
}

func openBPlusStorage() (*bplus.BPlusTreeStorage, func()) {
	store := bplus.NewBPlusTreeStorage()
	return store, func() {
		store.Close()
	}
}

func openBadgerStorage() (*badger.BadgerStorage, func()) {
	store := badger.NewBadgerStorage("/var/tmp/history_store_test.db")
	return store, func() {
		fmt.Println("Cleaning...")
		store.Close()
		deleteFile("/var/tmp/history_store_test.db")
	}
}

func deleteFile(path string) {
	err := os.RemoveAll(path)
	if err != nil {
		fmt.Printf("Unable to remove db file %s", err)
	}
}
