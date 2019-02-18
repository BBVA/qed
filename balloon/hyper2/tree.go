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

package hyper2

import (
	"sync"

	"github.com/bbva/qed/balloon/hyper2/navigation"

	"github.com/bbva/qed/balloon/cache"
	"github.com/bbva/qed/balloon/hyper2/pruning2"
	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/storage"
	"github.com/bbva/qed/util"
)

type HyperTree struct {
	store   storage.Store
	cache   cache.ModifiableCache
	hasherF func() hashing.Hasher

	hasher           hashing.Hasher
	cacheHeightLimit uint16
	defaultHashes    []hashing.Digest
	batchLoader      pruning2.BatchLoader

	sync.RWMutex
}

func NewHyperTree(hasherF func() hashing.Hasher, store storage.Store, cache cache.ModifiableCache) *HyperTree {

	hasher := hasherF()
	numBits := hasher.Len()
	cacheHeightLimit := numBits - min(24, (numBits/8)*4)

	tree := &HyperTree{
		store:            store,
		cache:            cache,
		hasherF:          hasherF,
		hasher:           hasher,
		cacheHeightLimit: cacheHeightLimit,
		defaultHashes:    make([]hashing.Digest, numBits),
		batchLoader:      pruning2.NewDefaultBatchLoader(store, cache, cacheHeightLimit),
	}

	tree.defaultHashes[0] = tree.hasher.Do([]byte{0x0}, []byte{0x0})
	for i := uint16(1); i < hasher.Len(); i++ {
		tree.defaultHashes[i] = tree.hasher.Do(tree.defaultHashes[i-1], tree.defaultHashes[i-1])
	}

	// warm-up cache
	//tree.RebuildCache()

	return tree
}

func (t *HyperTree) Add(eventDigest hashing.Digest, version uint64) (hashing.Digest, []*storage.Mutation, error) {
	t.Lock()
	defer t.Unlock()

	//log.Debugf("Adding new event digest %x with version %d", eventDigest, version)

	versionAsBytes := util.Uint64AsPaddedBytes(version, len(eventDigest))
	versionAsBytes = versionAsBytes[len(versionAsBytes)-len(eventDigest):]

	// build a stack of operations and then interpret it to generate the root hash
	ops := pruning2.PruneToInsert(eventDigest, versionAsBytes, t.cacheHeightLimit, t.batchLoader)
	ctx := &pruning2.Context{
		Hasher:        t.hasher,
		Cache:         t.cache,
		DefaultHashes: t.defaultHashes,
		Mutations:     make([]*storage.Mutation, 0),
	}

	rh := ops.Pop().Interpret(ops, ctx)

	// create a mutation for the new leaf
	leafMutation := storage.NewMutation(storage.IndexPrefix, eventDigest, versionAsBytes)

	// collect mutations
	mutations := append(ctx.Mutations, leafMutation)

	return rh, mutations, nil
}

func (t *HyperTree) QueryMembership(eventDigest hashing.Digest, version []byte) (proof *QueryProof, err error) {
	t.Lock()
	defer t.Unlock()

	//log.Debugf("Proving membership for index %d with version %d", eventDigest, version)

	// build a stack of operations and then interpret it to generate the audit path
	ops := pruning2.PruneToFind(eventDigest, t.batchLoader)
	ctx := &pruning2.Context{
		Hasher:        t.hasher,
		Cache:         t.cache,
		DefaultHashes: t.defaultHashes,
		AuditPath:     make(navigation.AuditPath, 0),
	}

	ops.Pop().Interpret(ops, ctx)

	return NewQueryProof(eventDigest, version, ctx.AuditPath, t.hasherF()), nil
}

func min(x, y uint16) uint16 {
	if x < y {
		return x
	}
	return y
}
