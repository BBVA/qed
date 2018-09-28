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

	assert "github.com/stretchr/testify/require"

	"github.com/bbva/qed/balloon/cache"
	"github.com/bbva/qed/balloon/navigator"
	"github.com/bbva/qed/balloon/visitor"
	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/storage"
	storage_utils "github.com/bbva/qed/testutils/storage"
)

var (
	FixedDigest = make([]byte, 8)
)

func pos(index byte, height uint16) navigator.Position {
	return NewPosition([]byte{index}, height)
}

func root(pos navigator.Position, left, right visitor.Visitable) *visitor.Root {
	return visitor.NewRoot(pos, left, right)
}

func node(pos navigator.Position, left, right visitor.Visitable) *visitor.Node {
	return visitor.NewNode(pos, left, right)
}

func leaf(pos navigator.Position, value byte) *visitor.Leaf {
	return visitor.NewLeaf(pos, []byte{value})
}

func cached(pos navigator.Position) *visitor.Cached {
	return visitor.NewCached(pos, hashing.Digest{0})
}

func collectable(underlying visitor.Visitable) *visitor.Collectable {
	return visitor.NewCollectable(underlying)
}

func cacheable(underlying visitor.Visitable) *visitor.Cacheable {
	return visitor.NewCacheable(underlying)
}

func TestInsertPruner(t *testing.T) {

	numBits := uint16(8)
	cacheLevel := uint16(4)

	testCases := []struct {
		key, value     []byte
		storeMutations []*storage.Mutation
		expectedPruned visitor.Visitable
	}{
		{
			key:            []byte{0},
			value:          []byte{0},
			storeMutations: []*storage.Mutation{},
			expectedPruned: root(pos(0, 8),
				cacheable(node(pos(0, 7),
					cacheable(node(pos(0, 6),
						collectable(cacheable(node(pos(0, 5),
							node(pos(0, 4),
								node(pos(0, 3),
									node(pos(0, 2),
										node(pos(0, 1),
											leaf(pos(0, 0), 0),
											cached(pos(1, 0))),
										cached(pos(2, 1))),
									cached(pos(4, 2))),
								cached(pos(8, 3))),
							cached(pos(16, 4))))),
						cached(pos(32, 5)))),
					cached(pos(64, 6)))),
				cached(pos(128, 7)),
			),
		},
		{
			key:   []byte{2},
			value: []byte{1},
			storeMutations: []*storage.Mutation{
				{storage.IndexPrefix, []byte{0}, []byte{0}},
			},
			expectedPruned: root(pos(0, 8),
				cacheable(node(pos(0, 7),
					cacheable(node(pos(0, 6),
						collectable(cacheable(node(pos(0, 5),
							node(pos(0, 4),
								node(pos(0, 3),
									node(pos(0, 2),
										node(pos(0, 1),
											leaf(pos(0, 0), 0),
											cached(pos(1, 0))),
										node(pos(2, 1),
											leaf(pos(2, 0), 1),
											cached(pos(3, 0)))),
									cached(pos(4, 2))),
								cached(pos(8, 3))),
							cached(pos(16, 4))))),
						cached(pos(32, 5)))),
					cached(pos(64, 6)))),
				cached(pos(128, 7)),
			),
		},
		{
			key:   []byte{255},
			value: []byte{2},
			storeMutations: []*storage.Mutation{
				{storage.IndexPrefix, []byte{0}, []byte{0}},
				{storage.IndexPrefix, []byte{2}, []byte{1}},
			},
			expectedPruned: root(pos(0, 8),
				cached(pos(0, 7)),
				cacheable(node(pos(128, 7),
					cached(pos(128, 6)),
					cacheable(node(pos(192, 6),
						cached(pos(192, 5)),
						collectable(cacheable(node(pos(224, 5),
							cached(pos(224, 4)),
							node(pos(240, 4),
								cached(pos(240, 3)),
								node(pos(248, 3),
									cached(pos(248, 2)),
									node(pos(252, 2),
										cached(pos(252, 1)),
										node(pos(254, 1),
											cached(pos(254, 0)),
											leaf(pos(255, 0), 2)))))))))),
				)),
			),
		},
	}

	for i, c := range testCases {
		store, closeF := storage_utils.OpenBPlusTreeStore()
		defer closeF()
		store.Mutate(c.storeMutations)

		cache := cache.NewSimpleCache(4)

		context := PruningContext{
			navigator:     NewHyperTreeNavigator(numBits),
			cacheResolver: NewSingleTargetedCacheResolver(numBits, cacheLevel, c.key),
			cache:         cache,
			store:         store,
			defaultHashes: []hashing.Digest{
				hashing.Digest{0}, hashing.Digest{0}, hashing.Digest{0}, hashing.Digest{0}, hashing.Digest{0},
				hashing.Digest{0}, hashing.Digest{0}, hashing.Digest{0}, hashing.Digest{0},
			},
		}

		pruned := NewInsertPruner(c.key, c.value, context).Prune()
		assert.Equalf(t, c.expectedPruned, pruned, "The pruned trees should match for test case %d", i)
	}
}

