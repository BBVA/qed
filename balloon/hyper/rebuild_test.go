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

package hyper

import (
	"testing"

	"github.com/bbva/qed/balloon/cache"
	"github.com/bbva/qed/hashing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPruneToRebuild(t *testing.T) {

	testCases := []struct {
		index, serializedBatch []byte
		cachedBatches          map[string][]byte
		expectedOps            []op
	}{
		{
			// insert index = 0 on empty cache
			index: []byte{0},
			serializedBatch: []byte{
				0xe0, 0x00, 0x00, 0x00, // bitmap: 11100000 00000000 00000000 00000000
				0x00, 0x01, // iBatch 0 -> hash=0x00 (shortcut index=0)
				0x00, 0x02, // iBatch 1 -> key=0x00
				0x00, 0x02, // iBatch 2 -> value=0x00
			},
			cachedBatches: map[string][]byte{},
			expectedOps: []op{
				{putInCacheCode, pos(0, 8)},
				{updateBatchNodeCode, pos(0, 8)},
				{innerHashCode, pos(0, 8)},
				{getDefaultHashCode, pos(128, 7)},
				{updateBatchNodeCode, pos(0, 7)},
				{innerHashCode, pos(0, 7)},
				{getDefaultHashCode, pos(64, 6)},
				{updateBatchNodeCode, pos(0, 6)},
				{innerHashCode, pos(0, 6)},
				{getDefaultHashCode, pos(32, 5)},
				{updateBatchNodeCode, pos(0, 5)},
				{innerHashCode, pos(0, 5)},
				{getDefaultHashCode, pos(16, 4)},
				{updateBatchNodeCode, pos(0, 4)},
				{useHashCode, pos(0, 4)},
			},
		},
		{
			// insert index = 1 on cache with one leaf (index=0)
			index: []byte{1},
			serializedBatch: []byte{
				0xd1, 0x01, 0x80, 0x00, // bitmap: 11010001 00000001 10000000 00000000
				0x01, 0x00, // iBatch 0 -> hash=0x01
				0x01, 0x00, // iBatch 1 -> hash=0x01
				0x01, 0x00, // iBatch 3 -> hash=0x01
				0x01, 0x00, // iBatch 7 -> hash=0x01
				0x00, 0x00, // iBatch 15 -> hash=0x00
				0x01, 0x00, // iBatch 16 -> hash=0x01
			},
			cachedBatches: map[string][]byte{
				pos(0, 8).StringId(): []byte{
					0xd1, 0x01, 0x00, 0x00, // bitmap: 11010001 00000001 00000000 00000000
					0x01, 0x00, // iBatch 0 -> hash=0x01
					0x01, 0x00, // iBatch 1 -> hash=0x01
					0x01, 0x00, // iBatch 3 -> hash=0x01
					0x01, 0x00, // iBatch 7 -> hash=0x01
					0x01, 0x00, // iBatch 15 -> hash=0x01
				},
			},
			expectedOps: []op{
				{putInCacheCode, pos(0, 8)},
				{updateBatchNodeCode, pos(0, 8)},
				{innerHashCode, pos(0, 8)},
				{getDefaultHashCode, pos(128, 7)},
				{updateBatchNodeCode, pos(0, 7)},
				{innerHashCode, pos(0, 7)},
				{getDefaultHashCode, pos(64, 6)},
				{updateBatchNodeCode, pos(0, 6)},
				{innerHashCode, pos(0, 6)},
				{getDefaultHashCode, pos(32, 5)},
				{updateBatchNodeCode, pos(0, 5)},
				{innerHashCode, pos(0, 5)},
				{getDefaultHashCode, pos(16, 4)},
				{updateBatchNodeCode, pos(0, 4)},
				{useHashCode, pos(0, 4)},
			},
		},
	}

	batchLevels := uint16(1)
	cacheHeightLimit := batchLevels * 4

	for i, c := range testCases {
		loader := newFakeBatchLoader(c.cachedBatches, nil, cacheHeightLimit)
		prunedOps := pruneToRebuild(c.index, c.serializedBatch, cacheHeightLimit, loader).List()
		require.Truef(t, len(c.expectedOps) == len(prunedOps), "The size of the pruned ops should match the expected for test case %d", i)
		for j := 0; j < len(prunedOps); j++ {
			assert.Equalf(t, c.expectedOps[j].Code, prunedOps[j].Code, "The pruned operation's code should match for test case %d", i)
			assert.Equalf(t, c.expectedOps[j].Pos, prunedOps[j].Pos, "The pruned operation's position should match for test case %d", i)
		}
	}

}

func TestRebuildInterpretation(t *testing.T) {

	testCases := []struct {
		index, serializedBatch []byte
		cachedBatches          map[string][]byte
		expectedElements       []*cachedElement
	}{
		{
			// insert index = 0 on empty cache
			index: []byte{0},
			serializedBatch: []byte{
				0xe0, 0x00, 0x00, 0x00, // bitmap: 11100000 00000000 00000000 00000000
				0x00, 0x01, // iBatch 0 -> hash=0x00 (shortcut index=0)
				0x00, 0x02, // iBatch 1 -> key=0x00
				0x00, 0x02, // iBatch 2 -> value=0x00
			},
			cachedBatches: map[string][]byte{},
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
			index: []byte{1},
			serializedBatch: []byte{
				0xd1, 0x01, 0x80, 0x00, // bitmap: 11010001 00000001 10000000 00000000
				0x01, 0x00, // iBatch 0 -> hash=0x01
				0x01, 0x00, // iBatch 1 -> hash=0x01
				0x01, 0x00, // iBatch 3 -> hash=0x01
				0x01, 0x00, // iBatch 7 -> hash=0x01
				0x00, 0x00, // iBatch 15 -> hash=0x00
				0x01, 0x00, // iBatch 16 -> hash=0x01
			},
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
	}

	batchLevels := uint16(1)
	cacheHeightLimit := batchLevels * 4
	defaultHashes := []hashing.Digest{{0}, {0}, {0}, {0}, {0}, {0}, {0}, {0}, {0}}

	for i, c := range testCases {
		cache := cache.NewFakeCache([]byte{0x0})
		batches := newFakeBatchLoader(c.cachedBatches, nil, cacheHeightLimit)

		ops := pruneToRebuild(c.index, c.serializedBatch, cacheHeightLimit, batches)
		ctx := &pruningContext{
			Hasher:        hashing.NewFakeXorHasher(),
			Cache:         cache,
			DefaultHashes: defaultHashes,
		}

		ops.Pop().Interpret(ops, ctx)

		for _, e := range c.expectedElements {
			v, _ := cache.Get(e.Pos.Bytes())
			assert.Equalf(t, e.Value, v, "The cached element %v should be cached in test case %d", e, i)
		}
	}
}
