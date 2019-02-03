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

package visit

import (
	"testing"

	"github.com/bbva/qed/balloon/history/navigation"
	"github.com/bbva/qed/hashing"
	"github.com/stretchr/testify/require"
)

func TestAuditPathVisitor(t *testing.T) {

	testCases := []struct {
		visitable         Visitable
		expectedAuditPath navigation.AuditPath
	}{
		{
			visitable:         leaf(pos(0, 0), 1),
			expectedAuditPath: navigation.AuditPath{},
		},
		{
			visitable: node(pos(0, 1),
				cached(pos(0, 0), 0),
				leaf(pos(1, 0), 1),
			),
			expectedAuditPath: navigation.AuditPath{},
		},
		{
			visitable: node(pos(0, 2),
				cached(pos(0, 1), 1),
				partialnode(pos(1, 1),
					leaf(pos(2, 0), 2),
				),
			),
			expectedAuditPath: navigation.AuditPath{},
		},
		{
			visitable: node(pos(0, 2),
				cached(pos(0, 1), 1),
				node(pos(1, 1),
					cached(pos(2, 0), 2),
					leaf(pos(3, 0), 3),
				),
			),
			expectedAuditPath: navigation.AuditPath{},
		},
		{
			visitable: node(pos(0, 2),
				collectable(cached(pos(0, 1), 1)),
				partialnode(pos(1, 1),
					leaf(pos(2, 0), 2),
				),
			),
			expectedAuditPath: navigation.AuditPath{
				pos(0, 1).FixedBytes(): hashing.Digest{0x1},
			},
		},
		{
			visitable: node(pos(0, 2),
				collectable(cached(pos(0, 1), 1)),
				node(pos(1, 1),
					collectable(cached(pos(2, 0), 2)),
					leaf(pos(3, 0), 3),
				),
			),
			expectedAuditPath: navigation.AuditPath{
				pos(0, 1).FixedBytes(): hashing.Digest{0x1},
				pos(2, 0).FixedBytes(): hashing.Digest{0x2},
			},
		},
	}

	for i, c := range testCases {
		visitor := NewAuditPathVisitor(NewComputeHashVisitor(hashing.NewFakeXorHasher()))
		c.visitable.PostOrder(visitor)
		auditPath := visitor.Result()
		require.Equalf(t, c.expectedAuditPath, auditPath, "The audit path %v should be equal to the expected %v in test case %d", auditPath, c.expectedAuditPath, i)
	}

}
