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
	"math/bits"

	"github.com/bbva/qed/util"
)

const KeySize = 10

type Position struct {
	Index  uint64
	Height uint16

	serialized [KeySize]byte
}

func NewPosition(index uint64, height uint16) *Position {
	var b [KeySize]byte // Size of the index plus 2 bytes for the height
	indexAsBytes := util.Uint64AsBytes(index)
	copy(b[:], indexAsBytes[:len(indexAsBytes)])
	copy(b[len(indexAsBytes):], util.Uint16AsBytes(height))
	return &Position{
		Index:      index,
		Height:     height,
		serialized: b, // memoized
	}
}

func NewRootPosition(version uint64) *Position {
	return NewPosition(0, uint16(bits.Len64(version)))
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
	return fmt.Sprintf("%d|%d", p.Index, p.Height)
}

func (p Position) Left() *Position {
	if p.IsLeaf() {
		return nil
	}
	return NewPosition(p.Index, p.Height-1)
}

func (p Position) Right() *Position {
	if p.IsLeaf() {
		return nil
	}
	return NewPosition(p.Index+1<<(p.Height-1), p.Height-1)
}

func (p Position) IsLeaf() bool {
	return p.Height == 0
}

func (p Position) FirstDescendant() *Position {
	if p.IsLeaf() {
		return &p
	}
	return NewPosition(p.Index, 0)
}

func (p Position) LastDescendant() *Position {
	if p.IsLeaf() {
		return &p
	}
	return NewPosition(p.Index+1<<p.Height-1, 0)
}
