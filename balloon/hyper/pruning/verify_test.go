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

package pruning

import (
	"testing"

	"github.com/bbva/qed/balloon/hyper/navigation"
	"github.com/bbva/qed/hashing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPruneToVerify(t *testing.T) {

	testCases := []struct {
		index, value []byte
		auditPath    navigation.AuditPath
		expectedOps  []op
	}{
		{
			// verify index=0 with empty audit path
			index:     []byte{0},
			value:     []byte{0},
			auditPath: navigation.AuditPath{},
			expectedOps: []op{
				{LeafHashCode, pos(0, 8)},
			},
		},
		{
			// verify index=0
			index: []byte{0},
			value: []byte{0},
			auditPath: navigation.AuditPath{
				pos(128, 7).StringId(): []byte{0x0},
				pos(64, 6).StringId():  []byte{0x0},
				pos(32, 5).StringId():  []byte{0x0},
				pos(16, 4).StringId():  []byte{0x0},
			},
			expectedOps: []op{
				{InnerHashCode, pos(0, 8)},
				{GetFromPathCode, pos(128, 7)},
				{InnerHashCode, pos(0, 7)},
				{GetFromPathCode, pos(64, 6)},
				{InnerHashCode, pos(0, 6)},
				{GetFromPathCode, pos(32, 5)},
				{InnerHashCode, pos(0, 5)},
				{GetFromPathCode, pos(16, 4)},
				{LeafHashCode, pos(0, 4)},
			},
		},
	}

	for i, c := range testCases {
		prunedOps := PruneToVerify(c.index, c.value, uint16(8-len(c.auditPath))).List()
		require.Truef(t, len(c.expectedOps) == len(prunedOps), "The size of the pruned ops should match the expected for test case %d", i)
		for j := 0; j < len(prunedOps); j++ {
			assert.Equalf(t, c.expectedOps[j].Code, prunedOps[j].Code, "The pruned operation's code should match for test case %d", i)
			assert.Equalf(t, c.expectedOps[j].Pos, prunedOps[j].Pos, "The pruned operation's position should match for test case %d", i)
		}
	}
}

func TestVerifyInterpretation(t *testing.T) {

	testCases := []struct {
		index, value     []byte
		auditPath        navigation.AuditPath
		expectedRootHash hashing.Digest
	}{
		{
			// verify index=0 with empty audit path
			index:            []byte{0},
			value:            []byte{0},
			auditPath:        navigation.AuditPath{},
			expectedRootHash: []byte{0},
		},
		{
			// verify index=0
			index: []byte{0},
			value: []byte{0},
			auditPath: navigation.AuditPath{
				pos(128, 7).StringId(): []byte{0x0},
				pos(64, 6).StringId():  []byte{0x1},
				pos(32, 5).StringId():  []byte{0x2},
				pos(16, 4).StringId():  []byte{0x3},
			},
			expectedRootHash: []byte{0},
		},
	}

	for i, c := range testCases {

		ops := PruneToVerify(c.index, c.value, uint16(8-len(c.auditPath)))
		ctx := &Context{
			Hasher:        hashing.NewFakeXorHasher(),
			Cache:         nil,
			DefaultHashes: nil,
			AuditPath:     c.auditPath,
		}

		rootHash := ops.Pop().Interpret(ops, ctx)
		assert.Equalf(t, c.expectedRootHash, rootHash, "The recomputed root hash should match for test case %d", i)

	}
}
