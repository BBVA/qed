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

// Package balloon implements the tree interface to interact with both hyper
// and history trees.
package proof

import (
	"bytes"

	"github.com/bbva/qed/balloon/position"
	"github.com/bbva/qed/hashing"
)

func NewProof(root position.Position, ap AuditPath, hasher hashing.Hasher) *Proof {
	return &Proof{
		root,
		ap,
		hashing.InteriorHasherF(hasher),
		hashing.LeafHasherF(hasher),
	}
}

type AuditPath map[string][]byte

type Proof struct {
	root      position.Position
	auditPath AuditPath
	ih        hashing.InteriorHasher
	lh        hashing.LeafHasher
}

func (p Proof) AuditPath() map[string][]byte {
	return p.auditPath
}

func (p Proof) LeafHash(id []byte, digest []byte) []byte {
	return p.lh(id, digest)
}

func (p Proof) InteriorHash(id []byte, h1 []byte, h2 []byte) []byte {
	return p.ih(id, h1, h2)
}

func computeHash(p Proof, pos position.Position, key, value []byte, auditPath map[string][]byte) []byte {

	var digest []byte
	direction := pos.Direction(key)

	switch {
	case direction == position.Halt && pos.IsLeaf():
		digest = auditPath[pos.StringId()]
	case direction == position.Left:
		digest = p.InteriorHash(pos.Id(), computeHash(p, pos.Left(), key, value, auditPath), auditPath[pos.Right().StringId()])
	case direction == position.Right:
		digest = p.InteriorHash(pos.Id(), auditPath[pos.Left().StringId()], computeHash(p, pos.Right(), key, value, auditPath))
	}

	return digest
}

// NewRootHyperPosition(p.treeId, p.numBits, 0)
// NewRootHistoryPosition(p.treeId, p.version, p.version)
func (p Proof) Verify(expectedDigest []byte, key, value []byte) bool {
	ap := p.AuditPath()
	if len(ap) == 0 {
		return false
	}
	recomputed := computeHash(p, p.root, key, value, ap)
	return bytes.Equal(expectedDigest, recomputed)
}
