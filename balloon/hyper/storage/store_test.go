/*
   Copyright 2018 Banco Bilbao Vizcaya Argentaria, S.A.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package storage

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
	b.N = 10000
	leaves := buildLeavesSlice(b.N)
	end := asBytes(b.N)
	left, _ := leaves.Split(end)
	if len(left) < b.N {
		b.Fatalf("Error splitting: left slice should have size %d but has %d", b.N, len(left))
	}
}

func asBytes(elem int) []byte {
	bytes := make([]byte, 4)
	binary.BigEndian.PutUint32(bytes, uint32(elem))
	return bytes
}

func buildLeavesSlice(numElems int) LeavesSlice {
	var leaves LeavesSlice
	// Initialize
	for i := 0; i < numElems; i++ {
		intBytes := make([]byte, 4)
		binary.BigEndian.PutUint32(intBytes, uint32(i))
		leaves = append(leaves, intBytes)
	}
	return leaves
}
