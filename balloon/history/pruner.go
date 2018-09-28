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
	"fmt"

	"github.com/bbva/qed/balloon/cache"
	"github.com/bbva/qed/balloon/navigator"
	"github.com/bbva/qed/balloon/visitor"
	"github.com/bbva/qed/hashing"
)

type PruningContext struct {
	navigator     navigator.TreeNavigator
	cacheResolver CacheResolver
	cache         cache.Cache
}

type Pruner interface {
	Prune() visitor.Visitable
}

type InsertPruner struct {
	version     uint64
	eventDigest hashing.Digest
	PruningContext
}

func NewInsertPruner(version uint64, eventDigest hashing.Digest, context PruningContext) *InsertPruner {
	return &InsertPruner{version, eventDigest, context}
}

func (p *InsertPruner) Prune() visitor.Visitable {
	return p.traverse(p.navigator.Root(), p.eventDigest)
}

func (p *InsertPruner) traverse(pos navigator.Position, eventDigest hashing.Digest) visitor.Visitable {
	if p.cacheResolver.ShouldGetFromCache(pos) {
		digest, ok := p.cache.Get(pos)
		if !ok {
			panic("this digest should be in cache")
		}
		return visitor.NewCached(pos, digest)
	}
	if p.navigator.IsLeaf(pos) {
		return visitor.NewCollectable(visitor.NewCacheable(visitor.NewLeaf(pos, eventDigest)))
	}
	// we do a post-order traversal
	left := p.traverse(p.navigator.GoToLeft(pos), eventDigest)
	rightPos := p.navigator.GoToRight(pos)
	if rightPos == nil {
		return visitor.NewPartialNode(pos, left)
	}
	right := p.traverse(rightPos, eventDigest)
	var result visitor.Visitable
	if p.navigator.IsRoot(pos) {
		result = visitor.NewRoot(pos, left, right)
	} else {
		result = visitor.NewNode(pos, left, right)
	}
	if p.shouldCollect(pos) {
		return visitor.NewCollectable(visitor.NewCacheable(result))
	}
	return result
}

func (p InsertPruner) shouldCollect(pos navigator.Position) bool {
	return p.version >= pos.IndexAsUint64()+1<<pos.Height()-1
}

type SearchPruner struct {
	PruningContext
}

func NewSearchPruner(context PruningContext) *SearchPruner {
	return &SearchPruner{context}
}

func (p *SearchPruner) Prune() visitor.Visitable {
	return p.traverse(p.navigator.Root())
}

func (p *SearchPruner) traverse(pos navigator.Position) visitor.Visitable {
	if p.cacheResolver.ShouldGetFromCache(pos) {
		digest, ok := p.cache.Get(pos)
		if !ok {
			panic("this digest should be in cache")
		}
		return visitor.NewCollectable(visitor.NewCached(pos, digest))
	}
	if p.navigator.IsLeaf(pos) {
		return visitor.NewLeaf(pos, nil)
	}
	// we do a post-order traversal
	left := p.traverse(p.navigator.GoToLeft(pos))
	rightPos := p.navigator.GoToRight(pos)
	if rightPos == nil {
		return visitor.NewPartialNode(pos, left)
	}
	right := p.traverse(rightPos)
	if p.navigator.IsRoot(pos) {
		return visitor.NewRoot(pos, left, right)
	}
	return visitor.NewNode(pos, left, right)
}

type VerifyPruner struct {
	eventDigest hashing.Digest
	PruningContext
}

func NewVerifyPruner(eventDigest hashing.Digest, context PruningContext) *VerifyPruner {
	return &VerifyPruner{eventDigest, context}
}

func (p *VerifyPruner) Prune() visitor.Visitable {
	return p.traverse(p.navigator.Root(), p.eventDigest)
}

func (p *VerifyPruner) traverse(pos navigator.Position, eventDigest hashing.Digest) visitor.Visitable {
	if p.cacheResolver.ShouldGetFromCache(pos) {
		digest, ok := p.cache.Get(pos)
		if !ok {
			panic(fmt.Sprintf("the digest in position %v must be in cache", pos)) // TODO return error instead of panic
		}
		return visitor.NewCached(pos, digest)
	}
	if p.navigator.IsLeaf(pos) {
		return visitor.NewLeaf(pos, eventDigest)
	}
	// we do a post-order traversal
	left := p.traverse(p.navigator.GoToLeft(pos), eventDigest)
	rightPos := p.navigator.GoToRight(pos)
	if rightPos == nil {
		return visitor.NewPartialNode(pos, left)
	}
	right := p.traverse(rightPos, eventDigest)
	if p.navigator.IsRoot(pos) {
		return visitor.NewRoot(pos, left, right)
	}
	return visitor.NewNode(pos, left, right)

}

type VerifyIncrementalPruner struct {
	PruningContext
}

func NewVerifyIncrementalPruner(context PruningContext) *VerifyIncrementalPruner {
	return &VerifyIncrementalPruner{context}
}

func (p *VerifyIncrementalPruner) Prune() visitor.Visitable {
	return p.traverse(p.navigator.Root())
}

func (p *VerifyIncrementalPruner) traverse(pos navigator.Position) visitor.Visitable {
	if p.cacheResolver.ShouldGetFromCache(pos) {
		digest, ok := p.cache.Get(pos)
		if !ok {
			panic("this digest should be in cache")
		}
		return visitor.NewCached(pos, digest)
	}
	if p.navigator.IsLeaf(pos) {
		panic("this digest should be in cache")
	}
	// we do a post-order traversal
	left := p.traverse(p.navigator.GoToLeft(pos))
	rightPos := p.navigator.GoToRight(pos)
	if rightPos == nil {
		return visitor.NewPartialNode(pos, left)
	}
	right := p.traverse(rightPos)
	if p.navigator.IsRoot(pos) {
		return visitor.NewRoot(pos, left, right)
	}
	return visitor.NewNode(pos, left, right)
}
