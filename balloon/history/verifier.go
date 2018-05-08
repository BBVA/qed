package history

import (
	"bytes"
	"fmt"
	"math"

	"verifiabledata/balloon/hashing"
	"verifiabledata/log"
)

type Proof struct {
	auditPath      []Node
	index          uint64
	leafHasher     leafHasher
	interiorHasher interiorHasher
}

func NewProof(auditPath []Node, index uint64, hasher hashing.Hasher) *Proof {
	return &Proof{
		auditPath,
		index,
		leafHasherF(hasher),
		interiorHasherF(hasher),
	}
}

func (p Proof) String() string {
	return fmt.Sprintf(`{"auditPathLen": "%d"}`, len(p.auditPath))
}

func (p *Proof) Verify(expectedDigest []byte, key []byte, version uint64) bool {
	log.Debugf("\nVerifying commitment %v with auditpath %v, key %x and version %v\n", expectedDigest, p.auditPath, key, version)
	depth := p.getDepth(version)
	pathMap := make(map[string][]byte)
	for _, n := range p.auditPath {
		pathMap[pathKey(n.Index, n.Layer)] = n.Digest
	}
	pathMap[pathKey(p.index, 0)] = p.leafHasher(Zero, key)
	recomputed := p.rootHash(pathMap, key, 0, depth, version)
	return bytes.Equal(expectedDigest, recomputed)
}

func (p *Proof) getDepth(index uint64) uint64 {
	return uint64(math.Ceil(math.Log2(float64(index + 1))))
}

func pathKey(index, layer uint64) string {
	return fmt.Sprintf("%d|%d", index, layer)
}

func (p *Proof) rootHash(auditPath map[string][]byte, key []byte, index, layer, version uint64) []byte {
	var digest []byte
	log.Debugf("Calling rootHash with auditpath %v, key %x, index %v, layer %v and version %v\n", auditPath, key, index, layer, version)

	digest, ok := auditPath[pathKey(index, layer)]
	if ok {
		log.Debug("found")
		return digest
	}

	switch {
	// we are at a leaf: A_v(i,0)
	case layer == 0 && version >= index:
		log.Debugf("Hashing leaf with key %x\n", key)
		digest = p.leafHasher(Zero, key)
		break
		// A_v(i,r) with one empty children
	case version < index+pow(2, layer-1):
		hash := p.rootHash(auditPath, key, index, layer-1, version)
		log.Debugf("Hashing node with empty at index %v and layer %v\n", index, layer)
		digest = p.leafHasher(One, hash)
		break
		// A_v(i,r) with no non-empty children
	case version >= index+pow(2, layer-1):
		hash1 := p.rootHash(auditPath, key, index, layer-1, version)
		hash2 := p.rootHash(auditPath, key, index+pow(2, layer-1), layer-1, version)
		log.Debugf("Hashing node at index %v and layer %v\n", index, layer)
		digest = p.interiorHasher(One, hash1, hash2)
		break
	}

	return digest
}
