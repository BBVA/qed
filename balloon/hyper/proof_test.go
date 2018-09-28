/*
   Copyright 2018 Banco Bilbao Vizcaya Argentaria, S.A.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package hyper

import (
	"testing"

	"github.com/bbva/qed/balloon/visitor"
	"github.com/bbva/qed/hashing"
	"github.com/stretchr/testify/assert"
)

func TestQueryProofVerify(t *testing.T) {
	testCases := []struct {
		key, value     []byte
		auditPath      visitor.AuditPath
		expectedDigest hashing.Digest
	}{
		{
			key:   []byte{0},
			value: []byte{0},
			auditPath: visitor.AuditPath{
				"01|0": hashing.Digest{0x0},
				"02|1": hashing.Digest{0x0},
				"04|2": hashing.Digest{0x0},
				"08|3": hashing.Digest{0x0},
				"10|4": hashing.Digest{0x0},
				"20|5": hashing.Digest{0x0},
				"40|6": hashing.Digest{0x0},
				"80|7": hashing.Digest{0x0},
			},
			expectedDigest: hashing.Digest{0},
		},
	}

	for i, c := range testCases {
		proof := NewQueryProof(c.key, c.value, c.auditPath, hashing.NewFakeXorHasher())
		correct := proof.Verify(c.key, c.expectedDigest)
		assert.Truef(t, correct, "Event should be a member for test case %d", i)
	}
}
