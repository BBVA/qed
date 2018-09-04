package history

import (
	"testing"

	"github.com/bbva/qed/balloon2/common"
	"github.com/stretchr/testify/assert"
)

func TestInsertPruner(t *testing.T) {

	cache := NewFakeCache(common.Digest{0x0})

	testCases := []struct {
		version        uint64
		eventDigest    common.Digest
		expectedPruned common.Visitable
	}{
		{
			version:     0,
			eventDigest: common.Digest{0x0},
			expectedPruned: common.NewCollectable(NewPosition(0, 0),
				common.NewLeaf(NewPosition(0, 0), common.Digest{0x0})),
		},
		{
			version:     1,
			eventDigest: common.Digest{0x1},
			expectedPruned: common.NewCollectable(NewPosition(0, 1),
				common.NewRoot(NewPosition(0, 1),
					common.NewCached(NewPosition(0, 0), cache.FixedDigest),
					common.NewCollectable(NewPosition(1, 0),
						common.NewLeaf(NewPosition(1, 0), common.Digest{0x1}))),
			),
		},
		{
			version:     2,
			eventDigest: common.Digest{0x2},
			expectedPruned: common.NewRoot(NewPosition(0, 2),
				common.NewCached(NewPosition(0, 1), cache.FixedDigest),
				common.NewPartialNode(NewPosition(2, 1),
					common.NewCollectable(NewPosition(2, 0),
						common.NewLeaf(NewPosition(2, 0), common.Digest{0x2}))),
			),
		},
		{
			version:     3,
			eventDigest: common.Digest{0x3},
			expectedPruned: common.NewCollectable(NewPosition(0, 2),
				common.NewRoot(NewPosition(0, 2),
					common.NewCached(NewPosition(0, 1), cache.FixedDigest),
					common.NewNode(NewPosition(2, 1),
						common.NewCached(NewPosition(2, 0), cache.FixedDigest),
						common.NewCollectable(NewPosition(3, 0),
							common.NewLeaf(NewPosition(3, 0), common.Digest{0x3})))),
			),
		},
		{
			version:     4,
			eventDigest: common.Digest{0x4},
			expectedPruned: common.NewRoot(NewPosition(0, 3),
				common.NewCached(NewPosition(0, 2), cache.FixedDigest),
				common.NewPartialNode(NewPosition(4, 2),
					common.NewPartialNode(NewPosition(4, 1),
						common.NewCollectable(NewPosition(4, 0),
							common.NewLeaf(NewPosition(4, 0), common.Digest{0x4})))),
			),
		},
		{
			version:     5,
			eventDigest: common.Digest{0x5},
			expectedPruned: common.NewRoot(NewPosition(0, 3),
				common.NewCached(NewPosition(0, 2), cache.FixedDigest),
				common.NewPartialNode(NewPosition(4, 2),
					common.NewNode(NewPosition(4, 1),
						common.NewCached(NewPosition(4, 0), cache.FixedDigest),
						common.NewCollectable(NewPosition(5, 0),
							common.NewLeaf(NewPosition(5, 0), common.Digest{0x5})))),
			),
		},
		{
			version:     6,
			eventDigest: common.Digest{0x6},
			expectedPruned: common.NewRoot(NewPosition(0, 3),
				common.NewCached(NewPosition(0, 2), cache.FixedDigest),
				common.NewNode(NewPosition(4, 2),
					common.NewCached(NewPosition(4, 1), cache.FixedDigest),
					common.NewPartialNode(NewPosition(6, 1),
						common.NewCollectable(NewPosition(6, 0),
							common.NewLeaf(NewPosition(6, 0), common.Digest{0x6})))),
			),
		},
		{
			version:     7,
			eventDigest: common.Digest{0x7},
			expectedPruned: common.NewCollectable(NewPosition(0, 3),
				common.NewRoot(NewPosition(0, 3),
					common.NewCached(NewPosition(0, 2), cache.FixedDigest),
					common.NewNode(NewPosition(4, 2),
						common.NewCached(NewPosition(4, 1), cache.FixedDigest),
						common.NewNode(NewPosition(6, 1),
							common.NewCached(NewPosition(6, 0), cache.FixedDigest),
							common.NewCollectable(NewPosition(7, 0),
								common.NewLeaf(NewPosition(7, 0), common.Digest{0x7}))))),
			),
		},
	}

	for i, c := range testCases {
		context := PruningContext{
			navigator:     NewHistoryTreeNavigator(c.version),
			cacheResolver: NewSingleTargetedCacheResolver(c.version),
			cache:         cache,
		}

		pruned := NewInsertPruner(c.version, c.eventDigest, context).Prune()

		assert.Equalf(t, c.expectedPruned, pruned, "The pruned trees should match for test case %d", i)
	}

}

