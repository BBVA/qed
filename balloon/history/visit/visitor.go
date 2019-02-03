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

type PostOrderVisitor interface {
	VisitNode(pos *navigation.Position, leftResult, rightResult hashing.Digest) hashing.Digest
	VisitPartialNode(pos *navigation.Position, leftResult hashing.Digest) hashing.Digest
	VisitLeaf(pos *navigation.Position, value []byte) hashing.Digest
	VisitCached(pos *navigation.Position, cachedDigest hashing.Digest) hashing.Digest
	VisitMutable(pos *navigation.Position, result hashing.Digest) hashing.Digest
	VisitCollectable(pos *navigation.Position, result hashing.Digest) hashing.Digest
	VisitCacheable(pos *navigation.Position, result hashing.Digest) hashing.Digest
}

type PreOrderVisitor interface {
	VisitNode(pos *navigation.Position)
	VisitPartialNode(pos *navigation.Position)
	VisitLeaf(pos *navigation.Position, value []byte)
	VisitCached(pos *navigation.Position, cachedDigest hashing.Digest)
	VisitMutable(pos *navigation.Position)
	VisitCollectable(pos *navigation.Position)
	VisitCacheable(pos *navigation.Position)
}
