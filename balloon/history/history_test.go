package history

import (
	"verifiabledata/balloon/hashing"
)

func fakeLeafHasherF(hasher hashing.Hasher) LeafHasher {
	return func(a, key []byte) []byte {
		return hasher(a, key)
	}
}

func fakeInteriorHasherF(hasher hashing.Hasher) InteriorHasher {
	return func(a, left, right []byte) []byte {
		return hasher(a, left, right)
	}
}
