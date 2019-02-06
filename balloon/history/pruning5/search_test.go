package pruning5

import (
	"testing"

	"github.com/bbva/qed/log"
	"github.com/stretchr/testify/assert"
)

func TestPruneToFind(t *testing.T) {

	testCases := []struct {
		version    uint64
		expectedOp Operation
	}{
		{
			version:    0,
			expectedOp: leafnil(pos(0, 0)),
		},
		{
			version: 1,
			expectedOp: inner(pos(0, 1),
				collect(getCache(pos(0, 0))),
				leafnil(pos(1, 0)),
			),
		},
		{
			version: 2,
			expectedOp: inner(pos(0, 2),
				collect(getCache(pos(0, 1))),
				partial(pos(2, 1),
					leafnil(pos(2, 0))),
			),
		},
		{
			version: 3,
			expectedOp: inner(pos(0, 2),
				collect(getCache(pos(0, 1))),
				inner(pos(2, 1),
					collect(getCache(pos(2, 0))),
					leafnil(pos(3, 0))),
			),
		},
		{
			version: 4,
			expectedOp: inner(pos(0, 3),
				collect(getCache(pos(0, 2))),
				partial(pos(4, 2),
					partial(pos(4, 1),
						leafnil(pos(4, 0)))),
			),
		},
		{
			version: 5,
			expectedOp: inner(pos(0, 3),
				collect(getCache(pos(0, 2))),
				partial(pos(4, 2),
					inner(pos(4, 1),
						collect(getCache(pos(4, 0))),
						leafnil(pos(5, 0)))),
			),
		},
		{
			version: 6,
			expectedOp: inner(pos(0, 3),
				collect(getCache(pos(0, 2))),
				inner(pos(4, 2),
					collect(getCache(pos(4, 1))),
					partial(pos(6, 1),
						leafnil(pos(6, 0)))),
			),
		},
		{
			version: 7,
			expectedOp: inner(pos(0, 3),
				collect(getCache(pos(0, 2))),
				inner(pos(4, 2),
					collect(getCache(pos(4, 1))),
					inner(pos(6, 1),
						collect(getCache(pos(6, 0))),
						leafnil(pos(7, 0)))),
			),
		},
	}

	for i, c := range testCases {
		prunedOp := PruneToFind(c.version)
		assert.Equalf(t, c.expectedOp, prunedOp, "The pruned operation should match for test case %d", i)
	}

}

func BenchmarkPruneToFind(b *testing.B) {

	log.SetLogger("BenchmarkPruneToFind", log.SILENT)

	b.ResetTimer()
	for i := uint64(0); i < uint64(b.N); i++ {
		pruned := PruneToFind(i)
		assert.NotNil(b, pruned)
	}

}
