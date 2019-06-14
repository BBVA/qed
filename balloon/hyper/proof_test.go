/*
   Copyright 2018-2019 Banco Bilbao Vizcaya Argentaria, S.A.

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

	"github.com/bbva/qed/crypto/hashing"
	"github.com/stretchr/testify/assert"
)

func TestProofVerify(t *testing.T) {

	testCases := []struct {
		key, value   []byte
		auditPath    AuditPath
		rootHash     hashing.Digest
		verifyResult bool
	}{
		{
			// verify key=0 with empty audit path
			key:          []byte{0},
			value:        []byte{0},
			auditPath:    AuditPath{},
			rootHash:     hashing.Digest{0x0},
			verifyResult: false,
		},
		{
			// verify key=0 with empty audit path
			key:   []byte{0},
			value: []byte{0},
			auditPath: AuditPath{
				"0x80|7": hashing.Digest{0x0},
				"0x40|6": hashing.Digest{0x0},
				"0x20|5": hashing.Digest{0x0},
				"0x10|4": hashing.Digest{0x0},
			},
			rootHash:     hashing.Digest{0x0},
			verifyResult: true,
		},
		{
			// verify key=0 with empty audit path
			key:   []byte{0},
			value: []byte{0},
			auditPath: AuditPath{
				"0x80|7": hashing.Digest{0x0},
				"0x40|6": hashing.Digest{0x0},
				"0x20|5": hashing.Digest{0x0},
				"0x10|4": hashing.Digest{0x0},
			},
			rootHash:     hashing.Digest{0x1},
			verifyResult: false,
		},
	}

	for i, c := range testCases {
		proof := NewQueryProof(c.key, c.value, c.auditPath, hashing.NewFakeXorHasher())
		correct := proof.Verify(c.key, c.rootHash)
		assert.Equalf(t, c.verifyResult, correct, "The verification result should match for test case %d", i)
	}

}
