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

package history

import (
	"testing"

	"github.com/bbva/qed/crypto/hashing"
	"github.com/bbva/qed/log"
	"github.com/stretchr/testify/assert"
)

func TestAuditPathSerialization(t *testing.T) {

	testCases := []struct {
		path     AuditPath
		expected map[string]hashing.Digest
	}{
		{
			AuditPath{},
			map[string]hashing.Digest{},
		},
		{
			AuditPath{
				newPosition(0, 0).FixedBytes(): []byte{0x0},
			},
			map[string]hashing.Digest{
				"0|0": []byte{0x0},
			},
		},
		{
			AuditPath{
				newPosition(0, 0).FixedBytes(): []byte{0x0},
				newPosition(2, 1).FixedBytes(): []byte{0x1},
				newPosition(4, 2).FixedBytes(): []byte{0x2},
			},
			map[string]hashing.Digest{
				"0|0": []byte{0x0},
				"2|1": []byte{0x1},
				"4|2": []byte{0x2},
			},
		},
	}

	for i, c := range testCases {
		serialized := c.path.Serialize()
		assert.Equalf(t, c.expected, serialized, "The serialized paths should match for test case %d", i)
	}

}

func TestParseAuditPath(t *testing.T) {

	testCases := []struct {
		path     map[string]hashing.Digest
		expected AuditPath
	}{
		{
			map[string]hashing.Digest{},
			AuditPath{},
		},
		{
			map[string]hashing.Digest{
				"0|0": []byte{0x0},
			},
			AuditPath{
				newPosition(0, 0).FixedBytes(): []byte{0x0},
			},
		},
		{
			map[string]hashing.Digest{
				"0|0": []byte{0x0},
				"2|1": []byte{0x1},
				"4|2": []byte{0x2},
			},
			AuditPath{
				newPosition(0, 0).FixedBytes(): []byte{0x0},
				newPosition(2, 1).FixedBytes(): []byte{0x1},
				newPosition(4, 2).FixedBytes(): []byte{0x2},
			},
		},
	}

	for i, c := range testCases {
		parsed := ParseAuditPath(c.path)
		assert.Equalf(t, c.expected, parsed, "The parsed paths should match for test case %d", i)
	}

}

