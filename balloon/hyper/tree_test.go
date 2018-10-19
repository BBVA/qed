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

package hyper

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bbva/qed/balloon/cache"
	"github.com/bbva/qed/balloon/visitor"
	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/log"
	"github.com/bbva/qed/storage"
	"github.com/bbva/qed/testutils/rand"
	storage_utils "github.com/bbva/qed/testutils/storage"
	"github.com/bbva/qed/util"
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
	simpleCache := cache.NewSimpleCache(10)
	tree := NewHyperTree(hashing.NewFakeXorHasher, store, simpleCache)

	for i, c := range testCases {
		index := uint64(i)
		commitment, mutations, err := tree.Add(c.eventDigest, index)
		tree.store.Mutate(mutations)
		require.NoErrorf(t, err, "This should not fail for index %d", i)
		assert.Equalf(t, c.expectedRootHash, commitment, "Incorrect root hash for index %d", i)

	}
}

func TestProveMembership(t *testing.T) {

	log.SetLogger("TestProveMembership", log.SILENT)

	hasher := hashing.NewFakeXorHasher()
	digest := hasher.Do(hashing.Digest{0x0})

	testCases := []struct {
		addOps            map[uint64]hashing.Digest
		expectedAuditPath visitor.AuditPath
	}{
		{
			addOps: map[uint64]hashing.Digest{
				uint64(0): {0},
			},
			expectedAuditPath: visitor.AuditPath{
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
				uint64(0): {0},
				uint64(1): {1},
				uint64(2): {2},
			},
			expectedAuditPath: visitor.AuditPath{
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

	for i, c := range testCases {
		store, closeF := storage_utils.OpenBPlusTreeStore()
		defer closeF()
		simpleCache := cache.NewSimpleCache(10)
		tree := NewHyperTree(hashing.NewFakeXorHasher, store, simpleCache)

		for index, digest := range c.addOps {
			_, mutations, err := tree.Add(digest, index)
			tree.store.Mutate(mutations)
			require.NoErrorf(t, err, "This should not fail for index %d", i)
		}
		leaf, err := store.Get(storage.IndexPrefix, digest)
		require.NoErrorf(t, err, "No leaf with digest %v", err)

		pf, err := tree.QueryMembership(leaf.Key, leaf.Value)
		require.NoErrorf(t, err, "Error adding to the tree: %v for index %d", err, i)
		assert.Equalf(t, c.expectedAuditPath, pf.AuditPath(), "Incorrect audit path for index %d", i)
	}
}

func TestAddAndVerify(t *testing.T) {

	log.SetLogger("TestAddAndVerify", log.SILENT)

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
		store, closeF := storage_utils.OpenBPlusTreeStore()
		defer closeF()
		simpleCache := cache.NewSimpleCache(10)
		tree := NewHyperTree(c.hasherF, store, simpleCache)

		key := hasher.Do(hashing.Digest("a test event"))
		commitment, mutations, err := tree.Add(key, value)
		tree.store.Mutate(mutations)
		require.NoErrorf(t, err, "This should not fail for index %d", i)

		leaf, err := store.Get(storage.IndexPrefix, key)
		require.NoErrorf(t, err, "No leaf with digest %v", err)

		proof, err := tree.QueryMembership(leaf.Key, leaf.Value)
		require.Nilf(t, err, "Error must be nil for index %d", i)
		assert.Equalf(t, util.Uint64AsBytes(value), proof.Value, "Incorrect actual value for index %d", i)

		correct := tree.VerifyMembership(proof, value, key, commitment)
		assert.Truef(t, correct, "Key %x should be a member for index %d", key, i)
	}
}

func TestDeterministicAdd(t *testing.T) {

	log.SetLogger("TestDeterministicAdd", log.SILENT)

	hasher := hashing.NewSha256Hasher()

	// create two trees
	cache1 := cache.NewSimpleCache(0)
	cache2 := cache.NewSimpleCache(0)
	store1, closeF1 := storage_utils.OpenBPlusTreeStore()
	store2, closeF2 := storage_utils.OpenBPlusTreeStore()
	defer closeF1()
	defer closeF2()
	tree1 := NewHyperTree(hashing.NewSha256Hasher, store1, cache1)
	tree2 := NewHyperTree(hashing.NewSha256Hasher, store2, cache2)

	// insert a bunch of events in both trees
	for i := 0; i < 100; i++ {
		event := rand.Bytes(32)
		eventDigest := hasher.Do(event)
		version := uint64(i)
		_, m1, _ := tree1.Add(eventDigest, version)
		store1.Mutate(m1)
		_, m2, _ := tree2.Add(eventDigest, version)
		store2.Mutate(m2)
	}

	// check index store equality
	reader11 := store1.GetAll(storage.IndexPrefix)
	reader21 := store2.GetAll(storage.IndexPrefix)
	defer reader11.Close()
	defer reader21.Close()
	buff11 := make([]*storage.KVPair, 0)
	buff21 := make([]*storage.KVPair, 0)
	for {
		b := make([]*storage.KVPair, 100)
		n, err := reader11.Read(b)
		if err != nil || n == 0 {
			break
		}
		buff11 = append(buff11, b...)
	}
	for {
		b := make([]*storage.KVPair, 100)
		n, err := reader21.Read(b)
		if err != nil || n == 0 {
			break
		}
		buff21 = append(buff21, b...)
	}
	require.Equalf(t, buff11, buff21, "The stored indexes should be equal")

	// check cache store equality
	reader12 := store1.GetAll(storage.HyperCachePrefix)
	reader22 := store2.GetAll(storage.HyperCachePrefix)
	defer reader12.Close()
	defer reader22.Close()
	buff12 := make([]*storage.KVPair, 0)
	buff22 := make([]*storage.KVPair, 0)
	for {
		b := make([]*storage.KVPair, 100)
		n, err := reader12.Read(b)
		if err != nil || n == 0 {
			break
		}
		buff12 = append(buff12, b...)
	}
	for {
		b := make([]*storage.KVPair, 100)
		n, err := reader22.Read(b)
		if err != nil || n == 0 {
			break
		}
		buff22 = append(buff22, b...)
	}
	require.Equalf(t, buff12, buff22, "The stored cached digests should be equal")

	// check cache equality
	require.True(t, cache1.Equal(cache2), "Both caches should be equal")
}

func TestRebuildCache(t *testing.T) {

	log.SetLogger("TestRebuildCache", log.SILENT)

	store, closeF := storage_utils.OpenBPlusTreeStore()
	defer closeF()
	hasherF := hashing.NewSha256Hasher
	hasher := hasherF()

	firstCache := cache.NewSimpleCache(10)
	tree := NewHyperTree(hasherF, store, firstCache)
	require.True(t, firstCache.Size() == 0, "The cache should be empty")

	// store multiple elements
	for i := 0; i < 1000; i++ {
		key := hasher.Do(rand.Bytes(32))
		_, mutations, _ := tree.Add(key, uint64(i))
		store.Mutate(mutations)
	}
	expectedSize := firstCache.Size()

	// Close tree and reopen with a new fresh cache
	tree.Close()
	secondCache := cache.NewSimpleCache(10)
	tree = NewHyperTree(hasherF, store, secondCache)

	require.Equal(t, expectedSize, secondCache.Size(), "The size of the caches should match")
	require.True(t, firstCache.Equal(secondCache), "The caches should be equal")
}

func BenchmarkAdd(b *testing.B) {

	log.SetLogger("BenchmarkAdd", log.SILENT)

	store, closeF := storage_utils.OpenBadgerStore(b, "/var/tmp/hyper_tree_test.db")
	defer closeF()

	hasher := hashing.NewSha256Hasher()
	simpleCache := cache.NewSimpleCache(0)
	tree := NewHyperTree(hashing.NewSha256Hasher, store, simpleCache)

	b.ResetTimer()
	b.N = 100000
	for i := 0; i < b.N; i++ {
		key := hasher.Do(rand.Bytes(32))
		_, mutations, _ := tree.Add(key, uint64(i))
		store.Mutate(mutations)
	}
}
