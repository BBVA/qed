package history

import (
	"testing"

	"github.com/bbva/qed/balloon2/common"
	"github.com/bbva/qed/db"
	"github.com/bbva/qed/db/bplus"
	"github.com/bbva/qed/log"
	"github.com/bbva/qed/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdd(t *testing.T) {

	log.SetLogger("TestAdd", log.INFO)

	testCases := []struct {
		eventDigest          common.Digest
		expectedRootHash     common.Digest
		expectedMutationsLen int
	}{
		{
			eventDigest:          common.Digest{0x0},
			expectedRootHash:     common.Digest{0x0},
			expectedMutationsLen: 1,
		},
		{
			eventDigest:          common.Digest{0x1},
			expectedRootHash:     common.Digest{0x1},
			expectedMutationsLen: 2,
		},
		{
			eventDigest:          common.Digest{0x2},
			expectedRootHash:     common.Digest{0x3},
			expectedMutationsLen: 1,
		},
		{
			eventDigest:          common.Digest{0x3},
			expectedRootHash:     common.Digest{0x0},
			expectedMutationsLen: 3,
		},
		{
			eventDigest:          common.Digest{0x4},
			expectedRootHash:     common.Digest{0x4},
			expectedMutationsLen: 1,
		},
		{
			eventDigest:          common.Digest{0x5},
			expectedRootHash:     common.Digest{0x1},
			expectedMutationsLen: 2,
		},
		{
			eventDigest:          common.Digest{0x6},
			expectedRootHash:     common.Digest{0x7},
			expectedMutationsLen: 1,
		},
		{
			eventDigest:          common.Digest{0x7},
			expectedRootHash:     common.Digest{0x0},
			expectedMutationsLen: 4,
		},
		{
			eventDigest:          common.Digest{0x8},
			expectedRootHash:     common.Digest{0x8},
			expectedMutationsLen: 1,
		},
		{
			eventDigest:          common.Digest{0x9},
			expectedRootHash:     common.Digest{0x1},
			expectedMutationsLen: 2,
		},
	}

	store := bplus.NewBPlusTreeStore()
	cache := common.NewPassThroughCache(db.HistoryCachePrefix, store)
	tree := NewHistoryTree(common.NewFakeXorHasher(), cache)

	for i, c := range testCases {
		index := uint64(i)
		rootHash, mutations, err := tree.Add(c.eventDigest, index)

		require.NoError(t, err)
		assert.Equalf(t, c.expectedRootHash, rootHash, "Incorrect root hash for test case %d", i)
		assert.Equalf(t, c.expectedMutationsLen, len(mutations), "The mutations should match for test case %d", i)

		store.Mutate(mutations...)
	}

}

func TestProveMembership(t *testing.T) {

	log.SetLogger("TestProveMembership", log.INFO)

	testCases := []struct {
		eventDigest       common.Digest
		expectedAuditPath common.AuditPath
	}{
		{
			common.Digest{0x0},
			common.AuditPath{},
		},
		{
			common.Digest{0x1},
			common.AuditPath{"0|0": common.Digest{0x0}},
		},
		{
			common.Digest{0x2},
			common.AuditPath{"0|1": common.Digest{0x1}},
		},
		{
			common.Digest{0x3},
			common.AuditPath{"0|1": common.Digest{0x1}, "2|0": common.Digest{0x2}},
		},
		{
			common.Digest{0x4},
			common.AuditPath{"0|2": common.Digest{0x0}},
		},
		{
			common.Digest{0x5},
			common.AuditPath{"0|2": common.Digest{0x0}, "4|0": common.Digest{0x4}},
		},
		{
			common.Digest{0x6},
			common.AuditPath{"0|2": common.Digest{0x0}, "4|1": common.Digest{0x1}},
		},
		{
			common.Digest{0x7},
			common.AuditPath{"0|2": common.Digest{0x0}, "4|1": common.Digest{0x1}, "6|0": common.Digest{0x6}},
		},
		{
			common.Digest{0x8},
			common.AuditPath{"0|3": common.Digest{0x0}},
		},
		{
			common.Digest{0x9},
			common.AuditPath{"0|3": common.Digest{0x0}, "8|0": common.Digest{0x8}},
		},
	}

	store := bplus.NewBPlusTreeStore()
	cache := common.NewPassThroughCache(db.HistoryCachePrefix, store)
	tree := NewHistoryTree(common.NewFakeXorHasher(), cache)

	for i, c := range testCases {

		index := uint64(i)

		_, mutations, _ := tree.Add(c.eventDigest, index)
		store.Mutate(mutations...)

		auditPath, err := tree.ProveMembership(index, index)
		require.NoError(t, err)
		assert.Equalf(t, c.expectedAuditPath, auditPath, "Incorrect audit path for index %d", i)
	}
}

func TestProveMembershipNonConsecutive(t *testing.T) {

	log.SetLogger("TestProveMembershipNonConsecutive", log.INFO)

	store := bplus.NewBPlusTreeStore()
	cache := common.NewPassThroughCache(db.HistoryCachePrefix, store)
	tree := NewHistoryTree(common.NewFakeXorHasher(), cache)

	// add nine events
	for i := uint64(0); i < 9; i++ {
		eventDigest := util.Uint64AsBytes(i)
		_, mutations, _ := tree.Add(eventDigest, i)
		store.Mutate(mutations...)
	}

	// query for membership with event 0 and version 8
	auditPath, err := tree.ProveMembership(0, 8)
	require.NoError(t, err)
	expectedAuditPath := common.AuditPath{"1|0": common.Digest{0x1}, "2|1": common.Digest{0x1}, "4|2": common.Digest{0x0}, "8|0": common.Digest{0x8}}
	assert.Equal(t, expectedAuditPath, auditPath, "Invalid audit path")
}
