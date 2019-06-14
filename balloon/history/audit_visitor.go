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
)

type auditPathVisitor struct {
	hasher hashing.Hasher
	cache  cache.Cache

	auditPath AuditPath
}

func newAuditPathVisitor(hasher hashing.Hasher, cache cache.Cache) *auditPathVisitor {
	return &auditPathVisitor{
		hasher:    hasher,
		cache:     cache,
		auditPath: make(AuditPath),
	}
}

func (v auditPathVisitor) Result() AuditPath {
	return v.auditPath
}

func (v *auditPathVisitor) VisitLeafHashOp(op leafHashOp) hashing.Digest {
	return v.hasher.Salted(op.Position().Bytes(), op.Value)
}

func (v *auditPathVisitor) VisitInnerHashOp(op innerHashOp) hashing.Digest {
	leftHash := op.Left.Accept(v)
	rightHash := op.Right.Accept(v)
	return v.hasher.Salted(op.Position().Bytes(), leftHash, rightHash)
}

func (v *auditPathVisitor) VisitPartialInnerHashOp(op partialInnerHashOp) hashing.Digest {
	leftHash := op.Left.Accept(v)
	return v.hasher.Salted(op.Position().Bytes(), leftHash)
}

func (v *auditPathVisitor) VisitGetCacheOp(op getCacheOp) hashing.Digest {
	hash, ok := v.cache.Get(op.Position().Bytes())
	if !ok {
		panic(fmt.Sprintf("Oops, something went wrong. There should be a cached element at position %v", op.Position()))
	}
	return hash
}

func (v *auditPathVisitor) VisitPutCacheOp(op putCacheOp) hashing.Digest {
	return op.operation.Accept(v)
}

func (v *auditPathVisitor) VisitMutateOp(op mutateOp) hashing.Digest {
	return op.operation.Accept(v)
}

func (v *auditPathVisitor) VisitCollectOp(op collectOp) hashing.Digest {
	hash := op.operation.Accept(v)
	v.auditPath[op.Position().FixedBytes()] = hash
	return hash
}
