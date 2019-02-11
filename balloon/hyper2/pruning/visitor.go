package pruning

import (
	"github.com/bbva/qed/balloon/cache"
	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/storage"
)

type InsertVisitor struct {
	cache         cache.ModifiableCache
	defaultHashes []hashing.Digest
	hasher        hashing.Hasher

	mutations []*storage.Mutation
}

func NewInsertVisitor(hasher hashing.Hasher, cache cache.ModifiableCache, defaultHashes []hashing.Digest) *InsertVisitor {
	return &InsertVisitor{
		cache:         cache,
		defaultHashes: defaultHashes,
		hasher:        hasher,
		mutations:     make([]*storage.Mutation, 0),
	}
}

func (v InsertVisitor) Result() []*storage.Mutation {
	return v.mutations
}

func (v *InsertVisitor) VisitShortcutLeafOp(op ShortcutLeafOp) hashing.Digest {
	hash := v.hasher.Salted(op.Position().Bytes(), op.Value)
	op.Batch.AddLeafAt(op.Idx, hash, op.Key, op.Value)
	return hash
}

func (v *InsertVisitor) VisitLeafOp(op LeafOp) hashing.Digest {
	hash := op.Operation.Accept(v)
	op.Batch.AddHashAt(op.Idx, hash)
	return hash
}

func (v *InsertVisitor) VisitInnerHashOp(op InnerHashOp) hashing.Digest {
	leftHash := op.Left.Accept(v)
	rightHash := op.Right.Accept(v)
	hash := v.hasher.Salted(op.Position().Bytes(), leftHash, rightHash)
	op.Batch.AddHashAt(op.Idx, hash)
	return hash
}

func (v *InsertVisitor) VisitGetDefaultOp(op GetDefaultOp) hashing.Digest {
	return v.defaultHashes[op.Position().Height]
}

func (v *InsertVisitor) VisitUseProvidedOp(op UseProvidedOp) hashing.Digest {
	return op.Batch.GetElementAt(op.Idx)
}

func (v *InsertVisitor) VisitPutBatchOp(op PutBatchOp) hashing.Digest {
	hash := op.Operation.Accept(v)
	v.cache.Put(op.Position().Bytes(), op.Batch.Serialize())
	return hash
}

func (v *InsertVisitor) VisitMutateBatchOp(op MutateBatchOp) hashing.Digest {
	hash := op.Operation.Accept(v)
	v.mutations = append(v.mutations, storage.NewMutation(storage.HyperCachePrefix, op.Position().Bytes(), op.Batch.Serialize()))
	return hash
}

func (v *InsertVisitor) VisitCollectOp(op CollectOp) hashing.Digest {
	return op.Operation.Accept(v)
}
