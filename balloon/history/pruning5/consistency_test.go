package pruning5

import (
	"testing"

	"github.com/bbva/qed/log"
	"github.com/stretchr/testify/assert"
)

func TestPruneToFindConsistent(t *testing.T) {

	testCases := []struct {
		index, version uint64
		expectedOp     Operation
	}{
		{
			index:      0,
			version:    0,
			expectedOp: leafnil(pos(0, 0)),
		},
		{
			index:   0,
			version: 1,
			expectedOp: inner(pos(0, 1),
				leafnil(pos(0, 0)),
				collect(getCache(pos(1, 0))),
			),
		},
		{
			index:   0,
			version: 2,
			expectedOp: inner(pos(0, 2),
				inner(pos(0, 1),
					leafnil(pos(0, 0)),
					collect(getCache(pos(1, 0))),
				),
				collect(partial(pos(2, 1),
					getCache(pos(2, 0))),
				),
			),
		},
		{
			index:   0,
			version: 3,
			expectedOp: inner(pos(0, 2),
				inner(pos(0, 1),
					leafnil(pos(0, 0)),
					collect(getCache(pos(1, 0))),
				),
				collect(inner(pos(2, 1),
					getCache(pos(2, 0)),
					getCache(pos(3, 0))),
				),
			),
		},
		{
			index:   0,
			version: 4,
			expectedOp: inner(pos(0, 3),
				inner(pos(0, 2),
					inner(pos(0, 1),
						leafnil(pos(0, 0)),
						collect(getCache(pos(1, 0))),
					),
					collect(getCache(pos(2, 1)))),
				collect(partial(pos(4, 2),
					partial(pos(4, 1),
						getCache(pos(4, 0)),
					),
				)),
			),
		},
		{
			index:   0,
			version: 5,
			expectedOp: inner(pos(0, 3),
				inner(pos(0, 2),
					inner(pos(0, 1),
						leafnil(pos(0, 0)),
						collect(getCache(pos(1, 0))),
					),
					collect(getCache(pos(2, 1)))),
				collect(partial(pos(4, 2),
					inner(pos(4, 1),
						getCache(pos(4, 0)),
						getCache(pos(5, 0)),
					),
				)),
			),
		},
		{
			index:   0,
			version: 6,
			expectedOp: inner(pos(0, 3),
				inner(pos(0, 2),
					inner(pos(0, 1),
						leafnil(pos(0, 0)),
						collect(getCache(pos(1, 0))),
					),
					collect(getCache(pos(2, 1)))),
				collect(inner(pos(4, 2),
					getCache(pos(4, 1)),
					partial(pos(6, 1),
						getCache(pos(6, 0)),
					),
				)),
			),
		},
		{
			index:   0,
			version: 7,
			expectedOp: inner(pos(0, 3),
				inner(pos(0, 2),
					inner(pos(0, 1),
						leafnil(pos(0, 0)),
						collect(getCache(pos(1, 0))),
					),
					collect(getCache(pos(2, 1)))),
				collect(inner(pos(4, 2),
					getCache(pos(4, 1)),
					inner(pos(6, 1),
						getCache(pos(6, 0)),
						getCache(pos(7, 0)),
					),
				)),
			),
		},
	}

	for i, c := range testCases {
		prunedOp := PruneToFindConsistent(c.index, c.version)
		assert.Equalf(t, c.expectedOp, prunedOp, "The pruned operation should match for test case %d", i)
	}

}

func TestPruneToCheckConsistency(t *testing.T) {

	testCases := []struct {
		start, end uint64
		expectedOp Operation
	}{
		{
			start:      0,
			end:        0,
			expectedOp: collect(getCache(pos(0, 0))),
		},
		{
			start: 0,
			end:   1,
			expectedOp: inner(pos(0, 1),
				collect(getCache(pos(0, 0))),
				collect(getCache(pos(1, 0))),
			),
		},
		{
			start: 0,
			end:   2,
			expectedOp: inner(pos(0, 2),
				inner(pos(0, 1),
					collect(getCache(pos(0, 0))),
					collect(getCache(pos(1, 0))),
				),
				partial(pos(2, 1),
					collect(getCache(pos(2, 0))),
				),
			),
		},
		{
			start: 0,
			end:   3,
			expectedOp: inner(pos(0, 2),
				inner(pos(0, 1),
					collect(getCache(pos(0, 0))),
					collect(getCache(pos(1, 0))),
				),
				inner(pos(2, 1),
					collect(getCache(pos(2, 0))),
					collect(getCache(pos(3, 0))),
				),
			),
		},
		{
			start: 0,
			end:   4,
			expectedOp: inner(pos(0, 3),
				inner(pos(0, 2),
					inner(pos(0, 1),
						collect(getCache(pos(0, 0))),
						collect(getCache(pos(1, 0))),
					),
					collect(getCache(pos(2, 1)))),
				partial(pos(4, 2),
					partial(pos(4, 1),
						collect(getCache(pos(4, 0))),
					),
				),
			),
		},
		{
			start: 0,
			end:   5,
			expectedOp: inner(pos(0, 3),
				inner(pos(0, 2),
					inner(pos(0, 1),
						collect(getCache(pos(0, 0))),
						collect(getCache(pos(1, 0))),
					),
					collect(getCache(pos(2, 1)))),
				partial(pos(4, 2),
					inner(pos(4, 1),
						collect(getCache(pos(4, 0))),
						collect(getCache(pos(5, 0))),
					),
				),
			),
		},
		{
			start: 0,
			end:   6,
			expectedOp: inner(pos(0, 3),
				inner(pos(0, 2),
					inner(pos(0, 1),
						collect(getCache(pos(0, 0))),
						collect(getCache(pos(1, 0))),
					),
					collect(getCache(pos(2, 1)))),
				inner(pos(4, 2),
					collect(getCache(pos(4, 1))),
					partial(pos(6, 1),
						collect(getCache(pos(6, 0))),
					),
				),
			),
		},
		{
			start: 0,
			end:   7,
			expectedOp: inner(pos(0, 3),
				inner(pos(0, 2),
					inner(pos(0, 1),
						collect(getCache(pos(0, 0))),
						collect(getCache(pos(1, 0))),
					),
					collect(getCache(pos(2, 1)))),
				inner(pos(4, 2),
					collect(getCache(pos(4, 1))),
					inner(pos(6, 1),
						collect(getCache(pos(6, 0))),
						collect(getCache(pos(7, 0))),
					),
				),
			),
		},
	}

	for i, c := range testCases {
		prunedOp := PruneToCheckConsistency(c.start, c.end)
		assert.Equalf(t, c.expectedOp, prunedOp, "The pruned operation should match for test case %d", i)
	}

}

func BenchmarkPruneToFindConsistent(b *testing.B) {

	log.SetLogger("BenchmarkPruneToFindConsistent", log.SILENT)

	b.ResetTimer()
	for i := uint64(0); i < uint64(b.N); i++ {
		pruned := PruneToFindConsistent(0, i)
		assert.NotNil(b, pruned)
	}

}

func BenchmarkPruneToCheckConsistency(b *testing.B) {

	log.SetLogger("BenchmarkPruneToCheckConsistency", log.SILENT)

	b.ResetTimer()
	for i := uint64(0); i < uint64(b.N); i++ {
		pruned := PruneToCheckConsistency(0, i)
		assert.NotNil(b, pruned)
	}

}
