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
	"errors"

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
	Prune() (visitor.Visitable, error)
}

var (
	ErrLeavesSlice    = errors.New("this should never happen (unsorted LeavesSlice or broken split?)")
	ErrWrongAuditPath = errors.New("this should never happen (wrong audit path)")
)

type InsertPruner struct {
	key   hashing.Digest
	value []byte
	PruningContext
}

func NewInsertPruner(key, value []byte, context PruningContext) *InsertPruner {
	return &InsertPruner{key, value, context}
}

func (p *InsertPruner) Prune() (visitor.Visitable, error) {
	leaves := storage.KVRange{storage.NewKVPair(p.key, p.value)}
	return p.traverse(p.navigator.Root(), leaves)
}

func (p *InsertPruner) traverse(pos navigator.Position, leaves storage.KVRange) (visitor.Visitable, error) {
	if p.cacheResolver.ShouldBeInCache(pos) {
		digest, ok := p.cache.Get(pos)
		if !ok {
			return visitor.NewCached(pos, p.defaultHashes[pos.Height()]), nil
		}
		return visitor.NewCached(pos, digest), nil
	}

	// if we are over the cache level, we need to do a range query to get the leaves
	var atLastLevel bool
	if atLastLevel = p.cacheResolver.ShouldCache(pos); atLastLevel {
		//fmt.Println(pos.Height())
		first := p.navigator.DescendToFirst(pos)
		last := p.navigator.DescendToLast(pos)

		kvRange, _ := p.store.GetRange(storage.IndexPrefix, first.Index(), last.Index())

		// replace leaves with new slice and append the previous to the new one
		for _, l := range leaves {
			kvRange = kvRange.InsertSorted(l)
		}
		leaves = kvRange
	}

	rightPos := p.navigator.GoToRight(pos)
	leftPos := p.navigator.GoToLeft(pos)
	leftSlice, rightSlice := leaves.Split(rightPos.Index())

	var left, right visitor.Visitable
	var err error
	if atLastLevel {
		left, err = p.traverseWithoutCache(leftPos, leftSlice)
		if err != nil {
			return nil, err
		}
		right, err = p.traverseWithoutCache(rightPos, rightSlice)

	} else {
		left, err = p.traverse(leftPos, leftSlice)
		if err != nil {
			return nil, err
		}
		right, err = p.traverse(rightPos, rightSlice)
	}
	if err != nil {
		return nil, err
	}

	if p.navigator.IsRoot(pos) {
		return visitor.NewRoot(pos, left, right), nil
	}

	result := visitor.NewCacheable(visitor.NewNode(pos, left, right))
	if p.cacheResolver.ShouldCollect(pos) {
		return visitor.NewCollectable(result), nil
	}

	return result, nil
}

func (p *InsertPruner) traverseWithoutCache(pos navigator.Position, leaves storage.KVRange) (visitor.Visitable, error) {
	if p.navigator.IsLeaf(pos) && len(leaves) == 1 {
		return visitor.NewLeaf(pos, leaves[0].Value), nil
	}
	if !p.navigator.IsRoot(pos) && len(leaves) == 0 {
		return visitor.NewCached(pos, p.defaultHashes[pos.Height()]), nil
	}
	if len(leaves) > 1 && p.navigator.IsLeaf(pos) {
		return nil, ErrLeavesSlice
	}

	// we do a post-order traversal

	// split leaves
	rightPos := p.navigator.GoToRight(pos)
	leftSlice, rightSlice := leaves.Split(rightPos.Index())
	left, err := p.traverseWithoutCache(p.navigator.GoToLeft(pos), leftSlice)
	if err != nil {
		return nil, ErrLeavesSlice
	}

	right, err := p.traverseWithoutCache(rightPos, rightSlice)
	if err != nil {
		return nil, ErrLeavesSlice
	}

	if p.navigator.IsRoot(pos) {
		return visitor.NewRoot(pos, left, right), nil
	}
	return visitor.NewNode(pos, left, right), nil
}

type SearchPruner struct {
	key []byte
	PruningContext
}

func NewSearchPruner(key []byte, context PruningContext) *SearchPruner {
	return &SearchPruner{key, context}
}

func (p *SearchPruner) Prune() (visitor.Visitable, error) {
	return p.traverseCache(p.navigator.Root(), storage.NewKVRange())
}

func (p *SearchPruner) traverseCache(pos navigator.Position, leaves storage.KVRange) (visitor.Visitable, error) {
	if p.cacheResolver.ShouldBeInCache(pos) {
		digest, ok := p.cache.Get(pos)
		if !ok {
			cached := visitor.NewCached(pos, p.defaultHashes[pos.Height()])
			return visitor.NewCollectable(cached), nil
		}
		return visitor.NewCollectable(visitor.NewCached(pos, digest)), nil
	}

	// if we are over the cache level, we need to do a range query to get the leaves
	var atLastLevel bool
	if atLastLevel = p.cacheResolver.ShouldCache(pos); atLastLevel {
		first := p.navigator.DescendToFirst(pos)
		last := p.navigator.DescendToLast(pos)
		kvRange, _ := p.store.GetRange(storage.IndexPrefix, first.Index(), last.Index())

		// replace leaves with new slice and append the previous to the new one
		for _, l := range leaves {
			kvRange = kvRange.InsertSorted(l)
		}
		leaves = kvRange
	}

	rightPos := p.navigator.GoToRight(pos)
	leftPos := p.navigator.GoToLeft(pos)
	leftSlice, rightSlice := leaves.Split(rightPos.Index())

	var left, right visitor.Visitable
	var err error
	if atLastLevel {
		left, err = p.traverse(leftPos, leftSlice)
		if err != nil {
			return nil, err
		}
		right, err = p.traverse(rightPos, rightSlice)
	} else {
		left, err = p.traverseCache(leftPos, leftSlice)
		if err != nil {
			return nil, err
		}
		right, err = p.traverseCache(rightPos, rightSlice)
	}
	if err != nil {
		return nil, err
	}

	if p.navigator.IsRoot(pos) {
		return visitor.NewRoot(pos, left, right), nil
	}

	return visitor.NewNode(pos, left, right), nil
}

