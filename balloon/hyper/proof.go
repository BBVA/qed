package hyper

import (
	"bytes"

	"github.com/bbva/qed/balloon/common"
	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/log"
)

type QueryProof struct {
	Key, Value []byte
	auditPath  common.AuditPath
	hasher     hashing.Hasher
}

func NewQueryProof(key, value []byte, auditPath common.AuditPath, hasher hashing.Hasher) *QueryProof {
	return &QueryProof{
		Key:       key,
		Value:     value,
		auditPath: auditPath,
		hasher:    hasher,
	}
}

func (p QueryProof) AuditPath() common.AuditPath {
	return p.auditPath
}

// Verify verifies a membership query for a provided key from an expected
// root hash that fixes the hyper tree. Returns true if the proof is valid,
// false otherwise.
func (p QueryProof) Verify(key []byte, expectedDigest hashing.Digest) (valid bool) {

	log.Debugf("Verifying membership query for key %x", key)

	if len(p.auditPath) == 0 {
		// and empty audit path shows non-membership for any key
		return p.Value == nil
	}

	// visitors
	computeHash := common.NewComputeHashVisitor(p.hasher)

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
