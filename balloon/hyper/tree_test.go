// Copyright © 2018 Banco Bilbao Vizcaya Argentaria S.A.  All rights reserved.
// Use of this source code is governed by an Apache 2 License
// that can be found in the LICENSE file

package hyper

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"os"
	"testing"
	"verifiabledata/balloon/hashing"
	"verifiabledata/balloon/storage/badger"
	"verifiabledata/balloon/storage/bplus"
	"verifiabledata/balloon/storage/cache"
)

func TestAdd(t *testing.T) {
	store, closeF := openBPlusStorage()
	defer closeF()

	cache := cache.NewSimpleCache(5000)
	hasher := hashing.XorHasher

	ht := NewTree(string(0x0), 2, cache, store, hasher, fakeLeafHasherF(hasher), fakeInteriorHasherF(hasher))

	key := []byte{0x5a}
	value := []byte{0x01}

	expectedCommitment := []byte{0x5a}
	commitment := <-ht.Add(key, value)

	if bytes.Compare(commitment, expectedCommitment) != 0 {
		t.Fatalf("Expected: %x, Actual: %x", expectedCommitment, commitment)
	}

}

// func TestExistentAuditPath(t *testing.T) {
// 	store, closeF := openBPlusStorage()
// 	defer closeF()

// 	cache := cache.NewSimpleCache(5000)

// 	id := "my test tree"
// 	ht := NewTree(id, cache, store, hashing.Sha256Hasher)

// 	//var commitment []byte
// 	// Add some elements in the tree
// 	for i := 0; i < 5; i++ {
// 		event := fmt.Sprintf("Hello World%d", i)
// 		key := hashing.Sha256Hasher([]byte(event))
// 		value := make([]byte, 8)
// 		binary.BigEndian.PutUint64(value, uint64(i))
// 		if i == 3 {
// 			//commitment = <-ht.Add(key, value)
// 		}
// 	}

// 	val := hashing.Sha256Hasher([]byte("Hello World3"))
// 	version := make([]byte, 8)
// 	binary.BigEndian.PutUint64(version, uint64(3))
// 	proof := <-ht.AuditPath(val)
// 	fmt.Printf("Path length: %v, actual value: %x\n", len(proof.AuditPath), proof.ActualValue)
// 	for i := 0; i < len(proof.AuditPath); i++ {
// 		fmt.Printf("Index %v - %x\n", i, proof.AuditPath[i])
// 	}

// 	// verifier := NewVerifier(id, hashing.Sha256Hasher)
// 	// correct, recomputed := verifier.Verify(commitment, proof.AuditPath, val, version)

// 	// if !correct {
// 	// 	t.Fatalf("Verify error: expected %x, actual %x", commitment, recomputed)
// 	// }

// }

func randomBytes(n int) []byte {
	bytes := make([]byte, n)
	_, err := rand.Read(bytes)
	if err != nil {
		panic(err)
	}

	return bytes
}

func BenchmarkAdd(b *testing.B) {
	store, closeF := openBadgerStorage()
	defer closeF()
	cache := cache.NewSimpleCache(5000000)
	hasher := hashing.Sha256Hasher
	ht := NewTree("my test tree", 30, cache, store, hasher, leafHashF(hasher), interiorHashF(hasher))
	b.N = 100000
	for i := 0; i < b.N; i++ {
		key := randomBytes(64)
		value := randomBytes(1)
		store.Add(key, value)
		<-ht.Add(key, value)
	}
	b.Logf("stats = %+v\n", ht.stats)
}

func openBPlusStorage() (*bplus.BPlusTreeStorage, func()) {
	store := bplus.NewBPlusTreeStorage()
	return store, func() {
		store.Close()
	}
}

func openBadgerStorage() (*badger.BadgerStorage, func()) {
	store := badger.NewBadgerStorage("/tmp/hyper_tree_test.db")
	return store, func() {
		store.Close()
		deleteFile("/tmp/hyper_tree_test.db")
	}
}

func deleteFile(path string) {
	err := os.RemoveAll(path)
	if err != nil {
		fmt.Printf("Unable to remove db file %s", err)
	}
}
