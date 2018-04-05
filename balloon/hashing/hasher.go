package hashing

import "crypto/sha256"

type Hasher func(...[]byte) []byte // computes the hash function

func Sha256Hasher() (Hasher, int) {
	var h Hasher
	h = func(data ...[]byte) []byte {
		hasher := sha256.New()

		for i := 0; i < len(data); i++ {
			hasher.Write(data[i])
		}

		return hasher.Sum(nil)[:]
	}
	return h, 256
}
