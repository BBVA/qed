package pruning

import (
	"testing"

	"github.com/bbva/qed/balloon/cache"
	"github.com/bbva/qed/balloon/history/navigation"
	"github.com/bbva/qed/hashing"
	"github.com/stretchr/testify/require"
)

func TestAuditPathVisitor(t *testing.T) {

	testCases := []struct {
		op                Operation
		expectedAuditPath navigation.AuditPath
	}{
		{
			op:                leafnil(pos(0, 0)),
			expectedAuditPath: navigation.AuditPath{},
		},
		{
			op: inner(pos(0, 1),
				collect(getCache(pos(0, 0))),
				leafnil(pos(1, 0)),
			),
			expectedAuditPath: navigation.AuditPath{
				pos(0, 0).FixedBytes(): hashing.Digest{0x0},
			},
		},
		{
			op: inner(pos(0, 2),
				collect(getCache(pos(0, 1))),
				partial(pos(2, 1),
					leafnil(pos(2, 0)),
				),
			),
			expectedAuditPath: navigation.AuditPath{
				pos(0, 1).FixedBytes(): hashing.Digest{0x0},
			},
		},
		{
			op: inner(pos(0, 2),
				collect(getCache(pos(0, 1))),
				inner(pos(2, 1),
					collect(getCache(pos(2, 0))),
					leafnil(pos(3, 0)),
				),
			),
			expectedAuditPath: navigation.AuditPath{
				pos(0, 1).FixedBytes(): hashing.Digest{0x0},
				pos(2, 0).FixedBytes(): hashing.Digest{0x0},
			},
		},
		{
			op: inner(pos(0, 3),
				collect(getCache(pos(0, 2))),
				partial(pos(4, 2),
					partial(pos(4, 1),
						leafnil(pos(4, 0)),
					),
				),
			),
			expectedAuditPath: navigation.AuditPath{
				pos(0, 2).FixedBytes(): hashing.Digest{0x0},
			},
		},
		{
			op: inner(pos(0, 3),
				collect(getCache(pos(0, 2))),
				partial(pos(4, 2),
					inner(pos(4, 1),
						collect(getCache(pos(4, 0))),
						leafnil(pos(5, 0)),
					),
				),
			),
			expectedAuditPath: navigation.AuditPath{
				pos(0, 2).FixedBytes(): hashing.Digest{0x0},
				pos(4, 0).FixedBytes(): hashing.Digest{0x0},
			},
		},
	}

	for i, c := range testCases {
		visitor := NewAuditPathVisitor(hashing.NewFakeXorHasher(), cache.NewFakeCache([]byte{0x0}))
		c.op.Accept(visitor)
		auditPath := visitor.Result()
		require.Equalf(t, c.expectedAuditPath, auditPath, "The audit path %v should be equal to the expected %v in test case %d", auditPath, c.expectedAuditPath, i)
	}

}
