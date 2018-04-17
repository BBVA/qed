package history

import (
	"verifiabledata/balloon/hashing"
)

// Constant Empty is a constant for empty leaves
var Empty = []byte{0x00}

// Constant Set is a constant for non-empty leaves
var Set = []byte{0x01}

type LeafHasher func([]byte, []byte) []byte
type InteriorHasher func([]byte, []byte, []byte) []byte

func LeafHasherF(hasher hashing.Hasher) LeafHasher {
	return func(a, key []byte) []byte {
		return hasher(a, key)
	}
}

func InteriorHasherF(hasher hashing.Hasher) InteriorHasher {
	return func(a, left, right []byte) []byte {
		return hasher(a, left, right)
	}
}
