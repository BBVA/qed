package pruning5

import (
	"fmt"

	"github.com/bbva/qed/balloon/cache"
	"github.com/bbva/qed/balloon/history/navigation"
	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/storage"
)

type OpVisitor interface {
	VisitLeafHashOp(op LeafHashOp) hashing.Digest
	VisitInnerHashOp(op InnerHashOp) hashing.Digest
	VisitPartialInnerHashOp(op PartialInnerHashOp) hashing.Digest
	VisitGetCacheOp(op GetCacheOp) hashing.Digest
	VisitPutCacheOp(op PutCacheOp) hashing.Digest
	VisitMutateOp(op MutateOp) hashing.Digest
	VisitCollectOp(op CollectOp) hashing.Digest
}

type ComputeHashVisitor struct {
	hasher hashing.Hasher
	cache  cache.Cache
}

func NewComputeHashVisitor(hasher hashing.Hasher, cache cache.Cache) *ComputeHashVisitor {
	return &ComputeHashVisitor{
		hasher: hasher,
		cache:  cache,
	}
}

func (v *ComputeHashVisitor) VisitLeafHashOp(op LeafHashOp) hashing.Digest {
	return v.hasher.Salted(op.Position().Bytes(), op.Value)
}

func (v *ComputeHashVisitor) VisitInnerHashOp(op InnerHashOp) hashing.Digest {
	leftHash := op.Left.Accept(v)
	rightHash := op.Right.Accept(v)
	return v.hasher.Salted(op.Position().Bytes(), leftHash, rightHash)
}

func (v *ComputeHashVisitor) VisitPartialInnerHashOp(op PartialInnerHashOp) hashing.Digest {
	leftHash := op.Left.Accept(v)
	return v.hasher.Salted(op.Position().Bytes(), leftHash)
}

func (v *ComputeHashVisitor) VisitGetCacheOp(op GetCacheOp) hashing.Digest {
	hash, ok := v.cache.Get(op.Position().Bytes())
	if !ok { // TODO maybe we should return an error
		panic(fmt.Sprintf("Oops, something wrong happend. There should be a cached element at position %v", op.Position()))
	}
	return hash
}

func (v *ComputeHashVisitor) VisitPutCacheOp(op PutCacheOp) hashing.Digest {
	return op.Operation.Accept(v)
}

func (v *ComputeHashVisitor) VisitMutateOp(op MutateOp) hashing.Digest {
	return op.Operation.Accept(v)
}

func (v *ComputeHashVisitor) VisitCollectOp(op CollectOp) hashing.Digest {
	return op.Operation.Accept(v)
}

type CollectMutationsVisitor struct {
	storagePrefix byte
	mutations     []*storage.Mutation
	OpVisitor
}

func NewCollectMutationsVisitor(decorated OpVisitor, storagePrefix byte) *CollectMutationsVisitor {
	return &CollectMutationsVisitor{
		OpVisitor:     decorated,
		storagePrefix: storagePrefix,
		mutations:     make([]*storage.Mutation, 0),
	}
}

func (v CollectMutationsVisitor) Result() []*storage.Mutation {
	return v.mutations
}

func (v *CollectMutationsVisitor) VisitMutateOp(op MutateOp) hashing.Digest {
	hash := op.Operation.Accept(v)
	v.mutations = append(v.mutations, storage.NewMutation(v.storagePrefix, op.Position().Bytes(), hash))
	return hash
}

type CachingVisitor struct {
	cache cache.ModifiableCache
	OpVisitor
}

func NewCachingVisitor(decorated OpVisitor, cache cache.ModifiableCache) *CachingVisitor {
	return &CachingVisitor{
		OpVisitor: decorated,
		cache:     cache,
	}
}

func (v *CachingVisitor) VisitPutCacheOp(op PutCacheOp) hashing.Digest {
	hash := op.Operation.Accept(v)
	v.cache.Put(op.Position().Bytes(), hash)
	return hash
}

type AuditPathVisitor struct {
	auditPath navigation.AuditPath
	OpVisitor
}

func NewAuditPathVisitor(decorated OpVisitor) *AuditPathVisitor {
	return &AuditPathVisitor{
		OpVisitor: decorated,
		auditPath: make(navigation.AuditPath),
	}
}

func (v AuditPathVisitor) Result() navigation.AuditPath {
	return v.auditPath
}

func (v *AuditPathVisitor) VisitCollectOp(op CollectOp) hashing.Digest {
	hash := op.Operation.Accept(v)
	v.auditPath[op.Position().FixedBytes()] = hash
	return hash
}
