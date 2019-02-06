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
	"github.com/stretchr/testify/assert"
)

func TestSearchPruner(t *testing.T) {

	cache := cache.NewFakeCache(hashing.Digest{0x0})

	testCases := []struct {
		version        uint64
		expectedPruned visit.Visitable
	}{
		{
			version:        0,
			expectedPruned: leafnil(pos(0, 0)),
		},
		{
			version: 1,
			expectedPruned: node(pos(0, 1),
				collectable(cached(pos(0, 0))),
				leafnil(pos(1, 0)),
			),
		},
		{
			version: 2,
			expectedPruned: node(pos(0, 2),
				collectable(cached(pos(0, 1))),
				partialnode(pos(2, 1),
					leafnil(pos(2, 0))),
			),
		},
		{
			version: 3,
			expectedPruned: node(pos(0, 2),
				collectable(cached(pos(0, 1))),
				node(pos(2, 1),
					collectable(cached(pos(2, 0))),
					leafnil(pos(3, 0))),
			),
		},
		{
			version: 4,
			expectedPruned: node(pos(0, 3),
				collectable(cached(pos(0, 2))),
				partialnode(pos(4, 2),
					partialnode(pos(4, 1),
						leafnil(pos(4, 0)))),
			),
		},
		{
			version: 5,
			expectedPruned: node(pos(0, 3),
				collectable(cached(pos(0, 2))),
				partialnode(pos(4, 2),
					node(pos(4, 1),
						collectable(cached(pos(4, 0))),
						leafnil(pos(5, 0)))),
			),
		},
		{
			version: 6,
			expectedPruned: node(pos(0, 3),
				collectable(cached(pos(0, 2))),
				node(pos(4, 2),
					collectable(cached(pos(4, 1))),
					partialnode(pos(6, 1),
						leafnil(pos(6, 0)))),
			),
		},
		{
			version: 7,
			expectedPruned: node(pos(0, 3),
				collectable(cached(pos(0, 2))),
				node(pos(4, 2),
					collectable(cached(pos(4, 1))),
					node(pos(6, 1),
						collectable(cached(pos(6, 0))),
						leafnil(pos(7, 0)))),
			),
		},
	}

	for i, c := range testCases {
		context := NewPruningContext(NewSingleTargetedCacheResolver(c.version), cache)
		pruned, _ := NewSearchPruner(c.version, context).Prune()
		assert.Equalf(t, c.expectedPruned, pruned, "The pruned trees should match for test case %d", i)
	}

}

func TestSearchPrunerConsistency(t *testing.T) {

	cache := cache.NewFakeCache(hashing.Digest{0x0})

	testCases := []struct {
		index, version uint64
		expectedPruned visit.Visitable
	}{
		{
			index:          0,
			version:        0,
			expectedPruned: leafnil(pos(0, 0)),
		},
		{
			index:   0,
			version: 1,
			expectedPruned: node(pos(0, 1),
				leafnil(pos(0, 0)),
				collectable(cached(pos(1, 0))),
			),
		},
		{
			index:   0,
			version: 2,
			expectedPruned: node(pos(0, 2),
				node(pos(0, 1),
					leafnil(pos(0, 0)),
					collectable(cached(pos(1, 0)))),
				partialnode(pos(2, 1),
					collectable(cached(pos(2, 0)))),
			),
		},
		{
			index:   0,
			version: 3,
			expectedPruned: node(pos(0, 2),
				node(pos(0, 1),
					leafnil(pos(0, 0)),
					collectable(cached(pos(1, 0)))),
				collectable(cached(pos(2, 1))),
			),
		},
		{
			index:   0,
			version: 4,
			expectedPruned: node(pos(0, 3),
				node(pos(0, 2),
					node(pos(0, 1),
						leafnil(pos(0, 0)),
						collectable(cached(pos(1, 0)))),
					collectable(cached(pos(2, 1)))),
				partialnode(pos(4, 2),
					partialnode(pos(4, 1),
						collectable(cached(pos(4, 0))))),
			),
		},
		{
			index:   0,
			version: 5,
			expectedPruned: node(pos(0, 3),
				node(pos(0, 2),
					node(pos(0, 1),
						leafnil(pos(0, 0)),
						collectable(cached(pos(1, 0)))),
					collectable(cached(pos(2, 1)))),
				partialnode(pos(4, 2),
					collectable(cached(pos(4, 1)))),
			),
		},
		{
			index:   0,
			version: 6,
			expectedPruned: node(pos(0, 3),
				node(pos(0, 2),
					node(pos(0, 1),
						leafnil(pos(0, 0)),
						collectable(cached(pos(1, 0)))),
					collectable(cached(pos(2, 1)))),
				node(pos(4, 2),
					collectable(cached(pos(4, 1))),
					partialnode(pos(6, 1),
						collectable(cached(pos(6, 0))))),
			),
		},
		{
			index:   0,
			version: 7,
			expectedPruned: node(pos(0, 3),
				node(pos(0, 2),
					node(pos(0, 1),
						leafnil(pos(0, 0)),
						collectable(cached(pos(1, 0)))),
					collectable(cached(pos(2, 1)))),
				collectable(cached(pos(4, 2))),
			),
		},
	}

	for i, c := range testCases {
		context := NewPruningContext(NewDoubleTargetedCacheResolver(c.index, c.version), cache)
		pruned, _ := NewSearchPruner(c.version, context).Prune()
		assert.Equalf(t, c.expectedPruned, pruned, "The pruned trees should match for test case %d", i)
	}

}

