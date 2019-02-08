package pruning

import (
	"testing"

	"github.com/bbva/qed/balloon/cache"
	"github.com/bbva/qed/hashing"
	"github.com/stretchr/testify/require"
)

func TestComputeHashVisitor(t *testing.T) {

	testCases := []struct {
		op             Operation
		expectedDigest hashing.Digest
	}{
		{
			op:             leaf(pos(0, 0), 1),
			expectedDigest: hashing.Digest{0x1},
		},
		{
			op: inner(pos(0, 1),
				getCache(pos(0, 0)),
				leaf(pos(1, 0), 1),
			),
			expectedDigest: hashing.Digest{0x1},
		},
		{
			op: inner(pos(0, 2),
				getCache(pos(0, 1)),
				partial(pos(1, 1),
					leaf(pos(2, 0), 2),
				),
			),
			expectedDigest: hashing.Digest{0x2},
		},
		{
			op: inner(pos(0, 2),
				getCache(pos(0, 1)),
				inner(pos(1, 1),
					getCache(pos(2, 0)),
					leaf(pos(3, 0), 3),
				),
			),
			expectedDigest: hashing.Digest{0x3},
		},
		{
			op: inner(pos(0, 2),
				getCache(pos(0, 1)),
				inner(pos(1, 1),
					getCache(pos(2, 0)),
					mutate(leaf(pos(3, 0), 3)),
				),
			),
			expectedDigest: hashing.Digest{0x3},
		},
	}

	visitor := NewComputeHashVisitor(hashing.NewFakeXorHasher(), cache.NewFakeCache([]byte{0x0}))

	for i, c := range testCases {
		digest := c.op.Accept(visitor)
		require.Equalf(t, c.expectedDigest, digest, "The computed digest %x should be equal to the expected %x in test case %d", digest, c.expectedDigest, i)
	}
}