func TestSearchPruner(t *testing.T) {

	cache := NewFakeCache(common.Digest{0x0})

	testCases := []struct {
		version        uint64
		expectedPruned common.Visitable
	}{
		{
			version:        0,
			expectedPruned: common.NewLeaf(NewPosition(0, 0), nil),
		},
		{
			version: 1,
			expectedPruned: common.NewRoot(NewPosition(0, 1),
				common.NewCollectable(NewPosition(0, 0),
					common.NewCached(NewPosition(0, 0), cache.FixedDigest)),
				common.NewLeaf(NewPosition(1, 0), nil),
			),
		},
		{
			version: 2,
			expectedPruned: common.NewRoot(NewPosition(0, 2),
				common.NewCollectable(NewPosition(0, 1),
					common.NewCached(NewPosition(0, 1), cache.FixedDigest)),
				common.NewPartialNode(NewPosition(2, 1),
					common.NewLeaf(NewPosition(2, 0), nil)),
			),
		},
		{
			version: 3,
			expectedPruned: common.NewRoot(NewPosition(0, 2),
				common.NewCollectable(NewPosition(0, 1),
					common.NewCached(NewPosition(0, 1), cache.FixedDigest)),
				common.NewNode(NewPosition(2, 1),
					common.NewCollectable(NewPosition(2, 0),
						common.NewCached(NewPosition(2, 0), cache.FixedDigest)),
					common.NewLeaf(NewPosition(3, 0), nil)),
			),
		},
		{
			version: 4,
			expectedPruned: common.NewRoot(NewPosition(0, 3),
				common.NewCollectable(NewPosition(0, 2),
					common.NewCached(NewPosition(0, 2), cache.FixedDigest)),
				common.NewPartialNode(NewPosition(4, 2),
					common.NewPartialNode(NewPosition(4, 1),
						common.NewLeaf(NewPosition(4, 0), nil))),
			),
		},
		{
			version: 5,
			expectedPruned: common.NewRoot(NewPosition(0, 3),
				common.NewCollectable(NewPosition(0, 2),
					common.NewCached(NewPosition(0, 2), cache.FixedDigest)),
				common.NewPartialNode(NewPosition(4, 2),
					common.NewNode(NewPosition(4, 1),
						common.NewCollectable(NewPosition(4, 0),
							common.NewCached(NewPosition(4, 0), cache.FixedDigest)),
						common.NewLeaf(NewPosition(5, 0), nil))),
			),
		},
		{
			version: 6,
			expectedPruned: common.NewRoot(NewPosition(0, 3),
				common.NewCollectable(NewPosition(0, 2),
					common.NewCached(NewPosition(0, 2), cache.FixedDigest)),
				common.NewNode(NewPosition(4, 2),
					common.NewCollectable(NewPosition(4, 1),
						common.NewCached(NewPosition(4, 1), cache.FixedDigest)),
					common.NewPartialNode(NewPosition(6, 1),
						common.NewLeaf(NewPosition(6, 0), nil))),
			),
		},
		{
			version: 7,
			expectedPruned: common.NewRoot(NewPosition(0, 3),
				common.NewCollectable(NewPosition(0, 2),
					common.NewCached(NewPosition(0, 2), cache.FixedDigest)),
				common.NewNode(NewPosition(4, 2),
					common.NewCollectable(NewPosition(4, 1),
						common.NewCached(NewPosition(4, 1), cache.FixedDigest)),
					common.NewNode(NewPosition(6, 1),
						common.NewCollectable(NewPosition(6, 0),
							common.NewCached(NewPosition(6, 0), cache.FixedDigest)),
						common.NewLeaf(NewPosition(7, 0), nil))),
			),
		},
	}

	for i, c := range testCases {
		context := PruningContext{
			navigator:     NewHistoryTreeNavigator(c.version),
			cacheResolver: NewSingleTargetedCacheResolver(c.version),
			cache:         cache,
		}

		pruned := NewSearchPruner(context).Prune()

		assert.Equalf(t, c.expectedPruned, pruned, "The pruned trees should match for test case %d", i)
	}

}