func TestSearchPruner(t *testing.T) {

	numBits := uint16(8)
	cacheLevel := uint16(4)

	testCases := []struct {
		key            []byte
		storeMutations []*storage.Mutation
		expectedPruned visitor.Visitable
	}{
		{
			key: []byte{0},
			storeMutations: []*storage.Mutation{
				{storage.IndexPrefix, []byte{0}, []byte{0}},
			},
			expectedPruned: root(pos(0, 8),
				node(pos(0, 7),
					node(pos(0, 6),
						node(pos(0, 5),
							node(pos(0, 4),
								node(pos(0, 3),
									node(pos(0, 2),
										node(pos(0, 1),
											leaf(pos(0, 0), 0),
											collectable(cached(pos(1, 0)))),
										collectable(cached(pos(2, 1)))),
									collectable(cached(pos(4, 2)))),
								collectable(cached(pos(8, 3)))),
							collectable(cached(pos(16, 4)))),
						collectable(cached(pos(32, 5)))),
					collectable(cached(pos(64, 6)))),
				collectable(cached(pos(128, 7))),
			),
		},
		{
			key: []byte{6},
			storeMutations: []*storage.Mutation{
				{storage.IndexPrefix, []byte{1}, []byte{1}},
				{storage.IndexPrefix, []byte{6}, []byte{6}},
			},
			expectedPruned: root(pos(0, 8),
				node(pos(0, 7),
					node(pos(0, 6),
						node(pos(0, 5),
							node(pos(0, 4),
								node(pos(0, 3),
									collectable(node(pos(0, 2),
										node(pos(0, 1),
											cached(pos(0, 0)),
											leaf(pos(1, 0), 1)),
										cached(pos(2, 1)))),
									node(pos(4, 2),
										collectable(cached(pos(4, 1))),
										node(pos(6, 1),
											leaf(pos(6, 0), 6),
											collectable(cached(pos(7, 0)))))),
								collectable(cached(pos(8, 3)))),
							collectable(cached(pos(16, 4)))),
						collectable(cached(pos(32, 5)))),
					collectable(cached(pos(64, 6)))),
				collectable(cached(pos(128, 7)))),
		},
	}

	for i, c := range testCases {
		store, closeF := storage_utils.OpenBPlusTreeStore()
		defer closeF()
		store.Mutate(c.storeMutations)

		cache := cache.NewSimpleCache(4)

		context := PruningContext{
			navigator:     NewHyperTreeNavigator(numBits),
			cacheResolver: NewSingleTargetedCacheResolver(numBits, cacheLevel, c.key),
			cache:         cache,
			store:         store,
			defaultHashes: []hashing.Digest{
				hashing.Digest{0}, hashing.Digest{0}, hashing.Digest{0}, hashing.Digest{0}, hashing.Digest{0},
				hashing.Digest{0}, hashing.Digest{0}, hashing.Digest{0}, hashing.Digest{0},
			},
		}

		pruned := NewSearchPruner(c.key, context).Prune()
		assert.Equalf(t, c.expectedPruned, pruned, "The pruned trees should match for test case %d", i)
	}
}
func TestVerifyPruner(t *testing.T) {

	numBits := uint16(8)
	cacheLevel := uint16(4)

	fakeCache := cache.NewFakeCache(hashing.Digest{0}) // Always return hashing.Digest{0}
	// Add element before verifying.
	store, closeF := storage_utils.OpenBPlusTreeStore()
	defer closeF()
	mutations := []*storage.Mutation{
		{storage.IndexPrefix, []byte{0}, []byte{0}},
	}
	store.Mutate(mutations)

	testCases := []struct {
		key, value     []byte
		expectedPruned visitor.Visitable
	}{
		{
			key:   []byte{0},
			value: []byte{0},
			expectedPruned: root(pos(0, 8),
				node(pos(0, 7),
					node(pos(0, 6),
						node(pos(0, 5),
							node(pos(0, 4),
								node(pos(0, 3),
									node(pos(0, 2),
										node(pos(0, 1),
											leaf(pos(0, 0), 0),
											cached(pos(1, 0))),
										cached(pos(2, 1))),
									cached(pos(4, 2))),
								cached(pos(8, 3))),
							cached(pos(16, 4))),
						cached(pos(32, 5))),
					cached(pos(64, 6))),
				cached(pos(128, 7))),
		},
	}

	for i, c := range testCases {
		context := PruningContext{
			navigator:     NewHyperTreeNavigator(numBits),
			cacheResolver: NewSingleTargetedCacheResolver(numBits, cacheLevel, c.key),
			cache:         fakeCache,
			store:         store,
			defaultHashes: []hashing.Digest{
				hashing.Digest{0}, hashing.Digest{0}, hashing.Digest{0}, hashing.Digest{0}, hashing.Digest{0},
				hashing.Digest{0}, hashing.Digest{0}, hashing.Digest{0}, hashing.Digest{0},
			},
		}

		pruned := NewVerifyPruner(c.key, c.value, context).Prune()
		assert.Equalf(t, c.expectedPruned, pruned, "The pruned trees should match for test case %d", i)
	}
}
