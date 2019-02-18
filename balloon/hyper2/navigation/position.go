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

package navigation

import (
	"fmt"

	"github.com/bbva/qed/util"
)

const KeySize = 34

type Position struct {
	Index  []byte
	Height uint16

	serialized [KeySize]byte
	numBits    uint16
}

func NewPosition(index []byte, height uint16) Position {
	var b [KeySize]byte // Size of the index plus 2 bytes for the height
	copy(b[:], index[:len(index)])
	copy(b[len(index):], util.Uint16AsBytes(height))
	return Position{
		Index:      index,
		Height:     height,
		serialized: b, // memoized
		numBits:    uint16(len(index)) * 8,
	}
}

func NewRootPosition(numBits uint16) Position {
	return NewPosition(make([]byte, numBits/8), numBits)
}

func (p Position) Bytes() []byte {
	return p.serialized[:]
}

func (p Position) FixedBytes() [KeySize]byte {
	return p.serialized
}

func (p Position) String() string {
	return fmt.Sprintf("Pos(%d, %d)", p.Index, p.Height)
}

func (p Position) StringId() string {
	return fmt.Sprintf("%#x|%d", p.Index, p.Height)
}

func (p Position) Left() Position {
	if p.IsLeaf() {
		return p
	}
	return NewPosition(p.Index, p.Height-1)
}

func (p Position) Right() Position {
	if p.IsLeaf() {
		return p
	}
	return NewPosition(p.splitBase(), p.Height-1)
}

func (p Position) IsLeaf() bool {
	return p.Height == 0
}

func (p Position) FirstDescendant() Position {
	if p.IsLeaf() {
		return p
	}
	return NewPosition(p.Index, 0)
}

func (p Position) LastDescendant() Position {
	if p.IsLeaf() {
		return p
	}
	index := make([]byte, p.numBits/8)
	copy(index, p.Index)
	for bit := p.numBits - p.Height; bit < p.numBits; bit++ {
		bitSet(index, bit)
	}
	return NewPosition(index, 0)
}

func (p Position) splitBase() []byte {
	splitBit := p.numBits - p.Height
	split := make([]byte, p.numBits/8)
	copy(split, p.Index)
	if splitBit < p.numBits {
		bitSet(split, splitBit)
	}
	return split
}

func bitSet(bits []byte, i uint16) {
	bits[i/8] |= 1 << uint(7-i%8)
}
