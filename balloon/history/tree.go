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
	"math"
	"sync"

	"github.com/bbva/qed/balloon/common"
	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/log"
	"github.com/bbva/qed/storage"
)

type HistoryTree struct {
	lock    sync.RWMutex
	hasherF func() hashing.Hasher
	cache   common.Cache
	hasher  hashing.Hasher
}

func NewHistoryTree(hasherF func() hashing.Hasher, cache common.Cache) *HistoryTree {
	var lock sync.RWMutex
	return &HistoryTree{lock, hasherF, cache, hasherF()}
}

func (t *HistoryTree) getDepth(version uint64) uint16 {
	return uint16(uint64(math.Ceil(math.Log2(float64(version + 1)))))
}

func (t *HistoryTree) Add(eventDigest hashing.Digest, version uint64) (hashing.Digest, []storage.Mutation, error) {
	t.lock.Lock() // TODO REMOVE THIS!!!
	defer t.lock.Unlock()

	log.Debugf("Adding event %b with version %d\n", eventDigest, version)

	// visitors
	computeHash := common.NewComputeHashVisitor(t.hasher)
	caching := common.NewCachingVisitor(computeHash)

	// build pruning context
	context := PruningContext{
		navigator:     NewHistoryTreeNavigator(version),
		cacheResolver: NewSingleTargetedCacheResolver(version),
		cache:         t.cache,
	}

	// traverse from root and generate a visitable pruned tree
	pruned := NewInsertPruner(version, eventDigest, context).Prune()

	// print := common.NewPrintVisitor(t.getDepth(version))
	// pruned.PreOrder(print)
	// log.Debugf("Pruned tree: %s", print.Result())

	// visit the pruned tree
	rh := pruned.PostOrder(caching).(hashing.Digest)

	// collect mutations
	cachedElements := caching.Result()
	mutations := make([]storage.Mutation, 0)
	for _, e := range cachedElements {
		mutation := storage.NewMutation(storage.HistoryCachePrefix, e.Pos.Bytes(), e.Digest)
		mutations = append(mutations, *mutation)
	}

	return rh, mutations, nil
}

func (t *HistoryTree) ProveMembership(index, version uint64) (*MembershipProof, error) {
	t.lock.Lock() // TODO REMOVE THIS!!!
	defer t.lock.Unlock()

	log.Debugf("Proving membership for index %d with version %d", index, version)

	// visitors
	computeHash := common.NewComputeHashVisitor(t.hasher)
	calcAuditPath := common.NewAuditPathVisitor(computeHash)

	// build pruning context
	var resolver CacheResolver
	switch index == version {
	case true:
		resolver = NewSingleTargetedCacheResolver(version)
	case false:
		resolver = NewDoubleTargetedCacheResolver(index, version)
	}
	context := PruningContext{
		navigator:     NewHistoryTreeNavigator(version),
		cacheResolver: resolver,
		cache:         t.cache,
	}

	// traverse from root and generate a visitable pruned tree
	pruned := NewSearchPruner(context).Prune()

	// print := common.NewPrintVisitor(t.getDepth(version))
	// pruned.PreOrder(print)
	// log.Debugf("Pruned tree: %s", print.Result())

	// visit the pruned tree
	pruned.PostOrder(calcAuditPath)

	proof := NewMembershipProof(index, version, calcAuditPath.Result(), t.hasherF())

	return proof, nil
}

func (t *HistoryTree) ProveConsistency(start, end uint64) (*IncrementalProof, error) {
	t.lock.Lock()
	defer t.lock.Unlock()

	log.Debugf("Proving consistency between versions %d and %d", start, end)

	// visitors
	computeHash := common.NewComputeHashVisitor(t.hasher)
	calcAuditPath := common.NewAuditPathVisitor(computeHash)

	// build pruning context
	context := PruningContext{
		navigator:     NewHistoryTreeNavigator(end),
		cacheResolver: NewIncrementalCacheResolver(start, end),
		cache:         t.cache,
	}

	// traverse from root and generate a visitable pruned tree
	pruned := NewSearchPruner(context).Prune()

	// visit the pruned tree
	pruned.PostOrder(calcAuditPath)
	proof := NewIncrementalProof(start, end, calcAuditPath.Result(), t.hasherF())

	return proof, nil
}

func (t *HistoryTree) Close() {
	t.cache = nil
	t.hasher = nil
}
