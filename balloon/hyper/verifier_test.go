package hyper

import (
	"bytes"
	"testing"
	"verifiabledata/balloon/hashing"
)

func TestVerify(t *testing.T) {
	hasher := hashing.XorHasher
	verifier := NewVerifier("test", hasher, fakeLeafHasherF(hasher), fakeInteriorHasherF(hasher))

	expectedCommitment := []byte{0x5a}
	key := []byte{0x5a}
	value := []byte{0x01}
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
