package pruning

import (
	"fmt"

	"github.com/bbva/qed/balloon/cache"
	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/storage"
)

type InsertVisitor struct {
	hasher        hashing.Hasher
	cache         cache.ModifiableCache
	storagePrefix byte // TODO shall i remove this?

	mutations []*storage.Mutation
}

func NewInsertVisitor(hasher hashing.Hasher, cache cache.ModifiableCache, storagePrefix byte) *InsertVisitor {
	return &InsertVisitor{
		hasher:        hasher,
		cache:         cache,
		storagePrefix: storagePrefix,
		mutations:     make([]*storage.Mutation, 0),
	}
}

func (v InsertVisitor) Result() []*storage.Mutation {
	return v.mutations
}

func (v *InsertVisitor) VisitLeafHashOp(op LeafHashOp) hashing.Digest {
	return v.hasher.Salted(op.Position().Bytes(), op.Value)
}

func (v *InsertVisitor) VisitInnerHashOp(op InnerHashOp) hashing.Digest {
	leftHash := op.Left.Accept(v)
	rightHash := op.Right.Accept(v)
	return v.hasher.Salted(op.Position().Bytes(), leftHash, rightHash)
}

func (v *InsertVisitor) VisitPartialInnerHashOp(op PartialInnerHashOp) hashing.Digest {
	leftHash := op.Left.Accept(v)
	return v.hasher.Salted(op.Position().Bytes(), leftHash)
}

func (v *InsertVisitor) VisitGetCacheOp(op GetCacheOp) hashing.Digest {
	hash, ok := v.cache.Get(op.Position().Bytes())
	if !ok {
		panic(fmt.Sprintf("Oops, something went wrong. There should be a cached element at position %v", op.Position()))
	}
	return hash
}

func (v *InsertVisitor) VisitPutCacheOp(op PutCacheOp) hashing.Digest {
	hash := op.Operation.Accept(v)
	v.cache.Put(op.Position().Bytes(), hash)
	return hash
}

func (v *InsertVisitor) VisitMutateOp(op MutateOp) hashing.Digest {
	hash := op.Operation.Accept(v)
	v.mutations = append(v.mutations, storage.NewMutation(v.storagePrefix, op.Position().Bytes(), hash))
	return hash
}

func (v *InsertVisitor) VisitCollectOp(op CollectOp) hashing.Digest {
	return op.Operation.Accept(v)
}
