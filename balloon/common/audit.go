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

package common

import (
	"github.com/bbva/qed/hashing"
)

type AuditPath map[string]hashing.Digest

func (p AuditPath) Get(pos Position) (hashing.Digest, bool) {
	digest, ok := p[pos.StringId()]
	return digest, ok
}

type Verifiable interface {
	Verify(key []byte, expectedDigest hashing.Digest) bool
	AuditPath() AuditPath
}

type FakeVerifiable struct {
	result bool
}

func NewFakeVerifiable(result bool) *FakeVerifiable {
	return &FakeVerifiable{result}
}

func (f FakeVerifiable) Verify(key []byte, commitment hashing.Digest) bool {
	return f.result
}

func (f FakeVerifiable) AuditPath() AuditPath {
	return make(AuditPath)
}

type AuditPathVisitor struct {
	decorated *ComputeHashVisitor
	auditPath AuditPath
}

func NewAuditPathVisitor(decorated *ComputeHashVisitor) *AuditPathVisitor {
	return &AuditPathVisitor{decorated, make(AuditPath)}
}

func (v AuditPathVisitor) Result() AuditPath {
	return v.auditPath
}

func (v *AuditPathVisitor) VisitRoot(pos Position, leftResult, rightResult interface{}) interface{} {
	// by-pass
	return v.decorated.VisitRoot(pos, leftResult, rightResult)
}

func (v *AuditPathVisitor) VisitNode(pos Position, leftResult, rightResult interface{}) interface{} {
	// by-pass
	return v.decorated.VisitNode(pos, leftResult, rightResult)
}

func (v *AuditPathVisitor) VisitPartialNode(pos Position, leftResult interface{}) interface{} {
	// by-pass
	return v.decorated.VisitPartialNode(pos, leftResult)
}

func (v *AuditPathVisitor) VisitLeaf(pos Position, eventDigest []byte) interface{} {
	// ignore. target leafs not included in path
	return v.decorated.VisitLeaf(pos, eventDigest)
}

func (v *AuditPathVisitor) VisitCached(pos Position, cachedDigest hashing.Digest) interface{} {
	// by-pass
	return v.decorated.VisitCached(pos, cachedDigest)
}

func (v *AuditPathVisitor) VisitCollectable(pos Position, result interface{}) interface{} {
	digest := v.decorated.VisitCollectable(pos, result).(hashing.Digest)
	v.auditPath[pos.StringId()] = digest
	return digest
}

func (v *AuditPathVisitor) VisitCacheable(pos Position, result interface{}) interface{} {
	// by-pass
	return v.decorated.VisitCacheable(pos, result)
}
