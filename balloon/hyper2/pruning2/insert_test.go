package pruning2

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPruneToInsert(t *testing.T) {

	testCases := []struct {
		index, value  []byte
		cachedBatches map[string][]byte
		storedBatches map[string][]byte
		expectedOps   []*Operation
	}{
		{
			// insert index = 0 on empty tree
			index:         []byte{0},
			value:         []byte{0},
			cachedBatches: map[string][]byte{},
			storedBatches: map[string][]byte{},
			expectedOps: []*Operation{
				put(pos(0, 8), []byte{0x00, 0x00, 0x00, 0x00}),
				updateNode(pos(0, 8), 0, []byte{0x00, 0x00, 0x00, 0x00}),
				inner(pos(0, 8)),
				getDefault(pos(128, 7)),
				updateNode(pos(0, 7), 1, []byte{0x00, 0x00, 0x00, 0x00}),
				inner(pos(0, 7)),
				getDefault(pos(64, 6)),
				updateNode(pos(0, 6), 3, []byte{0x00, 0x00, 0x00, 0x00}),
				inner(pos(0, 6)),
				getDefault(pos(32, 5)),
				updateNode(pos(0, 5), 7, []byte{0x00, 0x00, 0x00, 0x00}),
				inner(pos(0, 5)),
				getDefault(pos(16, 4)),
				updateNode(pos(0, 4), 15, []byte{0x00, 0x00, 0x00, 0x00}),
				mutate(pos(0, 4), []byte{0x00, 0x00, 0x00, 0x00}),
				updateLeaf(pos(0, 4), 0, []byte{0x00, 0x00, 0x00, 0x00}, []byte{0}, []byte{0}),
				shortcut(pos(0, 4), []byte{0}, []byte{0}),
			},
		},
	}

	batchLevels := uint16(1)
	cacheHeightLimit := batchLevels * 4

	for i, c := range testCases {
		loader := NewFakeBatchLoader(c.cachedBatches, c.storedBatches, cacheHeightLimit)
		prunedOps := PruneToInsert(c.index, c.value, cacheHeightLimit, loader).List()
		require.Truef(t, len(c.expectedOps) == len(prunedOps), "The size of the pruned ops should match the expected for test case %d", i)
		for i := 0; i < len(prunedOps); i++ {
			assert.Equalf(t, c.expectedOps[i].Code, prunedOps[i].Code, "The pruned operation's code should match for test case %d", i)
			assert.Equalf(t, c.expectedOps[i].Pos, prunedOps[i].Pos, "The pruned operation's position should match for test case %d", i)
		}
	}

}
