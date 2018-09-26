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
	auditPath AuditPath

	*ComputeHashVisitor
}

func NewAuditPathVisitor(decorated *ComputeHashVisitor) *AuditPathVisitor {
	return &AuditPathVisitor{ComputeHashVisitor: decorated, auditPath: make(AuditPath)}
}

func (v AuditPathVisitor) Result() AuditPath {
	return v.auditPath
}

func (v *AuditPathVisitor) VisitCollectable(pos Position, result interface{}) interface{} {
	digest := v.ComputeHashVisitor.VisitCollectable(pos, result).(hashing.Digest)
	v.auditPath[pos.StringId()] = digest
	return digest
}
