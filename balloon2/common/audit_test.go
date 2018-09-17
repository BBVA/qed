package common

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAuditPathVisitor(t *testing.T) {

	testCases := []struct {
		visitable         Visitable
		expectedAuditPath AuditPath
	}{
		{
			visitable:         NewLeaf(&FakePosition{[]byte{0x0}, 0}, []byte{0x1}),
			expectedAuditPath: AuditPath{},
		},
		{
			visitable: NewRoot(
				&FakePosition{[]byte{0x0}, 1},
				NewCached(&FakePosition{[]byte{0x0}, 0}, Digest{0x0}),
				NewLeaf(&FakePosition{[]byte{0x1}, 0}, Digest{0x1}),
			),
			expectedAuditPath: AuditPath{},
		},
		{
			visitable: NewRoot(
				&FakePosition{[]byte{0x0}, 2},
				NewCached(&FakePosition{[]byte{0x0}, 1}, Digest{0x1}),
				NewPartialNode(&FakePosition{[]byte{0x1}, 1},
					NewLeaf(&FakePosition{[]byte{0x2}, 0}, Digest{0x2}),
				),
			),
			expectedAuditPath: AuditPath{},
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
			expectedAuditPath: AuditPath{},
		},
		{
			visitable: NewRoot(
				&FakePosition{[]byte{0x0}, 2},
				NewCollectable(NewCached(&FakePosition{[]byte{0x0}, 1}, Digest{0x1})),
				NewPartialNode(&FakePosition{[]byte{0x1}, 1},
					NewLeaf(&FakePosition{[]byte{0x2}, 0}, Digest{0x2}),
				),
			),
			expectedAuditPath: AuditPath{
				"00|1": Digest{0x1},
			},
		},
		{
			visitable: NewRoot(
				&FakePosition{[]byte{0x0}, 2},
				NewCollectable(NewCached(&FakePosition{[]byte{0x0}, 1}, Digest{0x1})),
				NewNode(&FakePosition{[]byte{0x1}, 1},
					NewCollectable(NewCached(&FakePosition{[]byte{0x2}, 0}, Digest{0x2})),
					NewLeaf(&FakePosition{[]byte{0x3}, 0}, Digest{0x3}),
				),
			),
			expectedAuditPath: AuditPath{
				"00|1": Digest{0x1},
				"02|0": Digest{0x2},
			},
		},
	}

	for i, c := range testCases {
		visitor := NewAuditPathVisitor(NewComputeHashVisitor(NewFakeXorHasher()))
		c.visitable.PostOrder(visitor)
		auditPath := visitor.Result()
		require.Equalf(t, c.expectedAuditPath, auditPath, "The audit path %v should be equal to the expected %v in test case %d", auditPath, c.expectedAuditPath, i)
	}

}
