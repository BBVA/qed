package pruning

import (
	"github.com/bbva/qed/balloon/cache"
	"github.com/bbva/qed/balloon/hyper2/navigation"
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
	// fmt.Printf("Shortcut hash(%v): %x\n", op.Position(), hash)
	op.Batch.AddLeafAt(op.Idx, hash, op.Key, op.Value)
	return hash
}

func (v *InsertVisitor) VisitLeafOp(op LeafOp) hashing.Digest {
	hash := op.Operation.Accept(v)
	// fmt.Printf("Leaf hash(%v): %x\n", op.Position(), hash)
	op.Batch.AddHashAt(op.Idx, hash)
	return hash
}

func (v *InsertVisitor) VisitInnerHashOp(op InnerHashOp) hashing.Digest {
	leftHash := op.Left.Accept(v)
	rightHash := op.Right.Accept(v)
	hash := v.hasher.Salted(op.Position().Bytes(), leftHash, rightHash)
	// fmt.Printf("Inner hash(%v): %x, %x => %x\n", op.Position(), leftHash, rightHash, hash)
	op.Batch.AddHashAt(op.Idx, hash)
	return hash
}

func (v *InsertVisitor) VisitGetDefaultOp(op GetDefaultOp) hashing.Digest {
	hash := v.defaultHashes[op.Position().Height]
	// fmt.Printf("Default hash(%v): %x\n", op.Position(), hash)
	return hash
}

func (v *InsertVisitor) VisitUseProvidedOp(op UseProvidedOp) hashing.Digest {
	hash := op.Batch.GetElementAt(op.Idx)
	// fmt.Printf("Provided hash(%v): %x\n", op.Position(), hash)
	return hash
}

func (v *InsertVisitor) VisitPutBatchOp(op PutBatchOp) hashing.Digest {
	hash := op.Operation.Accept(v)
	v.cache.Put(op.Position().Bytes(), op.Batch.Serialize())
	// fmt.Printf("Put cache hash(%v) [%d]: %x\n", op.Position(), op.Batch.Serialize(), hash)
	return hash
}

func (v *InsertVisitor) VisitMutateBatchOp(op MutateBatchOp) hashing.Digest {
	hash := op.Operation.Accept(v)
	v.mutations = append(v.mutations, storage.NewMutation(storage.HyperCachePrefix, op.Position().Bytes(), op.Batch.Serialize()))
	// fmt.Printf("Mutate hash(%v) [%d]: %x\n", op.Position(), op.Batch.Serialize(), hash)
	return hash
}

func (v *InsertVisitor) VisitCollectOp(op CollectOp) hashing.Digest {
	return op.Operation.Accept(v)
}

type AuditPathVisitor struct {
	hasher        hashing.Hasher
	defaultHashes []hashing.Digest
	auditPath     navigation.AuditPath
}

func NewAuditPathVisitor(hasher hashing.Hasher, defaultHashes []hashing.Digest) *AuditPathVisitor {
	return &AuditPathVisitor{
		hasher:        hasher,
		defaultHashes: defaultHashes,
		auditPath:     navigation.NewAuditPath(),
	}
}

func (v AuditPathVisitor) Result() navigation.AuditPath {
	return v.auditPath
}

func (v *AuditPathVisitor) VisitShortcutLeafOp(op ShortcutLeafOp) hashing.Digest {
	return nil
}

func (v *AuditPathVisitor) VisitLeafOp(op LeafOp) hashing.Digest {
	return op.Operation.Accept(v)
}

func (v *AuditPathVisitor) VisitInnerHashOp(op InnerHashOp) hashing.Digest {
	op.Left.Accept(v)
	op.Right.Accept(v)
	return nil
}

func (v *AuditPathVisitor) VisitGetDefaultOp(op GetDefaultOp) hashing.Digest {
	return v.defaultHashes[op.Position().Height]
}

func (v *AuditPathVisitor) VisitUseProvidedOp(op UseProvidedOp) hashing.Digest {
	return op.Batch.GetElementAt(op.Idx)
}

func (v *AuditPathVisitor) VisitPutBatchOp(op PutBatchOp) hashing.Digest {
	return op.Operation.Accept(v)
}

func (v *AuditPathVisitor) VisitMutateBatchOp(op MutateBatchOp) hashing.Digest {
	return op.Operation.Accept(v)
}

func (v *AuditPathVisitor) VisitCollectOp(op CollectOp) hashing.Digest {
	hash := op.Operation.Accept(v)
	v.auditPath = append(v.auditPath, hash)
	return hash
}
