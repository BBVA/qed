package common

import (
	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/storage"
)

type CollectMutationsVisitor struct {
	storagePrefix byte
	mutations     []*storage.Mutation

	PostOrderVisitor
}

func NewCollectMutationsVisitor(decorated PostOrderVisitor, storagePrefix byte) *CollectMutationsVisitor {
	return &CollectMutationsVisitor{
		PostOrderVisitor: decorated,
		storagePrefix:    storagePrefix,
		mutations:        make([]*storage.Mutation, 0),
	}
}

func (v CollectMutationsVisitor) Result() []*storage.Mutation {
	return v.mutations
}

func (v *CollectMutationsVisitor) VisitCollectable(pos Position, result interface{}) interface{} {
	value := v.PostOrderVisitor.VisitCollectable(pos, result).(hashing.Digest)
	v.mutations = append(v.mutations, storage.NewMutation(v.storagePrefix, pos.Bytes(), value))
	return result
}
