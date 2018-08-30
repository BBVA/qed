package history

import (
	"sync"

	"github.com/bbva/qed/balloon2/common"
)

type HistoryTree struct {
	lock   sync.RWMutex
	cache  common.Cache
	hasher common.Hasher
}

func NewHistoryTree(hasher common.Hasher, cache common.Cache) *HistoryTree {
	var lock sync.RWMutex
	return &HistoryTree{lock, cache, hasher}
}