func TestVerifyMembershipProof(t *testing.T) {

	log.SetLogger("TestVerifyMembershipProof", log.INFO)

	testCases := []struct {
		index, version uint64
		auditPath      AuditPath
		eventDigest    hashing.Digest
		expectedDigest hashing.Digest
		verifies       bool
	}{
		{
			index:          0,
			version:        0,
			auditPath:      AuditPath{},
			eventDigest:    hashing.Digest{0x0},
			expectedDigest: hashing.Digest{0x0},
			verifies:       true,
		},
		{
			index:   1,
			version: 1,
			auditPath: AuditPath{
				pos(0, 0).FixedBytes(): hashing.Digest{0x0},
			},
			eventDigest:    hashing.Digest{0x1},
			expectedDigest: hashing.Digest{0x1},
			verifies:       true,
		},
		{
			index:   1,
			version: 1,
			auditPath: AuditPath{
				pos(0, 0).FixedBytes(): hashing.Digest{0x1},
			},
			eventDigest:    hashing.Digest{0x1},
			expectedDigest: hashing.Digest{0x1},
			verifies:       false,
		},
		{
			index:   2,
			version: 2,
			auditPath: AuditPath{
				pos(0, 1).FixedBytes(): hashing.Digest{0x1},
			},
			eventDigest:    hashing.Digest{0x2},
			expectedDigest: hashing.Digest{0x3},
			verifies:       true,
		},
		{
			index:   3,
			version: 3,
			auditPath: AuditPath{
				pos(0, 1).FixedBytes(): hashing.Digest{0x1},
				pos(2, 0).FixedBytes(): hashing.Digest{0x2},
			},
			eventDigest:    hashing.Digest{0x3},
			expectedDigest: hashing.Digest{0x0},
			verifies:       true,
		},
		{
			index:   4,
			version: 4,
			auditPath: AuditPath{
				pos(0, 2).FixedBytes(): hashing.Digest{0x0},
			},
			eventDigest:    hashing.Digest{0x4},
			expectedDigest: hashing.Digest{0x4},
			verifies:       true,
		},
		{
			index:   5,
			version: 5,
			auditPath: AuditPath{
				pos(0, 2).FixedBytes(): hashing.Digest{0x0},
				pos(4, 0).FixedBytes(): hashing.Digest{0x4},
			},
			eventDigest:    hashing.Digest{0x5},
			expectedDigest: hashing.Digest{0x1},
			verifies:       true,
		},
		{
			index:   6,
			version: 6,
			auditPath: AuditPath{
				pos(0, 2).FixedBytes(): hashing.Digest{0x0},
				pos(4, 1).FixedBytes(): hashing.Digest{0x1},
			},
			eventDigest:    hashing.Digest{0x6},
			expectedDigest: hashing.Digest{0x7},
			verifies:       true,
		},
		{
			index:   7,
			version: 7,
			auditPath: AuditPath{
				pos(0, 2).FixedBytes(): hashing.Digest{0x0},
				pos(4, 1).FixedBytes(): hashing.Digest{0x1},
				pos(6, 0).FixedBytes(): hashing.Digest{0x6},
			},
			eventDigest:    hashing.Digest{0x7},
			expectedDigest: hashing.Digest{0x0},
			verifies:       true,
		},
		{
			index:   8,
			version: 8,
			auditPath: AuditPath{
				pos(0, 3).FixedBytes(): hashing.Digest{0x0},
			},
			eventDigest:    hashing.Digest{0x8},
			expectedDigest: hashing.Digest{0x8},
			verifies:       true,
		},
		{
			index:   9,
			version: 9,
			auditPath: AuditPath{
				pos(0, 3).FixedBytes(): hashing.Digest{0x0},
				pos(8, 0).FixedBytes(): hashing.Digest{0x8},
			},
			eventDigest:    hashing.Digest{0x9},
			expectedDigest: hashing.Digest{0x1},
			verifies:       true,
		},
		{
			index:   0,
			version: 1,
			auditPath: AuditPath{
				pos(1, 0).FixedBytes(): hashing.Digest{0x1},
			},
			eventDigest:    hashing.Digest{0x0},
			expectedDigest: hashing.Digest{0x1},
			verifies:       true,
		},
		{
			index:   0,
			version: 1,
			auditPath: AuditPath{
				pos(1, 0).FixedBytes(): hashing.Digest{0x1},
			},
			eventDigest:    hashing.Digest{0x0},
			expectedDigest: hashing.Digest{0x1},
			verifies:       true,
		},
		{
			index:   0,
			version: 2,
			auditPath: AuditPath{
				pos(1, 0).FixedBytes(): hashing.Digest{0x1},
				pos(2, 1).FixedBytes(): hashing.Digest{0x2},
			},
			eventDigest:    hashing.Digest{0x0},
			expectedDigest: hashing.Digest{0x3},
			verifies:       true,
		},
		{
			index:   0,
			version: 3,
			auditPath: AuditPath{
				pos(1, 0).FixedBytes(): hashing.Digest{0x1},
				pos(2, 1).FixedBytes(): hashing.Digest{0x1},
			},
			eventDigest:    hashing.Digest{0x0},
			expectedDigest: hashing.Digest{0x0},
			verifies:       true,
		},
		{
			index:   0,
			version: 4,
			auditPath: AuditPath{
				pos(1, 0).FixedBytes(): hashing.Digest{0x1},
				pos(2, 1).FixedBytes(): hashing.Digest{0x1},
				pos(4, 2).FixedBytes(): hashing.Digest{0x4},
			},
			eventDigest:    hashing.Digest{0x0},
			expectedDigest: hashing.Digest{0x4},
			verifies:       true,
		},
		{
			index:   0,
			version: 5,
			auditPath: AuditPath{
				pos(1, 0).FixedBytes(): hashing.Digest{0x1},
				pos(2, 1).FixedBytes(): hashing.Digest{0x1},
				pos(4, 2).FixedBytes(): hashing.Digest{0x1},
			},
			eventDigest:    hashing.Digest{0x0},
			expectedDigest: hashing.Digest{0x1},
			verifies:       true,
		},
		{
			index:   0,
			version: 6,
			auditPath: AuditPath{
				pos(1, 0).FixedBytes(): hashing.Digest{0x1},
				pos(2, 1).FixedBytes(): hashing.Digest{0x1},
				pos(4, 2).FixedBytes(): hashing.Digest{0x7},
			},
			eventDigest:    hashing.Digest{0x0},
			expectedDigest: hashing.Digest{0x7},
			verifies:       true,
		},
		{
			index:   0,
			version: 7,
			auditPath: AuditPath{
				pos(1, 0).FixedBytes(): hashing.Digest{0x1},
				pos(2, 1).FixedBytes(): hashing.Digest{0x1},
				pos(4, 2).FixedBytes(): hashing.Digest{0x0},
			},
			eventDigest:    hashing.Digest{0x0},
			expectedDigest: hashing.Digest{0x0},
			verifies:       true,
		},
	}

	for _, c := range testCases {
		proof := NewMembershipProof(c.index, c.version, c.auditPath, hashing.NewFakeXorHasher())
		correct := proof.Verify(c.expectedDigest, c.eventDigest)
		assert.Equalf(t, c.verifies, correct, "The membership proof should be valid for test case with index %d and version %d", c.index, c.version)
	}

}

