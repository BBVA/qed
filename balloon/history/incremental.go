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

package history

import (
	"bytes"
	"encoding/binary"

	"github.com/bbva/qed/balloon/position"
	"github.com/bbva/qed/balloon/proof"
	"github.com/bbva/qed/hashing"
)

type IncrementalProof struct {
	start, end   uint64
	auditPath    proof.AuditPath
	interiorHash hashing.InteriorHasher
	leafHash     hashing.LeafHasher
}

func NewIncrementalProof(start, end uint64, ap proof.AuditPath, ih hashing.InteriorHasher, lh hashing.LeafHasher) *IncrementalProof {
	return &IncrementalProof{start, end, ap, ih, lh}
}

func (p IncrementalProof) AuditPath() proof.AuditPath {
	return p.auditPath
}

func (p IncrementalProof) Verify(startDigest, endDigest []byte) bool {
	startRoot := NewRootPosition(p.start)
	endRoot := NewRootPosition(p.end)

	startBytes := make([]byte, 8)
	endBytes := make([]byte, 8)
	binary.PutUvarint(startBytes, p.start)
	binary.PutUvarint(endBytes, p.end)

	startRecomputed := p.computeStartHash(startRoot, p.auditPath, startBytes)
	endRecomputed := p.computeEndHash(endRoot, p.auditPath, startBytes, endBytes)

	return bytes.Equal(startRecomputed, startDigest) && bytes.Equal(endRecomputed, endDigest)
}

func (p IncrementalProof) computeStartHash(pos position.Position, ap proof.AuditPath, index []byte) []byte {
	var digest []byte
	direction := pos.Direction(index)

	switch {
	case direction == position.Halt && pos.IsLeaf():
		digest = ap[pos.StringId()]
	case direction == position.Left:
		left := p.computeStartHash(pos.Left(), ap, index)
		digest = p.leafHash(pos.Id(), left)
	case direction == position.Right:
		left := ap[pos.Left().StringId()]
		right := p.computeStartHash(pos.Right(), ap, index)
		digest = p.interiorHash(pos.Id(), left, right)
	}

	return digest
}

func (p IncrementalProof) computeEndLeftHash(pos position.Position, ap proof.AuditPath, index []byte) []byte {
	var digest []byte
	direction := pos.Direction(index)

	switch {
	case direction == position.Halt && pos.IsLeaf():
		digest = ap[pos.StringId()]
	case direction == position.Left:
		left := p.computeEndLeftHash(pos.Left(), ap, index)
		if pos.Height() == 1 {
			digest = p.leafHash(pos.Id(), left)
		} else {
			right := ap[pos.Right().StringId()]
			digest = p.interiorHash(pos.Id(), left, right)
		}
	case direction == position.Right:
		left := ap[pos.Left().StringId()]
		right := p.computeEndLeftHash(pos.Right(), ap, index)
		digest = p.interiorHash(pos.Id(), left, right)
	}

	return digest
}

func (p IncrementalProof) computeEndHash(pos position.Position, ap proof.AuditPath, start, end []byte) []byte {
	var ok bool
	var left, right []byte
	var digest []byte
	direction := pos.Direction(end)

	switch {
	case direction == position.Halt && pos.IsLeaf():
		digest = ap[pos.StringId()]
	case direction == position.Left:
		left = p.computeEndHash(pos.Left(), ap, start, end)
		digest = p.leafHash(pos.Id(), left)
	case direction == position.Right:
		left, ok = ap[pos.Left().StringId()]
		if !ok {
			startIndex := binary.LittleEndian.Uint64(start)
			if startIndex%2 == 0 {
				nextIndex := make([]byte, 8)
				binary.LittleEndian.PutUint64(nextIndex, startIndex+1)
				left = p.computeEndLeftHash(pos.Left(), ap, nextIndex)
			} else {
				left = p.computeEndLeftHash(pos.Left(), ap, start)
			}
		}
		right = p.computeEndHash(pos.Right(), ap, start, end)
		digest = p.interiorHash(pos.Id(), left, right)
	}

	return digest
}
