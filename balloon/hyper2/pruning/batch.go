package pruning

import (
	"fmt"
	"strings"

	"github.com/bbva/qed/hashing"
)

type BatchNode struct {
	Batch    [][]byte
	nodeSize int // in bytes
}

func NewEmptyBatchNode(nodeSize int) *BatchNode {
	return &BatchNode{
		nodeSize: nodeSize,
		Batch:    make([][]byte, 31, 31),
	}
}

func NewBatchNode(nodeSize int, batch [][]byte) *BatchNode {
	return &BatchNode{
		nodeSize: nodeSize,
		Batch:    batch,
	}
}

func (b BatchNode) String() string {
	var strs []string
	for i, n := range b.Batch {
		strs = append(strs, fmt.Sprintf("[%d - %x]", i, n))
	}
	return strings.Join(strs, "\n")
}

func (b BatchNode) HasLeafAt(i int8) bool {
	return b.Batch[i][b.nodeSize] == byte(1)
}

func (b BatchNode) AddHashAt(i int8, value []byte) {
	b.Batch[i] = append(value, byte(0))
}

func (b BatchNode) AddLeafAt(i int8, hash hashing.Digest, key, value []byte) {
	b.Batch[i] = append(hash, byte(1))
	b.Batch[2*i+1] = append(key, byte(2))
	b.Batch[2*i+2] = append(value, byte(2))
}

func (b BatchNode) GetLeafKVAt(i int8) ([]byte, []byte) {
	return b.Batch[2*i+1][:b.nodeSize], b.Batch[2*i+2][:b.nodeSize]
}

func (b BatchNode) HasElementAt(i int8) bool {
	return len(b.Batch[i]) > 0
}

func (b BatchNode) GetElementAt(i int8) []byte {
	return b.Batch[i][:b.nodeSize]
}

func (b BatchNode) ResetElementAt(i int8) {
	b.Batch[i] = nil
}

func (b BatchNode) Serialize() []byte {
	serialized := make([]byte, 4)
	for i := 0; i < 31; i++ {
		if len(b.Batch[i]) != 0 {
			bitSet(serialized, i)
			serialized = append(serialized, b.Batch[i]...)
		}
	}
	return serialized
}

func ParseBatch(nodeSize int, value []byte) [][]byte {
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

func ParseBatchNode(nodeSize int, value []byte) *BatchNode {
	return NewBatchNode(nodeSize, ParseBatch(nodeSize, value))
}

func bitIsSet(bits []byte, i int) bool {
	return bits[i/8]&(1<<uint(7-i%8)) != 0
}

func bitSet(bits []byte, i int) {
	bits[i/8] |= 1 << uint(7-i%8)
}
