package pruning5

import (
	"testing"

	"github.com/bbva/qed/balloon/cache"
	"github.com/bbva/qed/hashing"
	"github.com/stretchr/testify/assert"
)

func BenchmarkComputeHashVisitor(b *testing.B) {

	cache := cache.NewFakeCache([]byte{0x0})
	visitor := NewComputeHashVisitor(hashing.NewFakeXorHasher(), cache)
	//prunedOp := PruneToFindConsistent(999999999999, 1000000000000)

	prunedOp := PruneToFind(10000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hash := prunedOp.Accept(visitor)
		assert.NotNil(b, hash)
	}
}
