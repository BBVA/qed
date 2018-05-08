package history

import (
	"fmt"
	"testing"

	"verifiabledata/balloon/hashing"
)

func TestVerify(t *testing.T) {

	var testCases = []struct {
		key                []byte
		auditPath          []Node
		version            uint64
		expectedCommitment []byte
	}{
		// INCREMENTAL
		{
			key:                []byte{0x0},
			auditPath:          []Node{},
			version:            0,
			expectedCommitment: []byte{0x0},
		},
		{
			key: []byte{0x1},
			auditPath: []Node{
				Node{[]byte{0x0}, 0, 0},
			},
			version:            1,
			expectedCommitment: []byte{0x0},
		},
		{
			key: []byte{0x2},
			auditPath: []Node{
				Node{[]byte{0x1}, 0, 1},
			},
			version:            2,
			expectedCommitment: []byte{0x3},
		},
		{
			key: []byte{0x3},
			auditPath: []Node{
				Node{[]byte{0x2}, 2, 0},
				Node{[]byte{0x1}, 0, 1},
			},
			version:            3,
			expectedCommitment: []byte{0x0},
		},
		{
			key: []byte{0x4},
			auditPath: []Node{
				Node{[]byte{0x1}, 0, 2},
			},
			version:            4,
			expectedCommitment: []byte{0x4},
		},

		// LONGER VERSION
		{
			key: []byte{0x0},
			auditPath: []Node{
				Node{[]byte{0x1}, 1, 0},
				Node{[]byte{0x0}, 2, 1},
				Node{[]byte{0x4}, 4, 2},
			},
			version:            4,
			expectedCommitment: []byte{0x4},
		},
		{
			key: []byte{0x2},
			auditPath: []Node{
				Node{[]byte{0x0}, 0, 1},
				Node{[]byte{0x3}, 3, 0},
				Node{[]byte{0x4}, 4, 2},
			},
			version:            4,
			expectedCommitment: []byte{0x4},
		},
	}

	hasher := hashing.XorHasher

	for i, c := range testCases {
		proof := NewProof(c.auditPath, c.version, hasher)
		correct := proof.Verify(c.expectedCommitment, c.key, c.version)

		if !correct {
			t.Fatalf("The verification of the test case #%d failed", i)
		}
	}
}

func TestAddAndVerify(t *testing.T) {

	store, closeF := openBPlusStorage()
	defer closeF()

	hasher := hashing.Sha256Hasher
	ht := NewFakeTree(store, hasher)

	key := []byte("I AM A STRANGE LOOP")
	var value uint64 = 0

	commitment := <-ht.Add(key, uInt64AsBytes(value))
	membershipProof := <-ht.Prove(key, 0, value)

	proof := NewProof(membershipProof.Nodes, value, hasher)
	correct := proof.Verify(commitment, key, value)

	if !correct {
		t.Errorf("incorrect")
	}
}

func TestAddAndVerify256(t *testing.T) {

	store, closeF := openBPlusStorage()
	defer closeF()

	hasher := hashing.Sha256Hasher
	ht := NewTree(store, hasher)

	for i := uint64(0); i < 256; i++ {
		key := randomBytes(128)
		value := i
		commitment := <-ht.Add(key, uInt64AsBytes(value))
		membershipProof := <-ht.Prove(key, i, value)
		proof := NewProof(membershipProof.Nodes, value, hasher)
		correct := proof.Verify(commitment, key, value)

		if !correct {
			fmt.Printf("C %+v\n", commitment)
			fmt.Printf("MP %+v\n", membershipProof)
			fmt.Printf("P %+v\n", proof)
			graphTree(ht, key, value)
			t.Fatalf("incorrect test case: %d", i)
		}
	}
}
