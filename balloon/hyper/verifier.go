package hyper

import (
	"bytes"
	"verifiabledata/balloon/hashing"
)

type Verifier struct {
	id           []byte
	digestLength int
	leafHash     LeafHasher
	interiorHash InteriorHasher
}

func NewVerifier(id string, hasher hashing.Hasher, leafHash LeafHasher, interiorHash InteriorHasher) *Verifier {
	digestLength := len(hasher([]byte("x"))) * 8
	return &Verifier{
		[]byte(id),
		digestLength,
		leafHash,
		interiorHash,
	}
}

func (v *Verifier) Verify(expectedDigest []byte, auditPath [][]byte, key, value []byte) (bool, []byte) {
	recomputed := v.rootHash(auditPath, rootPosition(v.digestLength), key, value)
	return bytes.Equal(expectedDigest, recomputed), recomputed
}

func (v *Verifier) rootHash(auditPath [][]byte, pos *Position, key, value []byte) []byte {
	if pos.height == 0 {
		return v.leafHash(v.id, value, pos.base)
	}
	if !bitIsSet(key, v.digestLength-pos.height) { // if k_j == 0
		left := v.rootHash(auditPath, pos.left(), key, value)
		right := auditPath[pos.height]
		next := pos.right()
		return v.interiorHash(left, right, next.base, next.heightBytes())
	}
	left := auditPath[pos.height]
	right := v.rootHash(auditPath, pos.right(), key, value)
	next := pos.left()
	return v.interiorHash(left, right, next.base, next.heightBytes())
}
