package history

import (
	"bytes"

	"github.com/bbva/qed/balloon2/common"
	"github.com/bbva/qed/log"
)

type MembershipProof struct {
	AuditPath      common.AuditPath
	Index, Version uint64
	hasher         common.Hasher // TODO should we include the eventDigest?
}

func NewMembershipProof(index, version uint64, auditPath common.AuditPath, hasher common.Hasher) *MembershipProof {
	return &MembershipProof{
		AuditPath: auditPath,
		Index:     index,
		Version:   version,
		hasher:    hasher,
	}
}

// Verify verifies a membership proof
func (p MembershipProof) Verify(expectedDigest, eventDigest common.Digest) (correct bool) {
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
		cache:         p.AuditPath,
	}

	// traverse from root and generate a visitable pruned tree
	pruned := NewVerifyPruner(eventDigest, context).Prune()

	// visit the pruned tree
	recomputed := pruned.PostOrder(computeHash).(common.Digest)

	return bytes.Equal(recomputed, expectedDigest)
}
