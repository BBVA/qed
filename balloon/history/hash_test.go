package history

import (
	"fmt"
	"verifiabledata/balloon/hashing"
)

func fakeLeafHasherF(hasher hashing.Hasher) LeafHasher {
	return func(a, key []byte) []byte {
		fmt.Printf("Hashing leaf: a-> %b key-> %b\n", a, key)
		return hasher(a, key)
	}
}

func fakeInteriorHasherF(hasher hashing.Hasher) InteriorHasher {
	return func(a, left, right []byte) []byte {
		fmt.Printf("Hashing interior: a-> %b left-> %b right-> %b\n", a, left, right)
		return hasher(a, left, right)
	}
}
