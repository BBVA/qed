package hyper

import (
	"testing"

	assert "github.com/stretchr/testify/require"

	"github.com/bbva/qed/balloon/common"
	"github.com/bbva/qed/db"
	"github.com/bbva/qed/db/bplus"
)

var (
	FixedDigest = make([]byte, 8)
)

func pos(index byte, height uint16) common.Position {
	return NewPosition([]byte{index}, height)
}

func root(pos common.Position, left, right common.Visitable) *common.Root {
	return common.NewRoot(pos, left, right)
}

func node(pos common.Position, left, right common.Visitable) *common.Node {
	return common.NewNode(pos, left, right)
}

func leaf(pos common.Position, value byte) *common.Leaf {
	return common.NewLeaf(pos, []byte{value})
}

func cached(pos common.Position) *common.Cached {
	return common.NewCached(pos, common.Digest{0})
}

func collectable(underlying common.Visitable) *common.Collectable {
	return common.NewCollectable(underlying)
}

func TestInsertPruner(t *testing.T) {

	numBits := uint16(8)
	cacheLevel := uint16(4)

	testCases := []struct {
		key, value     []byte
		storeMutations []db.Mutation
		expectedPruned common.Visitable
	}{
		{
			key:            []byte{0},
			value:          []byte{0},
			storeMutations: []db.Mutation{},
			expectedPruned: root(pos(0, 8),
				collectable(node(pos(0, 7),
					collectable(node(pos(0, 6),
						collectable(node(pos(0, 5),
							node(pos(0, 4),
								node(pos(0, 3),
									node(pos(0, 2),
										node(pos(0, 1),
											leaf(pos(0, 0), 0),
											cached(pos(1, 0))),
										cached(pos(2, 1))),
									cached(pos(4, 2))),
								cached(pos(8, 3))),
							cached(pos(16, 4)))),
						cached(pos(32, 5)))),
					cached(pos(64, 6)))),
				cached(pos(128, 7)),
			),
		},
		{
			key:   []byte{2},
			value: []byte{1},
			storeMutations: []db.Mutation{
				*db.NewMutation(db.IndexPrefix, []byte{0}, []byte{0}),
			},
			expectedPruned: root(pos(0, 8),
				collectable(node(pos(0, 7),
					collectable(node(pos(0, 6),
						collectable(node(pos(0, 5),
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
							cached(pos(16, 4)))),
						cached(pos(32, 5)))),
					cached(pos(64, 6)))),
				cached(pos(128, 7)),
			),
		},
		{
			key:   []byte{255},
			value: []byte{2},
			storeMutations: []db.Mutation{
				*db.NewMutation(db.IndexPrefix, []byte{0}, []byte{0}),
				*db.NewMutation(db.IndexPrefix, []byte{2}, []byte{1}),
			},
			expectedPruned: root(pos(0, 8),
				cached(pos(0, 7)),
				collectable(node(pos(128, 7),
					cached(pos(128, 6)),
					collectable(node(pos(192, 6),
						cached(pos(192, 5)),
						collectable(node(pos(224, 5),
							cached(pos(224, 4)),
							node(pos(240, 4),
								cached(pos(240, 3)),
								node(pos(248, 3),
									cached(pos(248, 2)),
									node(pos(252, 2),
										cached(pos(252, 1)),
										node(pos(254, 1),
											cached(pos(254, 0)),
											leaf(pos(255, 0), 2))))))))),
				),
				),
			),
		},
	}

	for i, c := range testCases {
		store := bplus.NewBPlusTreeStore()
		store.Mutate(c.storeMutations...)

		cache := common.NewSimpleCache(4)

		context := PruningContext{
			navigator:     NewHyperTreeNavigator(numBits),
			cacheResolver: NewSingleTargetedCacheResolver(numBits, cacheLevel, c.key),
			cache:         cache,
			store:         store,
			defaultHashes: []common.Digest{
				common.Digest{0}, common.Digest{0}, common.Digest{0}, common.Digest{0}, common.Digest{0},
				common.Digest{0}, common.Digest{0}, common.Digest{0}, common.Digest{0},
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
		storeMutations []db.Mutation
		expectedPruned common.Visitable
	}{
		{
			key: []byte{0},
			storeMutations: []db.Mutation{
				*db.NewMutation(db.IndexPrefix, []byte{0}, []byte{0}),
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
			storeMutations: []db.Mutation{
				*db.NewMutation(db.IndexPrefix, []byte{1}, []byte{1}),
				*db.NewMutation(db.IndexPrefix, []byte{6}, []byte{6}),
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
		store := bplus.NewBPlusTreeStore()
		store.Mutate(c.storeMutations...)

		cache := common.NewSimpleCache(4)

		context := PruningContext{
			navigator:     NewHyperTreeNavigator(numBits),
			cacheResolver: NewSingleTargetedCacheResolver(numBits, cacheLevel, c.key),
			cache:         cache,
			store:         store,
			defaultHashes: []common.Digest{
				common.Digest{0}, common.Digest{0}, common.Digest{0}, common.Digest{0}, common.Digest{0},
				common.Digest{0}, common.Digest{0}, common.Digest{0}, common.Digest{0},
			},
		}

		pruned := NewSearchPruner(c.key, context).Prune()
		assert.Equalf(t, c.expectedPruned, pruned, "The pruned trees should match for test case %d", i)
	}
}
func TestVerifyPruner(t *testing.T) {

	numBits := uint16(8)
	cacheLevel := uint16(4)

	fakeCache := common.NewFakeCache(common.Digest{0}) // Always return common.Digest{0}
	// Add element before verifying.
	store := bplus.NewBPlusTreeStore()
	mutations := db.Mutation{db.IndexPrefix, []byte{0}, []byte{0}}
	store.Mutate(mutations)

	testCases := []struct {
		key, value     []byte
		expectedPruned common.Visitable
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
			defaultHashes: []common.Digest{
				common.Digest{0}, common.Digest{0}, common.Digest{0}, common.Digest{0}, common.Digest{0},
				common.Digest{0}, common.Digest{0}, common.Digest{0}, common.Digest{0},
			},
		}

		pruned := NewVerifyPruner(c.key, c.value, context).Prune()
		assert.Equalf(t, c.expectedPruned, pruned, "The pruned trees should match for test case %d", i)
	}
}