func TestSearchPrunerConsistency(t *testing.T) {

	cache := NewFakeCache(common.Digest{0x0})

	testCases := []struct {
		index, version uint64
		expectedPruned common.Visitable
	}{
		{
			index:          0,
			version:        0,
			expectedPruned: common.NewLeaf(NewPosition(0, 0), nil),
		},
		{
			index:   0,
			version: 1,
			expectedPruned: common.NewRoot(NewPosition(0, 1),
				common.NewLeaf(NewPosition(0, 0), nil),
				common.NewCollectable(NewPosition(1, 0),
					common.NewCached(NewPosition(1, 0), cache.FixedDigest)),
			),
		},
		{
			index:   0,
			version: 2,
			expectedPruned: common.NewRoot(NewPosition(0, 2),
				common.NewNode(NewPosition(0, 1),
					common.NewLeaf(NewPosition(0, 0), nil),
					common.NewCollectable(NewPosition(1, 0),
						common.NewCached(NewPosition(1, 0), cache.FixedDigest))),
				common.NewPartialNode(NewPosition(2, 1),
					common.NewCollectable(NewPosition(2, 0),
						common.NewCached(NewPosition(2, 0), cache.FixedDigest))),
			),
		},
		{
			index:   0,
			version: 3,
			expectedPruned: common.NewRoot(NewPosition(0, 2),
				common.NewNode(NewPosition(0, 1),
					common.NewLeaf(NewPosition(0, 0), nil),
					common.NewCollectable(NewPosition(1, 0),
						common.NewCached(NewPosition(1, 0), cache.FixedDigest))),
				common.NewCollectable(NewPosition(2, 1),
					common.NewCached(NewPosition(2, 1), cache.FixedDigest)),
			),
		},
		{
			index:   0,
			version: 4,
			expectedPruned: common.NewRoot(NewPosition(0, 3),
				common.NewNode(NewPosition(0, 2),
					common.NewNode(NewPosition(0, 1),
						common.NewLeaf(NewPosition(0, 0), nil),
						common.NewCollectable(NewPosition(1, 0),
							common.NewCached(NewPosition(1, 0), cache.FixedDigest))),
					common.NewCollectable(NewPosition(2, 1),
						common.NewCached(NewPosition(2, 1), cache.FixedDigest))),
				common.NewPartialNode(NewPosition(4, 2),
					common.NewPartialNode(NewPosition(4, 1),
						common.NewCollectable(NewPosition(4, 0),
							common.NewCached(NewPosition(4, 0), cache.FixedDigest)))),
			),
		},
		{
			index:   0,
			version: 5,
			expectedPruned: common.NewRoot(NewPosition(0, 3),
				common.NewNode(NewPosition(0, 2),
					common.NewNode(NewPosition(0, 1),
						common.NewLeaf(NewPosition(0, 0), nil),
						common.NewCollectable(NewPosition(1, 0),
							common.NewCached(NewPosition(1, 0), cache.FixedDigest))),
					common.NewCollectable(NewPosition(2, 1),
						common.NewCached(NewPosition(2, 1), cache.FixedDigest))),
				common.NewPartialNode(NewPosition(4, 2),
					common.NewCollectable(NewPosition(4, 1),
						common.NewCached(NewPosition(4, 1), cache.FixedDigest))),
			),
		},
		{
			index:   0,
			version: 6,
			expectedPruned: common.NewRoot(NewPosition(0, 3),
				common.NewNode(NewPosition(0, 2),
					common.NewNode(NewPosition(0, 1),
						common.NewLeaf(NewPosition(0, 0), nil),
						common.NewCollectable(NewPosition(1, 0),
							common.NewCached(NewPosition(1, 0), cache.FixedDigest))),
					common.NewCollectable(NewPosition(2, 1),
						common.NewCached(NewPosition(2, 1), cache.FixedDigest))),
				common.NewNode(NewPosition(4, 2),
					common.NewCollectable(NewPosition(4, 1),
						common.NewCached(NewPosition(4, 1), cache.FixedDigest)),
					common.NewPartialNode(NewPosition(6, 1),
						common.NewCollectable(NewPosition(6, 0),
							common.NewCached(NewPosition(6, 0), cache.FixedDigest)))),
			),
		},
		{
			index:   0,
			version: 7,
			expectedPruned: common.NewRoot(NewPosition(0, 3),
				common.NewNode(NewPosition(0, 2),
					common.NewNode(NewPosition(0, 1),
						common.NewLeaf(NewPosition(0, 0), nil),
						common.NewCollectable(NewPosition(1, 0),
							common.NewCached(NewPosition(1, 0), cache.FixedDigest))),
					common.NewCollectable(NewPosition(2, 1),
						common.NewCached(NewPosition(2, 1), cache.FixedDigest))),
				common.NewCollectable(NewPosition(4, 2),
					common.NewCached(NewPosition(4, 2), cache.FixedDigest)),
			),
		},
	}

	for i, c := range testCases {
		context := PruningContext{
			navigator:     NewHistoryTreeNavigator(c.version),
			cacheResolver: NewDoubleTargetedCacheResolver(c.index, c.version),
			cache:         cache,
		}

		pruned := NewSearchPruner(context).Prune()

		assert.Equalf(t, c.expectedPruned, pruned, "The pruned trees should match for test case %d", i)
	}

}

