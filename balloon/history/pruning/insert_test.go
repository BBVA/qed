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
	"github.com/bbva/qed/balloon/history/visit"
	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/log"
	"github.com/bbva/qed/testutils/rand"
	"github.com/stretchr/testify/assert"
)

func TestInsertPruner(t *testing.T) {

	cache := cache.NewFakeCache(hashing.Digest{0x0})

	testCases := []struct {
		version        uint64
		eventDigest    hashing.Digest
		expectedPruned visit.Visitable
	}{
		{
			version:        0,
			eventDigest:    hashing.Digest{0x0},
			expectedPruned: mutable(cacheable(leaf(pos(0, 0), 0))),
		},
		{
			version:     1,
			eventDigest: hashing.Digest{0x1},
			expectedPruned: mutable(cacheable(node(pos(0, 1),
				cached(pos(0, 0)),
				mutable(cacheable(leaf(pos(1, 0), 1))))),
			),
		},
		{
			version:     2,
			eventDigest: hashing.Digest{0x2},
			expectedPruned: node(pos(0, 2),
				cached(pos(0, 1)),
				partialnode(pos(2, 1),
					mutable(cacheable(leaf(pos(2, 0), 2)))),
			),
		},
		{
			version:     3,
			eventDigest: hashing.Digest{0x3},
			expectedPruned: mutable(cacheable(node(pos(0, 2),
				cached(pos(0, 1)),
				mutable(cacheable(node(pos(2, 1),
					cached(pos(2, 0)),
					mutable(cacheable(leaf(pos(3, 0), 3)))))))),
			),
		},
		{
			version:     4,
			eventDigest: hashing.Digest{0x4},
			expectedPruned: node(pos(0, 3),
				cached(pos(0, 2)),
				partialnode(pos(4, 2),
					partialnode(pos(4, 1),
						mutable(cacheable(leaf(pos(4, 0), 4))))),
			),
		},
		{
			version:     5,
			eventDigest: hashing.Digest{0x5},
			expectedPruned: node(pos(0, 3),
				cached(pos(0, 2)),
				partialnode(pos(4, 2),
					mutable(cacheable(node(pos(4, 1),
						cached(pos(4, 0)),
						mutable(cacheable(leaf(pos(5, 0), 5))))))),
			),
		},
		{
			version:     6,
			eventDigest: hashing.Digest{0x6},
			expectedPruned: node(pos(0, 3),
				cached(pos(0, 2)),
				node(pos(4, 2),
					cached(pos(4, 1)),
					partialnode(pos(6, 1),
						mutable(cacheable(leaf(pos(6, 0), 6))))),
			),
		},
		{
			version:     7,
			eventDigest: hashing.Digest{0x7},
			expectedPruned: mutable(cacheable(node(pos(0, 3),
				cached(pos(0, 2)),
				mutable(cacheable(node(pos(4, 2),
					cached(pos(4, 1)),
					mutable(cacheable(node(pos(6, 1),
						cached(pos(6, 0)),
						mutable(cacheable(leaf(pos(7, 0), 7))))))))))),
			),
		},
	}

	for i, c := range testCases {
		context := NewPruningContext(NewSingleTargetedCacheResolver(c.version), cache)
		pruned, _ := NewInsertPruner(c.version, c.eventDigest, context).Prune()
		assert.Equalf(t, c.expectedPruned, pruned, "The pruned trees should match for test case %d", i)
	}

}

func BenchmarkInsertPruner(b *testing.B) {

	log.SetLogger("BenchmarkInsertPruner", log.SILENT)

	cache := cache.NewFakeCache(hashing.Digest{0x0})

	b.ResetTimer()
	for i := uint64(0); i < uint64(b.N); i++ {
		eventDigest := rand.Bytes(32)
		context := NewPruningContext(NewSingleTargetedCacheResolver(i), cache)
		_, err := NewInsertPruner(i, eventDigest, context).Prune()
		assert.NoError(b, err)
	}

}
