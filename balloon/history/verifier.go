package history

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"verifiabledata/log"
)

type Proof struct {
	auditPath      []Node
	leafHasher     LeafHasher
	interiorHasher InteriorHasher
	log            log.Logger
}

func NewProof(auditPath []Node, lh LeafHasher, ih InteriorHasher) *Proof {
	return &Proof{
		auditPath,
		lh,
		ih,
		log.NewError(os.Stdout, "HistoryProof", log.Ldate|log.Ltime|log.Lmicroseconds|log.Llongfile),
	}
}

func (p *Proof) Verify(expectedDigest []byte, key []byte, version uint) bool {
	p.log.Debug("\nVerifying commitment %v with auditpath %v, key %v and version %v\n", expectedDigest, p.auditPath, key, version)
	depth := p.getDepth(version)
	pathMap := make(map[string][]byte)
	for _, n := range p.auditPath {
		pathMap[pathKey(n.Index, n.Layer)] = n.Digest
	}
	recomputed := p.rootHash(pathMap, key, 0, depth, version)
	return bytes.Equal(expectedDigest, recomputed)
}

func (p *Proof) getDepth(index uint) uint {
	return uint(math.Ceil(math.Log2(float64(index + 1))))
}

func pathKey(index, layer uint) string {
	return fmt.Sprintf("%d|%d", index, layer)
}

func (p *Proof) rootHash(auditPath map[string][]byte, key []byte, index, layer, version uint) []byte {
	var digest []byte
	p.log.Debug("Calling rootHash with auditpath %v, key %v, index %v, layer %v and version %v\n", auditPath, key, index, layer, version)
	//if version >= index+pow(2, layer)-1 {
	p.log.Debug("Extracting hash from audit path at index %v and layer %v :=> ", index, layer)
	digest, ok := auditPath[pathKey(index, layer)]
	if ok {
		p.log.Debug("found")
		return digest
	}
	p.log.Info("not found")
	//	}

	switch {
	// we are at a leaf: A_v(i,0)
	case layer == 0 && version >= index:
		p.log.Debug("Hashing leaf with key %v\n", key)
		digest = p.leafHasher(Zero, key)
		break
		// A_v(i,r) with one empty children
	case version < index+pow(2, layer-1):
		hash := p.rootHash(auditPath, key, index, layer-1, version)
		p.log.Debug("Hashing node with empty at index %v and layer %v\n", index, layer)
		digest = p.leafHasher(One, hash)
		break
		// A_v(i,r) with no non-empty children
	case version >= index+pow(2, layer-1):
		hash1 := p.rootHash(auditPath, key, index, layer-1, version)
		hash2 := p.rootHash(auditPath, key, index+pow(2, layer-1), layer-1, version)
		p.log.Debug("Hashing node at index %v and layer %v\n", index, layer)
		digest = p.interiorHasher(One, hash1, hash2)
		break
	}

	return digest
}
