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

/* +build !release */

package hyper

import (
	"encoding/binary"
	"testing"

	"github.com/bbva/qed/balloon/proof"
	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/storage/cache"
	assert "github.com/stretchr/testify/require"
)

func TestAddAndVerifyXor(t *testing.T) {

	leaves, closeF := openBPlusStorage()
	defer closeF()

	hasher := new(hashing.XorHasher)
	ht := NewTree(string(0x0), cache.NewSimpleCache(0), leaves, hasher)

	key := hasher.Do([]byte("a test event"))
	value := uint64(0)

	valueBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(valueBytes, value)

	commitment, err := ht.Add(key, valueBytes)
	assert.Nil(t, err, "Error must be nil")
	membershipProof, actualValue, err := ht.ProveMembership(key)
	assert.Nil(t, err, "Error must be nil")

	assert.Equal(t, valueBytes, actualValue, "Incorrect actual value")

	rootPos := NewRootPosition(hasher.Len(), 0)
	proof := proof.NewProof(rootPos, membershipProof.AuditPath(), hasher)

	correct := proof.Verify(commitment, key, valueBytes)

	if !correct {
		t.Errorf("Key %x should be a member", key)
	}
}

func TestAddAndVerifyPearson(t *testing.T) {

	leaves, closeF := openBPlusStorage()
	defer closeF()

	hasher := new(hashing.PearsonHasher)
	ht := NewTree(string(0x0), cache.NewSimpleCache(0), leaves, hasher)

	key := hasher.Do([]byte("a test event"))
	value := uint64(0)

	valueBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(valueBytes, value)

	commitment, err := ht.Add(key, valueBytes)
	assert.Nil(t, err, "Error must be nil")
	membershipProof, actualValue, err := ht.ProveMembership(key)
	assert.Nil(t, err, "Error must be nil")

	assert.Equal(t, valueBytes, actualValue, "Incorrect actual value")

	rootPos := NewRootPosition(hasher.Len(), 0)
	proof := proof.NewProof(rootPos, membershipProof.AuditPath(), hasher)

	correct := proof.Verify(commitment, key, valueBytes)

	if !correct {
		t.Errorf("Key %x should be a member", key)
	}
}

func TestAddAndVerifySha256(t *testing.T) {

	leaves, closeF := openBPlusStorage()
	defer closeF()

	hasher := new(hashing.Sha256Hasher)
	ht := NewTree(string(0x0), cache.NewSimpleCache(0), leaves, hasher)

	key := hasher.Do([]byte("a test event"))
	value := uint64(0)

	valueBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(valueBytes, value)

	commitment, err := ht.Add(key, valueBytes)
	assert.Nil(t, err, "Error must be nil")
	membershipProof, actualValue, err := ht.ProveMembership(key)
	assert.Nil(t, err, "Error must be nil")

	assert.Equal(t, valueBytes, actualValue, "Incorrect actual value")

	rootPos := NewRootPosition(hasher.Len(), 0)
	proof := proof.NewProof(rootPos, membershipProof.AuditPath(), hasher)

	correct := proof.Verify(commitment, key, valueBytes)

	if !correct {
		t.Errorf("Key %x should be a member", key)
	}
}
