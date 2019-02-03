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

package pruning

import (
	"github.com/bbva/qed/balloon/history/navigation"
	"github.com/bbva/qed/balloon/history/visit"
	"github.com/bbva/qed/hashing"
)

func pos(index uint64, height uint16) *navigation.Position {
	return navigation.NewPosition(index, height)
}

func node(pos *navigation.Position, left, right visit.Visitable) *visit.Node {
	return visit.NewNode(pos, left, right)
}

func partialnode(pos *navigation.Position, left visit.Visitable) *visit.PartialNode {
	return visit.NewPartialNode(pos, left)
}

func leaf(pos *navigation.Position, value byte) *visit.Leaf {
	return visit.NewLeaf(pos, []byte{value})
}

func leafnil(pos *navigation.Position) *visit.Leaf {
	return visit.NewLeaf(pos, nil)
}

func cached(pos *navigation.Position) *visit.Cached {
	return visit.NewCached(pos, hashing.Digest{0})
}

func mutable(underlying visit.Visitable) *visit.Mutable {
	return visit.NewMutable(underlying)
}

func collectable(underlying visit.Visitable) *visit.Collectable {
	return visit.NewCollectable(underlying)
}

func cacheable(underlying visit.Visitable) *visit.Cacheable {
	return visit.NewCacheable(underlying)
}
