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
	"encoding/binary"
	"testing"

	"github.com/bbva/qed/balloon/proof"
	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/metrics"
	"github.com/bbva/qed/storage/cache"
	"github.com/bbva/qed/testutils/rand"
	assert "github.com/stretchr/testify/require"
)

func TestAdd(t *testing.T) {
	hasher := new(hashing.XorHasher)
	digest := hasher.Do([]byte{0x0})
	index := make([]byte, 8)
	binary.LittleEndian.PutUint64(index, 0)

	leaves, close := openBPlusStorage()
	defer close()
	cache := cache.NewSimpleCache(10)

	tree := NewTree(string(0x0), cache, leaves, hasher)
	expectedRH := []byte{0x08}

	rh, err := tree.Add(digest, index)
	assert.Nil(t, err, "Error adding to the tree: %v", err)
	assert.Equal(t, rh, expectedRH, "Incorrect root hash")
}

func TestClose(t *testing.T) {

}

func TestProveMembership(t *testing.T) {
	hasher := new(hashing.XorHasher)
	digest := hasher.Do([]byte{0x0})
	index := make([]byte, 8)
	binary.LittleEndian.PutUint64(index, 0)

	leaves, close := openBPlusStorage()
	defer close()
	cache := cache.NewSimpleCache(10)

	tree := NewTree(string(0x0), cache, leaves, hasher)
	rh, err := tree.Add(digest, index)
	assert.Nil(t, err, "Error adding to the tree: %v", err)
	assert.Equal(t, rh, []byte{0x08}, "Incorrect root hash")

	pf, _, err := tree.ProveMembership(digest)
	assert.Nil(t, err, "Error adding to the tree: %v", err)
	// pos := NewPosition([]byte{0x0}, []byte{0x00}, 8, 8, 3)
	// assert.Equal(t, pos, proof.Pos(), "Incorrect position")

	ap := proof.AuditPath{"10|4": []uint8{0x0}, "04|2": []uint8{0x0}, "00|0": []uint8{0x0}, "80|7": []uint8{0x0}, "40|6": []uint8{0x0}, "20|5": []uint8{0x0}, "08|3": []uint8{0x0}, "02|1": []uint8{0x0}, "01|0": []uint8{0x0}}
	assert.Equal(t, ap, pf.AuditPath(), "Incorrect audit path")

}

func BenchmarkAdd(b *testing.B) {
	store, closeF := openBadgerStorage("/var/tmp/hyper_tree_test.db") //openBoltStorage()
	defer closeF()

	hasher := new(hashing.Sha256Hasher)
	cache := cache.NewSimpleCache(1 << 25)
	ht := NewTree(string(0x0), cache, store, hasher)
	b.ResetTimer()
	b.N = 100000
	for i := 0; i < b.N; i++ {
		key := hasher.Do(rand.Bytes(32))
		value := rand.Bytes(1)
		store.Add(key, value)
		ht.Add(key, value)
	}
	b.Logf("stats = %+v\n", metrics.Hyper)
}
