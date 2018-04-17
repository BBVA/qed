package hyper

import (
	"bytes"
	"fmt"
	"verifiabledata/balloon/hashing"
)

func fakeLeafHasherF(hasher hashing.Hasher) LeafHasher {
	return func(id, value, base []byte) []byte {
		fmt.Printf("leafhash Base %b - value %b - empty %b - digest %b\n", base, value, Empty, hasher(base))
		if bytes.Equal(value, Empty) {
			return hasher(Empty)
		}
		return hasher(base)
	}
}

func fakeInteriorHasherF(hasher hashing.Hasher) InteriorHasher {
	return func(left, right, base, height []byte) []byte {
		fmt.Printf("interiorhash Left %b - right %b - digest %b\n", left, right, hasher(left, right))
		return hasher(left, right)
	}
}
