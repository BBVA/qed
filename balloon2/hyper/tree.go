package hyper

import (
	"math"
	"sync"

	"github.com/bbva/qed/balloon2/common"
	"github.com/bbva/qed/db"
)

type HyperTree struct {
	lock          sync.RWMutex
	store         db.Store
	cache         common.ModifiableCache
	hasher        common.Hasher
	cacheLevel    uint16
	defaultHashes []common.Digest
}

func NewHyperTree(hasher common.Hasher, store db.Store, cache common.ModifiableCache) *HyperTree {
	var lock sync.RWMutex
	cacheLevel := uint16(math.Max(float64(2), math.Floor(float64(hasher.Len())/10)))
	tree := &HyperTree{
		lock:          lock,
		store:         store,
		cache:         cache,
		hasher:        hasher,
		cacheLevel:    cacheLevel,
		defaultHashes: make([]common.Digest, hasher.Len()),
	}

	tree.defaultHashes[0] = tree.hasher.Do([]byte{0x0}, []byte{0x0})
	for i := uint16(1); i < hasher.Len(); i++ {
		tree.defaultHashes[i] = tree.hasher.Do(tree.defaultHashes[i-1], tree.defaultHashes[i-1])
	}
	return tree
}
