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

package history

import (
	"testing"

	"github.com/bbva/qed/balloon/history/navigation"
	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/log"
	"github.com/stretchr/testify/assert"
)

func TestVerifyMembershipProof(t *testing.T) {

	log.SetLogger("TestVerifyMembershipProof", log.INFO)

	testCases := []struct {
		index, version uint64
		auditPath      navigation.AuditPath
		eventDigest    hashing.Digest
		expectedDigest hashing.Digest
	}{
		{
			index:          0,
			version:        0,
			auditPath:      navigation.AuditPath{},
			eventDigest:    hashing.Digest{0x0},
			expectedDigest: hashing.Digest{0x0},
		},
		{
			index:          1,
			version:        1,
			auditPath:      navigation.AuditPath{pos(0, 0).FixedBytes(): hashing.Digest{0x0}},
			eventDigest:    hashing.Digest{0x1},
			expectedDigest: hashing.Digest{0x1},
		},
		{
			index:          2,
			version:        2,
			auditPath:      navigation.AuditPath{pos(0, 1).FixedBytes(): hashing.Digest{0x1}},
			eventDigest:    hashing.Digest{0x2},
			expectedDigest: hashing.Digest{0x3},
		},
		{
			index:          3,
			version:        3,
			auditPath:      navigation.AuditPath{pos(0, 1).FixedBytes(): hashing.Digest{0x1}, pos(2, 0).FixedBytes(): hashing.Digest{0x2}},
			eventDigest:    hashing.Digest{0x3},
			expectedDigest: hashing.Digest{0x0},
		},
		{
			index:          4,
			version:        4,
			auditPath:      navigation.AuditPath{pos(0, 2).FixedBytes(): hashing.Digest{0x0}},
			eventDigest:    hashing.Digest{0x4},
			expectedDigest: hashing.Digest{0x4},
		},
		{
			index:          5,
			version:        5,
			auditPath:      navigation.AuditPath{pos(0, 2).FixedBytes(): hashing.Digest{0x0}, pos(4, 0).FixedBytes(): hashing.Digest{0x4}},
			eventDigest:    hashing.Digest{0x5},
			expectedDigest: hashing.Digest{0x1},
		},
		{
			index:          6,
			version:        6,
			auditPath:      navigation.AuditPath{pos(0, 2).FixedBytes(): hashing.Digest{0x0}, pos(4, 1).FixedBytes(): hashing.Digest{0x1}},
			eventDigest:    hashing.Digest{0x6},
			expectedDigest: hashing.Digest{0x7},
		},
		{
			index:          7,
			version:        7,
			auditPath:      navigation.AuditPath{pos(0, 2).FixedBytes(): hashing.Digest{0x0}, pos(4, 1).FixedBytes(): hashing.Digest{0x1}, pos(6, 0).FixedBytes(): hashing.Digest{0x6}},
			eventDigest:    hashing.Digest{0x7},
			expectedDigest: hashing.Digest{0x0},
		},
		{
			index:          8,
			version:        8,
			auditPath:      navigation.AuditPath{pos(0, 3).FixedBytes(): hashing.Digest{0x0}},
			eventDigest:    hashing.Digest{0x8},
			expectedDigest: hashing.Digest{0x8},
		},
		{
			index:          9,
			version:        9,
			auditPath:      navigation.AuditPath{pos(0, 3).FixedBytes(): hashing.Digest{0x0}, pos(8, 0).FixedBytes(): hashing.Digest{0x8}},
			eventDigest:    hashing.Digest{0x9},
			expectedDigest: hashing.Digest{0x1},
		},
		{
			index:          0,
			version:        1,
			auditPath:      navigation.AuditPath{pos(1, 0).FixedBytes(): hashing.Digest{0x1}},
			eventDigest:    hashing.Digest{0x0},
			expectedDigest: hashing.Digest{0x1},
		},
		{
			index:          0,
			version:        1,
			auditPath:      navigation.AuditPath{pos(1, 0).FixedBytes(): hashing.Digest{0x1}},
			eventDigest:    hashing.Digest{0x0},
			expectedDigest: hashing.Digest{0x1},
		},
		{
			index:          0,
			version:        2,
			auditPath:      navigation.AuditPath{pos(1, 0).FixedBytes(): hashing.Digest{0x1}, pos(2, 0).FixedBytes(): hashing.Digest{0x2}},
			eventDigest:    hashing.Digest{0x0},
			expectedDigest: hashing.Digest{0x3},
		},
		{
			index:          0,
			version:        3,
			auditPath:      navigation.AuditPath{pos(1, 0).FixedBytes(): hashing.Digest{0x1}, pos(2, 1).FixedBytes(): hashing.Digest{0x1}},
			eventDigest:    hashing.Digest{0x0},
			expectedDigest: hashing.Digest{0x0},
		},
		{
			index:          0,
			version:        4,
			auditPath:      navigation.AuditPath{pos(1, 0).FixedBytes(): hashing.Digest{0x1}, pos(2, 1).FixedBytes(): hashing.Digest{0x1}, pos(4, 0).FixedBytes(): hashing.Digest{0x4}},
			eventDigest:    hashing.Digest{0x0},
			expectedDigest: hashing.Digest{0x4},
		},
		{
			index:          0,
			version:        5,
			auditPath:      navigation.AuditPath{pos(1, 0).FixedBytes(): hashing.Digest{0x1}, pos(2, 1).FixedBytes(): hashing.Digest{0x1}, pos(4, 1).FixedBytes(): hashing.Digest{0x1}},
			eventDigest:    hashing.Digest{0x0},
			expectedDigest: hashing.Digest{0x1},
		},
		{
			index:          0,
			version:        6,
			auditPath:      navigation.AuditPath{pos(1, 0).FixedBytes(): hashing.Digest{0x1}, pos(2, 1).FixedBytes(): hashing.Digest{0x1}, pos(4, 1).FixedBytes(): hashing.Digest{0x1}, pos(6, 0).FixedBytes(): hashing.Digest{0x6}},
			eventDigest:    hashing.Digest{0x0},
			expectedDigest: hashing.Digest{0x7},
		},
		{
			index:          0,
			version:        7,
			auditPath:      navigation.AuditPath{pos(1, 0).FixedBytes(): hashing.Digest{0x1}, pos(2, 1).FixedBytes(): hashing.Digest{0x1}, pos(4, 2).FixedBytes(): hashing.Digest{0x0}},
			eventDigest:    hashing.Digest{0x0},
			expectedDigest: hashing.Digest{0x0},
		},
	}

	for i, c := range testCases {
		proof := NewMembershipProof(c.index, c.version, c.auditPath, hashing.NewFakeXorHasher())
		correct := proof.Verify(c.expectedDigest, c.eventDigest)
		assert.Truef(t, correct, "Event should be a member for test case %d", i)
	}

}

