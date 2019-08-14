/*
   Copyright 2018-2019 Banco Bilbao Vizcaya Argentaria, S.A.

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

	"github.com/bbva/qed/crypto/hashing"
)

type AuditPath map[string]hashing.Digest

func (p AuditPath) Get(pos position) (hashing.Digest, bool) {
	digest, ok := p[pos.StringId()]
	return digest, ok
}

func NewAuditPath() AuditPath {
	return make(AuditPath, 0)
}

type QueryProof struct {
	AuditPath  AuditPath
	Key, Value []byte
	hasher     hashing.Hasher
}

func NewQueryProof(key, value []byte, auditPath AuditPath, hasher hashing.Hasher) *QueryProof {
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

	if len(p.AuditPath) == 0 {
		// an empty audit path (empty tree) shows non-membersip for any key
		return false
	}

	// build a stack of operations and then interpret it to recompute the root hash
	ops := pruneToVerify(key, p.Value, p.hasher.Len()-uint16(len(p.AuditPath)))
	ctx := &pruningContext{
		Hasher:    p.hasher,
		AuditPath: p.AuditPath,
	}
	recomputed, err := ops.Pop().Interpret(ops, ctx)
	if err != nil {
		panic(err)
	}

	return bytes.Equal(key, p.Key) && bytes.Equal(recomputed, expectedRootHash)

}
