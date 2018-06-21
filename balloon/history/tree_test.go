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

package history

import (
	"encoding/binary"
	"fmt"
	"os"
	"testing"

	"github.com/bbva/qed/balloon/proof"

	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/metrics"
	"github.com/bbva/qed/storage/badger"
	"github.com/bbva/qed/storage/bplus"
	"github.com/bbva/qed/testutils/rand"
	assert "github.com/stretchr/testify/require"
)

func TestAdd(t *testing.T) {
	hasher := new(hashing.XorHasher)
	digest := hasher.Do([]byte{0x0})
	index := make([]byte, 8)

	frozen, close := openBPlusStorage()
	defer close()

	expectedRH := []byte{0x00}

	tree := NewTree(string(0x0), frozen, hasher)
	rh, err := tree.Add(digest, index)
	assert.Nil(t, err, "Error adding to the tree: %v", err)
	assert.Equal(t, expectedRH, rh, "Incorrect root hash")
}

func TestClose(t *testing.T) {

}

func TestProveMembership(t *testing.T) {
	hasher := new(hashing.XorHasher)
	digest := hasher.Do([]byte{0x0})
	index := make([]byte, 8)

	frozen, close := openBPlusStorage()
	defer close()

	tree := NewTree(string(0x0), frozen, hasher)
	rh, err := tree.Add(digest, index)
	assert.Nil(t, err, "Error adding to the tree: %v", err)
	assert.Equal(t, []byte{0x00}, rh, "Incorrect root hash")

	pf, err := tree.ProveMembership(digest, 0, 0)
	assert.Nil(t, err, "Error adding to the tree: %v", err)
	pos := NewPosition(0, 0, 0)

	ap := make(proof.AuditPath)
	ap[pos.StringId()] = []byte{0x0}
	assert.Equal(t, pf.AuditPath(), ap, "Incorrect audit path")

}

func TestProveMembershipN(t *testing.T) {
	var err error
	var digest []byte
	var i uint64
	hasher := new(hashing.XorHasher)

	frozen, close := openBPlusStorage()
	defer close()

	tree := NewTree("treeId", frozen, hasher)

	for i = 0; i < 10; i++ {
		index := make([]byte, 8)
		binary.LittleEndian.PutUint64(index, i)
		digest = hasher.Do(index)
		tree.Add(digest, index)
	}

	pf, err := tree.ProveMembership(digest, 8, 9)

	assert.Nil(t, err, "Error adding to the tree: %v", err)

	ap := proof.AuditPath{"0|3": []uint8{0x7}, "12|2": []uint8{0xe}, "10|1": []uint8{0xb}, "9|0": []uint8{0x0}, "8|0": []uint8{0x0}}
	assert.Equal(t, ap, pf.AuditPath(), "Incorrect audit path")

}

func BenchmarkAdd(b *testing.B) {
	store, closeF := openBadgerStorage()
	defer closeF()
	ht := NewTree("treeId", store, new(hashing.Sha256Hasher))
	b.N = 100000
	for i := 0; i < b.N; i++ {
		key := rand.Bytes(64)
		index := make([]byte, 8)
		binary.LittleEndian.PutUint64(index, uint64(i))
		ht.Add(key, index)
	}
	b.Logf("stats = %+v\n", metrics.History)
}

func openBPlusStorage() (*bplus.BPlusTreeStorage, func()) {
	store := bplus.NewBPlusTreeStorage()
	return store, func() {
		store.Close()
	}
}

func openBadgerStorage() (*badger.BadgerStorage, func()) {
	store := badger.NewBadgerStorage("/var/tmp/history_store_test.db")
	return store, func() {
		fmt.Println("Cleaning...")
		store.Close()
		deleteFile("/var/tmp/history_store_test.db")
	}
}

func deleteFile(path string) {
	err := os.RemoveAll(path)
	if err != nil {
		fmt.Printf("Unable to remove db file %s", err)
	}
}
