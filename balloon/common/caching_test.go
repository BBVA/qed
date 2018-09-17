package common

import (
	"testing"

	"github.com/bbva/qed/hashing"
	"github.com/stretchr/testify/require"
)

func TestCachingVisitor(t *testing.T) {

	testCases := []struct {
		visitable        Visitable
		expectedElements []CachedElement
	}{
		{
			visitable: NewCollectable(
				NewLeaf(
					&FakePosition{[]byte{0x0}, 0},
					[]byte{0x0},
				),
			),
			expectedElements: []CachedElement{
				*NewCachedElement(
					&FakePosition{[]byte{0x0}, 0},
					hashing.Digest{0x0},
				),
			},
		},
		{
			visitable: NewCollectable(
				NewRoot(&FakePosition{[]byte{0x0}, 1},
					NewCached(&FakePosition{[]byte{0x0}, 0}, hashing.Digest{0x0}),
					NewCollectable(
						NewLeaf(&FakePosition{[]byte{0x1}, 0}, hashing.Digest{0x1})),
				)),
			expectedElements: []CachedElement{
				*NewCachedElement(
					&FakePosition{[]byte{0x1}, 0},
					hashing.Digest{0x1},
				),
				*NewCachedElement(
					&FakePosition{[]byte{0x0}, 1},
					hashing.Digest{0x1},
				),
			},
		},
		{
			visitable: NewRoot(
				&FakePosition{[]byte{0x0}, 2},
				NewCached(&FakePosition{[]byte{0x0}, 1}, hashing.Digest{0x1}),
				NewPartialNode(&FakePosition{[]byte{0x1}, 1},
					NewCollectable(
						NewLeaf(&FakePosition{[]byte{0x2}, 0}, hashing.Digest{0x2}),
					),
				),
			),
			expectedElements: []CachedElement{
				*NewCachedElement(
					&FakePosition{[]byte{0x2}, 0},
					hashing.Digest{0x2},
				),
			},
		},
		{
			visitable: NewCollectable(
				NewRoot(
					&FakePosition{[]byte{0x0}, 2},
					NewCached(&FakePosition{[]byte{0x0}, 1}, hashing.Digest{0x1}),
					NewCollectable(
						NewNode(&FakePosition{[]byte{0x2}, 1},
							NewCached(&FakePosition{[]byte{0x2}, 0}, hashing.Digest{0x2}),
							NewCollectable(
								NewLeaf(&FakePosition{[]byte{0x3}, 0}, hashing.Digest{0x3}),
							))),
				)),
			expectedElements: []CachedElement{
				*NewCachedElement(
					&FakePosition{[]byte{0x3}, 0},
					hashing.Digest{0x3},
				),
				*NewCachedElement(
					&FakePosition{[]byte{0x2}, 1},
					hashing.Digest{0x1},
				),
				*NewCachedElement(
					&FakePosition{[]byte{0x0}, 2},
					hashing.Digest{0x0},
				),
			},
		},
		{
			visitable: NewRoot(
				&FakePosition{[]byte{0x0}, 3},
				NewCached(&FakePosition{[]byte{0x0}, 2}, hashing.Digest{0x0}),
				NewPartialNode(&FakePosition{[]byte{0x4}, 2},
					NewPartialNode(&FakePosition{[]byte{0x4}, 1},
						NewCollectable(
							NewLeaf(&FakePosition{[]byte{0x4}, 0}, hashing.Digest{0x4}))),
				),
			),
			expectedElements: []CachedElement{
				*NewCachedElement(
					&FakePosition{[]byte{0x4}, 0},
					hashing.Digest{0x4},
				),
			},
		},
	}

	for i, c := range testCases {
		visitor := NewCachingVisitor(NewComputeHashVisitor(hashing.NewFakeXorHasher()))
		c.visitable.PostOrder(visitor)
		cachedElements := visitor.Result()
		require.Equalf(t, c.expectedElements, cachedElements, "The cached elements %v should be equal to the expected %v in test case %d", cachedElements, c.expectedElements, i)
	}

}
