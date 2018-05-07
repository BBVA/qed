// Copyright © 2018 Banco Bilbao Vizcaya Argentaria S.A.  All rights reserved.
// Use of this source code is governed by an Apache 2 License
// that can be found in the LICENSE file

package hyper

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"testing"

	"verifiabledata/balloon/hashing"
	"verifiabledata/balloon/storage"
	"verifiabledata/balloon/storage/cache"
	"verifiabledata/log"
)

func TestVerify(t *testing.T) {
	hasher := hashing.XorHasher

	expectedCommitment := []byte{0xff}
	key := []byte{0xff}
	value := uint64(1)
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

	proof := NewFakeProof("test", auditPath, hasher)

	correct := proof.Verify(expectedCommitment, key, value)

	if !correct {
		t.Fatalf("The verification failed")
	}
}

func TestAddAndVerify(t *testing.T) {

	store, closeF := openBPlusStorage()
	defer closeF()

	hasher := hashing.Sha256Hasher
	l := log.NewError(os.Stdout, "Server: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Llongfile)
	ht := NewTree("/tmp/balloon.db", cache.NewSimpleCache(storage.SIZE20), store, hasher, l)

	key := hasher([]byte("a test event"))
	value := uint64(0)

	valueBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(valueBytes, value)

	commitment := <-ht.Add(key, valueBytes)

	membershipProof := <-ht.Prove(key)

	if !bytes.Equal(membershipProof.ActualValue, valueBytes) {
		fmt.Errorf("Wrong proof: expected value %v, actual %v", value, membershipProof.ActualValue)
	}

	proof := NewProof("/tmp/balloon.db", membershipProof.AuditPath, hasher)

	correct := proof.Verify(commitment, key, value)

	if !correct {
		t.Errorf("Key %v should be a member", key)
	}

}

func TestAddAndVerifyXor(t *testing.T) {

	store, closeF := openBPlusStorage()
	defer closeF()

	hasher := hashing.XorHasher
	l := log.NewError(os.Stdout, "Server: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Llongfile)
	ht := NewTree("/tmp/balloon.db", cache.NewSimpleCache(0), store, hasher, l)

	key := hasher([]byte("a test event"))
	value := uint64(0)

	valueBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(valueBytes, value)

	commitment := <-ht.Add(key, valueBytes)
	membershipProof := <-ht.Prove(key)

	if !bytes.Equal(membershipProof.ActualValue, valueBytes) {
		fmt.Errorf("Wrong proof: expected value %v, actual %v", value, membershipProof.ActualValue)
	}

	proof := NewProof("/tmp/balloon.db", membershipProof.AuditPath, hasher)

	correct := proof.Verify(commitment, key, value)

	if !correct {
		t.Errorf("Key %v should be a member", key)
	}

}

func TestAddAndVerifyPearson(t *testing.T) {

	store, closeF := openBadgerStorage("/tmp/balloon.db") // openBPlusStorage()
	defer closeF()

	hasher := hashing.Pearson
	l := log.NewError(os.Stdout, "Server: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Llongfile)
	ht := NewTree("/tmp/balloon.db", cache.NewSimpleCache(0), store, hasher, l)

	key := hasher([]byte("a test event"))
	value := uint64(0)

	valueBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(valueBytes, value)

	commitment := <-ht.Add(key, valueBytes)

	membershipProof := <-ht.Prove(key)

	if !bytes.Equal(membershipProof.ActualValue, valueBytes) {
		fmt.Errorf("Wrong proof: expected value %v, actual %v", value, membershipProof.ActualValue)
	}

	proof := NewProof("/tmp/balloon.db", membershipProof.AuditPath, hasher)

	correct := proof.Verify(commitment, key, value)

	if !correct {
		t.Errorf("Key %v should be a member", key)
	}

}
