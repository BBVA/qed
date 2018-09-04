package history

import (
	"testing"

	"github.com/bbva/qed/balloon2/common"
	"github.com/bbva/qed/db"
	"github.com/bbva/qed/db/bplus"
	"github.com/bbva/qed/log"
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
