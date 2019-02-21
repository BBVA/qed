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

package pruning

import (
	"testing"

	"github.com/bbva/qed/balloon/cache"
	"github.com/bbva/qed/balloon/hyper2/navigation"
	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPruneToInsert(t *testing.T) {

	testCases := []struct {
		index, value  []byte
		cachedBatches map[string][]byte
		storedBatches map[string][]byte
		expectedOps   []op
	}{
		{
			// insert index = 0 on empty tree
			index:         []byte{0},
			value:         []byte{0},
			cachedBatches: map[string][]byte{},
			storedBatches: map[string][]byte{},
			expectedOps: []op{
				{PutInCacheCode, pos(0, 8)},
				{UpdateBatchNodeCode, pos(0, 8)},
				{InnerHashCode, pos(0, 8)},
				{GetDefaultHashCode, pos(128, 7)},
				{UpdateBatchNodeCode, pos(0, 7)},
				{InnerHashCode, pos(0, 7)},
				{GetDefaultHashCode, pos(64, 6)},
				{UpdateBatchNodeCode, pos(0, 6)},
				{InnerHashCode, pos(0, 6)},
				{GetDefaultHashCode, pos(32, 5)},
				{UpdateBatchNodeCode, pos(0, 5)},
				{InnerHashCode, pos(0, 5)},
				{GetDefaultHashCode, pos(16, 4)},
				{UpdateBatchNodeCode, pos(0, 4)},
				{MutateBatchCode, pos(0, 4)},
				{UpdateBatchShortcutCode, pos(0, 4)},
				{LeafHashCode, pos(0, 4)},
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
			expectedOps: []op{
				{PutInCacheCode, pos(0, 8)},
				{UpdateBatchNodeCode, pos(0, 8)},
				{InnerHashCode, pos(0, 8)},
				{GetDefaultHashCode, pos(128, 7)},
				{UpdateBatchNodeCode, pos(0, 7)},
				{InnerHashCode, pos(0, 7)},
				{GetDefaultHashCode, pos(64, 6)},
				{UpdateBatchNodeCode, pos(0, 6)},
				{InnerHashCode, pos(0, 6)},
				{GetDefaultHashCode, pos(32, 5)},
				{UpdateBatchNodeCode, pos(0, 5)},
				{InnerHashCode, pos(0, 5)},
				{GetDefaultHashCode, pos(16, 4)},
				{UpdateBatchNodeCode, pos(0, 4)},
				{MutateBatchCode, pos(0, 4)},
				{UpdateBatchShortcutCode, pos(0, 4)},
				{LeafHashCode, pos(0, 4)},
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
			expectedOps: []op{
				{PutInCacheCode, pos(0, 8)},
				{UpdateBatchNodeCode, pos(0, 8)},
				{InnerHashCode, pos(0, 8)},
				{GetDefaultHashCode, pos(128, 7)},
				{UpdateBatchNodeCode, pos(0, 7)},
				{InnerHashCode, pos(0, 7)},
				{GetDefaultHashCode, pos(64, 6)},
				{UpdateBatchNodeCode, pos(0, 6)},
				{InnerHashCode, pos(0, 6)},
				{GetDefaultHashCode, pos(32, 5)},
				{UpdateBatchNodeCode, pos(0, 5)},
				{InnerHashCode, pos(0, 5)},
				{GetDefaultHashCode, pos(16, 4)},
				{UpdateBatchNodeCode, pos(0, 4)},
				{MutateBatchCode, pos(0, 4)}, // reset previous shortcut
				{UpdateBatchNodeCode, pos(0, 4)},
				{InnerHashCode, pos(0, 4)},
				{GetDefaultHashCode, pos(8, 3)},
				{UpdateBatchNodeCode, pos(0, 3)},
				{InnerHashCode, pos(0, 3)},
				{GetDefaultHashCode, pos(4, 2)},
				{UpdateBatchNodeCode, pos(0, 2)},
				{InnerHashCode, pos(0, 2)},
				{GetDefaultHashCode, pos(2, 1)},
				{UpdateBatchNodeCode, pos(0, 1)},
				{InnerHashCode, pos(0, 1)},
				{UpdateBatchNodeCode, pos(1, 0)},
				{MutateBatchCode, pos(1, 0)}, // new batch
				{UpdateBatchShortcutCode, pos(1, 0)},
				{LeafHashCode, pos(1, 0)},
				{UpdateBatchNodeCode, pos(0, 0)},
				{MutateBatchCode, pos(0, 0)}, // new batch
				{UpdateBatchShortcutCode, pos(0, 0)},
				{LeafHashCode, pos(0, 0)},
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
			expectedOps: []op{
				{PutInCacheCode, pos(0, 8)},
				{UpdateBatchNodeCode, pos(0, 8)},
				{InnerHashCode, pos(0, 8)},
				{GetDefaultHashCode, pos(128, 7)},
				{UpdateBatchNodeCode, pos(0, 7)},
				{InnerHashCode, pos(0, 7)},
				{GetDefaultHashCode, pos(64, 6)},
				{UpdateBatchNodeCode, pos(0, 6)},
				{InnerHashCode, pos(0, 6)},
				{GetDefaultHashCode, pos(32, 5)},
				{UpdateBatchNodeCode, pos(0, 5)},
				{InnerHashCode, pos(0, 5)},
				{GetDefaultHashCode, pos(16, 4)},
				{UpdateBatchNodeCode, pos(0, 4)},
				{MutateBatchCode, pos(0, 4)}, // reset previous shortcut
				{UpdateBatchNodeCode, pos(0, 4)},
				{InnerHashCode, pos(0, 4)},
				{UpdateBatchShortcutCode, pos(8, 3)},
				{LeafHashCode, pos(8, 3)},
				{UpdateBatchShortcutCode, pos(0, 3)},
				{LeafHashCode, pos(0, 3)},
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
			expectedOps: []op{
				{PutInCacheCode, pos(0, 8)},
				{UpdateBatchNodeCode, pos(0, 8)},
				{InnerHashCode, pos(0, 8)},
				{GetDefaultHashCode, pos(128, 7)},
				{UpdateBatchNodeCode, pos(0, 7)},
				{InnerHashCode, pos(0, 7)},
				{GetDefaultHashCode, pos(64, 6)},
				{UpdateBatchNodeCode, pos(0, 6)},
				{InnerHashCode, pos(0, 6)},
				{GetDefaultHashCode, pos(32, 5)},
				{UpdateBatchNodeCode, pos(0, 5)},
				{InnerHashCode, pos(0, 5)},
				{GetDefaultHashCode, pos(16, 4)},
				{UpdateBatchNodeCode, pos(0, 4)},
				{MutateBatchCode, pos(0, 4)},
				{UpdateBatchNodeCode, pos(0, 4)},
				{InnerHashCode, pos(0, 4)},
				{UpdateBatchNodeCode, pos(8, 3)},
				{InnerHashCode, pos(8, 3)},
				{UpdateBatchShortcutCode, pos(12, 2)},
				{LeafHashCode, pos(12, 2)},
				{UpdateBatchShortcutCode, pos(8, 2)},
				{LeafHashCode, pos(8, 2)},
				{GetProvidedHashCode, pos(0, 3)},
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
			expectedOps: []op{
				{PutInCacheCode, pos(0, 8)},
				{UpdateBatchNodeCode, pos(0, 8)},
				{InnerHashCode, pos(0, 8)},
				{UpdateBatchNodeCode, pos(128, 7)},
				{InnerHashCode, pos(128, 7)},
				{GetDefaultHashCode, pos(192, 6)},
				{UpdateBatchNodeCode, pos(128, 6)},
				{InnerHashCode, pos(128, 6)},
				{GetDefaultHashCode, pos(160, 5)},
				{UpdateBatchNodeCode, pos(128, 5)},
				{InnerHashCode, pos(128, 5)},
				{GetDefaultHashCode, pos(144, 4)},
				{UpdateBatchNodeCode, pos(128, 4)},
				{MutateBatchCode, pos(128, 4)},
				{UpdateBatchShortcutCode, pos(128, 4)},
				{LeafHashCode, pos(128, 4)},
				{GetProvidedHashCode, pos(0, 7)},
			},
		},
	}

	batchLevels := uint16(1)
	cacheHeightLimit := batchLevels * 4

	for i, c := range testCases {
		loader := NewFakeBatchLoader(c.cachedBatches, c.storedBatches, cacheHeightLimit)
		prunedOps := PruneToInsert(c.index, c.value, cacheHeightLimit, loader).List()
		require.Truef(t, len(c.expectedOps) == len(prunedOps), "The size of the pruned ops should match the expected for test case %d", i)
		for j := 0; j < len(prunedOps); j++ {
			assert.Equalf(t, c.expectedOps[j].Code, prunedOps[j].Code, "The pruned operation's code should match for test case %d", i)
			assert.Equalf(t, c.expectedOps[j].Pos, prunedOps[j].Pos, "The pruned operation's position should match for test case %d", i)
		}
	}

}

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
