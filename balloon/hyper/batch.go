/*
   Copyright 2018-2019 Banco Bilbao Vizcaya Argentaria, S.A.

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

package hyper

import (
	"fmt"
	"strings"

	"github.com/bbva/qed/crypto/hashing"
)

const (
	DefaultBatchHeight uint16 = 4
	DefaultBatchLevels uint8  = 6
	DefaultBatchSize   uint64 = (31 * 33) + 4 // 31 nodes * 33 bytes/node + 4 bytes of bitmap
	MaxNumberOfBatches uint64 = 1118481       // for a 6 levels cache
)

type batchNode struct {
	batch    [][]byte
	nodeSize int // in bytes
}

func newEmptyBatchNode(nodeSize int) *batchNode {
	return &batchNode{
		nodeSize: nodeSize,
		batch:    make([][]byte, 31, 31),
	}
}

func newBatchNode(nodeSize int, batch [][]byte) *batchNode {
	return &batchNode{
		nodeSize: nodeSize,
		batch:    batch,
	}
}

func (b batchNode) String() string {
	var strs []string
	for i, n := range b.batch {
		strs = append(strs, fmt.Sprintf("[%d - %#x]", i, n))
	}
	return strings.Join(strs, "\n")
}

func (b batchNode) HasLeafAt(i int8) bool {
	return len(b.batch[i]) > 0 && b.batch[i][b.nodeSize] == byte(1)
}

func (b batchNode) AddHashAt(i int8, value []byte) {
	b.batch[i] = append(value, byte(0))
}

func (b batchNode) AddLeafAt(i int8, hash hashing.Digest, key, value []byte) {
	b.batch[i] = append(hash, byte(1))
	b.batch[2*i+1] = append(key, byte(2))
	b.batch[2*i+2] = append(value, byte(2))
}

func (b batchNode) GetLeafKVAt(i int8) ([]byte, []byte) {
	return b.batch[2*i+1][:b.nodeSize], b.batch[2*i+2][:b.nodeSize]
}

func (b batchNode) HasElementAt(i int8) bool {
	return len(b.batch[i]) > 0
}

func (b batchNode) GetElementAt(i int8) []byte {
	return b.batch[i][:b.nodeSize]
}

func (b batchNode) ResetElementAt(i int8) {
	b.batch[i] = nil
}

func (b batchNode) Serialize() []byte {
	serialized := make([]byte, 4)
	for i := uint16(0); i < 31; i++ {
		if len(b.batch[i]) != 0 {
			bitSet(serialized, i)
			serialized = append(serialized, b.batch[i]...)
		}
	}
	return serialized
}

func parseBatch(nodeSize int, value []byte) [][]byte {
	batch := make([][]byte, 31, 31) // 31 nodes (including the root)
	bitmap := value[:4]             // the first 4 bytes define the bitmap
	size := nodeSize + 1

	j := 0
	for i := 0; i < 31; i++ {
		if bitIsSet(bitmap, i) {
			batch[i] = value[4+size*j : 4+size*(j+1)]
			j++
		}
	}

	return batch
}

func parseBatchNode(nodeSize int, value []byte) *batchNode {
	return newBatchNode(nodeSize, parseBatch(nodeSize, value))
}

func bitIsSet(bits []byte, i int) bool {
	return bits[i/8]&(1<<uint(7-i%8)) != 0
}
