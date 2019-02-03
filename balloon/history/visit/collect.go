/*
   Copyright 2018 Banco Bilbao Vizcaya Argentaria, S.A.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package visit

import (
	"github.com/bbva/qed/balloon/history/navigation"
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

func (v *CollectMutationsVisitor) VisitMutable(pos *navigation.Position, result hashing.Digest) hashing.Digest {
	hash := v.PostOrderVisitor.VisitMutable(pos, result)
	v.mutations = append(v.mutations, storage.NewMutation(v.storagePrefix, pos.Bytes(), hash))
	return hash
}
