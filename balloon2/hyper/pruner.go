package hyper

import (
	"github.com/bbva/qed/balloon2/common"
	"github.com/bbva/qed/db"
)

type PruningContext struct {
	navigator     common.TreeNavigator
	cacheResolver CacheResolver
	cache         common.Cache
	store         db.Store
	defaultHashes []common.Digest
}

type Pruner interface {
	Prune() common.Visitable
}

type InsertPruner struct {
	key   common.Digest
	value []byte
	PruningContext
}

func NewInsertPruner(key, value []byte, context PruningContext) *InsertPruner {
	return &InsertPruner{key, value, context}
}

func (p *InsertPruner) Prune() common.Visitable {
	leaves := db.KVRange{db.NewKVPair(p.key, p.value)}
	return p.traverse(p.navigator.Root(), leaves)
}

func (p *InsertPruner) traverse(pos common.Position, leaves db.KVRange) common.Visitable {
	if p.cacheResolver.ShouldBeInCache(pos) {
		digest, ok := p.cache.Get(pos)
		if !ok {
			return common.NewCached(pos, p.defaultHashes[pos.Height()])
		}
		return common.NewCached(pos, digest)
	}

	// if we are over the cache level, we need to do a range query to get the leaves
	if !p.cacheResolver.ShouldCache(pos) {
		first := p.navigator.DescendToFirst(pos)
		last := p.navigator.DescendToLast(pos)

		kvRange, _ := p.store.GetRange(db.IndexPrefix, first.Index(), last.Index())

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
		return common.NewRoot(pos, left, right)
	}
	return common.NewCollectable(common.NewNode(pos, left, right))
}

func (p *InsertPruner) traverseWithoutCache(pos common.Position, leaves db.KVRange) common.Visitable {
	if p.navigator.IsLeaf(pos) && len(leaves) == 1 {
		return common.NewLeaf(pos, leaves[0].Value)
	}
	if !p.navigator.IsRoot(pos) && len(leaves) == 0 {
		return common.NewCached(pos, p.defaultHashes[pos.Height()])
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
		return common.NewRoot(pos, left, right)
	}
	return common.NewNode(pos, left, right)
}

type SearchPruner struct {
	key []byte
	PruningContext
}

func NewSearchPruner(key []byte, context PruningContext) *SearchPruner {
	return &SearchPruner{key, context}
}

func (p *SearchPruner) Prune() common.Visitable {
	return p.traverseCache(p.navigator.Root(), db.NewKVRange())
}

func (p *SearchPruner) traverseCache(pos common.Position, leaves db.KVRange) common.Visitable {
	if p.cacheResolver.ShouldBeInCache(pos) {
		digest, ok := p.cache.Get(pos)
		if !ok {
			cached := common.NewCached(pos, p.defaultHashes[pos.Height()])
			return common.NewCollectable(cached)
		}
		return common.NewCollectable(common.NewCached(pos, digest))
	}

	// if we are over the cache level, we need to do a range query to get the leaves
	if !p.cacheResolver.ShouldCache(pos) {
		first := p.navigator.DescendToFirst(pos)
		last := p.navigator.DescendToLast(pos)
		kvRange, _ := p.store.GetRange(db.IndexPrefix, first.Index(), last.Index())

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
		return common.NewRoot(pos, left, right)
	}
	return common.NewNode(pos, left, right)
}

func (p *SearchPruner) traverse(pos common.Position, leaves db.KVRange) common.Visitable {
	if p.navigator.IsLeaf(pos) && len(leaves) == 1 {
		leaf := common.NewLeaf(pos, leaves[0].Value)
		if !p.cacheResolver.IsOnPath(pos) {
			return common.NewCollectable(leaf)
		}
		return leaf
	}
	if !p.navigator.IsRoot(pos) && len(leaves) == 0 {
		cached := common.NewCached(pos, p.defaultHashes[pos.Height()])
		return common.NewCollectable(cached)
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
			return common.NewRoot(pos, left, right)
		}
		return common.NewCollectable(common.NewNode(pos, left, right))
	}

	left := p.traverse(p.navigator.GoToLeft(pos), leftSlice)
	right := p.traverse(rightPos, rightSlice)
	if p.navigator.IsRoot(pos) {
		return common.NewRoot(pos, left, right)
	}
	return common.NewNode(pos, left, right)
}

func (p *SearchPruner) traverseWithoutCaching(pos common.Position, leaves db.KVRange) common.Visitable {
	if p.navigator.IsLeaf(pos) && len(leaves) == 1 {
		return common.NewLeaf(pos, leaves[0].Value)
	}
	if !p.navigator.IsRoot(pos) && len(leaves) == 0 {
		return common.NewCached(pos, p.defaultHashes[pos.Height()])
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
		return common.NewRoot(pos, left, right)
	}
	return common.NewNode(pos, left, right)
}

type VerifyPruner struct {
	key   common.Digest
	value []byte
	PruningContext
}

func NewVerifyPruner(key, value []byte, context PruningContext) *VerifyPruner {
	return &VerifyPruner{key, value, context}
}

func (p *VerifyPruner) Prune() common.Visitable {
	leaves := db.KVRange{db.NewKVPair(p.key, p.value)}
	return p.traverse(p.navigator.Root(), leaves)
}

func (p *VerifyPruner) traverse(pos common.Position, leaves db.KVRange) common.Visitable {
	if p.navigator.IsLeaf(pos) && len(leaves) == 1 {
		return common.NewLeaf(pos, leaves[0].Value)
	}
	if !p.navigator.IsRoot(pos) && len(leaves) == 0 {
		digest, ok := p.cache.Get(pos)
		if !ok {
			panic("this should never happen (wrong audit path)")
		}
		return common.NewCached(pos, digest)
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
		return common.NewRoot(pos, left, right)
	}
	return common.NewNode(pos, left, right)
}
