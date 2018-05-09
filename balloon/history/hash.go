package history

import (
	"qed/balloon/hashing"
)

// Constant Zero is the 0x0 byte, and it is used as a prefix to know
// if a node has a zero digest.
var Zero = []byte{0x0}

// Constant One is the 0x1 byte, and it is used as a prefix to know
// if a node has a non-zero digest.
var One = []byte{0x1}

type leafHasher func([]byte, []byte) []byte
type interiorHasher func([]byte, []byte, []byte) []byte

func leafHasherF(hasher hashing.Hasher) leafHasher {
	return func(a, key []byte) []byte {
		return hasher(a, key)
	}
}

func interiorHasherF(hasher hashing.Hasher) interiorHasher {
	return func(a, left, right []byte) []byte {
		return hasher(a, left, right)
	}
}