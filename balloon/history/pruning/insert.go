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

type InsertPruner struct {
	version     uint64
	eventDigest hashing.Digest
	*PruningContext
}

func NewInsertPruner(version uint64, eventDigest hashing.Digest, context *PruningContext) *InsertPruner {
	return &InsertPruner{
		version:        version,
		eventDigest:    eventDigest,
		PruningContext: context,
	}
}

func (p *InsertPruner) Prune() (visit.Visitable, error) {
	return p.traverse(navigation.NewRootPosition(p.version), p.eventDigest)
}

func (p *InsertPruner) traverse(pos *navigation.Position, eventDigest hashing.Digest) (visit.Visitable, error) {

	if p.cacheResolver.ShouldGetFromCache(pos) {
		frozen, ok := p.cache.Get(pos.Bytes())
		if !ok {
			return nil, ErrCacheNotFound
		}
		return visit.NewCached(pos, frozen), nil
	}

	if pos.IsLeaf() {
		return visit.NewMutable(visit.NewCacheable(visit.NewLeaf(pos, eventDigest))), nil
	}

	// we do a post-order traversal
	left, err := p.traverse(pos.Left(), eventDigest)
	if err != nil {
		return nil, err
	}

	rightPos := pos.Right()
	if rightPos.Index > p.version {
		return visit.NewPartialNode(pos, left), nil
	}
	right, err := p.traverse(rightPos, eventDigest)
	if err != nil {
		return nil, err
	}

	result := visit.NewNode(pos, left, right)

	if p.shouldFreeze(pos) {
		return visit.NewMutable(visit.NewCacheable(result)), nil
	}
	return result, nil
}

func (p InsertPruner) shouldFreeze(pos *navigation.Position) bool {
	return p.version >= pos.Index+1<<pos.Height-1
}
