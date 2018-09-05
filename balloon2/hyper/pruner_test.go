package hyper

import (
	"testing"

	assert "github.com/stretchr/testify/require"

	"github.com/bbva/qed/balloon2/common"
)

var (
	FixedDigest = make([]byte, 8)
)

func TestInsertPruner(t *testing.T) {

	numBits := uint16(8)
	cacheLevel := uint16(4)
	fakeCache := common.NewSimpleCache(4)

	store, closeF := common.OpenBadgerStore("/var/tmp/balloon.hyper.test")
	defer closeF()

	testCases := []struct {
		key, value     []byte
		expectedPruned common.Visitable
	}{
		{
			key:   []byte{0},
			value: []byte{0},
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

		pruned := NewInsertPruner(c.key, c.value, context).Prune()
		assert.Equalf(t, c.expectedPruned, pruned, "The pruned trees should match for test case %d", i)
	}
}

func TestSearchPruner(t *testing.T) {

	numBits := uint16(8)
	cacheLevel := uint16(4)
	fakeCache := common.NewSimpleCache(4)

	store, closeF := common.OpenBadgerStore("/var/tmp/balloon.hyper.test")
	defer closeF()

	testCases := []struct {
		key            []byte
		expectedPruned common.Visitable
	}{
		{
			key: []byte{0},
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

		pruned := NewSearchPruner(c.key, context).Prune()
		assert.Equalf(t, c.expectedPruned, pruned, "The pruned trees should match for test case %d", i)
	}
}
