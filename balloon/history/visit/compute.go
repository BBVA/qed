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
)

type ComputeHashVisitor struct {
	hasher hashing.Hasher
}

func NewComputeHashVisitor(hasher hashing.Hasher) *ComputeHashVisitor {
	return &ComputeHashVisitor{hasher}
}

func (v *ComputeHashVisitor) VisitNode(pos *navigation.Position, leftResult, rightResult hashing.Digest) hashing.Digest {
	return v.hasher.Salted(pos.Bytes(), leftResult, rightResult)
}

func (v *ComputeHashVisitor) VisitPartialNode(pos *navigation.Position, leftResult hashing.Digest) hashing.Digest {
	return v.hasher.Salted(pos.Bytes(), leftResult)
}

func (v *ComputeHashVisitor) VisitLeaf(pos *navigation.Position, value []byte) hashing.Digest {
	return v.hasher.Salted(pos.Bytes(), value)
}

func (v *ComputeHashVisitor) VisitCached(pos *navigation.Position, cachedDigest hashing.Digest) hashing.Digest {
	return cachedDigest
}

func (v *ComputeHashVisitor) VisitMutable(pos *navigation.Position, result hashing.Digest) hashing.Digest {
	return result
}

func (v *ComputeHashVisitor) VisitCacheable(pos *navigation.Position, result hashing.Digest) hashing.Digest {
	return result
}

func (v *ComputeHashVisitor) VisitCollectable(pos *navigation.Position, result hashing.Digest) hashing.Digest {
	return result
}
