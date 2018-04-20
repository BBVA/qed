package history

import (
	"verifiabledata/balloon/hashing"
)

// Constant Zero is the 0x0 byte, and it is used as a prefix to know
// if a node has a zero digest.
var Zero = []byte{0x0}

// Constant One is the 0x1 byte, and it is used as a prefix to know
// if a node has a non-zero digest.
var One = []byte{0x1}

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
