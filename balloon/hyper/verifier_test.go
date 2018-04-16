package hyper

import (
	"bytes"
	"testing"
	"verifiabledata/balloon/hashing"
)

func TestVerify(t *testing.T) {
	verifier := NewVerifier("test", ByteHasher, fakeLeafHasherF(ByteHasher), fakeInteriorHasherF(ByteHasher))

	expectedCommitment := []byte{0x5a}
	key := []byte{0x5a}
	value := []byte{0x00}
	auditPath := [][]byte{
		[]byte{0x5a},
		[]byte{0x00},
		[]byte{0x00},
		[]byte{0x00},
		[]byte{0x00},
		[]byte{0x00},
		[]byte{0x00},
		[]byte{0x00},
		[]byte{0x00},
	}

	correct, recomputed := verifier.Verify(expectedCommitment, auditPath, key, value)

	if !correct {
		t.Fatalf("The verification failed")
	}

	if bytes.Compare(recomputed, expectedCommitment) != 0 {
		t.Fatalf("Expected: %x, Actual: %x", expectedCommitment, recomputed)
	}
}

func fakeLeafHasherF(hasher hashing.Hasher) LeafHasher {
	return func(id, value, base []byte) []byte {
		return hasher(base)
	}
}

func fakeInteriorHasherF(hasher hashing.Hasher) InteriorHasher {
	return func(left, right, base, height []byte) []byte {
		return hasher(left, right)
	}
}

func ByteHasher(data ...[]byte) []byte {
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
