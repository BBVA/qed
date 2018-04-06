package hashing

import "crypto/sha256"


type Hasher func(...[]byte) []byte // computes the hash function

var Sha256Hasher Hasher = func(data ...[]byte) []byte {
		hasher := sha256.New()

		for i := 0; i < len(data); i++ {
			hasher.Write(data[i])
		}

		return hasher.Sum(nil)[:]
	}
