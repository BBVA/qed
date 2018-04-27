// Copyright © 2018 Banco Bilbao Vizcaya Argentaria S.A.  All rights reserved.
// Use of this source code is governed by an Apache 2 License
// that can be found in the LICENSE file

package hyper

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"testing"

	"verifiabledata/balloon/hashing"
	"verifiabledata/balloon/storage/cache"
)

func TestVerify(t *testing.T) {
	hasher := hashing.XorHasher

	expectedCommitment := []byte{0x5a}
	key := []byte{0x5a}
	value := uint(1)
	auditPath := [][]byte{
		[]byte{0x00},
		[]byte{0x00},
		[]byte{0x00},
		[]byte{0x00},
		[]byte{0x00},
		[]byte{0x00},
		[]byte{0x00},
		[]byte{0x00},
	}

	proof := NewProof("test", auditPath, FakeLeafHasherF(hasher), FakeInteriorHasherF(hasher))

	correct := proof.Verify(expectedCommitment, key, value)

	if !correct {
		t.Fatalf("The verification failed")
	}
}

func TestAddAndVerify(t *testing.T) {

	store, closeF := openBPlusStorage()
	defer closeF()

	hasher := hashing.Sha256Hasher
	ht := NewTree(string(0x0), 30, cache.NewSimpleCache(5000000), store, hasher, LeafHasherF(hasher), InteriorHasherF(hasher))

	key := hasher([]byte("test event 1"))
	value := uint(0)

	valueBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(valueBytes, uint64(value))

	commitment := <-ht.Add(key, valueBytes)
	membershipProof := <-ht.Prove(key)

	if !bytes.Equal(membershipProof.ActualValue, valueBytes) {
		fmt.Errorf("Wrong proof: expected value %v, actual %v", value, membershipProof.ActualValue)
	}

	proof := NewProof(string(0x0), membershipProof.AuditPath, LeafHasherF(hasher), InteriorHasherF(hasher))

	correct := proof.Verify(commitment, key, value)

	if !correct {
		fmt.Errorf("Key %v should be a member", key)
	}

}
