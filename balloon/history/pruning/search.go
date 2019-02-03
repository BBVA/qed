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
)

type SearchPruner struct {
	version uint64
	*PruningContext
}

func NewSearchPruner(version uint64, context *PruningContext) *SearchPruner {
	return &SearchPruner{
		version:        version,
		PruningContext: context,
	}
}

func (p *SearchPruner) Prune() (visit.Visitable, error) {
	return p.traverse(navigation.NewRootPosition(p.version))
}

func (p *SearchPruner) traverse(pos *navigation.Position) (visit.Visitable, error) {
	if p.cacheResolver.ShouldGetFromCache(pos) {
		digest, ok := p.cache.Get(pos.Bytes())
		if !ok {
			return nil, ErrCacheNotFound
		}
		return visit.NewCollectable(visit.NewCached(pos, digest)), nil
	}

	if pos.IsLeaf() {
		return visit.NewLeaf(pos, nil), nil
	}

	// we do a post-order traversal
	left, err := p.traverse(pos.Left())
	if err != nil {
		return nil, err
	}

	rightPos := pos.Right()
	if rightPos.Index > p.version {
		return visit.NewPartialNode(pos, left), nil
	}

	right, err := p.traverse(rightPos)
	if err != nil {
		return nil, err
	}

	return visit.NewNode(pos, left, right), nil
}