func TestSearchPrunerIncremental(t *testing.T) {

	cache := NewFakeCache(common.Digest{0x0})

	testCases := []struct {
		start, end     uint64
		expectedPruned common.Visitable
	}{
		{
			start: 0,
			end:   0,
			expectedPruned: common.NewCollectable(NewPosition(0, 0),
				common.NewCached(NewPosition(0, 0), cache.FixedDigest)),
		},
		{
			start: 0,
			end:   1,
			expectedPruned: common.NewRoot(NewPosition(0, 1),
				common.NewCollectable(NewPosition(0, 0),
					common.NewCached(NewPosition(0, 0), cache.FixedDigest)),
				common.NewCollectable(NewPosition(1, 0),
					common.NewCached(NewPosition(1, 0), cache.FixedDigest)),
			),
		},
		{
			start: 0,
			end:   2,
			expectedPruned: common.NewRoot(NewPosition(0, 2),
				common.NewNode(NewPosition(0, 1),
					common.NewCollectable(NewPosition(0, 0),
						common.NewCached(NewPosition(0, 0), cache.FixedDigest)),
					common.NewCollectable(NewPosition(1, 0),
						common.NewCached(NewPosition(1, 0), cache.FixedDigest))),
				common.NewPartialNode(NewPosition(2, 1),
					common.NewCollectable(NewPosition(2, 0),
						common.NewCached(NewPosition(2, 0), cache.FixedDigest))),
			),
		},
		{
			start: 0,
			end:   3,
			expectedPruned: common.NewRoot(NewPosition(0, 2),
				common.NewNode(NewPosition(0, 1),
					common.NewCollectable(NewPosition(0, 0),
						common.NewCached(NewPosition(0, 0), cache.FixedDigest)),
					common.NewCollectable(NewPosition(1, 0),
						common.NewCached(NewPosition(1, 0), cache.FixedDigest))),
				common.NewCollectable(NewPosition(2, 1),
					common.NewCached(NewPosition(2, 1), cache.FixedDigest)),
			),
		},
		{
			start: 0,
			end:   4,
			expectedPruned: common.NewRoot(NewPosition(0, 3),
				common.NewNode(NewPosition(0, 2),
					common.NewNode(NewPosition(0, 1),
						common.NewCollectable(NewPosition(0, 0),
							common.NewCached(NewPosition(0, 0), cache.FixedDigest)),
						common.NewCollectable(NewPosition(1, 0),
							common.NewCached(NewPosition(1, 0), cache.FixedDigest))),
					common.NewCollectable(NewPosition(2, 1),
						common.NewCached(NewPosition(2, 1), cache.FixedDigest))),
				common.NewPartialNode(NewPosition(4, 2),
					common.NewPartialNode(NewPosition(4, 1),
						common.NewCollectable(NewPosition(4, 0),
							common.NewCached(NewPosition(4, 0), cache.FixedDigest)))),
			),
		},
		{
			start: 0,
			end:   5,
			expectedPruned: common.NewRoot(NewPosition(0, 3),
				common.NewNode(NewPosition(0, 2),
					common.NewNode(NewPosition(0, 1),
						common.NewCollectable(NewPosition(0, 0),
							common.NewCached(NewPosition(0, 0), cache.FixedDigest)),
						common.NewCollectable(NewPosition(1, 0),
							common.NewCached(NewPosition(1, 0), cache.FixedDigest))),
					common.NewCollectable(NewPosition(2, 1),
						common.NewCached(NewPosition(2, 1), cache.FixedDigest))),
				common.NewPartialNode(NewPosition(4, 2),
					common.NewCollectable(NewPosition(4, 1),
						common.NewCached(NewPosition(4, 1), cache.FixedDigest))),
			),
		},
		{
			start: 0,
			end:   6,
			expectedPruned: common.NewRoot(NewPosition(0, 3),
				common.NewNode(NewPosition(0, 2),
					common.NewNode(NewPosition(0, 1),
						common.NewCollectable(NewPosition(0, 0),
							common.NewCached(NewPosition(0, 0), cache.FixedDigest)),
						common.NewCollectable(NewPosition(1, 0),
							common.NewCached(NewPosition(1, 0), cache.FixedDigest))),
					common.NewCollectable(NewPosition(2, 1),
						common.NewCached(NewPosition(2, 1), cache.FixedDigest))),
				common.NewNode(NewPosition(4, 2),
					common.NewCollectable(NewPosition(4, 1),
						common.NewCached(NewPosition(4, 1), cache.FixedDigest)),
					common.NewPartialNode(NewPosition(6, 1),
						common.NewCollectable(NewPosition(6, 0),
							common.NewCached(NewPosition(6, 0), cache.FixedDigest)))),
			),
		},
		{
			start: 0,
			end:   7,
			expectedPruned: common.NewRoot(NewPosition(0, 3),
				common.NewNode(NewPosition(0, 2),
					common.NewNode(NewPosition(0, 1),
						common.NewCollectable(NewPosition(0, 0),
							common.NewCached(NewPosition(0, 0), cache.FixedDigest)),
						common.NewCollectable(NewPosition(1, 0),
							common.NewCached(NewPosition(1, 0), cache.FixedDigest))),
					common.NewCollectable(NewPosition(2, 1),
						common.NewCached(NewPosition(2, 1), cache.FixedDigest))),
				common.NewCollectable(NewPosition(4, 2),
					common.NewCached(NewPosition(4, 2), cache.FixedDigest)),
			),
		},
	}

	for i, c := range testCases {
		context := PruningContext{
			navigator:     NewHistoryTreeNavigator(c.end),
			cacheResolver: NewIncrementalCacheResolver(c.start, c.end),
			cache:         cache,
		}

		pruned := NewSearchPruner(context).Prune()

		assert.Equalf(t, c.expectedPruned, pruned, "The pruned trees should match for test case %d", i)
	}

}

