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

package hyper

import (
	"github.com/bbva/qed/balloon/cache"
	"github.com/bbva/qed/balloon/navigator"
	"github.com/bbva/qed/balloon/visitor"
	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/storage"
)

type PruningContext struct {
	navigator     navigator.TreeNavigator
	cacheResolver CacheResolver
	cache         cache.Cache
	store         storage.Store
	defaultHashes []hashing.Digest
}

type Pruner interface {
	Prune() visitor.Visitable
}

type InsertPruner struct {
	key   hashing.Digest
	value []byte
	PruningContext
}

func NewInsertPruner(key, value []byte, context PruningContext) *InsertPruner {
	return &InsertPruner{key, value, context}
}

func (p *InsertPruner) Prune() visitor.Visitable {
	leaves := storage.KVRange{storage.NewKVPair(p.key, p.value)}
	return p.traverse(p.navigator.Root(), leaves)
}

func (p *InsertPruner) traverse(pos navigator.Position, leaves storage.KVRange) visitor.Visitable {
	if p.cacheResolver.ShouldBeInCache(pos) {
		digest, ok := p.cache.Get(pos)
		if !ok {
			return visitor.NewCached(pos, p.defaultHashes[pos.Height()])
		}
		return visitor.NewCached(pos, digest)
	}

	// if we are over the cache level, we need to do a range query to get the leaves
	if !p.cacheResolver.ShouldCache(pos) {
		first := p.navigator.DescendToFirst(pos)
		last := p.navigator.DescendToLast(pos)

		kvRange, _ := p.store.GetRange(storage.IndexPrefix, first.Index(), last.Index())

		// replace leaves with new slice and append the previous to the new one
		for _, l := range leaves {
			kvRange = kvRange.InsertSorted(l)
		}
		leaves = kvRange

		return p.traverseWithoutCache(pos, leaves)
	}

	rightPos := p.navigator.GoToRight(pos)
	leftSlice, rightSlice := leaves.Split(rightPos.Index())
	left := p.traverse(p.navigator.GoToLeft(pos), leftSlice)
	right := p.traverse(rightPos, rightSlice)
	if p.navigator.IsRoot(pos) {
		return visitor.NewRoot(pos, left, right)
	}
	result := visitor.NewCacheable(visitor.NewNode(pos, left, right))
	if p.cacheResolver.ShouldCollect(pos) {
		return visitor.NewCollectable(result)
	}
	return result
}

func (p *InsertPruner) traverseWithoutCache(pos navigator.Position, leaves storage.KVRange) visitor.Visitable {
	if p.navigator.IsLeaf(pos) && len(leaves) == 1 {
		return visitor.NewLeaf(pos, leaves[0].Value)
	}
	if !p.navigator.IsRoot(pos) && len(leaves) == 0 {
		return visitor.NewCached(pos, p.defaultHashes[pos.Height()])
	}
	if len(leaves) > 1 && p.navigator.IsLeaf(pos) {
		panic("this should never happen (unsorted LeavesSlice or broken split?)")
	}

	// we do a post-order traversal

	// split leaves
	rightPos := p.navigator.GoToRight(pos)
	leftSlice, rightSlice := leaves.Split(rightPos.Index())
	left := p.traverseWithoutCache(p.navigator.GoToLeft(pos), leftSlice)
	right := p.traverseWithoutCache(rightPos, rightSlice)
	if p.navigator.IsRoot(pos) {
		return visitor.NewRoot(pos, left, right)
	}
	return visitor.NewNode(pos, left, right)
}

type SearchPruner struct {
	key []byte
	PruningContext
}

func NewSearchPruner(key []byte, context PruningContext) *SearchPruner {
	return &SearchPruner{key, context}
}

func (p *SearchPruner) Prune() visitor.Visitable {
	return p.traverseCache(p.navigator.Root(), storage.NewKVRange())
}

func (p *SearchPruner) traverseCache(pos navigator.Position, leaves storage.KVRange) visitor.Visitable {
	if p.cacheResolver.ShouldBeInCache(pos) {
		digest, ok := p.cache.Get(pos)
		if !ok {
			cached := visitor.NewCached(pos, p.defaultHashes[pos.Height()])
			return visitor.NewCollectable(cached)
		}
		return visitor.NewCollectable(visitor.NewCached(pos, digest))
	}

	// if we are over the cache level, we need to do a range query to get the leaves
	if !p.cacheResolver.ShouldCache(pos) {
		first := p.navigator.DescendToFirst(pos)
		last := p.navigator.DescendToLast(pos)
		kvRange, _ := p.store.GetRange(storage.IndexPrefix, first.Index(), last.Index())

		// replace leaves with new slice and append the previous to the new one
		for _, l := range leaves {
			kvRange = kvRange.InsertSorted(l)
		}
		leaves = kvRange
		return p.traverse(pos, leaves)
	}

	rightPos := p.navigator.GoToRight(pos)
	leftSlice, rightSlice := leaves.Split(rightPos.Index())
	left := p.traverseCache(p.navigator.GoToLeft(pos), leftSlice)
	right := p.traverseCache(rightPos, rightSlice)
	if p.navigator.IsRoot(pos) {
		return visitor.NewRoot(pos, left, right)
	}
	return visitor.NewNode(pos, left, right)
}

