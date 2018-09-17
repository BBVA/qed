package common

import "github.com/bbva/qed/hashing"

type FakeCache struct {
	FixedDigest hashing.Digest
}

func NewFakeCache(fixedDigest hashing.Digest) *FakeCache {
	return &FakeCache{fixedDigest}
}

func (c FakeCache) Get(Position) (hashing.Digest, bool) {
	return hashing.Digest{0x0}, true
}
