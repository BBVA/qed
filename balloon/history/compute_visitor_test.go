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

func TestComputeHashVisitor(t *testing.T) {

	testCases := []struct {
		op             operation
		expectedDigest hashing.Digest
	}{
		{
			op:             leaf(pos(0, 0), 1),
			expectedDigest: hashing.Digest{0x1},
		},
		{
			op: inner(pos(0, 1),
				getCache(pos(0, 0)),
				leaf(pos(1, 0), 1),
			),
			expectedDigest: hashing.Digest{0x1},
		},
		{
			op: inner(pos(0, 2),
				getCache(pos(0, 1)),
				partial(pos(1, 1),
					leaf(pos(2, 0), 2),
				),
			),
			expectedDigest: hashing.Digest{0x2},
		},
		{
			op: inner(pos(0, 2),
				getCache(pos(0, 1)),
				inner(pos(1, 1),
					getCache(pos(2, 0)),
					leaf(pos(3, 0), 3),
				),
			),
			expectedDigest: hashing.Digest{0x3},
		},
		{
			op: inner(pos(0, 2),
				getCache(pos(0, 1)),
				inner(pos(1, 1),
					getCache(pos(2, 0)),
					mutate(leaf(pos(3, 0), 3)),
				),
			),
			expectedDigest: hashing.Digest{0x3},
		},
	}

	visitor := newComputeHashVisitor(hashing.NewFakeXorHasher(), cache.NewFakeCache([]byte{0x0}))

	for i, c := range testCases {
		digest := c.op.Accept(visitor)
		require.Equalf(t, c.expectedDigest, digest, "The computed digest %x should be equal to the expected %x in test case %d", digest, c.expectedDigest, i)
	}
}
