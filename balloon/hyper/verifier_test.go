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
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/bbva/qed/balloon/hashing"
	"github.com/bbva/qed/balloon/storage"
	"github.com/bbva/qed/storage/cache"
)

func TestVerify(t *testing.T) {
	hasher := hashing.XorHasher

	expectedCommitment := []byte{0xff}
	key := []byte{0xff}
	value := uint64(1)
	auditPath := [][]byte{
		{0x00},
		{0x00},
		{0x00},
		{0x00},
		{0x00},
		{0x00},
		{0x00},
		{0x00},
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

	membershipProof := <-ht.ProveMembership(key)

	if !bytes.Equal(membershipProof.ActualValue, valueBytes) {
		t.Errorf("Wrong proof: expected value %v, actual %v", value, membershipProof.ActualValue)
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
	membershipProof := <-ht.ProveMembership(key)

	if !bytes.Equal(membershipProof.ActualValue, valueBytes) {
		t.Errorf("Wrong proof: expected value %v, actual %v", value, membershipProof.ActualValue)
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

	hasher := hashing.PearsonHasher
	ht := NewTree("/var/tmp/balloon.db", cache.NewSimpleCache(0), store, hasher)

	key := hasher([]byte("a test event"))
	value := uint64(0)

	valueBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(valueBytes, value)

	commitment := <-ht.Add(key, valueBytes)

	membershipProof := <-ht.ProveMembership(key)

	if !bytes.Equal(membershipProof.ActualValue, valueBytes) {
		t.Errorf("Wrong proof: expected value %v, actual %v", value, membershipProof.ActualValue)
	}

	proof := NewProof("/var/tmp/balloon.db", membershipProof.AuditPath, hasher)

	correct := proof.Verify(commitment, key, value)

	if !correct {
		t.Errorf("Key %v should be a member", key)
	}

}
