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

	"github.com/bbva/qed/balloon/cache"
	"github.com/bbva/qed/crypto/hashing"
	"github.com/stretchr/testify/require"
)

func TestAuditPathVisitor(t *testing.T) {

	testCases := []struct {
		op                operation
		expectedAuditPath AuditPath
	}{
		{
			op:                leafnil(pos(0, 0)),
			expectedAuditPath: AuditPath{},
		},
		{
			op: inner(pos(0, 1),
				collect(getCache(pos(0, 0))),
				leafnil(pos(1, 0)),
			),
			expectedAuditPath: AuditPath{
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
			expectedAuditPath: AuditPath{
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
			expectedAuditPath: AuditPath{
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
			expectedAuditPath: AuditPath{
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
			expectedAuditPath: AuditPath{
				pos(0, 2).FixedBytes(): hashing.Digest{0x0},
				pos(4, 0).FixedBytes(): hashing.Digest{0x0},
			},
		},
	}

	for i, c := range testCases {
		visitor := newAuditPathVisitor(hashing.NewFakeXorHasher(), cache.NewFakeCache([]byte{0x0}))
		c.op.Accept(visitor)
		auditPath := visitor.Result()
		require.Equalf(t, c.expectedAuditPath, auditPath, "The audit path %v should be equal to the expected %v in test case %d", auditPath, c.expectedAuditPath, i)
	}

}
