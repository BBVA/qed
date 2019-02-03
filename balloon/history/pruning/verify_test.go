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

	"github.com/bbva/qed/balloon/history/visit"

	"github.com/bbva/qed/balloon/cache"
	"github.com/bbva/qed/hashing"
	"github.com/stretchr/testify/assert"
)

func TestVerifyPruner(t *testing.T) {

	cache := cache.NewFakeCache(hashing.Digest{0x0})

	testCases := []struct {
		index, version uint64
		eventDigest    hashing.Digest
		expectedPruned visit.Visitable
	}{
		{
			index:          0,
			version:        0,
			eventDigest:    hashing.Digest{0x0},
			expectedPruned: leaf(pos(0, 0), 0),
		},
		{
			index:       0,
			version:     1,
			eventDigest: hashing.Digest{0x0},
			expectedPruned: node(pos(0, 1),
				leaf(pos(0, 0), 0),
				cached(pos(1, 0))),
		},
		{
			index:       1,
			version:     1,
			eventDigest: hashing.Digest{0x1},
			expectedPruned: node(pos(0, 1),
				cached(pos(0, 0)),
				leaf(pos(1, 0), 1)),
		},
		{
			index:       1,
			version:     2,
			eventDigest: hashing.Digest{0x1},
			expectedPruned: node(pos(0, 2),
				node(pos(0, 1),
					cached(pos(0, 0)),
					leaf(pos(1, 0), 1)),
				partialnode(pos(2, 1),
					cached(pos(2, 0)))),
		},
		{
			index:       6,
			version:     6,
			eventDigest: hashing.Digest{0x6},
			expectedPruned: node(pos(0, 3),
				cached(pos(0, 2)),
				node(pos(4, 2),
					cached(pos(4, 1)),
					partialnode(pos(6, 1),
						leaf(pos(6, 0), 6)))),
		},
		{
			index:       1,
			version:     7,
			eventDigest: hashing.Digest{0x1},
			expectedPruned: node(pos(0, 3),
				node(pos(0, 2),
					node(pos(0, 1),
						cached(pos(0, 0)),
						leaf(pos(1, 0), 1)),
					cached(pos(2, 1))),
				cached(pos(4, 2))),
		},
	}

	for i, c := range testCases {

		var cacheResolver CacheResolver
		if c.index == c.version {
			cacheResolver = NewSingleTargetedCacheResolver(c.version)
		} else {
			cacheResolver = NewDoubleTargetedCacheResolver(c.index, c.version)
		}
		context := NewPruningContext(cacheResolver, cache)
		pruned, _ := NewVerifyPruner(c.version, c.eventDigest, context).Prune()
		assert.Equalf(t, c.expectedPruned, pruned, "The pruned trees should match for test case %d", i)
	}

}

func TestVerifyPrunerIncremental(t *testing.T) {

	cache := cache.NewFakeCache(hashing.Digest{0x0})

	testCases := []struct {
		start, end     uint64
		expectedPruned visit.Visitable
	}{
		{
			start:          0,
			end:            0,
			expectedPruned: cached(pos(0, 0)),
		},
		{
			start: 0,
			end:   1,
			expectedPruned: node(pos(0, 1),
				cached(pos(0, 0)),
				cached(pos(1, 0)),
			),
		},
		{
			start: 0,
			end:   2,
			expectedPruned: node(pos(0, 2),
				node(pos(0, 1),
					cached(pos(0, 0)),
					cached(pos(1, 0))),
				partialnode(pos(2, 1),
					cached(pos(2, 0))),
			),
		},
		{
			start: 0,
			end:   3,
			expectedPruned: node(pos(0, 2),
				node(pos(0, 1),
					cached(pos(0, 0)),
					cached(pos(1, 0))),
				cached(pos(2, 1)),
			),
		},
		{
			start: 0,
			end:   4,
			expectedPruned: node(pos(0, 3),
				node(pos(0, 2),
					node(pos(0, 1),
						cached(pos(0, 0)),
						cached(pos(1, 0))),
					cached(pos(2, 1))),
				partialnode(pos(4, 2),
					partialnode(pos(4, 1),
						cached(pos(4, 0)))),
			),
		},
		{
			start: 0,
			end:   5,
			expectedPruned: node(pos(0, 3),
				node(pos(0, 2),
					node(pos(0, 1),
						cached(pos(0, 0)),
						cached(pos(1, 0))),
					cached(pos(2, 1))),
				partialnode(pos(4, 2),
					cached(pos(4, 1))),
			),
		},
		{
			start: 0,
			end:   6,
			expectedPruned: node(pos(0, 3),
				node(pos(0, 2),
					node(pos(0, 1),
						cached(pos(0, 0)),
						cached(pos(1, 0))),
					cached(pos(2, 1))),
				node(pos(4, 2),
					cached(pos(4, 1)),
					partialnode(pos(6, 1),
						cached(pos(6, 0)))),
			),
		},
		{
			start: 0,
			end:   7,
			expectedPruned: node(pos(0, 3),
				node(pos(0, 2),
					node(pos(0, 1),
						cached(pos(0, 0)),
						cached(pos(1, 0))),
					cached(pos(2, 1))),
				cached(pos(4, 2)),
			),
		},
	}

	for i, c := range testCases {
		context := NewPruningContext(NewIncrementalCacheResolver(c.start, c.end), cache)
		pruned, _ := NewVerifyIncrementalPruner(c.end, context).Prune()
		assert.Equalf(t, c.expectedPruned, pruned, "The pruned trees should match for test case %d", i)
	}

}
