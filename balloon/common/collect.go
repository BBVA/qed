package common

import (
	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/storage"
)

type CollectMutationsVisitor struct {
	decorated     PostOrderVisitor
	storagePrefix byte
	mutations     []*storage.Mutation
}

func NewCollectMutationsVisitor(decorated PostOrderVisitor, storagePrefix byte) *CollectMutationsVisitor {
	return &CollectMutationsVisitor{
		decorated:     decorated,
		storagePrefix: storagePrefix,
		mutations:     make([]*storage.Mutation, 0),
	}
}

func (v CollectMutationsVisitor) Result() []*storage.Mutation {
	return v.mutations
}

func (v *CollectMutationsVisitor) VisitRoot(pos Position, leftResult, rightResult interface{}) interface{} {
	// by-pass
	return v.decorated.VisitRoot(pos, leftResult, rightResult)
}

func (v *CollectMutationsVisitor) VisitNode(pos Position, leftResult, rightResult interface{}) interface{} {
	// by-pass
	return v.decorated.VisitNode(pos, leftResult, rightResult)
}

func (v *CollectMutationsVisitor) VisitPartialNode(pos Position, leftResult interface{}) interface{} {
	// by-pass
	return v.decorated.VisitPartialNode(pos, leftResult)
}

func (v *CollectMutationsVisitor) VisitLeaf(pos Position, eventDigest []byte) interface{} {
	// ignore. target leafs not included in path
	return v.decorated.VisitLeaf(pos, eventDigest)
}

func (v *CollectMutationsVisitor) VisitCached(pos Position, cachedDigest hashing.Digest) interface{} {
	// by-pass
	return v.decorated.VisitCached(pos, cachedDigest)
}

func (v *CollectMutationsVisitor) VisitCollectable(pos Position, result interface{}) interface{} {
	value := v.decorated.VisitCollectable(pos, result).(hashing.Digest)
	v.mutations = append(v.mutations, storage.NewMutation(v.storagePrefix, pos.Bytes(), value))
	return result
}

func (v *CollectMutationsVisitor) VisitCacheable(pos Position, result interface{}) interface{} {
	// by-pass
	return v.decorated.VisitCacheable(pos, result)
}
