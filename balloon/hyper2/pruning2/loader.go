package pruning2

import (
	"github.com/bbva/qed/log"

	"github.com/bbva/qed/balloon/cache"
	"github.com/bbva/qed/balloon/hyper2/navigation"
	"github.com/bbva/qed/storage"
)

type BatchLoader interface {
	Load(pos navigation.Position) *BatchNode
}

// TODO maybe use a function
type DefaultBatchLoader struct {
	cacheHeightLimit uint16
	cache            cache.Cache
	store            storage.Store
}

func NewDefaultBatchLoader(store storage.Store, cache cache.Cache, cacheHeightLimit uint16) *DefaultBatchLoader {
	return &DefaultBatchLoader{
		cacheHeightLimit: cacheHeightLimit,
		cache:            cache,
		store:            store,
	}
}

func (l DefaultBatchLoader) Load(pos navigation.Position) *BatchNode {
	if pos.Height > l.cacheHeightLimit {
		return l.loadBatchFromCache(pos)
	}
	return l.loadBatchFromStore(pos)
}

func (l DefaultBatchLoader) loadBatchFromCache(pos navigation.Position) *BatchNode {
	value, ok := l.cache.Get(pos.Bytes())
	if !ok {
		return NewEmptyBatchNode(len(pos.Index))
	}
	batch := ParseBatchNode(len(pos.Index), value)
	return batch
}

func (l DefaultBatchLoader) loadBatchFromStore(pos navigation.Position) *BatchNode {
	kv, err := l.store.Get(storage.HyperCachePrefix, pos.Bytes())
	if err != nil {
		if err == storage.ErrKeyNotFound {
			return NewEmptyBatchNode(len(pos.Index))
		}
		log.Fatalf("Oops, something went wrong. Unable to load batch: %v", err)
	}
	batch := ParseBatchNode(len(pos.Index), kv.Value)
	return batch
}