func TestSearchPrunerIncremental(t *testing.T) {

	cache := cache.NewFakeCache(hashing.Digest{0x0})

	testCases := []struct {
		start, end     uint64
		expectedPruned visit.Visitable
	}{
		{
			start:          0,
			end:            0,
			expectedPruned: collectable(cached(pos(0, 0))),
		},
		{
			start: 0,
			end:   1,
			expectedPruned: node(pos(0, 1),
				collectable(cached(pos(0, 0))),
				collectable(cached(pos(1, 0))),
			),
		},
		{
			start: 0,
			end:   2,
			expectedPruned: node(pos(0, 2),
				node(pos(0, 1),
					collectable(cached(pos(0, 0))),
					collectable(cached(pos(1, 0)))),
				partialnode(pos(2, 1),
					collectable(cached(pos(2, 0)))),
			),
		},
		{
			start: 0,
			end:   3,
			expectedPruned: node(pos(0, 2),
				node(pos(0, 1),
					collectable(cached(pos(0, 0))),
					collectable(cached(pos(1, 0)))),
				collectable(cached(pos(2, 1))),
			),
		},
		{
			start: 0,
			end:   4,
			expectedPruned: node(pos(0, 3),
				node(pos(0, 2),
					node(pos(0, 1),
						collectable(cached(pos(0, 0))),
						collectable(cached(pos(1, 0)))),
					collectable(cached(pos(2, 1)))),
				partialnode(pos(4, 2),
					partialnode(pos(4, 1),
						collectable(cached(pos(4, 0))))),
			),
		},
		{
			start: 0,
			end:   5,
			expectedPruned: node(pos(0, 3),
				node(pos(0, 2),
					node(pos(0, 1),
						collectable(cached(pos(0, 0))),
						collectable(cached(pos(1, 0)))),
					collectable(cached(pos(2, 1)))),
				partialnode(pos(4, 2),
					collectable(cached(pos(4, 1)))),
			),
		},
		{
			start: 0,
			end:   6,
			expectedPruned: node(pos(0, 3),
				node(pos(0, 2),
					node(pos(0, 1),
						collectable(cached(pos(0, 0))),
						collectable(cached(pos(1, 0)))),
					collectable(cached(pos(2, 1)))),
				node(pos(4, 2),
					collectable(cached(pos(4, 1))),
					partialnode(pos(6, 1),
						collectable(cached(pos(6, 0))))),
			),
		},
		{
			start: 0,
			end:   7,
			expectedPruned: node(pos(0, 3),
				node(pos(0, 2),
					node(pos(0, 1),
						collectable(cached(pos(0, 0))),
						collectable(cached(pos(1, 0)))),
					collectable(cached(pos(2, 1)))),
				collectable(cached(pos(4, 2))),
			),
		},
	}

	for i, c := range testCases {
		context := NewPruningContext(NewIncrementalCacheResolver(c.start, c.end), cache)
		pruned, _ := NewSearchPruner(c.end, context).Prune()
		assert.Equalf(t, c.expectedPruned, pruned, "The pruned trees should match for test case %d", i)
	}

}

func BenchmarkSearchPruner(b *testing.B) {

	log.SetLogger("BenchmarkSearchPruner", log.SILENT)

	cache := cache.NewFakeCache(hashing.Digest{0x0})

	b.ResetTimer()
	for i := uint64(0); i < uint64(b.N); i++ {
		context := NewPruningContext(NewDoubleTargetedCacheResolver(0, i), cache)
		_, err := NewSearchPruner(i, context).Prune()
		assert.NoError(b, err)
	}

}