func TestVerifyIncrementalProof(t *testing.T) {

	log.SetLogger("TestVerifyIncrementalProof", log.INFO)

	testCases := []struct {
		auditPath           AuditPath
		start               uint64
		end                 uint64
		expectedStartDigest hashing.Digest
		expectedEndDigest   hashing.Digest
		verifies            bool
	}{
		{
			auditPath: AuditPath{
				pos(0, 1).FixedBytes(): hashing.Digest{0x1},
				pos(2, 0).FixedBytes(): hashing.Digest{0x2},
				pos(3, 0).FixedBytes(): hashing.Digest{0x3},
				pos(4, 1).FixedBytes(): hashing.Digest{0x1},
				pos(6, 0).FixedBytes(): hashing.Digest{0x6},
			},
			start:               2,
			end:                 6,
			expectedStartDigest: hashing.Digest{0x3},
			expectedEndDigest:   hashing.Digest{0x7},
			verifies:            true,
		},
		{
			auditPath: AuditPath{
				pos(0, 1).FixedBytes(): hashing.Digest{0x1},
				pos(2, 0).FixedBytes(): hashing.Digest{0x2},
				pos(3, 0).FixedBytes(): hashing.Digest{0x3},
				pos(4, 1).FixedBytes(): hashing.Digest{0x1},
				pos(6, 0).FixedBytes(): hashing.Digest{0x6},
				pos(7, 0).FixedBytes(): hashing.Digest{0x7},
			},
			start:               2,
			end:                 7,
			expectedStartDigest: hashing.Digest{0x3},
			expectedEndDigest:   hashing.Digest{0x0},
			verifies:            true,
		},
		{
			auditPath: AuditPath{
				pos(0, 2).FixedBytes(): hashing.Digest{0x0},
				pos(4, 0).FixedBytes(): hashing.Digest{0x4},
				pos(5, 0).FixedBytes(): hashing.Digest{0x5},
				pos(6, 0).FixedBytes(): hashing.Digest{0x6},
			},
			start:               4,
			end:                 6,
			expectedStartDigest: hashing.Digest{0x4},
			expectedEndDigest:   hashing.Digest{0x7},
			verifies:            true,
		},
		{
			auditPath: AuditPath{
				pos(0, 2).FixedBytes(): hashing.Digest{0x0},
				pos(4, 0).FixedBytes(): hashing.Digest{0x4},
				pos(5, 0).FixedBytes(): hashing.Digest{0x5},
				pos(6, 0).FixedBytes(): hashing.Digest{0x6},
				pos(7, 0).FixedBytes(): hashing.Digest{0x7},
			},
			start:               4,
			end:                 7,
			expectedStartDigest: hashing.Digest{0x4},
			expectedEndDigest:   hashing.Digest{0x0},
			verifies:            true,
		},
		{
			auditPath: AuditPath{
				pos(0, 1).FixedBytes(): hashing.Digest{0x1},
				pos(2, 0).FixedBytes(): hashing.Digest{0x2},
				pos(3, 0).FixedBytes(): hashing.Digest{0x3},
				pos(4, 0).FixedBytes(): hashing.Digest{0x4},
			},
			start:               2,
			end:                 4,
			expectedStartDigest: hashing.Digest{0x3},
			expectedEndDigest:   hashing.Digest{0x4},
			verifies:            true,
		},
		{
			auditPath: AuditPath{
				pos(0, 2).FixedBytes(): hashing.Digest{0x0},
				pos(4, 1).FixedBytes(): hashing.Digest{0x1},
				pos(6, 0).FixedBytes(): hashing.Digest{0x6},
				pos(7, 0).FixedBytes(): hashing.Digest{0x7},
			},
			start:               6,
			end:                 7,
			expectedStartDigest: hashing.Digest{0x7},
			expectedEndDigest:   hashing.Digest{0x0},
			verifies:            true,
		},
	}

	hasher := hashing.NewFakeXorHasher()
	for _, c := range testCases {
		proof := NewIncrementalProof(c.start, c.end, c.auditPath, hasher)
		correct := proof.Verify(c.expectedStartDigest, c.expectedEndDigest)
		assert.Equalf(t, c.verifies, correct, "The incremental proof between %d and %d should be valid", c.start, c.end)
	}
}
