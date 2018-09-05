package history

import (
	"testing"

	"github.com/bbva/qed/balloon2/common"
	"github.com/bbva/qed/log"
	"github.com/stretchr/testify/assert"
)

func TestVerify(t *testing.T) {

	log.SetLogger("TestVerify", log.INFO)

	testCases := []struct {
		index, version uint64
		auditPath      common.AuditPath
		eventDigest    common.Digest
		expectedDigest common.Digest
	}{
		{
			index:          0,
			version:        0,
			auditPath:      common.AuditPath{},
			eventDigest:    common.Digest{0x0},
			expectedDigest: common.Digest{0x0},
		},
		{
			index:          1,
			version:        1,
			auditPath:      common.AuditPath{"0|0": common.Digest{0x0}},
			eventDigest:    common.Digest{0x1},
			expectedDigest: common.Digest{0x1},
		},
		{
			index:          2,
			version:        2,
			auditPath:      common.AuditPath{"0|1": common.Digest{0x1}},
			eventDigest:    common.Digest{0x2},
			expectedDigest: common.Digest{0x3},
		},
		{
			index:          3,
			version:        3,
			auditPath:      common.AuditPath{"0|1": common.Digest{0x1}, "2|0": common.Digest{0x2}},
			eventDigest:    common.Digest{0x3},
			expectedDigest: common.Digest{0x0},
		},
		{
			index:          4,
			version:        4,
			auditPath:      common.AuditPath{"0|2": common.Digest{0x0}},
			eventDigest:    common.Digest{0x4},
			expectedDigest: common.Digest{0x4},
		},
		{
			index:          5,
			version:        5,
			auditPath:      common.AuditPath{"0|2": common.Digest{0x0}, "4|0": common.Digest{0x4}},
			eventDigest:    common.Digest{0x5},
			expectedDigest: common.Digest{0x1},
		},
		{
			index:          6,
			version:        6,
			auditPath:      common.AuditPath{"0|2": common.Digest{0x0}, "4|1": common.Digest{0x1}},
			eventDigest:    common.Digest{0x6},
			expectedDigest: common.Digest{0x7},
		},
		{
			index:          7,
			version:        7,
			auditPath:      common.AuditPath{"0|2": common.Digest{0x0}, "4|1": common.Digest{0x1}, "6|0": common.Digest{0x6}},
			eventDigest:    common.Digest{0x7},
			expectedDigest: common.Digest{0x0},
		},
		{
			index:          8,
			version:        8,
			auditPath:      common.AuditPath{"0|3": common.Digest{0x0}},
			eventDigest:    common.Digest{0x8},
			expectedDigest: common.Digest{0x8},
		},
		{
			index:          9,
			version:        9,
			auditPath:      common.AuditPath{"0|3": common.Digest{0x0}, "8|0": common.Digest{0x8}},
			eventDigest:    common.Digest{0x9},
			expectedDigest: common.Digest{0x1},
		},
		{
			index:          0,
			version:        1,
			auditPath:      common.AuditPath{"1|0": common.Digest{0x1}},
			eventDigest:    common.Digest{0x0},
			expectedDigest: common.Digest{0x1},
		},
		{
			index:          0,
			version:        1,
			auditPath:      common.AuditPath{"1|0": common.Digest{0x1}},
			eventDigest:    common.Digest{0x0},
			expectedDigest: common.Digest{0x1},
		},
		{
			index:          0,
			version:        2,
			auditPath:      common.AuditPath{"1|0": common.Digest{0x1}, "2|0": common.Digest{0x2}},
			eventDigest:    common.Digest{0x0},
			expectedDigest: common.Digest{0x3},
		},
		{
			index:          0,
			version:        3,
			auditPath:      common.AuditPath{"1|0": common.Digest{0x1}, "2|1": common.Digest{0x1}},
			eventDigest:    common.Digest{0x0},
			expectedDigest: common.Digest{0x0},
		},
		{
			index:          0,
			version:        4,
			auditPath:      common.AuditPath{"1|0": common.Digest{0x1}, "2|1": common.Digest{0x1}, "4|0": common.Digest{0x4}},
			eventDigest:    common.Digest{0x0},
			expectedDigest: common.Digest{0x4},
		},
		{
			index:          0,
			version:        5,
			auditPath:      common.AuditPath{"1|0": common.Digest{0x1}, "2|1": common.Digest{0x1}, "4|1": common.Digest{0x1}},
			eventDigest:    common.Digest{0x0},
			expectedDigest: common.Digest{0x1},
		},
		{
			index:          0,
			version:        6,
			auditPath:      common.AuditPath{"1|0": common.Digest{0x1}, "2|1": common.Digest{0x1}, "4|1": common.Digest{0x1}, "6|0": common.Digest{0x6}},
			eventDigest:    common.Digest{0x0},
			expectedDigest: common.Digest{0x7},
		},
		{
			index:          0,
			version:        7,
			auditPath:      common.AuditPath{"1|0": common.Digest{0x1}, "2|1": common.Digest{0x1}, "4|2": common.Digest{0x0}},
			eventDigest:    common.Digest{0x0},
			expectedDigest: common.Digest{0x0},
		},
	}

	for i, c := range testCases {
		proof := NewMembershipProof(c.index, c.version, c.auditPath, common.NewFakeXorHasher())
		correct := proof.Verify(c.expectedDigest, c.eventDigest)
		assert.Truef(t, correct, "Event should be a member for test case %d", i)
	}

}
