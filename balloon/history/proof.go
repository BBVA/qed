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
	"github.com/bbva/qed/balloon/history/visit"
	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/log"
)

type MembershipProof struct {
	auditPath      navigation.AuditPath
	Index, Version uint64
	hasher         hashing.Hasher // TODO should we remove this and pass as an argument when verifying?
	// TODO should we include the eventDigest?
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
func (p MembershipProof) Verify(eventDigest []byte, expectedDigest hashing.Digest) (correct bool) {

	// visitors
	computeHash := visit.NewComputeHashVisitor(p.hasher)

	// build pruning context
	var cacheResolver pruning.CacheResolver
	if p.Index == p.Version {
		cacheResolver = pruning.NewSingleTargetedCacheResolver(p.Version)
	} else {
		cacheResolver = pruning.NewDoubleTargetedCacheResolver(p.Index, p.Version)
	}
	context := pruning.NewPruningContext(
		cacheResolver,
		p.auditPath,
	)

	// traverse from root and generate a visitable pruned tree
	pruned, err := pruning.NewVerifyPruner(p.Version, eventDigest, context).Prune()
	if err != nil {
		return false
	}

	// visit the pruned tree
	recomputed := pruned.PostOrder(computeHash)

	return bytes.Equal(recomputed, expectedDigest)
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

	log.Debugf("Verifying incremental between versions %d and %d", p.StartVersion, p.EndVersion)

	// visitors
	computeHash := visit.NewComputeHashVisitor(p.hasher)

	// build pruning context
	startContext := pruning.NewPruningContext(
		pruning.NewIncrementalCacheResolver(p.StartVersion, p.EndVersion),
		p.AuditPath,
	)
	endContext := pruning.NewPruningContext(
		pruning.NewIncrementalCacheResolver(p.StartVersion, p.EndVersion),
		p.AuditPath,
	)

	// traverse from root and generate a visitable pruned tree
	startPruned, err := pruning.NewVerifyPruner(p.StartVersion, startDigest, startContext).Prune()
	if err != nil {
		return false
	}
	endPruned, err := pruning.NewVerifyPruner(p.EndVersion, endDigest, endContext).Prune()
	if err != nil {
		return false
	}

	// visit the pruned trees
	startRecomputed := startPruned.PostOrder(computeHash)
	endRecomputed := endPruned.PostOrder(computeHash)

	return bytes.Equal(startRecomputed, startDigest) && bytes.Equal(endRecomputed, endDigest)

}
