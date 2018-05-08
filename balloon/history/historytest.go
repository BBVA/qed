// Copyright © 2018 Banco Bilbao Vizcaya Argentaria S.A.  All rights reserved.
// Use of this source code is governed by an Apache 2 License
// that can be found in the LICENSE file
package history

import (
	"fmt"
	"runtime"

	"verifiabledata/balloon/hashing"
	"verifiabledata/balloon/storage"
	"verifiabledata/log"
)

func fakeLeafHasherF(hasher hashing.Hasher) leafHasher {
	return func(a, key []byte) []byte {
		digest := hasher(a, key)
		log.Debug("Hashing leaf: a-> %b key-> %b :=> %b\n", a, key, digest)
		return digest
	}
}

func fakeInteriorHasherF(hasher hashing.Hasher) interiorHasher {
	return func(a, left, right []byte) []byte {
		digest := hasher(a, left, right)
		log.Debug("Hashing interior: a-> %b left-> %b right-> %b :=> %b\n", a, left, right, digest)
		return digest
	}
}

func fakeLeafHasherCleanF(hasher hashing.Hasher) leafHasher {
	return func(a, key []byte) []byte {
		return hasher(key)
	}
}

func fakeInteriorHasherCleanF(hasher hashing.Hasher) interiorHasher {
	return func(a, left, right []byte) []byte {
		return hasher(left, right)
	}
}

func NewFakeTree(frozen storage.Store, hasher hashing.Hasher) *Tree {

	tree := NewTree(frozen, hasher)
	tree.leafHasher = fakeLeafHasherF(hasher)
	tree.interiorHasher = fakeInteriorHasherF(hasher)

	return tree
}

func NewFakeCleanTree(frozen storage.Store, hasher hashing.Hasher) *Tree {

	tree := NewTree(frozen, hasher)
	tree.leafHasher = fakeLeafHasherCleanF(hasher)
	tree.interiorHasher = fakeInteriorHasherCleanF(hasher)

	return tree
}

func NewFakeProof(auditPath []Node, index uint64, hasher hashing.Hasher) *Proof {

	proof := NewProof(auditPath, index, hasher)
	proof.leafHasher = fakeLeafHasherF(hasher)
	proof.interiorHasher = fakeInteriorHasherF(hasher)

	return proof
}

func where(calldepth int) string {
	_, file, line, ok := runtime.Caller(calldepth)
	if !ok {
		file = "???"
		line = 0
	}
	return fmt.Sprintf("%s:%d", file, line)
}
