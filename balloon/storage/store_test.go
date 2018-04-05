package storage

import (
	"encoding/binary"
	"fmt"
	"testing"
	"verifiabledata/util"
)

func TestSplit(t *testing.T) {
	tests := []struct {
		numElems   int
		searchElem int
		leftSize   int
		rightSize  int
	}{
		{numElems: 4000, searchElem: 2000, leftSize: 2000, rightSize: 2000},
		{numElems: 4000, searchElem: 3500, leftSize: 3500, rightSize: 500},
		{numElems: 4000, searchElem: 4001, leftSize: 4000, rightSize: 0},
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

func BenchmarkSplitAtTheEnd(b *testing.B) {
	N := 1000
	leaves := buildLeavesSlice(N)
	fmt.Println(len(leaves))
	end := asBytes(N - 1)
	fmt.Println(end)
	left, _ := leaves.Split(end)
	if len(left) < N {
		b.Fatalf("Error splitting: left slice should have size %d but has %d", N, len(left))
	}
}

func asBytes(elem int) []byte {
	bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytes, uint32(elem))
	return bytes
}

func buildLeavesSlice(numElems int) LeavesSlice {
	var leaves LeavesSlice
	hasher := util.Hash256()
	// Initialize
	for i := 0; i < numElems; i++ {
		intBytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(intBytes, uint32(i))
		leaves = append(leaves, hasher.Do(intBytes))
	}
	return leaves
}
