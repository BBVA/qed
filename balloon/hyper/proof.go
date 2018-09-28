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

	"github.com/bbva/qed/balloon/visitor"
	"github.com/bbva/qed/hashing"
)

type QueryProof struct {
	Key, Value []byte
	auditPath  visitor.AuditPath
	hasher     hashing.Hasher
}

func NewQueryProof(key, value []byte, auditPath visitor.AuditPath, hasher hashing.Hasher) *QueryProof {
	return &QueryProof{
		Key:       key,
		Value:     value,
		auditPath: auditPath,
		hasher:    hasher,
	}
}

func (p QueryProof) AuditPath() visitor.AuditPath {
	return p.auditPath
}

// Verify verifies a membership query for a provided key from an expected
// root hash that fixes the hyper tree. Returns true if the proof is valid,
// false otherwise.
func (p QueryProof) Verify(key []byte, expectedDigest hashing.Digest) (valid bool) {

	if len(p.auditPath) == 0 {
		// and empty audit path shows non-membership for any key
		return p.Value == nil
	}

	// visitors
	computeHash := visitor.NewComputeHashVisitor(p.hasher)

	// build pruning context
	context := PruningContext{
		navigator:     NewHyperTreeNavigator(p.hasher.Len()),
		cacheResolver: nil,
		cache:         p.auditPath,
		store:         nil,
		defaultHashes: nil,
	}

	// traverse from root and generate a visitable pruned tree
	pruned := NewVerifyPruner(key, p.Value, context).Prune()

	// visit the pruned tree
	recomputed := pruned.PostOrder(computeHash).(hashing.Digest)

	return bytes.Equal(key, p.Key) && bytes.Equal(recomputed, expectedDigest)
}
