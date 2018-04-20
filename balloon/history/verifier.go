package history

import (
	"bytes"
	"fmt"
	"math"
)

type Verifier struct {
	leafHasher     LeafHasher
	interiorHasher InteriorHasher
}

func NewVerifier(leafHasher LeafHasher, interiorHasher InteriorHasher) *Verifier {
	return &Verifier{
		leafHasher,
		interiorHasher,
	}
}

func (v *Verifier) Verify(expectedDigest []byte, auditPath []Node, key []byte, version uint) (bool, []byte) {
	fmt.Printf("\nVerifying commitment %v with auditpath %v, key %v and version %v\n", expectedDigest, auditPath, key, version)
	depth := v.getDepth(version)
	pathMap := make(map[string][]byte)
	for _, n := range auditPath {
		pathMap[pathKey(n.Index, n.Layer)] = n.Digest
	}
	recomputed := v.rootHash(pathMap, key, 0, depth, version)
	return bytes.Equal(expectedDigest, recomputed), recomputed
}

func (v *Verifier) getDepth(index uint) uint {
	return uint(math.Ceil(math.Log2(float64(index + 1))))
}

func pathKey(index, layer uint) string {
	return fmt.Sprintf("%s|%s", index, layer)
}

func (v *Verifier) rootHash(auditPath map[string][]byte, key []byte, index, layer, version uint) []byte {
	var digest []byte
	fmt.Printf("Calling rootHash with auditpath %v, key %v, index %v, layer %v and version %v\n", auditPath, key, index, layer, version)
	//if version >= index+pow(2, layer)-1 {
	fmt.Printf("Extracting hash from audit path at index %v and layer %v :=> ", index, layer)
	digest, ok := auditPath[pathKey(index, layer)]
	if ok {
		fmt.Println("found")
		return digest
	}
	fmt.Println("not found")
	//	}

	switch {
	// we are at a leaf: A_v(i,0)
	case layer == 0 && version >= index:
		fmt.Printf("Hashing leaf with key %v\n", key)
		digest = v.leafHasher(Zero, key)
		break
		// A_v(i,r) with one empty children
	case version < index+pow(2, layer-1):
		hash := v.rootHash(auditPath, key, index, layer-1, version)
		fmt.Printf("Hashing node with empty at index %v and layer %v\n", index, layer)
		digest = v.leafHasher(One, hash)
		break
		// A_v(i,r) with no non-empty children
	case version >= index+pow(2, layer-1):
		hash1 := v.rootHash(auditPath, key, index, layer-1, version)
		hash2 := v.rootHash(auditPath, key, index+pow(2, layer-1), layer-1, version)
		fmt.Printf("Hashing node at index %v and layer %v\n", index, layer)
		digest = v.interiorHasher(One, hash1, hash2)
		break
	}

	return digest
}
