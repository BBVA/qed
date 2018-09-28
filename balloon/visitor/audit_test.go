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

package visitor

import (
	"testing"

	"github.com/bbva/qed/balloon/navigator"
	"github.com/bbva/qed/hashing"
	"github.com/stretchr/testify/require"
)

func TestAuditPathVisitor(t *testing.T) {

	testCases := []struct {
		visitable         Visitable
		expectedAuditPath AuditPath
	}{
		{
			visitable:         NewLeaf(&navigator.FakePosition{[]byte{0x0}, 0}, []byte{0x1}),
			expectedAuditPath: AuditPath{},
		},
		{
			visitable: NewRoot(
				&navigator.FakePosition{[]byte{0x0}, 1},
				NewCached(&navigator.FakePosition{[]byte{0x0}, 0}, hashing.Digest{0x0}),
				NewLeaf(&navigator.FakePosition{[]byte{0x1}, 0}, hashing.Digest{0x1}),
			),
			expectedAuditPath: AuditPath{},
		},
		{
			visitable: NewRoot(
				&navigator.FakePosition{[]byte{0x0}, 2},
				NewCached(&navigator.FakePosition{[]byte{0x0}, 1}, hashing.Digest{0x1}),
				NewPartialNode(&navigator.FakePosition{[]byte{0x1}, 1},
					NewLeaf(&navigator.FakePosition{[]byte{0x2}, 0}, hashing.Digest{0x2}),
				),
			),
			expectedAuditPath: AuditPath{},
		},
		{
			visitable: NewRoot(
				&navigator.FakePosition{[]byte{0x0}, 2},
				NewCached(&navigator.FakePosition{[]byte{0x0}, 1}, hashing.Digest{0x1}),
				NewNode(&navigator.FakePosition{[]byte{0x1}, 1},
					NewCached(&navigator.FakePosition{[]byte{0x2}, 0}, hashing.Digest{0x2}),
					NewLeaf(&navigator.FakePosition{[]byte{0x3}, 0}, hashing.Digest{0x3}),
				),
			),
			expectedAuditPath: AuditPath{},
		},
		{
			visitable: NewRoot(
				&navigator.FakePosition{[]byte{0x0}, 2},
				NewCollectable(NewCached(&navigator.FakePosition{[]byte{0x0}, 1}, hashing.Digest{0x1})),
				NewPartialNode(&navigator.FakePosition{[]byte{0x1}, 1},
					NewLeaf(&navigator.FakePosition{[]byte{0x2}, 0}, hashing.Digest{0x2}),
				),
			),
			expectedAuditPath: AuditPath{
				"00|1": hashing.Digest{0x1},
			},
		},
		{
			visitable: NewRoot(
				&navigator.FakePosition{[]byte{0x0}, 2},
				NewCollectable(NewCached(&navigator.FakePosition{[]byte{0x0}, 1}, hashing.Digest{0x1})),
				NewNode(&navigator.FakePosition{[]byte{0x1}, 1},
					NewCollectable(NewCached(&navigator.FakePosition{[]byte{0x2}, 0}, hashing.Digest{0x2})),
					NewLeaf(&navigator.FakePosition{[]byte{0x3}, 0}, hashing.Digest{0x3}),
				),
			),
			expectedAuditPath: AuditPath{
				"00|1": hashing.Digest{0x1},
				"02|0": hashing.Digest{0x2},
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
