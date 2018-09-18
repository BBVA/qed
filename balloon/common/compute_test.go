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

package common

import (
	"testing"

	"github.com/bbva/qed/hashing"
	"github.com/stretchr/testify/require"
)

func TestComputeHashVisitor(t *testing.T) {

	testCases := []struct {
		visitable      Visitable
		expectedDigest hashing.Digest
	}{
		{
			visitable:      NewLeaf(&FakePosition{[]byte{0x0}, 0}, []byte{0x1}),
			expectedDigest: hashing.Digest{0x1},
		},
		{
			visitable: NewRoot(
				&FakePosition{[]byte{0x0}, 1},
				NewCached(&FakePosition{[]byte{0x0}, 0}, hashing.Digest{0x0}),
				NewLeaf(&FakePosition{[]byte{0x1}, 0}, hashing.Digest{0x1}),
			),
			expectedDigest: hashing.Digest{0x1},
		},
		{
			visitable: NewRoot(
				&FakePosition{[]byte{0x0}, 2},
				NewCached(&FakePosition{[]byte{0x0}, 1}, hashing.Digest{0x1}),
				NewPartialNode(&FakePosition{[]byte{0x1}, 1},
					NewLeaf(&FakePosition{[]byte{0x2}, 0}, hashing.Digest{0x2}),
				),
			),
			expectedDigest: hashing.Digest{0x3},
		},
		{
			visitable: NewRoot(
				&FakePosition{[]byte{0x0}, 2},
				NewCached(&FakePosition{[]byte{0x0}, 1}, hashing.Digest{0x1}),
				NewNode(&FakePosition{[]byte{0x1}, 1},
					NewCached(&FakePosition{[]byte{0x2}, 0}, hashing.Digest{0x2}),
					NewLeaf(&FakePosition{[]byte{0x3}, 0}, hashing.Digest{0x3}),
				),
			),
			expectedDigest: hashing.Digest{0x0},
		},
		{
			visitable: NewRoot(
				&FakePosition{[]byte{0x0}, 2},
				NewCached(&FakePosition{[]byte{0x0}, 1}, hashing.Digest{0x1}),
				NewNode(&FakePosition{[]byte{0x1}, 1},
					NewCached(&FakePosition{[]byte{0x2}, 0}, hashing.Digest{0x2}),
					NewCollectable(NewLeaf(&FakePosition{[]byte{0x3}, 0}, hashing.Digest{0x3})),
				),
			),
			expectedDigest: hashing.Digest{0x0},
		},
	}

	visitor := NewComputeHashVisitor(hashing.NewFakeXorHasher())

	for i, c := range testCases {
		digest := c.visitable.PostOrder(visitor)
		require.Equalf(t, c.expectedDigest, digest, "The computed digest %x should be equal to the expected %x in test case %d", digest, c.expectedDigest, i)
	}
}
