package history

import (
	"github.com/bbva/qed/balloon2/common"
)

type CacheResolver interface {
	ShouldGetFromCache(pos common.Position) bool
}

type SingleTargetedCacheResolver struct {
	version uint64
}

func NewSingleTargetedCacheResolver(version uint64) *SingleTargetedCacheResolver {
	return &SingleTargetedCacheResolver{version}
}

func (r SingleTargetedCacheResolver) ShouldGetFromCache(pos common.Position) bool {
	return r.version > pos.IndexAsUint64()+pow(2, pos.Height())-1
}

type DoubleTargetedCacheResolver struct {
	start, end uint64
}

func NewDoubleTargetedCacheResolver(start, end uint64) *DoubleTargetedCacheResolver {
	return &DoubleTargetedCacheResolver{start, end}
}

func (r DoubleTargetedCacheResolver) ShouldGetFromCache(pos common.Position) bool {
	if pos.Height() == 0 && pos.IndexAsUint64() == r.start { // TODO THIS SHOULD BE TRUE for inc proofs but not for membership
		return false
	}
	lastDescendantIndex := pos.IndexAsUint64() + pow(2, pos.Height()) - 1
	if r.start > lastDescendantIndex && r.end > lastDescendantIndex {
		return true
	}
	return pos.IndexAsUint64() > r.start && lastDescendantIndex <= r.end
}

type IncrementalCacheResolver struct {
	start, end uint64
}

func NewIncrementalCacheResolver(start, end uint64) *IncrementalCacheResolver {
	return &IncrementalCacheResolver{start, end}
}

func (r IncrementalCacheResolver) ShouldGetFromCache(pos common.Position) bool {
	if pos.Height() == 0 && pos.IndexAsUint64() == r.start {
		return true
	}
	lastDescendantIndex := pos.IndexAsUint64() + pow(2, pos.Height()) - 1
	if r.start > lastDescendantIndex && r.end > lastDescendantIndex {
		return true
	}
	return pos.IndexAsUint64() > r.start && lastDescendantIndex <= r.end
}
