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

package hyper

import (
	"encoding/binary"
	"fmt"

	"github.com/bbva/qed/balloon/position"
)

type Position struct {
	base       []byte
	height     uint64
	numBits    uint64
	cacheLevel uint64
}

func NewPosition(base []byte, height, numBits, cacheLevel uint64) position.Position {
	return Position{base, height, numBits, cacheLevel}
}

func NewRootPosition(numBits, cacheLevel uint64) position.Position {
	base := make([]byte, numBits/8)
	return NewPosition(base, numBits, numBits, cacheLevel)
}

func (p Position) String() string {
	return fmt.Sprintf("base: %b , split: %b , height: %d , numBits: %d", p.base, p.Split(), p.height, p.numBits)
}

func (p Position) Key() []byte {
	return p.base
}

func (p Position) Height() uint64 {
	return p.height
}

func (p Position) Id() []byte {
	id := make([]byte, p.len()+8)
	copy(id, p.base)
	copy(id[p.len():], p.heightBytes())
	return id
}

func (p Position) StringId() string {
	return fmt.Sprintf("%x|%d", p.base, p.height)
}

func (p Position) IsLeaf() bool {
	return p.height == 0
}

func (p Position) Direction(target []byte) position.Direction {
	if (p.height) == 0 {
		return position.Halt
	}
	if !bitIsSet(target, p.numBits-p.height) {
		return position.Left
	}
	return position.Right
}

func (p Position) Left() position.Position {
	return NewPosition(p.base, p.height-1, p.numBits, p.cacheLevel)
}

func (p Position) Right() position.Position {
	return NewPosition(p.Split(), p.height-1, p.numBits, p.cacheLevel)
}

func (p Position) FirstLeaf() position.Position {
	return NewPosition(p.base, 0, p.numBits, p.cacheLevel)
}

func (p Position) LastLeaf() position.Position {
	layer := p.numBits - p.height
	base := make([]byte, p.len())
	copy(base, p.base)
	for bit := layer; bit < p.numBits; bit++ {
		bitSet(base, bit)
	}
	return NewPosition(base, 0, p.numBits, p.cacheLevel)
}

func (p Position) Split() []byte {
	splitBit := p.numBits - p.height
	split := make([]byte, p.len())
	copy(split, p.base)
	if splitBit < p.numBits {
		bitSet(split, splitBit)
	}
	return split
}

func (p Position) ShouldBeCached() bool {
	if p.height > p.cacheLevel {
		return true
	}
	return false
}

func (p Position) heightBytes() []byte {
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, p.height)
	return bytes
}

func (p Position) len() uint64 {
	return p.numBits / 8
}

func bitIsSet(bits []byte, i uint64) bool { return bits[i/8]&(1<<uint(7-i%8)) != 0 }
func bitSet(bits []byte, i uint64)        { bits[i/8] |= 1 << uint(7-i%8) }
func bitUnset(bits []byte, i uint64)      { bits[i/8] &= 0 << uint(7-i%8) }
