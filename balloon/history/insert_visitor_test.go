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
	"github.com/bbva/qed/storage"
	"github.com/stretchr/testify/assert"
)

func TestInsertVisitor(t *testing.T) {

	testCases := []struct {
		op                operation
		expectedMutations []*storage.Mutation
		expectedElements  []*cachedElement
	}{
		{
			op: mutate(putCache(inner(pos(0, 3),
				getCache(pos(0, 2)),
				mutate(putCache(inner(pos(4, 2),
					getCache(pos(4, 1)),
					mutate(putCache(inner(pos(6, 1),
						getCache(pos(6, 0)),
						mutate(putCache(leaf(pos(7, 0), 7))),
					))),
				))),
			))),
			expectedMutations: []*storage.Mutation{
				{
					Table: storage.HistoryTable,
					Key:   pos(7, 0).Bytes(),
					Value: []byte{7},
				},
				{
					Table: storage.HistoryTable,
					Key:   pos(6, 1).Bytes(),
					Value: []byte{7},
				},
				{
					Table: storage.HistoryTable,
					Key:   pos(4, 2).Bytes(),
					Value: []byte{7},
				},
				{
					Table: storage.HistoryTable,
					Key:   pos(0, 3).Bytes(),
					Value: []byte{7},
				},
			},
			expectedElements: []*cachedElement{
				newCachedElement(pos(7, 0), []byte{7}),
				newCachedElement(pos(6, 1), []byte{7}),
				newCachedElement(pos(4, 2), []byte{7}),
				newCachedElement(pos(0, 3), []byte{7}),
			},
		},
	}

	for i, c := range testCases {
		cache := cache.NewFakeCache([]byte{0x0})
		visitor := newInsertVisitor(hashing.NewFakeXorHasher(), cache, storage.HistoryTable)

		c.op.Accept(visitor)

		mutations := visitor.Result()
		assert.ElementsMatchf(t, c.expectedMutations, mutations, "Mutation error in test case %d", i)
		for _, e := range c.expectedElements {
			v, _ := cache.Get(e.Pos.Bytes())
			assert.Equalf(t, e.Digest, v, "The cached element %v should be cached in test case %d", e, i)
		}
	}
}

type cachedElement struct {
	Pos    *position
	Digest []byte
}

func newCachedElement(pos *position, digest []byte) *cachedElement {
	return &cachedElement{pos, digest}
}