func TestVerifyPruner(t *testing.T) {

	cache := NewFakeCache(common.Digest{0x0})

	testCases := []struct {
		index, version uint64
		eventDigest    common.Digest
		expectedPruned common.Visitable
	}{
		{
			index:          0,
			version:        0,
			eventDigest:    common.Digest{0x0},
			expectedPruned: common.NewLeaf(NewPosition(0, 0), common.Digest{0x0}),
		},
		{
			index:       0,
			version:     1,
			eventDigest: common.Digest{0x0},
			expectedPruned: common.NewRoot(NewPosition(0, 1),
				common.NewLeaf(NewPosition(0, 0), common.Digest{0x0}),
				common.NewCached(NewPosition(1, 0), cache.FixedDigest)),
		},
		{
			index:       1,
			version:     1,
			eventDigest: common.Digest{0x1},
			expectedPruned: common.NewRoot(NewPosition(0, 1),
				common.NewCached(NewPosition(0, 0), cache.FixedDigest),
				common.NewLeaf(NewPosition(1, 0), common.Digest{0x1})),
		},
		{
			index:       1,
			version:     2,
			eventDigest: common.Digest{0x1},
			expectedPruned: common.NewRoot(NewPosition(0, 2),
				common.NewNode(NewPosition(0, 1),
					common.NewCached(NewPosition(0, 0), cache.FixedDigest),
					common.NewLeaf(NewPosition(1, 0), common.Digest{0x1})),
				common.NewPartialNode(NewPosition(2, 1),
					common.NewCached(NewPosition(2, 0), cache.FixedDigest))),
		},
		{
			index:       6,
			version:     6,
			eventDigest: common.Digest{0x6},
			expectedPruned: common.NewRoot(NewPosition(0, 3),
				common.NewCached(NewPosition(0, 2), cache.FixedDigest),
				common.NewNode(NewPosition(4, 2),
					common.NewCached(NewPosition(4, 1), cache.FixedDigest),
					common.NewPartialNode(NewPosition(6, 1),
						common.NewLeaf(NewPosition(6, 0), common.Digest{0x6})))),
		},
		{
			index:       1,
			version:     7,
			eventDigest: common.Digest{0x1},
			expectedPruned: common.NewRoot(NewPosition(0, 3),
				common.NewNode(NewPosition(0, 2),
					common.NewNode(NewPosition(0, 1),
						common.NewCached(NewPosition(0, 0), cache.FixedDigest),
						common.NewLeaf(NewPosition(1, 0), common.Digest{0x1})),
					common.NewCached(NewPosition(2, 1), cache.FixedDigest)),
				common.NewCached(NewPosition(4, 2), cache.FixedDigest)),
		},
	}

	for i, c := range testCases {

		var cacheResolver CacheResolver
		if c.index == c.version {
			cacheResolver = NewSingleTargetedCacheResolver(c.version)
		} else {
			cacheResolver = NewDoubleTargetedCacheResolver(c.index, c.version)
		}
		context := PruningContext{
			navigator:     NewHistoryTreeNavigator(c.version),
			cacheResolver: cacheResolver,
			cache:         cache,
		}

		pruned := NewVerifyPruner(c.eventDigest, context).Prune()

		assert.Equalf(t, c.expectedPruned, pruned, "The pruned trees should match for test case %d", i)
	}

}

