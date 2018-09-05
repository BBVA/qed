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
		index, version    uint64
		eventDigest       common.Digest
		expectedAuditPath common.AuditPath
	}{
		{
			index:             0,
			version:           0,
			eventDigest:       common.Digest{0x0},
			expectedAuditPath: common.AuditPath{},
		},
		{
			index:             1,
			version:           1,
			eventDigest:       common.Digest{0x1},
			expectedAuditPath: common.AuditPath{"0|0": common.Digest{0x0}},
		},
		{
			index:             2,
			version:           2,
			eventDigest:       common.Digest{0x2},
			expectedAuditPath: common.AuditPath{"0|1": common.Digest{0x1}},
		},
		{
			index:             3,
			version:           3,
			eventDigest:       common.Digest{0x3},
			expectedAuditPath: common.AuditPath{"0|1": common.Digest{0x1}, "2|0": common.Digest{0x2}},
		},
		{
			index:             4,
			version:           4,
			eventDigest:       common.Digest{0x4},
			expectedAuditPath: common.AuditPath{"0|2": common.Digest{0x0}},
		},
		{
			index:             5,
			version:           5,
			eventDigest:       common.Digest{0x5},
			expectedAuditPath: common.AuditPath{"0|2": common.Digest{0x0}, "4|0": common.Digest{0x4}},
		},
		{
			index:             6,
			version:           6,
			eventDigest:       common.Digest{0x6},
			expectedAuditPath: common.AuditPath{"0|2": common.Digest{0x0}, "4|1": common.Digest{0x1}},
		},
		{
			index:             7,
			version:           7,
			eventDigest:       common.Digest{0x7},
			expectedAuditPath: common.AuditPath{"0|2": common.Digest{0x0}, "4|1": common.Digest{0x1}, "6|0": common.Digest{0x6}},
		},
		{
			index:             8,
			version:           8,
			eventDigest:       common.Digest{0x8},
			expectedAuditPath: common.AuditPath{"0|3": common.Digest{0x0}},
		},
		{
			index:             9,
			version:           9,
			eventDigest:       common.Digest{0x9},
			expectedAuditPath: common.AuditPath{"0|3": common.Digest{0x0}, "8|0": common.Digest{0x8}},
		},
		{
			index:             0,
			version:           1,
			eventDigest:       common.Digest{0x0},
			expectedAuditPath: common.AuditPath{"1|0": common.Digest{0x1}},
		},
		{
			index:             0,
			version:           1,
			eventDigest:       common.Digest{0x0},
			expectedAuditPath: common.AuditPath{"1|0": common.Digest{0x1}},
		},
		{
			index:             0,
			version:           2,
			eventDigest:       common.Digest{0x0},
			expectedAuditPath: common.AuditPath{"1|0": common.Digest{0x1}, "2|0": common.Digest{0x2}},
		},
		{
			index:             0,
			version:           3,
			eventDigest:       common.Digest{0x0},
			expectedAuditPath: common.AuditPath{"1|0": common.Digest{0x1}, "2|1": common.Digest{0x1}},
		},
		{
			index:             0,
			version:           4,
			eventDigest:       common.Digest{0x0},
			expectedAuditPath: common.AuditPath{"1|0": common.Digest{0x1}, "2|1": common.Digest{0x1}, "4|0": common.Digest{0x4}},
		},
		{
			index:             0,
			version:           5,
			eventDigest:       common.Digest{0x0},
			expectedAuditPath: common.AuditPath{"1|0": common.Digest{0x1}, "2|1": common.Digest{0x1}, "4|1": common.Digest{0x1}},
		},
		{
			index:             0,
			version:           6,
			eventDigest:       common.Digest{0x0},
			expectedAuditPath: common.AuditPath{"1|0": common.Digest{0x1}, "2|1": common.Digest{0x1}, "4|1": common.Digest{0x1}, "6|0": common.Digest{0x6}},
		},
		{
			index:             0,
			version:           7,
			eventDigest:       common.Digest{0x0},
			expectedAuditPath: common.AuditPath{"1|0": common.Digest{0x1}, "2|1": common.Digest{0x1}, "4|2": common.Digest{0x0}},
		},
	}

	store := bplus.NewBPlusTreeStore()
	cache := common.NewPassThroughCache(db.HistoryCachePrefix, store)
	tree := NewHistoryTree(common.NewFakeXorHasher(), cache)

	for i, c := range testCases {
		_, mutations, _ := tree.Add(c.eventDigest, c.index)
		store.Mutate(mutations...)

		mp, err := tree.ProveMembership(c.index, c.version)
		require.NoError(t, err)
		assert.Equalf(t, c.expectedAuditPath, mp.AuditPath, "Incorrect audit path for index %d", i)
		assert.Equal(t, c.index, mp.Index, "The index should math")
		assert.Equal(t, c.version, mp.Version, "The version should match")
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
	index := uint64(0)
	version := uint64(8)
	expectedAuditPath := common.AuditPath{"1|0": common.Digest{0x1}, "2|1": common.Digest{0x1}, "4|2": common.Digest{0x0}, "8|0": common.Digest{0x8}}
	mp, err := tree.ProveMembership(index, version)

	require.NoError(t, err)
	assert.Equal(t, expectedAuditPath, mp.AuditPath, "Invalid audit path")
	assert.Equal(t, index, mp.Index, "The index should math")
	assert.Equal(t, version, mp.Version, "The version should match")

}
