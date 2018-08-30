package common

import (
	"crypto/sha256"
	"hash"
)

type Digest []byte

type Hasher interface {
	Salted([]byte, ...[]byte) Digest
	Do(...[]byte) Digest
	Len() uint16
}

type XorHasher struct{}

func (x XorHasher) Salted(salt []byte, data ...[]byte) Digest {
	data = append(data, salt)
	return x.Do(data...)
}

func (x XorHasher) Do(data ...[]byte) Digest {
	var result byte
	for _, elem := range data {
		var sum byte
		for _, b := range elem {
			sum = sum ^ b
		}
		result = result ^ sum
	}
	return []byte{result}
}
func (s XorHasher) Len() uint16 { return uint16(8) }

type Sha256Hasher struct {
	underlying hash.Hash
}

func NewSha256Hasher() *Sha256Hasher {
	return &Sha256Hasher{underlying: sha256.New()}
}

func (s *Sha256Hasher) Salted(salt []byte, data ...[]byte) Digest {
	data = append(data, salt)
	return s.Do(data...)
}

func (s *Sha256Hasher) Do(data ...[]byte) Digest {
	s.underlying.Reset()
	for i := 0; i < len(data); i++ {
		s.underlying.Write(data[i])
	}
	return s.underlying.Sum(nil)[:]
}

func (s Sha256Hasher) Len() uint16 { return uint16(256) }

type FakeHasher struct {
	underlying Hasher
}

func (h *FakeHasher) Salted(salt []byte, data ...[]byte) Digest {
	return h.underlying.Do(data...)
}

func (h *FakeHasher) Do(data ...[]byte) Digest {
	return h.underlying.Do(data...)
}

func (h FakeHasher) Len() uint16 {
	return h.underlying.Len()
}
