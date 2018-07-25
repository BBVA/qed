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

package history

import (
	"testing"

	"github.com/bbva/qed/balloon/proof"
	"github.com/bbva/qed/hashing"
	assert "github.com/stretchr/testify/require"
)

func TestAddAndVerifyXor(t *testing.T) {

	frozen, closeF := openBPlusStorage()
	defer closeF()

	hasher := new(hashing.XorHasher)
	ht := NewTree(string(0x0), frozen, hasher)

	key := hasher.Do([]byte("a test event"))

	index := make([]byte, 8)

	commitment, err := ht.Add(key, index)

	assert.Nil(t, err, "Error must be nil")

	membershipProof, err := ht.ProveMembership(key, 0, 0)
	assert.Nil(t, err, "Error must be nil")

	rootPos := NewRootPosition(0)
	proof := proof.NewProof(rootPos, membershipProof.AuditPath(), hasher)

	correct := proof.Verify(commitment, index, key)

	if !correct {
		t.Errorf("Key %x should be a member", key)
	}
}

func TestAddAndVerifyPearson(t *testing.T) {

	frozen, closeF := openBPlusStorage()
	defer closeF()

	hasher := new(hashing.PearsonHasher)
	ht := NewTree(string(0x0), frozen, hasher)

	key := hasher.Do([]byte("a test event"))

	index := make([]byte, 8)

	commitment, err := ht.Add(key, index)

	assert.Nil(t, err, "Error must be nil")

	membershipProof, err := ht.ProveMembership(key, 0, 0)
	assert.Nil(t, err, "Error must be nil")

	rootPos := NewRootPosition(0)
	proof := proof.NewProof(rootPos, membershipProof.AuditPath(), hasher)

	correct := proof.Verify(commitment, index, key)

	if !correct {
		t.Errorf("Key %x should be a member", key)
	}
}

func TestAddAndVerifySha256(t *testing.T) {

	frozen, closeF := openBPlusStorage()
	defer closeF()

	hasher := hashing.NewSha256Hasher()
	ht := NewTree(string(0x0), frozen, hasher)

	key := hasher.Do([]byte("a test event"))

	index := make([]byte, 8)

	commitment, err := ht.Add(key, index)

	assert.Nil(t, err, "Error must be nil")

	membershipProof, err := ht.ProveMembership(key, 0, 0)
	assert.Nil(t, err, "Error must be nil")

	rootPos := NewRootPosition(0)
	proof := proof.NewProof(rootPos, membershipProof.AuditPath(), hasher)

	correct := proof.Verify(commitment, index, key)

	if !correct {
		t.Errorf("Key %x should be a member", key)
	}
}
