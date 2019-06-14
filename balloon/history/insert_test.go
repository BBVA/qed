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
	"github.com/bbva/qed/testutils/rand"
	"github.com/stretchr/testify/assert"
)

func TestPruneToInsert(t *testing.T) {

	testCases := []struct {
		version     uint64
		eventDigest hashing.Digest
		expectedOp  operation
	}{
		{
			version:     0,
			eventDigest: hashing.Digest{0x0},
			expectedOp:  mutate(putCache(leaf(pos(0, 0), 0))),
		},
		{
			version:     1,
			eventDigest: hashing.Digest{0x1},
			expectedOp: mutate(putCache(inner(pos(0, 1),
				getCache(pos(0, 0)),
				mutate(putCache(leaf(pos(1, 0), 1))),
			))),
		},
		{
			version:     2,
			eventDigest: hashing.Digest{0x2},
			expectedOp: inner(pos(0, 2),
				getCache(pos(0, 1)),
				partial(pos(2, 1),
					mutate(putCache(leaf(pos(2, 0), 2))),
				),
			),
		},
		{
			version:     3,
			eventDigest: hashing.Digest{0x3},
			expectedOp: mutate(putCache(inner(pos(0, 2),
				getCache(pos(0, 1)),
				mutate(putCache(inner(pos(2, 1),
					getCache(pos(2, 0)),
					mutate(putCache(leaf(pos(3, 0), 3)))),
				)),
			))),
		},
		{
			version:     4,
			eventDigest: hashing.Digest{0x4},
			expectedOp: inner(pos(0, 3),
				getCache(pos(0, 2)),
				partial(pos(4, 2),
					partial(pos(4, 1),
						mutate(putCache(leaf(pos(4, 0), 4))),
					),
				),
			),
		},
		{
			version:     5,
			eventDigest: hashing.Digest{0x5},
			expectedOp: inner(pos(0, 3),
				getCache(pos(0, 2)),
				partial(pos(4, 2),
					mutate(putCache(inner(pos(4, 1),
						getCache(pos(4, 0)),
						mutate(putCache(leaf(pos(5, 0), 5))),
					))),
				),
			),
		},
		{
			version:     6,
			eventDigest: hashing.Digest{0x6},
			expectedOp: inner(pos(0, 3),
				getCache(pos(0, 2)),
				inner(pos(4, 2),
					getCache(pos(4, 1)),
					partial(pos(6, 1),
						mutate(putCache(leaf(pos(6, 0), 6))),
					),
				),
			),
		},
		{
			version:     7,
			eventDigest: hashing.Digest{0x7},
			expectedOp: mutate(putCache(inner(pos(0, 3),
				getCache(pos(0, 2)),
				mutate(putCache(inner(pos(4, 2),
					getCache(pos(4, 1)),
					mutate(putCache(inner(pos(6, 1),
						getCache(pos(6, 0)),
						mutate(putCache(leaf(pos(7, 0), 7))),
					))),
				))),
			))),
		},
	}

	for i, c := range testCases {
		prunedOp := pruneToInsert(c.version, c.eventDigest)
		assert.Equalf(t, c.expectedOp, prunedOp, "The pruned operation should match for test case %d", i)
	}

}

func BenchmarkPruneToInsert(b *testing.B) {

	log.SetLogger("BenchmarkPruneToInsert", log.SILENT)

	b.ResetTimer()
	for i := uint64(0); i < uint64(b.N); i++ {
		pruned := pruneToInsert(i, rand.Bytes(32))
		assert.NotNil(b, pruned)
	}

}
