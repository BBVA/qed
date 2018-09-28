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

package visitor

import (
	"github.com/bbva/qed/balloon/navigator"
	"github.com/bbva/qed/hashing"
)

type ComputeHashVisitor struct {
	hasher hashing.Hasher
}

func NewComputeHashVisitor(hasher hashing.Hasher) *ComputeHashVisitor {
	return &ComputeHashVisitor{hasher}
}

func (v *ComputeHashVisitor) VisitRoot(pos navigator.Position, leftResult, rightResult interface{}) interface{} {
	return v.hasher.Salted(pos.Bytes(), leftResult.(hashing.Digest), rightResult.(hashing.Digest))
}

func (v *ComputeHashVisitor) VisitNode(pos navigator.Position, leftResult, rightResult interface{}) interface{} {
	return v.hasher.Salted(pos.Bytes(), leftResult.(hashing.Digest), rightResult.(hashing.Digest))
}

func (v *ComputeHashVisitor) VisitPartialNode(pos navigator.Position, leftResult interface{}) interface{} {
	return v.hasher.Salted(pos.Bytes(), leftResult.(hashing.Digest))
}

func (v *ComputeHashVisitor) VisitLeaf(pos navigator.Position, value []byte) interface{} {
	return v.hasher.Salted(pos.Bytes(), value)
}

func (v *ComputeHashVisitor) VisitCached(pos navigator.Position, cachedDigest hashing.Digest) interface{} {
	return cachedDigest
}

func (v *ComputeHashVisitor) VisitCollectable(pos navigator.Position, result interface{}) interface{} {
	return result
}

func (v *ComputeHashVisitor) VisitCacheable(pos navigator.Position, result interface{}) interface{} {
	return result
}
