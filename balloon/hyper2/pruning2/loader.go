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
