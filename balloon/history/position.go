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

package history

import (
	"fmt"
	"math/bits"

	"github.com/bbva/qed/util"
)

const keySize = 10

type position struct {
	Index  uint64
	Height uint16

	serialized [keySize]byte
}

func newPosition(index uint64, height uint16) *position {
	var b [keySize]byte // Size of the index plus 2 bytes for the height
	indexAsBytes := util.Uint64AsBytes(index)
	copy(b[:], indexAsBytes[:len(indexAsBytes)])
	copy(b[len(indexAsBytes):], util.Uint16AsBytes(height))
	return &position{
		Index:      index,
		Height:     height,
		serialized: b, // memoized
	}
}

func newRootPosition(version uint64) *position {
	return newPosition(0, uint16(bits.Len64(version)))
}

func (p position) Bytes() []byte {
	return p.serialized[:]
}

func (p position) FixedBytes() [keySize]byte {
	return p.serialized
}

func (p position) String() string {
	return fmt.Sprintf("Pos(%d, %d)", p.Index, p.Height)
}

func (p position) StringId() string {
	return fmt.Sprintf("%d|%d", p.Index, p.Height)
}

func (p position) Left() *position {
	if p.IsLeaf() {
		return nil
	}
	return newPosition(p.Index, p.Height-1)
}

func (p position) Right() *position {
	if p.IsLeaf() {
		return nil
	}
	return newPosition(p.Index+1<<(p.Height-1), p.Height-1)
}

func (p position) IsLeaf() bool {
	return p.Height == 0
}

func (p position) FirstDescendant() *position {
	if p.IsLeaf() {
		return &p
	}
	return newPosition(p.Index, 0)
}

func (p position) LastDescendant() *position {
	if p.IsLeaf() {
		return &p
	}
	return newPosition(p.Index+1<<p.Height-1, 0)
}
