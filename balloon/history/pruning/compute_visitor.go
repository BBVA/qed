package pruning

import (
	"fmt"

	"github.com/bbva/qed/balloon/cache"
	"github.com/bbva/qed/hashing"
)

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
		panic(fmt.Sprintf("Oops, something went wrong. There should be a cached element at position %v", op.Position()))
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
