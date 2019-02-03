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

package pruning

import (
	"errors"

	"github.com/bbva/qed/balloon/history/visit"

	"github.com/bbva/qed/balloon/cache"
)

type PruningContext struct {
	cacheResolver CacheResolver
	cache         cache.Cache
}

func NewPruningContext(cacheResolver CacheResolver, cache cache.Cache) *PruningContext {
	return &PruningContext{
		cacheResolver: cacheResolver,
		cache:         cache,
	}
}

type Pruner interface {
	Prune() (visit.Visitable, error)
}

var (
	ErrCacheNotFound = errors.New("this digest should be in cache")
)
