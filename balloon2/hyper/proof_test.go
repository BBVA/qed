package hyper

import (
	"testing"

	"github.com/bbva/qed/balloon2/common"
	"github.com/stretchr/testify/assert"
)

func TestQueryProofVerify(t *testing.T) {
	testCases := []struct {
		key, value     []byte
		auditPath      common.AuditPath
		expectedDigest common.Digest
	}{
		{
			key:   []byte{0},
			value: []byte{0},
			auditPath: common.AuditPath{
				"01|0": common.Digest{0x0},
				"02|1": common.Digest{0x0},
				"04|2": common.Digest{0x0},
				"08|3": common.Digest{0x0},
				"10|4": common.Digest{0x0},
				"20|5": common.Digest{0x0},
				"40|6": common.Digest{0x0},
				"80|7": common.Digest{0x0},
			},
			expectedDigest: common.Digest{0},
		},
	}

	for i, c := range testCases {
		proof := NewQueryProof(c.key, c.value, c.auditPath, common.NewFakeXorHasher())
		correct := proof.Verify(c.key, c.expectedDigest)
		assert.Truef(t, correct, "Event should be a member for test case %d", i)
	}
}
