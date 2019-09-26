/*
   Copyright 2018-2019 Banco Bilbao Vizcaya Argentaria, S.A.

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
	"github.com/bbva/qed/log"

	"github.com/bbva/qed/balloon/cache"
	"github.com/bbva/qed/storage"
)

type batchLoader interface {
	Load(pos position) *batchNode
}

// TODO maybe use a function
type defaultBatchLoader struct {
	cacheHeightLimit uint16
	cache            cache.Cache
	store            storage.Store

	log log.Logger
}

func NewDefaultBatchLoader(store storage.Store, cache cache.Cache, cacheHeightLimit uint16) *defaultBatchLoader {
	return &defaultBatchLoader{
		cacheHeightLimit: cacheHeightLimit,
		cache:            cache,
		store:            store,
		log:              log.L(),
	}
}

func NewDefaultBatchLoaderWithLogger(store storage.Store, cache cache.Cache, cacheHeightLimit uint16, logger log.Logger) *defaultBatchLoader {
	return &defaultBatchLoader{
		cacheHeightLimit: cacheHeightLimit,
		cache:            cache,
		store:            store,
		log:              logger,
	}
}

func (l defaultBatchLoader) Load(pos position) *batchNode {
	if pos.Height > l.cacheHeightLimit {
		return l.loadBatchFromCache(pos)
	}
	return l.loadBatchFromStore(pos)
}

func (l defaultBatchLoader) loadBatchFromCache(pos position) *batchNode {
	value, ok := l.cache.Get(pos.Bytes())
	if !ok {
		return newEmptyBatchNode(len(pos.Index))
	}
	batch := parseBatchNode(len(pos.Index), value)
	return batch
}

func (l defaultBatchLoader) loadBatchFromStore(pos position) *batchNode {
	kv, err := l.store.Get(storage.HyperTable, pos.Bytes())
	if err != nil {
		if err == storage.ErrKeyNotFound {
			return newEmptyBatchNode(len(pos.Index))
		}
		l.log.Fatalf("Oops, something went wrong. Unable to load batch: %v", err)
	}
	batch := parseBatchNode(len(pos.Index), kv.Value)
	return batch
}
