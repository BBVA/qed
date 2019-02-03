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

	"github.com/stretchr/testify/require"

	"github.com/bbva/qed/balloon/cache"
	"github.com/bbva/qed/balloon/history/navigation"
	"github.com/bbva/qed/hashing"
)

func TestCachingVisitor(t *testing.T) {

	testCases := []struct {
		visitable        Visitable
		expectedElements []*cachedElement
	}{
		{
			visitable: mutable(cacheable(leaf(pos(0, 0), 0))),
			expectedElements: []*cachedElement{
				newCachedElement(pos(0, 0), []byte{0x0}),
			},
		},
		{
			visitable: mutable(cacheable(node(pos(0, 1),
				cached(pos(0, 0), 0),
				mutable(cacheable(leaf(pos(1, 0), 1))),
			))),
			expectedElements: []*cachedElement{
				newCachedElement(pos(1, 0), hashing.Digest{0x1}),
				newCachedElement(pos(0, 1), hashing.Digest{0x1}),
			},
		},
		{
			visitable: node(pos(0, 2),
				cached(pos(0, 1), 1),
				partialnode(pos(1, 1),
					mutable(cacheable(leaf(pos(2, 0), 2))),
				),
			),
			expectedElements: []*cachedElement{
				newCachedElement(pos(2, 0), hashing.Digest{0x2}),
			},
		},
		{
			visitable: mutable(cacheable(node(pos(0, 2),
				cached(pos(0, 1), 1),
				mutable(cacheable(node(pos(2, 1),
					cached(pos(2, 0), 2),
					mutable(cacheable(leaf(pos(3, 0), 3)))))),
			))),
			expectedElements: []*cachedElement{
				newCachedElement(pos(3, 0), hashing.Digest{0x3}),
				newCachedElement(pos(2, 1), hashing.Digest{0x1}),
				newCachedElement(pos(0, 2), hashing.Digest{0x0}),
			},
		},
		{
			visitable: node(pos(0, 3),
				cached(pos(0, 2), 0),
				partialnode(pos(4, 2),
					partialnode(pos(4, 1),
						mutable(cacheable(leaf(pos(4, 0), 4)))),
				),
			),
			expectedElements: []*cachedElement{
				newCachedElement(pos(4, 0), hashing.Digest{0x4}),
			},
		},
	}

	for i, c := range testCases {
		cache := cache.NewSimpleCache(0)
		visitor := NewCachingVisitor(NewComputeHashVisitor(hashing.NewFakeXorHasher()), cache)
		c.visitable.PostOrder(visitor)
		for _, e := range c.expectedElements {
			v, _ := cache.Get(e.Pos.Bytes())
			require.Equalf(t, e.Digest, v, "The cached element %+v should be cached in test case %d", e, i)
		}
	}

}

type cachedElement struct {
	Pos    *navigation.Position
	Digest []byte
}

func newCachedElement(pos *navigation.Position, digest []byte) *cachedElement {
	return &cachedElement{pos, digest}
}
