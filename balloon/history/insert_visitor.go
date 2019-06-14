/*
   Copyright 2018-2019 Banco Bilbao Vizcaya Argentaria, S.A.

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

package history

import (
	"fmt"

	"github.com/bbva/qed/balloon/cache"
	"github.com/bbva/qed/crypto/hashing"
	"github.com/bbva/qed/storage"
)

type insertVisitor struct {
	hasher       hashing.Hasher
	cache        cache.ModifiableCache
	storageTable storage.Table // TODO shall i remove this?

	mutations []*storage.Mutation
}

func newInsertVisitor(hasher hashing.Hasher, cache cache.ModifiableCache, storageTable storage.Table) *insertVisitor {
	return &insertVisitor{
		hasher:       hasher,
		cache:        cache,
		storageTable: storageTable,
		mutations:    make([]*storage.Mutation, 0),
	}
}

func (v insertVisitor) Result() []*storage.Mutation {
	return v.mutations
}

func (v *insertVisitor) VisitLeafHashOp(op leafHashOp) hashing.Digest {
	return v.hasher.Salted(op.Position().Bytes(), op.Value)
}

func (v *insertVisitor) VisitInnerHashOp(op innerHashOp) hashing.Digest {
	leftHash := op.Left.Accept(v)
	rightHash := op.Right.Accept(v)
	return v.hasher.Salted(op.Position().Bytes(), leftHash, rightHash)
}

func (v *insertVisitor) VisitPartialInnerHashOp(op partialInnerHashOp) hashing.Digest {
	leftHash := op.Left.Accept(v)
	return v.hasher.Salted(op.Position().Bytes(), leftHash)
}

func (v *insertVisitor) VisitGetCacheOp(op getCacheOp) hashing.Digest {
	hash, ok := v.cache.Get(op.Position().Bytes())
	if !ok {
		panic(fmt.Sprintf("Oops, something went wrong. There should be a cached element at position %v", op.Position()))
	}
	return hash
}

func (v *insertVisitor) VisitPutCacheOp(op putCacheOp) hashing.Digest {
	hash := op.operation.Accept(v)
	v.cache.Put(op.Position().Bytes(), hash)
	return hash
}

func (v *insertVisitor) VisitMutateOp(op mutateOp) hashing.Digest {
	hash := op.operation.Accept(v)
	v.mutations = append(v.mutations, storage.NewMutation(v.storageTable, op.Position().Bytes(), hash))
	return hash
}

func (v *insertVisitor) VisitCollectOp(op collectOp) hashing.Digest {
	return op.operation.Accept(v)
}
