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
	"errors"

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
	Prune() (visitor.Visitable, error)
}

var (
	ErrCacheNotFound = errors.New("this digest should be in cache")
)

type InsertPruner struct {
	version     uint64
	eventDigest hashing.Digest
	PruningContext
}

func NewInsertPruner(version uint64, eventDigest hashing.Digest, context PruningContext) *InsertPruner {
	return &InsertPruner{version, eventDigest, context}
}

func (p *InsertPruner) Prune() (visitor.Visitable, error) {
	return p.traverse(p.navigator.Root(), p.eventDigest)
}

func (p *InsertPruner) traverse(pos navigator.Position, eventDigest hashing.Digest) (visitor.Visitable, error) {
	if p.cacheResolver.ShouldGetFromCache(pos) {
		digest, ok := p.cache.Get(pos)
		if !ok {
			return nil, ErrCacheNotFound
		}
		return visitor.NewCached(pos, digest), nil
	}
	if p.navigator.IsLeaf(pos) {
		return visitor.NewCollectable(visitor.NewCacheable(visitor.NewLeaf(pos, eventDigest))), nil
	}
	// we do a post-order traversal
	left, err := p.traverse(p.navigator.GoToLeft(pos), eventDigest)
	if err != nil {
		return nil, err
	}

	rightPos := p.navigator.GoToRight(pos)
	if rightPos == nil {
		return visitor.NewPartialNode(pos, left), nil
	}

	right, err := p.traverse(rightPos, eventDigest)
	if err != nil {
		return nil, err
	}

	var result visitor.Visitable
	if p.navigator.IsRoot(pos) {
		result = visitor.NewRoot(pos, left, right)
	} else {
		result = visitor.NewNode(pos, left, right)
	}

	if p.shouldCollect(pos) {
		return visitor.NewCollectable(visitor.NewCacheable(result)), nil
	}

	return result, nil
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

func (p *SearchPruner) Prune() (visitor.Visitable, error) {
	return p.traverse(p.navigator.Root())
}

func (p *SearchPruner) traverse(pos navigator.Position) (visitor.Visitable, error) {
	if p.cacheResolver.ShouldGetFromCache(pos) {
		digest, ok := p.cache.Get(pos)
		if !ok {
			return nil, ErrCacheNotFound
		}
		return visitor.NewCollectable(visitor.NewCached(pos, digest)), nil
	}

	if p.navigator.IsLeaf(pos) {
		return visitor.NewLeaf(pos, nil), nil
	}

	// we do a post-order traversal
	left, err := p.traverse(p.navigator.GoToLeft(pos))
	if err != nil {
		return nil, err
	}

	rightPos := p.navigator.GoToRight(pos)
	if rightPos == nil {
		return visitor.NewPartialNode(pos, left), nil
	}

	right, err := p.traverse(rightPos)
	if err != nil {
		return nil, err
	}

	if p.navigator.IsRoot(pos) {
		return visitor.NewRoot(pos, left, right), nil
	}

	return visitor.NewNode(pos, left, right), nil
}

type VerifyPruner struct {
	eventDigest hashing.Digest
	PruningContext
}

func NewVerifyPruner(eventDigest hashing.Digest, context PruningContext) *VerifyPruner {
	return &VerifyPruner{eventDigest, context}
}

func (p *VerifyPruner) Prune() (visitor.Visitable, error) {
	return p.traverse(p.navigator.Root(), p.eventDigest)
}

func (p *VerifyPruner) traverse(pos navigator.Position, eventDigest hashing.Digest) (visitor.Visitable, error) {
	if p.cacheResolver.ShouldGetFromCache(pos) {
		digest, ok := p.cache.Get(pos)
		if !ok {
			return nil, ErrCacheNotFound
		}
		return visitor.NewCached(pos, digest), nil
	}
	if p.navigator.IsLeaf(pos) {
		return visitor.NewLeaf(pos, eventDigest), nil
	}

	// we do a post-order traversal
	left, err := p.traverse(p.navigator.GoToLeft(pos), eventDigest)
	if err != nil {
		return nil, err
	}

	rightPos := p.navigator.GoToRight(pos)
	if rightPos == nil {
		return visitor.NewPartialNode(pos, left), nil
	}

	right, err := p.traverse(rightPos, eventDigest)
	if err != nil {
		return nil, err
	}

	if p.navigator.IsRoot(pos) {
		return visitor.NewRoot(pos, left, right), nil
	}

	return visitor.NewNode(pos, left, right), nil

}

type VerifyIncrementalPruner struct {
	PruningContext
}

func NewVerifyIncrementalPruner(context PruningContext) *VerifyIncrementalPruner {
	return &VerifyIncrementalPruner{context}
}

func (p *VerifyIncrementalPruner) Prune() (visitor.Visitable, error) {
	return p.traverse(p.navigator.Root())
}

func (p *VerifyIncrementalPruner) traverse(pos navigator.Position) (visitor.Visitable, error) {
	if p.cacheResolver.ShouldGetFromCache(pos) {
		digest, ok := p.cache.Get(pos)
		if !ok {
			return nil, ErrCacheNotFound
		}
		return visitor.NewCached(pos, digest), nil
	}

	if p.navigator.IsLeaf(pos) {
		return nil, ErrCacheNotFound
	}

	// we do a post-order traversal
	left, err := p.traverse(p.navigator.GoToLeft(pos))
	if err != nil {
		return nil, err
	}

	rightPos := p.navigator.GoToRight(pos)
	if rightPos == nil {
		return visitor.NewPartialNode(pos, left), nil
	}
	right, err := p.traverse(rightPos)
	if err != nil {
		return nil, err
	}

	if p.navigator.IsRoot(pos) {
		return visitor.NewRoot(pos, left, right), nil
	}

	return visitor.NewNode(pos, left, right), nil
}
