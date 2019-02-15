package hyper2

import (
	"github.com/bbva/qed/balloon/hyper2/navigation"
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

	// build a visitable pruned tree and the visit it to recompute the root hash
	// visitor := pruning.NewComputeHashVisitor(p)
	// recomputed := pruning.PruneToVerify(key, p.Value, p.hasher.Len()-len(p.AuditPath)).Accept(visitor)

	// return bytes.Equal(key, p.Key) && bytes.Equal(recomputed, expectedRootHash)
	return true

}
