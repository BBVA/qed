package history

import (
	"math"
	"sync"

	"github.com/bbva/qed/balloon2/common"
	"github.com/bbva/qed/db"
	"github.com/bbva/qed/log"
)

type HistoryTree struct {
	lock   sync.RWMutex
	hasher common.Hasher
	cache  common.Cache
}

func NewHistoryTree(hasher common.Hasher, cache common.Cache) *HistoryTree {
	var lock sync.RWMutex
	return &HistoryTree{lock, hasher, cache}
}

func (t *HistoryTree) getDepth(version uint64) uint16 {
	return uint16(uint64(math.Ceil(math.Log2(float64(version + 1)))))
}

func (t *HistoryTree) Add(eventDigest common.Digest, version uint64) (common.Digest, []db.Mutation, error) {
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
	rh := pruned.PostOrder(caching).(common.Digest)

	// collect mutations
	cachedElements := caching.Result()
	mutations := make([]db.Mutation, 0)
	for _, e := range cachedElements {
		mutation := db.NewMutation(db.HistoryCachePrefix, e.Pos.Bytes(), e.Digest)
		mutations = append(mutations, *mutation)
	}

	return rh, mutations, nil
}

func (t *HistoryTree) ProveMembership(index, version uint64) (common.AuditPath, error) {
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

	return calcAuditPath.Result(), nil
}
