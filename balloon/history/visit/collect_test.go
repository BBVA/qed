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

	assert "github.com/stretchr/testify/require"

	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/storage"
)

func TestCollect(t *testing.T) {

	testCases := []struct {
		visitable         Visitable
		expectedMutations []*storage.Mutation
	}{
		{
			visitable: node(pos(0, 8),
				node(pos(0, 9),
					mutable(cached(pos(0, 7), 0)), node(pos(0, 9),
						mutable(cached(pos(0, 6), 1)),
						leaf(pos(0, 1), 0),
					),
				),
				node(pos(0, 9),
					leaf(pos(0, 1), 0),
					mutable(cached(pos(0, 8), 2)),
				),
			),
			expectedMutations: []*storage.Mutation{
				{storage.HyperCachePrefix, pos(0, 7).Bytes(), []byte{0}},
				{storage.HyperCachePrefix, pos(0, 6).Bytes(), []byte{1}},
				{storage.HyperCachePrefix, pos(0, 8).Bytes(), []byte{2}},
			},
		},
	}

	for i, c := range testCases {
		decorated := NewComputeHashVisitor(hashing.NewFakeXorHasher())
		visitor := NewCollectMutationsVisitor(decorated, storage.HyperCachePrefix)
		c.visitable.PostOrder(visitor)

		mutations := visitor.Result()
		assert.ElementsMatchf(
			t,
			mutations,
			c.expectedMutations,
			"Mutation error in test case %d",
			i,
		)
	}
}
