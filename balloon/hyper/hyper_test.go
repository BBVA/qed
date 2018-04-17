package hyper

import (
	"bytes"
	"verifiabledata/balloon/hashing"
)

func fakeLeafHasherF(hasher hashing.Hasher) LeafHasher {
	return func(id, value, base []byte) []byte {
		if bytes.Equal(value, Empty) {
			return hasher(Empty)
		}
		return hasher(base)
	}
}

func fakeInteriorHasherF(hasher hashing.Hasher) InteriorHasher {
	return func(left, right, base, height []byte) []byte {
		return hasher(left, right)
	}
}
