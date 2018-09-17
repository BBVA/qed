package hyper

import (
	"testing"

	assert "github.com/stretchr/testify/require"

	"github.com/bbva/qed/balloon/common"
	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/log"
	"github.com/bbva/qed/testutils/rand"
	storage_utils "github.com/bbva/qed/testutils/storage"
	"github.com/bbva/qed/util"
)

func TestAdd(t *testing.T) {

	log.SetLogger("TestAdd", log.INFO)

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

	store, closeF := storage_utils.NewBPlusTreeStore()
	defer closeF()
	simpleCache := common.NewSimpleCache(10)
	tree := NewHyperTree(hashing.NewFakeXorHasher, store, simpleCache)

	for i, c := range testCases {
		index := uint64(i)
		commitment, mutations, err := tree.Add(c.eventDigest, index)
		tree.store.Mutate(mutations...)
		assert.NoErrorf(t, err, "This should not fail for index %d", i)
		assert.Equalf(t, c.expectedRootHash, commitment, "Incorrect root hash for index %d", i)

	}
}

func TestProveMembership(t *testing.T) {

	hasher := hashing.NewFakeXorHasher()
	digest := hasher.Do(hashing.Digest{0x0})

	testCases := []struct {
		addOps            map[uint64]hashing.Digest
		expectedAuditPath common.AuditPath
	}{
		{
			addOps: map[uint64]hashing.Digest{
				uint64(0): hashing.Digest{0},
			},
			expectedAuditPath: common.AuditPath{
				"10|4": hashing.Digest{0x0},
				"04|2": hashing.Digest{0x0},
				"80|7": hashing.Digest{0x0},
				"40|6": hashing.Digest{0x0},
				"20|5": hashing.Digest{0x0},
				"08|3": hashing.Digest{0x0},
				"02|1": hashing.Digest{0x0},
				"01|0": hashing.Digest{0x0},
			},
		},
		{
			addOps: map[uint64]hashing.Digest{
				uint64(0): hashing.Digest{0},
				uint64(1): hashing.Digest{1},
				uint64(2): hashing.Digest{2},
			},
			expectedAuditPath: common.AuditPath{
				"10|4": hashing.Digest{0x0},
				"04|2": hashing.Digest{0x0},
				"80|7": hashing.Digest{0x0},
				"40|6": hashing.Digest{0x0},
				"20|5": hashing.Digest{0x0},
				"08|3": hashing.Digest{0x0},
				"02|1": hashing.Digest{0x2},
				"01|0": hashing.Digest{0x1},
			},
		},
	}

	log.SetLogger("TestProveMembership", log.INFO)

	for i, c := range testCases {
		store, closeF := storage_utils.NewBPlusTreeStore()
		defer closeF()
		simpleCache := common.NewSimpleCache(10)
		tree := NewHyperTree(hashing.NewFakeXorHasher, store, simpleCache)

		for index, digest := range c.addOps {
			_, mutations, err := tree.Add(digest, index)
			tree.store.Mutate(mutations...)
			assert.NoErrorf(t, err, "This should not fail for index %d", i)
		}

		pf, err := tree.QueryMembership(digest)
		assert.NoErrorf(t, err, "Error adding to the tree: %v for index %d", err, i)
		assert.Equalf(t, c.expectedAuditPath, pf.AuditPath(), "Incorrect audit path for index %d", i)
	}
}

func TestAddAndVerify(t *testing.T) {

	log.SetLogger("TestAddAndVerify", log.INFO)

	value := uint64(0)

	testCases := []struct {
		hasherF func() hashing.Hasher
	}{
		{hasherF: hashing.NewXorHasher},
		{hasherF: hashing.NewSha256Hasher},
		{hasherF: hashing.NewPearsonHasher},
	}

	for i, c := range testCases {
		hasher := c.hasherF()
		store, closeF := storage_utils.NewBPlusTreeStore()
		defer closeF()
		simpleCache := common.NewSimpleCache(10)
		tree := NewHyperTree(c.hasherF, store, simpleCache)

		key := hasher.Do(hashing.Digest("a test event"))
		commitment, mutations, err := tree.Add(key, value)
		tree.store.Mutate(mutations...)
		assert.NoErrorf(t, err, "This should not fail for index %d", i)

		proof, err := tree.QueryMembership(key)
		assert.Nilf(t, err, "Error must be nil for index %d", i)
		assert.Equalf(t, util.Uint64AsBytes(value), proof.Value, "Incorrect actual value for index %d", i)

		correct := tree.VerifyMembership(proof, value, key, commitment)
		assert.Truef(t, correct, "Key %x should be a member for index %d", key, i)
	}
}

func BenchmarkAdd(b *testing.B) {

	log.SetLogger("BenchmarkAdd", log.SILENT)

	store, closeF := common.OpenBadgerStore("/var/tmp/hyper_tree_test.db")
	defer closeF()

	hasher := hashing.NewSha256Hasher()
	simpleCache := common.NewSimpleCache(0)
	tree := NewHyperTree(hashing.NewSha256Hasher, store, simpleCache)

	b.ResetTimer()
	b.N = 100000
	for i := 0; i < b.N; i++ {
		key := hasher.Do(rand.Bytes(32))
		_, mutations, _ := tree.Add(key, uint64(i))
		store.Mutate(mutations...)
	}
}
