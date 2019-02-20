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

package hyper2

import (
	"bytes"

	"github.com/bbva/qed/balloon/hyper2/navigation"
	"github.com/bbva/qed/balloon/hyper2/pruning2"
	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/log"
)

type QueryProof struct {
	AuditPath  navigation.AuditPath
	Key, Value []byte
	hasher     hashing.Hasher
}

func NewQueryProof(key, value []byte, auditPath navigation.AuditPath, hasher hashing.Hasher) *QueryProof {
	return &QueryProof{
		Key:       key,
		Value:     value,
		AuditPath: auditPath,
		hasher:    hasher,
	}
}

// Verify verifies a membership query for a provided key from an expected
// root hash that fixes the hyper tree. Returns true if the proof is valid,
// false otherwise.
func (p QueryProof) Verify(key []byte, expectedRootHash hashing.Digest) (valid bool) {

	log.Debugf("Verifying query proof for key %d", p.Key)

	if len(p.AuditPath) == 0 {
		// an empty audit path (empty tree) shows non-membersip for any key
		return false
	}

	// build a stack of operations and then interpret it to recompute the root hash
	ops := pruning2.PruneToVerify(key, p.Value, p.hasher.Len()-uint16(len(p.AuditPath)))
	ctx := &pruning2.Context{
		Hasher:    p.hasher,
		AuditPath: p.AuditPath,
	}
	recomputed := ops.Pop().Interpret(ops, ctx)

	return bytes.Equal(key, p.Key) && bytes.Equal(recomputed, expectedRootHash)

}
