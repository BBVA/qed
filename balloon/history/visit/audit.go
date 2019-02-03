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

type AuditPathVisitor struct {
	auditPath navigation.AuditPath
	PostOrderVisitor
}

func NewAuditPathVisitor(decorated PostOrderVisitor) *AuditPathVisitor {
	return &AuditPathVisitor{
		PostOrderVisitor: decorated,
		auditPath:        make(navigation.AuditPath),
	}
}

func (v AuditPathVisitor) Result() navigation.AuditPath {
	return v.auditPath
}

func (v *AuditPathVisitor) VisitCollectable(pos *navigation.Position, result hashing.Digest) hashing.Digest {
	hash := v.PostOrderVisitor.VisitCollectable(pos, result)
	v.auditPath[pos.FixedBytes()] = hash
	return hash
}
