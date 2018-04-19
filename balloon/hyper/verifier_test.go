// Copyright © 2018 Banco Bilbao Vizcaya Argentaria S.A.  All rights reserved.
// Use of this source code is governed by an Apache 2 License
// that can be found in the LICENSE file

package hyper

import (
	"bytes"
	"fmt"
	"testing"
	"verifiabledata/balloon/hashing"
	"verifiabledata/balloon/storage/cache"
)

func TestVerify(t *testing.T) {
	hasher := hashing.XorHasher
	verifier := NewVerifier("test", hasher, fakeLeafHasherF(hasher), fakeInteriorHasherF(hasher))

	expectedCommitment := []byte{0x5a}
	key := []byte{0x5a}
	value := []byte{0x01}
	auditPath := [][]byte{
		[]byte{0x5a},
		[]byte{0x00},
		[]byte{0x00},
		[]byte{0x00},
		[]byte{0x00},
		[]byte{0x00},
		[]byte{0x00},
		[]byte{0x00},
		[]byte{0x00},
	}

	correct, recomputed := verifier.Verify(expectedCommitment, auditPath, key, value)

	if !correct {
		t.Fatalf("The verification failed")
	}

	if bytes.Compare(recomputed, expectedCommitment) != 0 {
		t.Fatalf("Expected: %x, Actual: %x", expectedCommitment, recomputed)
	}
}

func TestAddAndVerify(t *testing.T) {

	store, closeF := openBPlusStorage()
	defer closeF()

	hasher := hashing.Sha256Hasher
	ht := NewTree(string(0x0), 30, cache.NewSimpleCache(5000000), store, hasher, LeafHasherF(hasher), InteriorHasherF(hasher))
	verifier := NewVerifier(string(0x0), hasher, LeafHasherF(hasher), InteriorHasherF(hasher))

	key := hasher([]byte("test event 1"))
	value := hasher([]byte{0})

	commitment := <-ht.Add(key, value)
	proof := <-ht.Prove(key)

	if !bytes.Equal(proof.ActualValue, value) {
		fmt.Errorf("Wrong proof: expected value %v , actual %v", value, proof.ActualValue)
	}

	fullPath := make([][]byte, len(proof.AuditPath)+1)
	copy(fullPath[:1], [][]byte{key})
	copy(fullPath[1:], proof.AuditPath)

	isMembership, recomputed := verifier.Verify(commitment, fullPath, key, value)

	if !isMembership {
		fmt.Errorf("Key %v should be a member", key)
	}

	if !bytes.Equal(recomputed, commitment) {
		fmt.Errorf("Commitments don't match: expected %v , actual %v", commitment, recomputed)
	}

}