func TestVerifyIncrementalProof(t *testing.T) {

	log.SetLogger("TestVerifyIncrementalProof", log.INFO)

	testCases := []struct {
		auditPath           navigation.AuditPath
		start               uint64
		end                 uint64
		expectedStartDigest hashing.Digest
		expectedEndDigest   hashing.Digest
	}{
		{
			auditPath:           navigation.AuditPath{pos(0, 1).FixedBytes(): hashing.Digest{0x1}, pos(2, 0).FixedBytes(): hashing.Digest{0x2}, pos(3, 0).FixedBytes(): hashing.Digest{0x3}, pos(4, 1).FixedBytes(): hashing.Digest{0x1}, pos(6, 0).FixedBytes(): hashing.Digest{0x6}},
			start:               2,
			end:                 6,
			expectedStartDigest: hashing.Digest{0x3},
			expectedEndDigest:   hashing.Digest{0x7},
		},
		{
			auditPath:           navigation.AuditPath{pos(0, 1).FixedBytes(): hashing.Digest{0x1}, pos(2, 0).FixedBytes(): hashing.Digest{0x2}, pos(3, 0).FixedBytes(): hashing.Digest{0x3}, pos(4, 2).FixedBytes(): hashing.Digest{0x0}},
			start:               2,
			end:                 7,
			expectedStartDigest: hashing.Digest{0x3},
			expectedEndDigest:   hashing.Digest{0x0},
		},
		{
			auditPath:           navigation.AuditPath{pos(0, 2).FixedBytes(): hashing.Digest{0x0}, pos(4, 0).FixedBytes(): hashing.Digest{0x4}, pos(5, 0).FixedBytes(): hashing.Digest{0x5}, pos(6, 0).FixedBytes(): hashing.Digest{0x6}},
			start:               4,
			end:                 6,
			expectedStartDigest: hashing.Digest{0x4},
			expectedEndDigest:   hashing.Digest{0x7},
		},
		{
			auditPath:           navigation.AuditPath{pos(0, 2).FixedBytes(): hashing.Digest{0x0}, pos(4, 0).FixedBytes(): hashing.Digest{0x4}, pos(5, 0).FixedBytes(): hashing.Digest{0x5}, pos(6, 1).FixedBytes(): hashing.Digest{0x1}},
			start:               4,
			end:                 7,
			expectedStartDigest: hashing.Digest{0x4},
			expectedEndDigest:   hashing.Digest{0x0},
		},
		{
			auditPath:           navigation.AuditPath{pos(2, 0).FixedBytes(): hashing.Digest{0x2}, pos(3, 0).FixedBytes(): hashing.Digest{0x3}, pos(4, 0).FixedBytes(): hashing.Digest{0x4}, pos(0, 1).FixedBytes(): hashing.Digest{0x1}},
			start:               2,
			end:                 4,
			expectedStartDigest: hashing.Digest{0x3},
			expectedEndDigest:   hashing.Digest{0x4},
		},
		{
			auditPath:           navigation.AuditPath{pos(0, 2).FixedBytes(): hashing.Digest{0x0}, pos(4, 1).FixedBytes(): hashing.Digest{0x1}, pos(6, 0).FixedBytes(): hashing.Digest{0x6}, pos(7, 0).FixedBytes(): hashing.Digest{0x7}},
			start:               6,
			end:                 7,
			expectedStartDigest: hashing.Digest{0x7},
			expectedEndDigest:   hashing.Digest{0x0},
		},
	}

	hasher := hashing.NewFakeXorHasher()
	for _, c := range testCases {
		proof := NewIncrementalProof(c.start, c.end, c.auditPath, hasher)
		correct := proof.Verify(c.expectedStartDigest, c.expectedEndDigest)
		assert.Truef(t, correct, "Events between %d and %d should be consistent", c.start, c.end)
	}
}