func (p *SearchPruner) traverse(pos navigator.Position, leaves storage.KVRange) visitor.Visitable {
	if p.navigator.IsLeaf(pos) && len(leaves) == 1 {
		leaf := visitor.NewLeaf(pos, leaves[0].Value)
		if !p.cacheResolver.IsOnPath(pos) {
			return visitor.NewCollectable(leaf)
		}
		return leaf
	}
	if !p.navigator.IsRoot(pos) && len(leaves) == 0 {
		cached := visitor.NewCached(pos, p.defaultHashes[pos.Height()])
		return visitor.NewCollectable(cached)
	}
	if len(leaves) > 1 && p.navigator.IsLeaf(pos) {
		panic("this should never happen (unsorted LeavesSlice or broken split?)")
	}

	// we do a post-order traversal

	// split leaves
	rightPos := p.navigator.GoToRight(pos)
	leftSlice, rightSlice := leaves.Split(rightPos.Index())

	if !p.cacheResolver.IsOnPath(pos) {
		left := p.traverseWithoutCaching(p.navigator.GoToLeft(pos), leftSlice)
		right := p.traverseWithoutCaching(rightPos, rightSlice)
		if p.navigator.IsRoot(pos) {
			return visitor.NewRoot(pos, left, right)
		}
		return visitor.NewCollectable(visitor.NewNode(pos, left, right))
	}

	left := p.traverse(p.navigator.GoToLeft(pos), leftSlice)
	right := p.traverse(rightPos, rightSlice)
	if p.navigator.IsRoot(pos) {
		return visitor.NewRoot(pos, left, right)
	}
	return visitor.NewNode(pos, left, right)
}

func (p *SearchPruner) traverseWithoutCaching(pos navigator.Position, leaves storage.KVRange) visitor.Visitable {
	if p.navigator.IsLeaf(pos) && len(leaves) == 1 {
		return visitor.NewLeaf(pos, leaves[0].Value)
	}
	if !p.navigator.IsRoot(pos) && len(leaves) == 0 {
		return visitor.NewCached(pos, p.defaultHashes[pos.Height()])
	}
	if len(leaves) > 1 && p.navigator.IsLeaf(pos) {
		panic("this should never happen (unsorted LeavesSlice or broken split?)")
	}

	// we do a post-order traversal

	// split leaves
	rightPos := p.navigator.GoToRight(pos)
	leftSlice, rightSlice := leaves.Split(rightPos.Index())
	left := p.traverseWithoutCaching(p.navigator.GoToLeft(pos), leftSlice)
	right := p.traverseWithoutCaching(rightPos, rightSlice)
	if p.navigator.IsRoot(pos) {
		return visitor.NewRoot(pos, left, right)
	}
	return visitor.NewNode(pos, left, right)
}

type VerifyPruner struct {
	key   hashing.Digest
	value []byte
	PruningContext
}

func NewVerifyPruner(key, value []byte, context PruningContext) *VerifyPruner {
	return &VerifyPruner{key, value, context}
}

func (p *VerifyPruner) Prune() visitor.Visitable {
	leaves := storage.KVRange{storage.NewKVPair(p.key, p.value)}
	return p.traverse(p.navigator.Root(), leaves)
}

func (p *VerifyPruner) traverse(pos navigator.Position, leaves storage.KVRange) visitor.Visitable {
	if p.navigator.IsLeaf(pos) && len(leaves) == 1 {
		return visitor.NewLeaf(pos, leaves[0].Value)
	}
	if !p.navigator.IsRoot(pos) && len(leaves) == 0 {
		digest, ok := p.cache.Get(pos)
		if !ok {
			panic("this should never happen (wrong audit path)")
		}
		return visitor.NewCached(pos, digest)
	}
	if len(leaves) > 1 && p.navigator.IsLeaf(pos) {
		panic("this should never happen (unsorted LeavesSlice or broken split?)")
	}

	// we do a post-order traversal

	// split leaves
	rightPos := p.navigator.GoToRight(pos)
	leftSlice, rightSlice := leaves.Split(rightPos.Index())
	left := p.traverse(p.navigator.GoToLeft(pos), leftSlice)
	right := p.traverse(rightPos, rightSlice)
	if p.navigator.IsRoot(pos) {
		return visitor.NewRoot(pos, left, right)
	}
	return visitor.NewNode(pos, left, right)
}
