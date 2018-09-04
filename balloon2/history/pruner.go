package history

import "github.com/bbva/qed/balloon2/common"

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
	eventDigest common.Digest
	PruningContext
}

func NewInsertPruner(version uint64, eventDigest common.Digest, context PruningContext) *InsertPruner {
	return &InsertPruner{version, eventDigest, context}
}

func (p *InsertPruner) Prune() common.Visitable {
	return p.traverse(p.navigator.Root(), p.eventDigest)
}

func (p *InsertPruner) traverse(pos common.Position, eventDigest common.Digest) common.Visitable {
	if p.cacheResolver.ShouldGetFromCache(pos) {
		digest, ok := p.cache.Get(pos)
		if !ok {
			panic("this digest should be in cache")
		}
		return common.NewCached(pos, digest)
	}
	if p.navigator.IsLeaf(pos) {
		return common.NewCollectable(pos, common.NewLeaf(pos, eventDigest))
	}
	// we do a post-order traversal
	left := p.traverse(p.navigator.GoToLeft(pos), eventDigest)
	rightPos := p.navigator.GoToRight(pos)
	if rightPos == nil {
		return common.NewPartialNode(pos, left)
	}
	right := p.traverse(rightPos, eventDigest)
	if p.navigator.IsRoot(pos) {
		result := common.NewRoot(pos, left, right)
		if p.shouldCollect(pos) {
			return common.NewCollectable(pos, result)
		}
		return result
	}
	return common.NewNode(pos, left, right)
}

func (p InsertPruner) shouldCollect(pos common.Position) bool {
	return p.version >= pos.IndexAsUint64()+pow(2, pos.Height())-1
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
		return common.NewCollectable(pos, common.NewCached(pos, digest))
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
	eventDigest common.Digest
	PruningContext
}

func NewVerifyPruner(eventDigest common.Digest, context PruningContext) *VerifyPruner {
	return &VerifyPruner{eventDigest, context}
}

func (p *VerifyPruner) Prune() common.Visitable {
	return p.traverse(p.navigator.Root(), p.eventDigest)
}

func (p *VerifyPruner) traverse(pos common.Position, eventDigest common.Digest) common.Visitable {
	if p.cacheResolver.ShouldGetFromCache(pos) {
		digest, ok := p.cache.Get(pos)
		if !ok {
			panic("this digest should be in cache")
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
