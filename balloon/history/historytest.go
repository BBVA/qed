// Copyright © 2018 Banco Bilbao Vizcaya Argentaria S.A.  All rights reserved.
// Use of this source code is governed by an Apache 2 License
// that can be found in the LICENSE file
package history

import (
	"fmt"
	"verifiabledata/balloon/hashing"
)

func FakeLeafHasherF(hasher hashing.Hasher) LeafHasher {
	return func(a, key []byte) []byte {
		digest := hasher(a, key)
		fmt.Printf("Hashing leaf: a-> %b key-> %b :=> %b\n", a, key, digest)
		return digest
	}
}

func FakeInteriorHasherF(hasher hashing.Hasher) InteriorHasher {
	return func(a, left, right []byte) []byte {
		digest := hasher(a, left, right)
		fmt.Printf("Hashing interior: a-> %b left-> %b right-> %b :=> %b\n", a, left, right, digest)
		return digest
	}
}

func FakeLeafHasherCleanF(hasher hashing.Hasher) LeafHasher {
	return func(a, key []byte) []byte {
		return hasher(key)
	}
}

func FakeInteriorHasherCleanF(hasher hashing.Hasher) InteriorHasher {
	return func(a, left, right []byte) []byte {
		return hasher(left, right)
	}
}
