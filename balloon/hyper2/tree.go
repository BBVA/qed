package hyper2

import (
	"sync"

	"github.com/bbva/qed/balloon/cache"
	"github.com/bbva/qed/balloon/hyper2/pruning"
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
	batchLoader      pruning.BatchLoader

	sync.RWMutex
}

func NewHyperTree(hasherF func() hashing.Hasher, store storage.Store, cache cache.ModifiableCache) *HyperTree {

	hasher := hasherF()
	numBits := hasher.Len()
	cacheHeightLimit := numBits - min(24, numBits/8*4)

	tree := &HyperTree{
		store:            store,
		cache:            cache,
		hasherF:          hasherF,
		hasher:           hasher,
		cacheHeightLimit: cacheHeightLimit,
		defaultHashes:    make([]hashing.Digest, numBits),
		batchLoader:      pruning.NewDefaultBatchLoader(store, cache, cacheHeightLimit),
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

	// build a visitable pruned tree and then visit it to generate the root hash
	versionAsBytes := util.Uint64AsPaddedBytes(version, len(eventDigest))
	versionAsBytes = versionAsBytes[len(versionAsBytes)-len(eventDigest):]
	visitor := pruning.NewInsertVisitor(t.hasher, t.cache, t.defaultHashes)
	op, err := pruning.PruneToInsert(eventDigest, versionAsBytes, t.cacheHeightLimit, t.batchLoader)
	if err != nil {
		return nil, nil, err
	}

	rh := op.Accept(visitor)

	return rh, visitor.Result(), nil
}

func min(x, y uint16) uint16 {
	if x < y {
		return x
	}
	return y
}
