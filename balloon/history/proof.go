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

package history

import (
	"bytes"

	"github.com/bbva/qed/balloon/history/navigation"
	"github.com/bbva/qed/balloon/history/pruning"
	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/log"
)

type MembershipProof struct {
	auditPath      navigation.AuditPath
	Index, Version uint64
	hasher         hashing.Hasher // TODO should we remove this and pass as an argument when verifying?
}

func NewMembershipProof(index, version uint64, auditPath navigation.AuditPath, hasher hashing.Hasher) *MembershipProof {
	return &MembershipProof{
		auditPath: auditPath,
		Index:     index,
		Version:   version,
		hasher:    hasher,
	}
}

func (p MembershipProof) AuditPath() navigation.AuditPath {
	return p.auditPath
}

// Verify verifies a membership proof
func (p MembershipProof) Verify(eventDigest []byte, expectedRootHash hashing.Digest) (correct bool) {

	log.Debugf("Verifying membership proof for index %d and version %d", p.Index, p.Version)

	// build a visitable pruned tree and then visit it to recompute root hash
	visitor := pruning.NewComputeHashVisitor(p.hasher, p.auditPath)
	recomputed := pruning.PruneToVerify(p.Index, p.Version, eventDigest).Accept(visitor)

	return bytes.Equal(recomputed, expectedRootHash)
}

type IncrementalProof struct {
	AuditPath                navigation.AuditPath
	StartVersion, EndVersion uint64
	hasher                   hashing.Hasher
}

func NewIncrementalProof(start, end uint64, auditPath navigation.AuditPath, hasher hashing.Hasher) *IncrementalProof {
	return &IncrementalProof{
		AuditPath:    auditPath,
		StartVersion: start,
		EndVersion:   end,
		hasher:       hasher,
	}
}

func (p IncrementalProof) Verify(startDigest, endDigest hashing.Digest) (correct bool) {

	log.Debugf("Verifying incremental proof between versions %d and %d", p.StartVersion, p.EndVersion)

	// build two visitable pruned trees and then visit them to recompute root hash
	visitor := pruning.NewComputeHashVisitor(p.hasher, p.AuditPath)
	startRecomputed := pruning.PruneToVerifyIncrementalStart(p.StartVersion).Accept(visitor)
	endRecomputed := pruning.PruneToVerifyIncrementalEnd(p.StartVersion, p.EndVersion).Accept(visitor)

	return bytes.Equal(startRecomputed, startDigest) && bytes.Equal(endRecomputed, endDigest)

}
