package pruning

import (
	"fmt"

	"github.com/bbva/qed/balloon/cache"
	"github.com/bbva/qed/balloon/history/navigation"
	"github.com/bbva/qed/hashing"
)

type AuditPathVisitor struct {
	hasher hashing.Hasher
	cache  cache.Cache

	auditPath navigation.AuditPath
}

func NewAuditPathVisitor(hasher hashing.Hasher, cache cache.Cache) *AuditPathVisitor {
	return &AuditPathVisitor{
		hasher:    hasher,
		cache:     cache,
		auditPath: make(navigation.AuditPath),
	}
}

func (v AuditPathVisitor) Result() navigation.AuditPath {
	return v.auditPath
}

func (v *AuditPathVisitor) VisitLeafHashOp(op LeafHashOp) hashing.Digest {
	return v.hasher.Salted(op.Position().Bytes(), op.Value)
}

func (v *AuditPathVisitor) VisitInnerHashOp(op InnerHashOp) hashing.Digest {
	leftHash := op.Left.Accept(v)
	rightHash := op.Right.Accept(v)
	return v.hasher.Salted(op.Position().Bytes(), leftHash, rightHash)
}

func (v *AuditPathVisitor) VisitPartialInnerHashOp(op PartialInnerHashOp) hashing.Digest {
	leftHash := op.Left.Accept(v)
	return v.hasher.Salted(op.Position().Bytes(), leftHash)
}

func (v *AuditPathVisitor) VisitGetCacheOp(op GetCacheOp) hashing.Digest {
	hash, ok := v.cache.Get(op.Position().Bytes())
	if !ok {
		panic(fmt.Sprintf("Oops, something went wrong. There should be a cached element at position %v", op.Position()))
	}
	return hash
}

func (v *AuditPathVisitor) VisitPutCacheOp(op PutCacheOp) hashing.Digest {
	return op.Operation.Accept(v)
}

func (v *AuditPathVisitor) VisitMutateOp(op MutateOp) hashing.Digest {
	return op.Operation.Accept(v)
}

func (v *AuditPathVisitor) VisitCollectOp(op CollectOp) hashing.Digest {
	hash := op.Operation.Accept(v)
	v.auditPath[op.Position().FixedBytes()] = hash
	return hash
}
