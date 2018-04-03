package hyper

import (
	"encoding/binary"
	"testing"
)

func TestSplit(t *testing.T) {
	tests := []struct {
		numElems   int
		searchElem int
		leftSize   int
		rightSize  int
	}{
		{numElems: 40, searchElem: 20, leftSize: 20, rightSize: 20},
		{numElems: 40, searchElem: 35, leftSize: 35, rightSize: 5},
		{numElems: 40, searchElem: 41, leftSize: 40, rightSize: 0},
		{numElems: 0, searchElem: 1, leftSize: 0, rightSize: 0},
	}

	for _, test := range tests {
		leaves := buildLeavesSlice(test.numElems)
		left, right := leaves.Split(asBytes(test.searchElem))
		if len(left) != test.leftSize {
			t.Fatalf("Error splitting: left slice should have size %d but has %d", test.leftSize, len(left))
		}
		if len(right) != test.rightSize {
			t.Fatalf("Error splitting: right slice should have size %d but has %d", test.rightSize, len(right))
		}
	}
}

func asBytes(elem int) []byte {
	bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytes, uint32(elem))
	return bytes
}

func buildLeavesSlice(numElems int) LeavesSlice {
	var leaves LeavesSlice
	// Initialize
	for i := 0; i < numElems; i++ {
		intBytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(intBytes, uint32(i))
		leaves = append(leaves, intBytes)
	}
	return leaves
}