func TestVerifyPrunerIncremental(t *testing.T) {

	cache := NewFakeCache(common.Digest{0x0})

	testCases := []struct {
		start, end     uint64
		expectedPruned common.Visitable
	}{
		{
			start:          0,
			end:            0,
			expectedPruned: common.NewCached(NewPosition(0, 0), cache.FixedDigest),
		},
		{
			start: 0,
			end:   1,
			expectedPruned: common.NewRoot(NewPosition(0, 1),
				common.NewCached(NewPosition(0, 0), cache.FixedDigest),
				common.NewCached(NewPosition(1, 0), cache.FixedDigest),
			),
		},
		{
			start: 0,
			end:   2,
			expectedPruned: common.NewRoot(NewPosition(0, 2),
				common.NewNode(NewPosition(0, 1),
					common.NewCached(NewPosition(0, 0), cache.FixedDigest),
					common.NewCached(NewPosition(1, 0), cache.FixedDigest)),
				common.NewPartialNode(NewPosition(2, 1),
					common.NewCached(NewPosition(2, 0), cache.FixedDigest)),
			),
		},
		{
			start: 0,
			end:   3,
			expectedPruned: common.NewRoot(NewPosition(0, 2),
				common.NewNode(NewPosition(0, 1),
					common.NewCached(NewPosition(0, 0), cache.FixedDigest),
					common.NewCached(NewPosition(1, 0), cache.FixedDigest)),
				common.NewCached(NewPosition(2, 1), cache.FixedDigest),
			),
		},
		{
			start: 0,
			end:   4,
			expectedPruned: common.NewRoot(NewPosition(0, 3),
				common.NewNode(NewPosition(0, 2),
					common.NewNode(NewPosition(0, 1),
						common.NewCached(NewPosition(0, 0), cache.FixedDigest),
						common.NewCached(NewPosition(1, 0), cache.FixedDigest)),
					common.NewCached(NewPosition(2, 1), cache.FixedDigest)),
				common.NewPartialNode(NewPosition(4, 2),
					common.NewPartialNode(NewPosition(4, 1),
						common.NewCached(NewPosition(4, 0), cache.FixedDigest))),
			),
		},
		{
			start: 0,
			end:   5,
			expectedPruned: common.NewRoot(NewPosition(0, 3),
				common.NewNode(NewPosition(0, 2),
					common.NewNode(NewPosition(0, 1),
						common.NewCached(NewPosition(0, 0), cache.FixedDigest),
						common.NewCached(NewPosition(1, 0), cache.FixedDigest)),
					common.NewCached(NewPosition(2, 1), cache.FixedDigest)),
				common.NewPartialNode(NewPosition(4, 2),
					common.NewCached(NewPosition(4, 1), cache.FixedDigest)),
			),
		},
		{
			start: 0,
			end:   6,
			expectedPruned: common.NewRoot(NewPosition(0, 3),
				common.NewNode(NewPosition(0, 2),
					common.NewNode(NewPosition(0, 1),
						common.NewCached(NewPosition(0, 0), cache.FixedDigest),
						common.NewCached(NewPosition(1, 0), cache.FixedDigest)),
					common.NewCached(NewPosition(2, 1), cache.FixedDigest)),
				common.NewNode(NewPosition(4, 2),
					common.NewCached(NewPosition(4, 1), cache.FixedDigest),
					common.NewPartialNode(NewPosition(6, 1),
						common.NewCached(NewPosition(6, 0), cache.FixedDigest))),
			),
		},
		{
			start: 0,
			end:   7,
			expectedPruned: common.NewRoot(NewPosition(0, 3),
				common.NewNode(NewPosition(0, 2),
					common.NewNode(NewPosition(0, 1),
						common.NewCached(NewPosition(0, 0), cache.FixedDigest),
						common.NewCached(NewPosition(1, 0), cache.FixedDigest)),
					common.NewCached(NewPosition(2, 1), cache.FixedDigest)),
				common.NewCached(NewPosition(4, 2), cache.FixedDigest),
			),
		},
	}

	for i, c := range testCases {
		context := PruningContext{
			navigator:     NewHistoryTreeNavigator(c.end),
			cacheResolver: NewIncrementalCacheResolver(c.start, c.end),
			cache:         cache,
		}

		pruned := NewVerifyIncrementalPruner(context).Prune()

		assert.Equalf(t, c.expectedPruned, pruned, "The pruned trees should match for test case %d", i)
	}

}

type FakeCache struct {
	FixedDigest common.Digest
}

func NewFakeCache(fixedDigest common.Digest) *FakeCache {
	return &FakeCache{fixedDigest}
}

func (c FakeCache) Get(common.Position) (common.Digest, bool) {
	return common.Digest{0x0}, true
}
