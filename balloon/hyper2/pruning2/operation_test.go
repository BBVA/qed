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
					Value: []byte{
						0xe0, 0x00, 0x00, 0x00, // bitmap: 11100000 00000000 00000000 00000000
						0x00, 0x01, // iBatch 0 -> hash=0x00 (shortcut index=0)
						0x00, 0x02, // iBatch 1 -> key=0x00
						0x00, 0x02, // iBatch 2 -> value=0x00
					},
				},
			},
			expectedElements: []*cachedElement{
				{
					Pos: pos(0, 8),
					Value: []byte{
						0xd1, 0x01, 0x00, 0x00, // bitmap: 11010001 00000001 00000000 00000000
						0x00, 0x00, // iBatch 0 -> hash=0x00
						0x00, 0x00, // iBatch 1 -> hash=0x00
						0x00, 0x00, // iBatch 3 -> hash=0x00
						0x00, 0x00, // iBatch 7 -> hash=0x00
						0x00, 0x00, // iBatch 15 -> hash=0x00
					},
				},
			},
		},
		{
			// update index = 0 on tree with only one leaf
			index: []byte{0},
			value: []byte{0},
			cachedBatches: map[string][]byte{
				pos(0, 8).StringId(): []byte{
					0xd1, 0x01, 0x00, 0x00, // bitmap: 11010001 00000001 00000000 00000000
					0x00, 0x00, // iBatch 0 -> hash=0x00
					0x00, 0x00, // iBatch 1 -> hash=0x00
					0x00, 0x00, // iBatch 3 -> hash=0x00
					0x00, 0x00, // iBatch 7 -> hash=0x00
					0x00, 0x00, // iBatch 15 -> hash=0x00
				},
			},
			storedBatches: map[string][]byte{
				pos(0, 4).StringId(): []byte{
					0xe0, 0x00, 0x00, 0x00, // bitmap: 11100000 00000000 00000000 00000000
					0x00, 0x01, // iBatch 0 -> hash=0x00 (shortcut index=0)
					0x00, 0x02, // iBatch 1 -> key=0x00
					0x00, 0x02, // iBatch 2 -> value=0x00
				},
			},
			expectedMutations: []*storage.Mutation{
				{
					Prefix: storage.HyperCachePrefix,
					Key:    pos(0, 4).Bytes(),
					Value: []byte{
						0xe0, 0x00, 0x00, 0x00, // bitmap: 11100000 00000000 00000000 00000000
						0x00, 0x01, // iBatch 0 -> hash=0x00 (shortcut index=0)
						0x00, 0x02, // iBatch 1 -> key=0x00
						0x00, 0x02, // iBatch 2 -> value=0x00
					},
				},
			},
			expectedElements: []*cachedElement{
				{
					Pos: pos(0, 8),
					Value: []byte{
						0xd1, 0x01, 0x00, 0x00, // bitmap: 11010001 00000001 00000000 00000000
						0x00, 0x00, // iBatch 0 -> hash=0x00
						0x00, 0x00, // iBatch 1 -> hash=0x00
						0x00, 0x00, // iBatch 3 -> hash=0x00
						0x00, 0x00, // iBatch 7 -> hash=0x00
						0x00, 0x00, // iBatch 15 -> hash=0x00
					},
				},
			},
		},
		{
			// insert index=1 on tree with 1 leaf (index: 0, value: 0)
			// it should push down the previous leaf to the last level
			index: []byte{1},
			value: []byte{1},
			cachedBatches: map[string][]byte{
				pos(0, 8).StringId(): []byte{
					0xd1, 0x01, 0x00, 0x00, // bitmap: 11010001 00000001 00000000 00000000
					0x00, 0x00, // iBatch 0 -> hash=0x00
					0x00, 0x00, // iBatch 1 -> hash=0x00
					0x00, 0x00, // iBatch 3 -> hash=0x00
					0x00, 0x00, // iBatch 7 -> hash=0x00
					0x00, 0x00, // iBatch 15 -> hash=0x00
				},
			},
			storedBatches: map[string][]byte{
				pos(0, 4).StringId(): []byte{
					0xe0, 0x00, 0x00, 0x00, // bitmap: 11100000 00000000 00000000 00000000
					0x00, 0x01, // iBatch 0 -> hash=0x00 (shortcut index=0)
					0x00, 0x02, // iBatch 1 -> key=0x00
					0x00, 0x02, // iBatch 2 -> value=0x00
				},
			},
			expectedMutations: []*storage.Mutation{
				{
					Prefix: storage.HyperCachePrefix,
					Key:    pos(1, 0).Bytes(),
					Value: []byte{
						0xe0, 0x00, 0x00, 0x00, // bitmap: 11100000 00000000 00000000 00000000
						0x01, 0x01, // iBatch 0 -> hash=0x01 (shortcut index=0)
						0x01, 0x02, // iBatch 1 -> key=0x01
						0x01, 0x02, // iBatch 2 -> value=0x01
					},
				},
				{
					Prefix: storage.HyperCachePrefix,
					Key:    pos(0, 0).Bytes(),
					Value: []byte{
						0xe0, 0x00, 0x00, 0x00, // bitmap: 11100000 00000000 00000000 00000000
						0x00, 0x01, // iBatch 0 -> hash=0x00 (shortcut index=0)
						0x00, 0x02, // iBatch 1 -> key=0x00
						0x00, 0x02, // iBatch 2 -> value=0x00
					},
				},
				{
					Prefix: storage.HyperCachePrefix,
					Key:    pos(0, 4).Bytes(),
					Value: []byte{
						0xd1, 0x01, 0x80, 0x00, // bitmap: 11010001 00000001 10000000 00000000
						0x01, 0x00, // iBatch 0 -> hash=0x01
						0x01, 0x00, // iBatch 1 -> hash=0x01
						0x01, 0x00, // iBatch 3 -> hash=0x01
						0x01, 0x00, // iBatch 7 -> hash=0x01
						0x00, 0x00, // iBatch 15 -> hash=0x00
						0x01, 0x00, // iBatch 16 -> hash=0x01
					},
				},
			},
			expectedElements: []*cachedElement{
				{
					Pos: pos(0, 8),
					Value: []byte{
						0xd1, 0x01, 0x00, 0x00, // bitmap: 11010001 00000001 00000000 00000000
						0x01, 0x00, // iBatch 0 -> hash=0x01
						0x01, 0x00, // iBatch 1 -> hash=0x01
						0x01, 0x00, // iBatch 3 -> hash=0x01
						0x01, 0x00, // iBatch 7 -> hash=0x01
						0x01, 0x00, // iBatch 15 -> hash=0x01
					},
				},
			},
		},
		{
			// insert index=8 on tree with 1 leaf (index: 0, value: 0)
			// it should push down the previous leaf to the next subtree
			index: []byte{8},
			value: []byte{8},
			cachedBatches: map[string][]byte{
				pos(0, 8).StringId(): []byte{
					0xd1, 0x01, 0x00, 0x00, // bitmap: 11010001 00000001 00000000 00000000
					0x00, 0x00, // iBatch 0 -> hash=0x00
					0x00, 0x00, // iBatch 1 -> hash=0x00
					0x00, 0x00, // iBatch 3 -> hash=0x00
					0x00, 0x00, // iBatch 7 -> hash=0x00
					0x00, 0x00, // iBatch 15 -> hash=0x00
				},
			},
			storedBatches: map[string][]byte{
				pos(0, 4).StringId(): []byte{
					0xe0, 0x00, 0x00, 0x00, // bitmap: 11100000 00000000 00000000 00000000
					0x00, 0x01, // iBatch 0 -> hash=0x00 (shortcut index=0)
					0x00, 0x02, // iBatch 1 -> key=0x00
					0x00, 0x02, // iBatch 2 -> value=0x00
				},
			},
			expectedMutations: []*storage.Mutation{
				{
					Prefix: storage.HyperCachePrefix,
					Key:    pos(0, 4).Bytes(),
					Value: []byte{
						0xfe, 0x00, 0x00, 0x00, // bitmap: 11111110 00000000 00000000 00000000
						0x08, 0x00, // iBatch 0 -> hash=0x08
						0x00, 0x01, // iBatch 1 -> hash=0x00 (shortcut index=0)
						0x08, 0x01, // iBatch 2 -> hash=0x08 (shortcut index=8)
						0x00, 0x02, // iBatch 3 -> key=0x00
						0x00, 0x02, // iBatch 4 -> value=0x00
						0x08, 0x02, // iBatch 5 -> key=0x08
						0x08, 0x02, // iBatch 6 -> value=0x08
					},
				},
			},
			expectedElements: []*cachedElement{
				{
					Pos: pos(0, 8),
					Value: []byte{
						0xd1, 0x01, 0x00, 0x00, // bitmap: 11010001 00000001 00000000 00000000
						0x08, 0x00, // iBatch 0 -> hash=0x08
						0x08, 0x00, // iBatch 1 -> hash=0x08
						0x08, 0x00, // iBatch 3 -> hash=0x08
						0x08, 0x00, // iBatch 7 -> hash=0x08
						0x08, 0x00, // iBatch 15 -> hash=0x08
					},
				},
			},
		},
		{
			// insert index=12 on tree with 2 leaves ([index:0, value:0], [index:8, value:8])
			// it should push down the leaf with index=8 to the next level
			index: []byte{12},
			value: []byte{12},
			cachedBatches: map[string][]byte{
				pos(0, 8).StringId(): []byte{
					0xd1, 0x01, 0x00, 0x00, // bitmap: 11010001 00000001 00000000 00000000
					0x08, 0x00, // iBatch 0 -> hash=0x08
					0x08, 0x00, // iBatch 1 -> hash=0x08
					0x08, 0x00, // iBatch 3 -> hash=0x08
					0x08, 0x00, // iBatch 7 -> hash=0x08
					0x08, 0x00, // iBatch 15 -> hash=0x08
				},
			},
			storedBatches: map[string][]byte{
				pos(0, 4).StringId(): []byte{
					0xfe, 0x00, 0x00, 0x00, // bitmap: 11111110 00000000 00000000 00000000
					0x08, 0x00, // iBatch 0 -> hash=0x08
					0x00, 0x01, // iBatch 1 -> hash=0x00 (shortcut index=0)
					0x08, 0x01, // iBatch 2 -> hash=0x08 (shortcut index=8)
					0x00, 0x02, // iBatch 3 -> key=0x00
					0x00, 0x02, // iBatch 4 -> value=0x00
					0x08, 0x02, // iBatch 5 -> key=0x08
					0x08, 0x02, // iBatch 6 -> value=0x08
				},
			},
			expectedMutations: []*storage.Mutation{
				{
					Prefix: storage.HyperCachePrefix,
					Key:    pos(0, 4).Bytes(),
					Value: []byte{
						0xfe, 0x1e, 0x00, 0x00, // bitmap: 11111110 00011110 00000000 00000000
						0x04, 0x00, // iBatch 0 -> hash=0x08
						0x00, 0x01, // iBatch 1 -> hash=0x00 (shortcut index=0)
						0x04, 0x00, // iBatch 2 -> hash=0x04
						0x00, 0x02, // iBatch 3 -> key=0x00
						0x00, 0x02, // iBatch 4 -> value=0x00
						0x08, 0x01, // iBatch 5 -> hash=0x08 (shortcut index=8)
						0x0c, 0x01, // iBatch 6 -> hash=0x0c (shortcut index=12)
						0x08, 0x02, // iBatch 11 -> key=0x08
						0x08, 0x02, // iBatch 12 -> value=0x08
						0x0c, 0x02, // iBatch 13 -> key=0x0c
						0x0c, 0x02, // iBatch 14 -> value=0x0c
					},
				},
			},
			expectedElements: []*cachedElement{
				{
					Pos: pos(0, 8),
					Value: []byte{
						0xd1, 0x01, 0x00, 0x00, // bitmap: 11010001 00000001 00000000 00000000
						0x04, 0x00, // iBatch 0 -> hash=0x04
						0x04, 0x00, // iBatch 1 -> hash=0x04
						0x04, 0x00, // iBatch 3 -> hash=0x04
						0x04, 0x00, // iBatch 7 -> hash=0x04
						0x04, 0x00, // iBatch 15 -> hash=0x04
					},
				},
			},
		},
		{
			// insert index=128 on tree with one leaf ([index:0, value:0]
			index: []byte{128},
			value: []byte{128},
			cachedBatches: map[string][]byte{
				pos(0, 8).StringId(): []byte{
					0xd1, 0x01, 0x00, 0x00, // bitmap: 11010001 00000001 00000000 00000000
					0x00, 0x00, // iBatch 0 -> hash=0x00
					0x00, 0x00, // iBatch 1 -> hash=0x00
					0x00, 0x00, // iBatch 3 -> hash=0x00
					0x00, 0x00, // iBatch 7 -> hash=0x00
					0x00, 0x00, // iBatch 15 -> hash=0x00
				},
			},
			storedBatches: map[string][]byte{
				pos(0, 4).StringId(): []byte{
					0xe0, 0x00, 0x00, 0x00, // bitmap: 11100000 00000000 00000000 00000000
					0x00, 0x01, // iBatch 0 -> hash=0x00 (shortcut index=0)
					0x00, 0x02, // iBatch 1 -> key=0x00
					0x00, 0x02, // iBatch 2 -> value=0x00
				},
			},
			expectedMutations: []*storage.Mutation{
				{
					Prefix: storage.HyperCachePrefix,
					Key:    pos(128, 4).Bytes(),
					Value: []byte{
						0xe0, 0x00, 0x00, 0x00, // bitmap: 11100000 00000000 00000000 00000000
						0x80, 0x01, // iBatch 0 -> hash=0x80 (shortcut index=128)
						0x80, 0x02, // iBatch 1 -> key=0x80
						0x80, 0x02, // iBatch 2 -> value=0x80
					},
				},
			},
			expectedElements: []*cachedElement{
				{
					Pos: pos(0, 8),
					Value: []byte{
						0xf5, 0x11, 0x01, 0x00, // bitmap: 11110101 00010001 00000001 00000000
						0x80, 0x00, // iBatch 0 -> hash=0x80
						0x00, 0x00, // iBatch 1 -> hash=0x00
						0x80, 0x00, // iBatch 2 -> hash=0x80
						0x00, 0x00, // iBatch 3 -> hash=0x00
						0x80, 0x00, // iBatch 5 -> hash=0x80
						0x00, 0x00, // iBatch 7 -> hash=0x00
						0x80, 0x00, // iBatch 11 -> hash=0x80
						0x00, 0x00, // iBatch 15 -> hash=0x00
						0x80, 0x00, // iBatch 23 -> hash=0x80
					},
				},
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
