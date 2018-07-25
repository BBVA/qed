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

package history

import (
	"testing"

	"github.com/bbva/qed/hashing"

	"github.com/bbva/qed/balloon/proof"
	assert "github.com/stretchr/testify/require"
)

func TestVerifyIncremental(t *testing.T) {
	testCases := []struct {
		auditPath   proof.AuditPath
		start       uint64
		end         uint64
		startDigest []byte
		endDigest   []byte
	}{
		{
			proof.AuditPath{"0|1": []uint8{0x1}, "2|0": []uint8{0x2}, "3|0": []uint8{0x3}, "4|1": []uint8{0x1}, "6|0": []uint8{0x6}},
			2, 6, []byte{0x3}, []byte{0x7},
		},
		{
			proof.AuditPath{"0|1": []uint8{0x1}, "2|0": []uint8{0x2}, "3|0": []uint8{0x3}, "4|1": []uint8{0x1}, "6|0": []uint8{0x6}, "7|0": []uint8{0x7}},
			2, 7, []byte{0x3}, []byte{0x0},
		},
		{
			proof.AuditPath{"0|2": []uint8{0x0}, "4|0": []uint8{0x4}, "5|0": []uint8{0x5}, "6|0": []uint8{0x6}},
			4, 6, []byte{0x4}, []byte{0x7},
		},
		{
			proof.AuditPath{"0|2": []uint8{0x0}, "4|0": []uint8{0x4}, "5|0": []uint8{0x5}, "6|0": []uint8{0x6}, "7|0": []uint8{0x7}},
			4, 7, []byte{0x4}, []byte{0x0},
		},
		{
			proof.AuditPath{"2|0": []uint8{0x2}, "3|0": []uint8{0x3}, "4|0": []uint8{0x4}, "0|1": []uint8{0x1}},
			2, 4, []byte{0x3}, []byte{0x4},
		},
		{
			proof.AuditPath{"0,2": []uint8{0x0}, "4|1": []uint8{0x1}, "6|0": []uint8{0x6}, "7|0": []uint8{0x7}},
			6, 7, []byte{0x7}, []byte{0x0},
		},
	}

	lh := fakeLeafHasherCleanF(new(hashing.XorHasher))
	ih := fakeInteriorHasherCleanF(new(hashing.XorHasher))

	for _, c := range testCases {
		proof := IncrementalProof{c.start, c.end, c.auditPath, ih, lh}
		assert.True(t, proof.Verify(c.startDigest, c.endDigest))
	}
}

func TestProveAndVerifyConsecutivelyN(t *testing.T) {
	frozen, close := openBPlusStorage()
	defer close()

	tree := NewTree("treeId", frozen, hashing.NewSha256Hasher())

	const size int = 10
	eventDigests := make([][]byte, size)
	digests := make([][]byte, size)
	var err error

	for i := 0; i < size; i++ {
		index := uint64(i)
		eventDigests[i] = uint64AsBytes(index)
		digests[i], err = tree.Add(eventDigests[i], uint64AsBytes(index))
		assert.NoError(t, err, "Error while adding to the tree")

		pf, err := tree.ProveIncremental(eventDigests[max(0, i-1)], eventDigests[i], uint64(max(0, i-1)), uint64(i))
		assert.NoError(t, err, "Error while querying for incremental proof")
		assert.True(t, pf.Verify(digests[max(0, i-1)], digests[i]), "The proof should verfify correctly")
	}
}

func TestProveAndVerifyNonConsecutively(t *testing.T) {
	frozen, close := openBPlusStorage()
	defer close()

	tree := NewTree("treeId", frozen, hashing.NewSha256Hasher())

	const size int = 10
	eventDigests := make([][]byte, size)
	digests := make([][]byte, size)
	var err error

	for i := 0; i < size; i++ {
		index := uint64(i)
		eventDigests[i] = uint64AsBytes(index)
		digests[i], err = tree.Add(eventDigests[i], uint64AsBytes(index))
		assert.NoError(t, err, "Error while adding to the tree")
	}

	for i := 0; i < size-1; i++ {
		for j := i + 1; j < size; j++ {
			pf, err := tree.ProveIncremental(eventDigests[2], eventDigests[8], uint64(2), uint64(8))
			assert.NoErrorf(t, err, "Error while querying for incremental proof between versions %d and %d", i, j)
			assert.Truef(t, pf.Verify(digests[2], digests[8]), "The proof should verfify correctly for versions %d, and %d", i, j)
		}
	}

}
