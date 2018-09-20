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

package history

import (
	"github.com/bbva/qed/balloon/common"
	"github.com/bbva/qed/hashing"
)

type PruningContext struct {
	navigator     common.TreeNavigator
	cacheResolver CacheResolver
	cache         common.Cache
}

type Pruner interface {
	Prune() common.Visitable
}

type InsertPruner struct {
	version     uint64
	eventDigest hashing.Digest
	PruningContext
}

func NewInsertPruner(version uint64, eventDigest hashing.Digest, context PruningContext) *InsertPruner {
	return &InsertPruner{version, eventDigest, context}
}

func (p *InsertPruner) Prune() common.Visitable {
	return p.traverse(p.navigator.Root(), p.eventDigest)
}

func (p *InsertPruner) traverse(pos common.Position, eventDigest hashing.Digest) common.Visitable {
	if p.cacheResolver.ShouldGetFromCache(pos) {
		digest, ok := p.cache.Get(pos)
		if !ok {
			panic("this digest should be in cache")
		}
		return common.NewCached(pos, digest)
	}
	if p.navigator.IsLeaf(pos) {
		return common.NewCollectable(common.NewLeaf(pos, eventDigest))
	}
	// we do a post-order traversal
	left := p.traverse(p.navigator.GoToLeft(pos), eventDigest)
	rightPos := p.navigator.GoToRight(pos)
	if rightPos == nil {
		return common.NewPartialNode(pos, left)
	}
	right := p.traverse(rightPos, eventDigest)
	var result common.Visitable
	if p.navigator.IsRoot(pos) {
		result = common.NewRoot(pos, left, right)
	} else {
		result = common.NewNode(pos, left, right)
	}
	if p.shouldCollect(pos) {
		return common.NewCollectable(result)
	}
	return result
}

func (p InsertPruner) shouldCollect(pos common.Position) bool {
	return p.version >= pos.IndexAsUint64()+1<<pos.Height()-1
}

type SearchPruner struct {
	PruningContext
}

func NewSearchPruner(context PruningContext) *SearchPruner {
	return &SearchPruner{context}
}

func (p *SearchPruner) Prune() common.Visitable {
	return p.traverse(p.navigator.Root())
}

func (p *SearchPruner) traverse(pos common.Position) common.Visitable {
	if p.cacheResolver.ShouldGetFromCache(pos) {
		digest, ok := p.cache.Get(pos)
		if !ok {
			panic("this digest should be in cache")
		}
		return common.NewCollectable(common.NewCached(pos, digest))
	}
	if p.navigator.IsLeaf(pos) {
		return common.NewLeaf(pos, nil)
	}
	// we do a post-order traversal
	left := p.traverse(p.navigator.GoToLeft(pos))
	rightPos := p.navigator.GoToRight(pos)
	if rightPos == nil {
		return common.NewPartialNode(pos, left)
	}
	right := p.traverse(rightPos)
	if p.navigator.IsRoot(pos) {
		return common.NewRoot(pos, left, right)
	}
	return common.NewNode(pos, left, right)
}

type VerifyPruner struct {
	eventDigest hashing.Digest
	PruningContext
}

func NewVerifyPruner(eventDigest hashing.Digest, context PruningContext) *VerifyPruner {
	return &VerifyPruner{eventDigest, context}
}

func (p *VerifyPruner) Prune() common.Visitable {
	return p.traverse(p.navigator.Root(), p.eventDigest)
}

func (p *VerifyPruner) traverse(pos common.Position, eventDigest hashing.Digest) common.Visitable {
	if p.cacheResolver.ShouldGetFromCache(pos) {
		digest, ok := p.cache.Get(pos)
		if !ok {
			panic("this digest should be in cache") // TODO return error instead of panic
		}
		return common.NewCached(pos, digest)
	}
	if p.navigator.IsLeaf(pos) {
		return common.NewLeaf(pos, eventDigest)
	}
	// we do a post-order traversal
	left := p.traverse(p.navigator.GoToLeft(pos), eventDigest)
	rightPos := p.navigator.GoToRight(pos)
	if rightPos == nil {
		return common.NewPartialNode(pos, left)
	}
	right := p.traverse(rightPos, eventDigest)
	if p.navigator.IsRoot(pos) {
		return common.NewRoot(pos, left, right)
	}
	return common.NewNode(pos, left, right)

}

type VerifyIncrementalPruner struct {
	PruningContext
}

func NewVerifyIncrementalPruner(context PruningContext) *VerifyIncrementalPruner {
	return &VerifyIncrementalPruner{context}
}

func (p *VerifyIncrementalPruner) Prune() common.Visitable {
	return p.traverse(p.navigator.Root())
}

func (p *VerifyIncrementalPruner) traverse(pos common.Position) common.Visitable {
	if p.cacheResolver.ShouldGetFromCache(pos) {
		digest, ok := p.cache.Get(pos)
		if !ok {
			panic("this digest should be in cache")
		}
		return common.NewCached(pos, digest)
	}
	if p.navigator.IsLeaf(pos) {
		panic("this digest should be in cache")
	}
	// we do a post-order traversal
	left := p.traverse(p.navigator.GoToLeft(pos))
	rightPos := p.navigator.GoToRight(pos)
	if rightPos == nil {
		return common.NewPartialNode(pos, left)
	}
	right := p.traverse(rightPos)
	if p.navigator.IsRoot(pos) {
		return common.NewRoot(pos, left, right)
	}
	return common.NewNode(pos, left, right)
}
