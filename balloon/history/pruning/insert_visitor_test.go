package pruning

import (
	"testing"

	"github.com/bbva/qed/balloon/cache"
	"github.com/bbva/qed/balloon/history/navigation"
	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/storage"
	"github.com/stretchr/testify/assert"
)

func TestInsertVisitor(t *testing.T) {

	testCases := []struct {
		op                Operation
		expectedMutations []*storage.Mutation
		expectedElements  []*cachedElement
	}{
		{
			op: mutate(putCache(inner(pos(0, 3),
				getCache(pos(0, 2)),
				mutate(putCache(inner(pos(4, 2),
					getCache(pos(4, 1)),
					mutate(putCache(inner(pos(6, 1),
						getCache(pos(6, 0)),
						mutate(putCache(leaf(pos(7, 0), 7))),
					))),
				))),
			))),
			expectedMutations: []*storage.Mutation{
				{
					Prefix: storage.HistoryCachePrefix,
					Key:    pos(7, 0).Bytes(),
					Value:  []byte{7},
				},
				{
					Prefix: storage.HistoryCachePrefix,
					Key:    pos(6, 1).Bytes(),
					Value:  []byte{7},
				},
				{
					Prefix: storage.HistoryCachePrefix,
					Key:    pos(4, 2).Bytes(),
					Value:  []byte{7},
				},
				{
					Prefix: storage.HistoryCachePrefix,
					Key:    pos(0, 3).Bytes(),
					Value:  []byte{7},
				},
			},
			expectedElements: []*cachedElement{
				newCachedElement(pos(7, 0), []byte{7}),
				newCachedElement(pos(6, 1), []byte{7}),
				newCachedElement(pos(4, 2), []byte{7}),
				newCachedElement(pos(0, 3), []byte{7}),
			},
		},
	}

	for i, c := range testCases {
		cache := cache.NewFakeCache([]byte{0x0})
		visitor := NewInsertVisitor(hashing.NewFakeXorHasher(), cache, storage.HistoryCachePrefix)

		c.op.Accept(visitor)

		mutations := visitor.Result()
		assert.ElementsMatchf(t, c.expectedMutations, mutations, "Mutation error in test case %d", i)
		for _, e := range c.expectedElements {
			v, _ := cache.Get(e.Pos.Bytes())
			assert.Equalf(t, e.Digest, v, "The cached element %v should be cached in test case %d", e, i)
		}
	}
}

type cachedElement struct {
	Pos    *navigation.Position
	Digest []byte
}

func newCachedElement(pos *navigation.Position, digest []byte) *cachedElement {
	return &cachedElement{pos, digest}
}
