package common

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestComputeHashVisitor(t *testing.T) {

	testCases := []struct {
		visitable      Visitable
		expectedDigest Digest
	}{
		{
			visitable:      NewLeaf(&FakePosition{[]byte{0x0}, 0}, []byte{0x1}),
			expectedDigest: Digest{0x1},
		},
		{
			visitable: NewRoot(
				&FakePosition{[]byte{0x0}, 1},
				NewCached(&FakePosition{[]byte{0x0}, 0}, Digest{0x0}),
				NewLeaf(&FakePosition{[]byte{0x1}, 0}, Digest{0x1}),
			),
			expectedDigest: Digest{0x1},
		},
		{
			visitable: NewRoot(
				&FakePosition{[]byte{0x0}, 2},
				NewCached(&FakePosition{[]byte{0x0}, 1}, Digest{0x1}),
				NewPartialNode(&FakePosition{[]byte{0x1}, 1},
					NewLeaf(&FakePosition{[]byte{0x2}, 0}, Digest{0x2}),
				),
			),
			expectedDigest: Digest{0x3},
		},
		{
			visitable: NewRoot(
				&FakePosition{[]byte{0x0}, 2},
				NewCached(&FakePosition{[]byte{0x0}, 1}, Digest{0x1}),
				NewNode(&FakePosition{[]byte{0x1}, 1},
					NewCached(&FakePosition{[]byte{0x2}, 0}, Digest{0x2}),
					NewLeaf(&FakePosition{[]byte{0x3}, 0}, Digest{0x3}),
				),
			),
			expectedDigest: Digest{0x0},
		},
		{
			visitable: NewRoot(
				&FakePosition{[]byte{0x0}, 2},
				NewCached(&FakePosition{[]byte{0x0}, 1}, Digest{0x1}),
				NewNode(&FakePosition{[]byte{0x1}, 1},
					NewCached(&FakePosition{[]byte{0x2}, 0}, Digest{0x2}),
					NewCollectable(&FakePosition{[]byte{0x3}, 0},
						NewLeaf(&FakePosition{[]byte{0x3}, 0}, Digest{0x3})),
				),
			),
			expectedDigest: Digest{0x0},
		},
	}

	visitor := NewComputeHashVisitor(NewFakeXorHasher())

	for i, c := range testCases {
		digest := c.visitable.PostOrder(visitor)
		require.Equalf(t, c.expectedDigest, digest, "The computed digest %x should be equal to the expected %x in test case %d", digest, c.expectedDigest, i)
	}
}
