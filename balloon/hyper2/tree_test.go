package hyper2

import (
	"encoding/binary"
	"testing"

	"github.com/bbva/qed/balloon/cache"
	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/log"
	"github.com/bbva/qed/testutils/rand"
	storage_utils "github.com/bbva/qed/testutils/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdd(t *testing.T) {

	log.SetLogger("TestAdd", log.SILENT)

	testCases := []struct {
		eventDigest      hashing.Digest
		expectedRootHash hashing.Digest
	}{
		{hashing.Digest{0x0}, hashing.Digest{0x0}},
		{hashing.Digest{0x1}, hashing.Digest{0x1}},
		{hashing.Digest{0x2}, hashing.Digest{0x3}},
		{hashing.Digest{0x3}, hashing.Digest{0x0}},
		{hashing.Digest{0x4}, hashing.Digest{0x4}},
		{hashing.Digest{0x5}, hashing.Digest{0x1}},
		{hashing.Digest{0x6}, hashing.Digest{0x7}},
		{hashing.Digest{0x7}, hashing.Digest{0x0}},
		{hashing.Digest{0x8}, hashing.Digest{0x8}},
		{hashing.Digest{0x9}, hashing.Digest{0x1}},
	}

	store, closeF := storage_utils.OpenBPlusTreeStore()
	defer closeF()

	tree := NewHyperTree(hashing.NewFakeXorHasher, store, cache.NewSimpleCache(10))

	for i, c := range testCases {
		version := uint64(i)
		rootHash, mutations, err := tree.Add(c.eventDigest, version)
		require.NoErrorf(t, err, "This should not fail for version %d", i)
		tree.store.Mutate(mutations)
		assert.Equalf(t, c.expectedRootHash, rootHash, "Incorrect root hash for index %d", i)
	}
}

func BenchmarkAdd(b *testing.B) {

	log.SetLogger("BenchmarkAdd", log.SILENT)

	store, closeF := storage_utils.OpenBadgerStore(b, "/var/tmp/hyper_tree_test.db")
	defer closeF()

	hasher := hashing.NewSha256Hasher()
	freeCache := cache.NewFreeCache(2000 * (1 << 20))

	tree := NewHyperTree(hashing.NewSha256Hasher, store, freeCache)

	b.ResetTimer()
	b.N = 100000
	for i := 0; i < b.N; i++ {
		index := make([]byte, 8)
		binary.LittleEndian.PutUint64(index, uint64(i))
		elem := append(rand.Bytes(32), index...)
		_, mutations, err := tree.Add(hasher.Do(elem), uint64(i))
		if err != nil {
			b.Fatal(err)
		}
		store.Mutate(mutations)
	}
}
