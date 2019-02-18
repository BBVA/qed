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
	"encoding/binary"
	"testing"

	"github.com/bbva/qed/balloon/cache"
	"github.com/bbva/qed/balloon/hyper2/navigation"
	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/log"
	"github.com/bbva/qed/storage"
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

func TestProveMembership(t *testing.T) {

	log.SetLogger("TestProveMembership", log.SILENT)

	testCases := []struct {
		addedKeys         map[uint64]hashing.Digest
		expectedAuditPath navigation.AuditPath
	}{
		{
			addedKeys: map[uint64]hashing.Digest{
				uint64(0): {0x0},
			},
			expectedAuditPath: navigation.AuditPath{
				"0x80|7": hashing.Digest{0x0},
				"0x40|6": hashing.Digest{0x0},
				"0x20|5": hashing.Digest{0x0},
				"0x10|4": hashing.Digest{0x0},
			},
		},
		{
			addedKeys: map[uint64]hashing.Digest{
				uint64(0): {0x0},
				uint64(1): {0x1},
				uint64(2): {0x2},
			},
			expectedAuditPath: navigation.AuditPath{
				"0x80|7": hashing.Digest{0x0},
				"0x40|6": hashing.Digest{0x0},
				"0x20|5": hashing.Digest{0x0},
				"0x10|4": hashing.Digest{0x0},
				"0x08|3": hashing.Digest{0x0},
				"0x04|2": hashing.Digest{0x0},
				"0x02|1": hashing.Digest{0x2},
				"0x01|0": hashing.Digest{0x1},
			},
		},
	}

	hasher := hashing.NewFakeXorHasher()
	searchedDigest := hasher.Do(hashing.Digest{0x0})

	for i, c := range testCases {
		store, closeF := storage_utils.OpenBPlusTreeStore()
		defer closeF()
		simpleCache := cache.NewSimpleCache(10)
		tree := NewHyperTree(hashing.NewFakeXorHasher, store, simpleCache)

		for index, digest := range c.addedKeys {
			_, mutations, err := tree.Add(digest, index)
			tree.store.Mutate(mutations)
			require.NoErrorf(t, err, "This should not fail for index %d", i)
		}

		leaf, err := store.Get(storage.IndexPrefix, searchedDigest)
		require.NoErrorf(t, err, "No leaf with digest %v", err)

		proof, err := tree.QueryMembership(leaf.Key, leaf.Value)
		require.NoErrorf(t, err, "Error adding to the tree: %v for index %d", err, i)
		assert.Equalf(t, c.expectedAuditPath, proof.AuditPath, "Incorrect audit path for index %d", i)
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
