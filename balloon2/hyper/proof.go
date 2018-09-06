package hyper

import (
	"bytes"

	"github.com/bbva/qed/balloon2/common"
	"github.com/bbva/qed/log"
)

type QueryProof struct {
	Key, Value []byte
	AuditPath  common.AuditPath
	hasher     common.Hasher
}

func NewQueryProof(key, value []byte, auditPath common.AuditPath, hasher common.Hasher) *QueryProof {
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
func (p QueryProof) Verify(key []byte, expectedDigest common.Digest) (valid bool) {

	log.Debugf("Verifying membership query for key %x", key)

	if len(p.AuditPath) == 0 {
		// and empty audit path shows non-membership for any key
		return p.Value == nil
	}

	// visitors
	computeHash := common.NewComputeHashVisitor(p.hasher)

	// build pruning context
	context := PruningContext{
		navigator:     NewHyperTreeNavigator(p.hasher.Len()),
		cacheResolver: nil,
		cache:         p.AuditPath,
		store:         nil,
		defaultHashes: nil,
	}

	// traverse from root and generate a visitable pruned tree
	pruned := NewVerifyPruner(key, p.Value, context).Prune()

	// visit the pruned tree
	recomputed := pruned.PostOrder(computeHash).(common.Digest)

	return bytes.Equal(key, p.Key) && bytes.Equal(recomputed, expectedDigest)
}
