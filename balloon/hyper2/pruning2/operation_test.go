package pruning2

import (
	"testing"

	"github.com/bbva/qed/balloon/cache"
	"github.com/bbva/qed/balloon/hyper2/navigation"
	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/storage"
	"github.com/stretchr/testify/assert"
)

func TestInsertInterpretation(t *testing.T) {

	testCases := []struct {
		index, value      []byte
		cachedBatches     map[string][]byte
		storedBatches     map[string][]byte
		expectedMutations []*storage.Mutation
		expectedElements  []*cachedElement
	}{
		{
			// insert index = 0 on empty tree
			index:         []byte{0},
			value:         []byte{0},
			cachedBatches: map[string][]byte{},
			storedBatches: map[string][]byte{},
			expectedMutations: []*storage.Mutation{
				{
					Prefix: storage.HyperCachePrefix,
					Key:    pos(0, 4).Bytes(),
					Value:  []byte{0xe0, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x02, 0x00, 0x02},
				},
			},
			expectedElements: []*cachedElement{
				newCachedElement(pos(0, 8), []byte{0xd1, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}),
			},
		},
	}

	batchLevels := uint16(1)
	cacheHeightLimit := batchLevels * 4
	defaultHashes := []hashing.Digest{{0}, {0}, {0}, {0}, {0}, {0}, {0}, {0}, {0}}

	for i, c := range testCases {
		cache := cache.NewFakeCache([]byte{0x0})
		batches := NewFakeBatchLoader(c.cachedBatches, c.storedBatches, cacheHeightLimit)

		ops := PruneToInsert(c.index, c.value, cacheHeightLimit, batches)
		ctx := &Context{
			Hasher:        hashing.NewFakeXorHasher(),
			Cache:         cache,
			DefaultHashes: defaultHashes,
			Mutations:     make([]*storage.Mutation, 0),
		}

		ops.Pop().Interpret(ops, ctx)

		assert.ElementsMatchf(t, c.expectedMutations, ctx.Mutations, "Mutation error in test case %d", i)
		for _, e := range c.expectedElements {
			v, _ := cache.Get(e.Pos.Bytes())
			assert.Equalf(t, e.Value, v, "The cached element %v should be cached in test case %d", e, i)
		}
	}

}

type cachedElement struct {
	Pos   navigation.Position
	Value []byte
}

func newCachedElement(pos navigation.Position, value []byte) *cachedElement {
	return &cachedElement{pos, value}
}
