package hyper

import (
	"testing"

	assert "github.com/stretchr/testify/require"

	"github.com/bbva/qed/balloon2/common"
	"github.com/bbva/qed/db"
	"github.com/bbva/qed/db/bplus"
)

var (
	FixedDigest = make([]byte, 8)
)

func TestInsertPruner(t *testing.T) {

	numBits := uint16(8)
	cacheLevel := uint16(4)

	testCases := []struct {
		key, value     []byte
		storeMutation  []db.Mutation
		expectedPruned common.Visitable
	}{
		{
			key:           []byte{0},
			value:         []byte{0},
			storeMutation: []db.Mutation{},
			expectedPruned: common.NewRoot(NewPosition([]byte{0}, 8),
				common.NewCollectable(NewPosition([]byte{0}, 7),
					common.NewNode(NewPosition([]byte{0}, 7),
						common.NewCollectable(NewPosition([]byte{0}, 6),
							common.NewNode(NewPosition([]byte{0}, 6),
								common.NewCollectable(NewPosition([]byte{0}, 5),
									common.NewNode(NewPosition([]byte{0}, 5),
										common.NewNode(NewPosition([]byte{0}, 4),
											common.NewNode(NewPosition([]byte{0}, 3),
												common.NewNode(NewPosition([]byte{0}, 2),
													common.NewNode(NewPosition([]byte{0}, 1),
														common.NewLeaf(NewPosition([]byte{0}, 0), []byte{0}),
														common.NewCached(NewPosition([]byte{1}, 0), common.Digest{0})),
													common.NewCached(NewPosition([]byte{2}, 1), common.Digest{0})),
												common.NewCached(NewPosition([]byte{4}, 2), common.Digest{0})),
											common.NewCached(NewPosition([]byte{8}, 3), common.Digest{0})),
										common.NewCached(NewPosition([]byte{16}, 4), common.Digest{0}))),
								common.NewCached(NewPosition([]byte{32}, 5), common.Digest{0}))),
						common.NewCached(NewPosition([]byte{64}, 6), common.Digest{0}))),
				common.NewCached(NewPosition([]byte{128}, 7), common.Digest{0}),
			),
		},
		{
			key:   []byte{2},
			value: []byte{1},
			storeMutation: []db.Mutation{
				*db.NewMutation(db.IndexPrefix, []byte{0}, []byte{0}),
			},
			expectedPruned: common.NewRoot(NewPosition([]byte{0}, 8),
				common.NewCollectable(NewPosition([]byte{0}, 7),
					common.NewNode(NewPosition([]byte{0}, 7),
						common.NewCollectable(NewPosition([]byte{0}, 6),
							common.NewNode(NewPosition([]byte{0}, 6),
								common.NewCollectable(NewPosition([]byte{0}, 5),
									common.NewNode(NewPosition([]byte{0}, 5),
										common.NewNode(NewPosition([]byte{0}, 4),
											common.NewNode(NewPosition([]byte{0}, 3),
												common.NewNode(NewPosition([]byte{0}, 2),
													common.NewNode(NewPosition([]byte{0}, 1),
														common.NewLeaf(NewPosition([]byte{0}, 0), []byte{0}),
														common.NewCached(NewPosition([]byte{1}, 0), common.Digest{0})),
													common.NewNode(NewPosition([]byte{2}, 1),
														common.NewLeaf(NewPosition([]byte{2}, 0), []byte{1}),
														common.NewCached(NewPosition([]byte{3}, 0), common.Digest{0}))),
												common.NewCached(NewPosition([]byte{4}, 2), common.Digest{0})),
											common.NewCached(NewPosition([]byte{8}, 3), common.Digest{0})),
										common.NewCached(NewPosition([]byte{16}, 4), common.Digest{0}))),
								common.NewCached(NewPosition([]byte{32}, 5), common.Digest{0}))),
						common.NewCached(NewPosition([]byte{64}, 6), common.Digest{0}))),
				common.NewCached(NewPosition([]byte{128}, 7), common.Digest{0}),
			),
		},
		{
			key:   []byte{255},
			value: []byte{2},
			storeMutation: []db.Mutation{
				*db.NewMutation(db.IndexPrefix, []byte{0}, []byte{0}),
				*db.NewMutation(db.IndexPrefix, []byte{2}, []byte{1}),
			},
			expectedPruned: common.NewRoot(NewPosition([]byte{0}, 8),
				common.NewCached(NewPosition([]byte{0}, 7), common.Digest{0}),
				common.NewCollectable(NewPosition([]byte{128}, 7),
					common.NewNode(NewPosition([]byte{128}, 7),
						common.NewCached(NewPosition([]byte{128}, 6), common.Digest{0}),
						common.NewCollectable(NewPosition([]byte{192}, 6),
							common.NewNode(NewPosition([]byte{192}, 6),
								common.NewCached(NewPosition([]byte{192}, 5), common.Digest{0}),
								common.NewCollectable(NewPosition([]byte{224}, 5),
									common.NewNode(NewPosition([]byte{224}, 5),
										common.NewCached(NewPosition([]byte{224}, 4), common.Digest{0}),
										common.NewNode(NewPosition([]byte{240}, 4),
											common.NewCached(NewPosition([]byte{240}, 3), common.Digest{0}),
											common.NewNode(NewPosition([]byte{248}, 3),
												common.NewCached(NewPosition([]byte{248}, 2), common.Digest{0}),
												common.NewNode(NewPosition([]byte{252}, 2),
													common.NewCached(NewPosition([]byte{252}, 1), common.Digest{0}),
													common.NewNode(NewPosition([]byte{254}, 1),
														common.NewCached(NewPosition([]byte{254}, 0), common.Digest{0}),
														common.NewLeaf(NewPosition([]byte{255}, 0), []byte{2}))))))))),
					),
				),
			),
		},
	}

	for i, c := range testCases {
		store := bplus.NewBPlusTreeStore()
		store.Mutate(c.storeMutation...)

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

	cache := common.NewSimpleCache(4)
	// Add element before searching.
	// BPlus storage (memmory) instead of Badger (disk) for ease.
	store := bplus.NewBPlusTreeStore()
	mutations := db.Mutation{db.IndexPrefix, []byte{0}, []byte{0}}
	store.Mutate(mutations)

	testCases := []struct {
		key            []byte
		expectedPruned common.Visitable
	}{
		{
			key: []byte{0},
			expectedPruned: common.NewRoot(NewPosition([]byte{0}, 8),
				common.NewNode(NewPosition([]byte{0}, 7),
					common.NewNode(NewPosition([]byte{0}, 6),
						common.NewNode(NewPosition([]byte{0}, 5),
							common.NewNode(NewPosition([]byte{0}, 4),
								common.NewNode(NewPosition([]byte{0}, 3),
									common.NewNode(NewPosition([]byte{0}, 2),
										common.NewNode(NewPosition([]byte{0}, 1),
											common.NewLeaf(NewPosition([]byte{0}, 0), []byte{0}),
											common.NewCollectable(NewPosition([]byte{1}, 0),
												common.NewCached(NewPosition([]byte{1}, 0), common.Digest{0}))),
										common.NewCollectable(NewPosition([]byte{2}, 1),
											common.NewCached(NewPosition([]byte{2}, 1), common.Digest{0}))),
									common.NewCollectable(NewPosition([]byte{4}, 2),
										common.NewCached(NewPosition([]byte{4}, 2), common.Digest{0}))),
								common.NewCollectable(NewPosition([]byte{8}, 3),
									common.NewCached(NewPosition([]byte{8}, 3), common.Digest{0}))),
							common.NewCollectable(NewPosition([]byte{16}, 4),
								common.NewCached(NewPosition([]byte{16}, 4), common.Digest{0}))),
						common.NewCollectable(NewPosition([]byte{32}, 5),
							common.NewCached(NewPosition([]byte{32}, 5), common.Digest{0}))),
					common.NewCollectable(NewPosition([]byte{64}, 6),
						common.NewCached(NewPosition([]byte{64}, 6), common.Digest{0}))),
				common.NewCollectable(NewPosition([]byte{128}, 7),
					common.NewCached(NewPosition([]byte{128}, 7), common.Digest{0})),
			),
		},
	}

	for i, c := range testCases {
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
			expectedPruned: common.NewRoot(NewPosition([]byte{0}, 8),
				common.NewNode(NewPosition([]byte{0}, 7),
					common.NewNode(NewPosition([]byte{0}, 6),
						common.NewNode(NewPosition([]byte{0}, 5),
							common.NewNode(NewPosition([]byte{0}, 4),
								common.NewNode(NewPosition([]byte{0}, 3),
									common.NewNode(NewPosition([]byte{0}, 2),
										common.NewNode(NewPosition([]byte{0}, 1),
											common.NewLeaf(NewPosition([]byte{0}, 0), []byte{0}),
											common.NewCached(NewPosition([]byte{1}, 0), common.Digest{0})),
										common.NewCached(NewPosition([]byte{2}, 1), common.Digest{0})),
									common.NewCached(NewPosition([]byte{4}, 2), common.Digest{0})),
								common.NewCached(NewPosition([]byte{8}, 3), common.Digest{0})),
							common.NewCached(NewPosition([]byte{16}, 4), common.Digest{0})),
						common.NewCached(NewPosition([]byte{32}, 5), common.Digest{0})),
					common.NewCached(NewPosition([]byte{64}, 6), common.Digest{0})),
				common.NewCached(NewPosition([]byte{128}, 7), common.Digest{0})),
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
