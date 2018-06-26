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

func (p IncrementalProof) Verify(startDigest, endDigest []byte) bool {
	startRoot := NewRootPosition(p.start)
	endRoot := NewRootPosition(p.end)

	startBytes := make([]byte, 8)
	endBytes := make([]byte, 8)
	binary.PutUvarint(startBytes, p.start)
	binary.PutUvarint(endBytes, p.end)

	startRecomputed := p.computeStartHash(startRoot, p.auditPath, startBytes)
	endRecomputed := p.computeEndHash(endRoot, p.auditPath, endBytes)

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
		if pos.Height() == 1 {
			digest = p.leafHash(pos.Id(), left)
		} else {
			right := ap[pos.Right().StringId()]
			digest = p.interiorHash(pos.Id(), left, right)
		}
	case direction == position.Right:
		left := ap[pos.Left().StringId()]
		right := p.computeStartHash(pos.Right(), ap, index)
		digest = p.interiorHash(pos.Id(), left, right)
	}

	return digest
}

func (p IncrementalProof) computeEndHash(pos position.Position, ap proof.AuditPath, index []byte) []byte {
	var ok bool
	var left, right []byte
	var digest []byte
	direction := pos.Direction(index)

	switch {
	case direction == position.Halt && pos.IsLeaf():
		digest = ap[pos.StringId()]
	case direction == position.Left:
		right, ok = ap[pos.Right().StringId()]
		if !ok {
			right = p.computeEndHash(pos.Right(), ap, index)
		}
		left = p.computeEndHash(pos.Left(), ap, index)
		digest = p.interiorHash(pos.Id(), left, right)
	case direction == position.Right:
		left, ok = ap[pos.Left().StringId()]
		if !ok {
			left = p.computeEndHash(pos.Left(), ap, index)
		}
		right = p.computeEndHash(pos.Right(), ap, index)
		digest = p.interiorHash(pos.Id(), left, right)
	}

	return digest
}