func (p *SearchPruner) traverse(pos navigator.Position, leaves storage.KVRange) (visitor.Visitable, error) {
	if p.navigator.IsLeaf(pos) && len(leaves) == 1 {
		leaf := visitor.NewLeaf(pos, leaves[0].Value)
		if !p.cacheResolver.IsOnPath(pos) {
			return visitor.NewCollectable(leaf), nil
		}
		return leaf, nil
	}
	if !p.navigator.IsRoot(pos) && len(leaves) == 0 {
		cached := visitor.NewCached(pos, p.defaultHashes[pos.Height()])
		return visitor.NewCollectable(cached), nil
	}
	if len(leaves) > 1 && p.navigator.IsLeaf(pos) {
		return nil, ErrLeavesSlice
	}

	// we do a post-order traversal

	// split leaves
	rightPos := p.navigator.GoToRight(pos)
	leftSlice, rightSlice := leaves.Split(rightPos.Index())

	if !p.cacheResolver.IsOnPath(pos) {
		left, err := p.traverseWithoutCollecting(p.navigator.GoToLeft(pos), leftSlice)
		if err != nil {
			return nil, err
		}

		right, err := p.traverseWithoutCollecting(rightPos, rightSlice)
		if err != nil {
			return nil, err
		}

		if p.navigator.IsRoot(pos) {
			return visitor.NewRoot(pos, left, right), nil
		}
		return visitor.NewCollectable(visitor.NewNode(pos, left, right)), nil
	}

	left, err := p.traverse(p.navigator.GoToLeft(pos), leftSlice)
	if err != nil {
		return nil, err
	}

	right, err := p.traverse(rightPos, rightSlice)
	if err != nil {
		return nil, err
	}

	if p.navigator.IsRoot(pos) {
		return visitor.NewRoot(pos, left, right), nil
	}
	return visitor.NewNode(pos, left, right), nil
}

func (p *SearchPruner) traverseWithoutCollecting(pos navigator.Position, leaves storage.KVRange) (visitor.Visitable, error) {
	if p.navigator.IsLeaf(pos) && len(leaves) == 1 {
		return visitor.NewLeaf(pos, leaves[0].Value), nil
	}
	if !p.navigator.IsRoot(pos) && len(leaves) == 0 {
		return visitor.NewCached(pos, p.defaultHashes[pos.Height()]), nil
	}
	if len(leaves) > 1 && p.navigator.IsLeaf(pos) {
		return nil, ErrLeavesSlice
	}

	// we do a post-order traversal

	// split leaves
	rightPos := p.navigator.GoToRight(pos)
	leftSlice, rightSlice := leaves.Split(rightPos.Index())
	left, err := p.traverseWithoutCollecting(p.navigator.GoToLeft(pos), leftSlice)
	if err != nil {
		return nil, ErrLeavesSlice
	}
	right, err := p.traverseWithoutCollecting(rightPos, rightSlice)
	if err != nil {
		return nil, ErrLeavesSlice
	}

	if p.navigator.IsRoot(pos) {
		return visitor.NewRoot(pos, left, right), nil
	}
	return visitor.NewNode(pos, left, right), nil
}

type VerifyPruner struct {
	key   hashing.Digest
	value []byte
	PruningContext
}

func NewVerifyPruner(key, value []byte, context PruningContext) *VerifyPruner {
	return &VerifyPruner{key, value, context}
}

func (p *VerifyPruner) Prune() (visitor.Visitable, error) {
	leaves := storage.KVRange{storage.NewKVPair(p.key, p.value)}
	return p.traverse(p.navigator.Root(), leaves)
}

func (p *VerifyPruner) traverse(pos navigator.Position, leaves storage.KVRange) (visitor.Visitable, error) {
	if p.navigator.IsLeaf(pos) && len(leaves) == 1 {
		return visitor.NewLeaf(pos, leaves[0].Value), nil
	}
	if !p.navigator.IsRoot(pos) && len(leaves) == 0 {
		digest, ok := p.cache.Get(pos)
		if !ok {
			return nil, ErrWrongAuditPath
		}
		return visitor.NewCached(pos, digest), nil
	}
	if len(leaves) > 1 && p.navigator.IsLeaf(pos) {
		return nil, ErrLeavesSlice
	}

	// we do a post-order traversal

	// split leaves
	rightPos := p.navigator.GoToRight(pos)
	leftSlice, rightSlice := leaves.Split(rightPos.Index())
	left, err := p.traverse(p.navigator.GoToLeft(pos), leftSlice)
	if err != nil {
		return nil, err
	}
	right, err := p.traverse(rightPos, rightSlice)
	if err != nil {
		return nil, err
	}
	if p.navigator.IsRoot(pos) {
		return visitor.NewRoot(pos, left, right), nil
	}
	return visitor.NewNode(pos, left, right), nil
}
