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

	"github.com/stretchr/testify/require"

	"github.com/bbva/qed/balloon/cache"
	"github.com/bbva/qed/balloon/navigator"
	"github.com/bbva/qed/hashing"
)

func TestCachingVisitor(t *testing.T) {

	testCases := []struct {
		visitable        Visitable
		expectedElements []CachedElement
	}{
		{
			visitable: NewCollectable(
				NewCacheable(
					NewLeaf(
						&navigator.FakePosition{[]byte{0x0}, 0},
						[]byte{0x0},
					),
				)),
			expectedElements: []CachedElement{
				*NewCachedElement(
					&navigator.FakePosition{[]byte{0x0}, 0},
					hashing.Digest{0x0},
				),
			},
		},
		{
			visitable: NewCollectable(
				NewCacheable(
					NewRoot(&navigator.FakePosition{[]byte{0x0}, 1},
						NewCached(&navigator.FakePosition{[]byte{0x0}, 0}, hashing.Digest{0x0}),
						NewCollectable(
							NewCacheable(
								NewLeaf(&navigator.FakePosition{[]byte{0x1}, 0}, hashing.Digest{0x1}))),
					))),
			expectedElements: []CachedElement{
				*NewCachedElement(
					&navigator.FakePosition{[]byte{0x1}, 0},
					hashing.Digest{0x1},
				),
				*NewCachedElement(
					&navigator.FakePosition{[]byte{0x0}, 1},
					hashing.Digest{0x1},
				),
			},
		},
		{
			visitable: NewRoot(
				&navigator.FakePosition{[]byte{0x0}, 2},
				NewCached(&navigator.FakePosition{[]byte{0x0}, 1}, hashing.Digest{0x1}),
				NewPartialNode(&navigator.FakePosition{[]byte{0x1}, 1},
					NewCollectable(
						NewCacheable(
							NewLeaf(&navigator.FakePosition{[]byte{0x2}, 0}, hashing.Digest{0x2}),
						)),
				),
			),
			expectedElements: []CachedElement{
				*NewCachedElement(
					&navigator.FakePosition{[]byte{0x2}, 0},
					hashing.Digest{0x2},
				),
			},
		},
		{
			visitable: NewCollectable(
				NewCacheable(
					NewRoot(
						&navigator.FakePosition{[]byte{0x0}, 2},
						NewCached(&navigator.FakePosition{[]byte{0x0}, 1}, hashing.Digest{0x1}),
						NewCollectable(
							NewCacheable(
								NewNode(&navigator.FakePosition{[]byte{0x2}, 1},
									NewCached(&navigator.FakePosition{[]byte{0x2}, 0}, hashing.Digest{0x2}),
									NewCollectable(
										NewCacheable(
											NewLeaf(&navigator.FakePosition{[]byte{0x3}, 0}, hashing.Digest{0x3}),
										))))),
					))),
			expectedElements: []CachedElement{
				*NewCachedElement(
					&navigator.FakePosition{[]byte{0x3}, 0},
					hashing.Digest{0x3},
				),
				*NewCachedElement(
					&navigator.FakePosition{[]byte{0x2}, 1},
					hashing.Digest{0x1},
				),
				*NewCachedElement(
					&navigator.FakePosition{[]byte{0x0}, 2},
					hashing.Digest{0x0},
				),
			},
		},
		{
			visitable: NewRoot(
				&navigator.FakePosition{[]byte{0x0}, 3},
				NewCached(&navigator.FakePosition{[]byte{0x0}, 2}, hashing.Digest{0x0}),
				NewPartialNode(&navigator.FakePosition{[]byte{0x4}, 2},
					NewPartialNode(&navigator.FakePosition{[]byte{0x4}, 1},
						NewCollectable(
							NewCacheable(
								NewLeaf(&navigator.FakePosition{[]byte{0x4}, 0}, hashing.Digest{0x4})))),
				),
			),
			expectedElements: []CachedElement{
				*NewCachedElement(
					&navigator.FakePosition{[]byte{0x4}, 0},
					hashing.Digest{0x4},
				),
			},
		},
	}

	for i, c := range testCases {
		cache := cache.NewSimpleCache(0)
		visitor := NewCachingVisitor(NewComputeHashVisitor(hashing.NewFakeXorHasher()), cache)
		c.visitable.PostOrder(visitor)
		for _, e := range c.expectedElements {
			v, _ := cache.Get(e.Pos)
			require.Equalf(t, e.Digest, v, "The cached element %v should be cached in test case %d", e, i)
		}
	}

}

type CachedElement struct {
	Pos    navigator.Position
	Digest hashing.Digest
}

func NewCachedElement(pos navigator.Position, digest hashing.Digest) *CachedElement {
	return &CachedElement{pos, digest}
}
