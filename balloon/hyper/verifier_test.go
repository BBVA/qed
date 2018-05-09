package hyper

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"testing"

	"qed/balloon/hashing"
	"qed/balloon/storage"
	"qed/balloon/storage/cache"
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
	ht := NewTree("/var/tmp/balloon.db", cache.NewSimpleCache(storage.SIZE20), store, hasher)

	key := hasher([]byte("a test event"))
	value := uint64(0)

	valueBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(valueBytes, value)

	commitment := <-ht.Add(key, valueBytes)

	membershipProof := <-ht.Prove(key)

	if !bytes.Equal(membershipProof.ActualValue, valueBytes) {
		fmt.Errorf("Wrong proof: expected value %v, actual %v", value, membershipProof.ActualValue)
	}

	proof := NewProof("/var/tmp/balloon.db", membershipProof.AuditPath, hasher)

	correct := proof.Verify(commitment, key, value)

	if !correct {
		t.Errorf("Key %v should be a member", key)
	}

}

func TestAddAndVerifyXor(t *testing.T) {

	store, closeF := openBPlusStorage()
	defer closeF()

	hasher := hashing.XorHasher
	ht := NewTree("/var/tmp/balloon.db", cache.NewSimpleCache(0), store, hasher)

	key := hasher([]byte("a test event"))
	value := uint64(0)

	valueBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(valueBytes, value)

	commitment := <-ht.Add(key, valueBytes)
	membershipProof := <-ht.Prove(key)

	if !bytes.Equal(membershipProof.ActualValue, valueBytes) {
		fmt.Errorf("Wrong proof: expected value %v, actual %v", value, membershipProof.ActualValue)
	}

	proof := NewProof("/var/tmp/balloon.db", membershipProof.AuditPath, hasher)

	correct := proof.Verify(commitment, key, value)

	if !correct {
		t.Errorf("Key %v should be a member", key)
	}

}

func TestAddAndVerifyPearson(t *testing.T) {

	store, closeF := openBadgerStorage("/var/tmp/balloon.db") // openBPlusStorage()
	defer closeF()

	hasher := hashing.Pearson
	ht := NewTree("/var/tmp/balloon.db", cache.NewSimpleCache(0), store, hasher)

	key := hasher([]byte("a test event"))
	value := uint64(0)

	valueBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(valueBytes, value)

	commitment := <-ht.Add(key, valueBytes)

	membershipProof := <-ht.Prove(key)

	if !bytes.Equal(membershipProof.ActualValue, valueBytes) {
		fmt.Errorf("Wrong proof: expected value %v, actual %v", value, membershipProof.ActualValue)
	}

	proof := NewProof("/var/tmp/balloon.db", membershipProof.AuditPath, hasher)

	correct := proof.Verify(commitment, key, value)

	if !correct {
		t.Errorf("Key %v should be a member", key)
	}

}
