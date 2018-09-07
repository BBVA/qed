package history

import (
	"bytes"

	"github.com/bbva/qed/balloon2/common"
	"github.com/bbva/qed/log"
)

type MembershipProof struct {
	auditPath      common.AuditPath
	Index, Version uint64
	hasher         common.Hasher // TODO should we remove this and pass as an argument when verifying?
	// TODO should we include the eventDigest?
}

func NewMembershipProof(index, version uint64, auditPath common.AuditPath, hasher common.Hasher) *MembershipProof {
	return &MembershipProof{
		auditPath: auditPath,
		Index:     index,
		Version:   version,
		hasher:    hasher,
	}
}

func (p MembershipProof) AuditPath() common.AuditPath {
	return p.auditPath
}

// Verify verifies a membership proof
func (p MembershipProof) Verify(eventDigest []byte, expectedDigest common.Digest) (correct bool) {
	log.Debugf("Verifying membership for version %d", p.Version)

	// visitors
	computeHash := common.NewComputeHashVisitor(p.hasher)

	// build pruning context
	var cacheResolver CacheResolver
	if p.Index == p.Version {
		cacheResolver = NewSingleTargetedCacheResolver(p.Version)
	} else {
		cacheResolver = NewDoubleTargetedCacheResolver(p.Index, p.Version)
	}
	context := PruningContext{
		navigator:     NewHistoryTreeNavigator(p.Version),
		cacheResolver: cacheResolver,
		cache:         p.auditPath,
	}

	// traverse from root and generate a visitable pruned tree
	pruned := NewVerifyPruner(eventDigest, context).Prune()

	// visit the pruned tree
	recomputed := pruned.PostOrder(computeHash).(common.Digest)

	return bytes.Equal(recomputed, expectedDigest)
}

type IncrementalProof struct {
	AuditPath                common.AuditPath
	StartVersion, EndVersion uint64
	hasher                   common.Hasher
}

func NewIncrementalProof(start, end uint64, auditPath common.AuditPath, hasher common.Hasher) *IncrementalProof {
	return &IncrementalProof{
		AuditPath:    auditPath,
		StartVersion: start,
		EndVersion:   end,
		hasher:       hasher,
	}
}

func (p IncrementalProof) Verify(startDigest, endDigest common.Digest) (correct bool) {

	log.Debugf("Verifying incremental between versions %d and %d", p.StartVersion, p.EndVersion)

	// visitors
	computeHash := common.NewComputeHashVisitor(p.hasher)

	// build pruning context
	startContext := PruningContext{
		navigator:     NewHistoryTreeNavigator(p.StartVersion),
		cacheResolver: NewIncrementalCacheResolver(p.StartVersion, p.EndVersion),
		cache:         p.AuditPath,
	}
	endContext := PruningContext{
		navigator:     NewHistoryTreeNavigator(p.EndVersion),
		cacheResolver: NewIncrementalCacheResolver(p.StartVersion, p.EndVersion),
		cache:         p.AuditPath,
	}

	// traverse from root and generate a visitable pruned tree
	startPruned := NewVerifyPruner(startDigest, startContext).Prune()
	endPruned := NewVerifyPruner(endDigest, endContext).Prune()

	// visit the pruned trees
	startRecomputed := startPruned.PostOrder(computeHash).(common.Digest)
	endRecomputed := endPruned.PostOrder(computeHash).(common.Digest)

	return bytes.Equal(startRecomputed, startDigest) && bytes.Equal(endRecomputed, endDigest)

}
