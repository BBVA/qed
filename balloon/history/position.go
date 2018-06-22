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
	"encoding/binary"
	"fmt"
	"math"

	"github.com/bbva/qed/balloon/position"
)

type Position struct {
	index   uint64
	layer   uint64
	version uint64
}

func NewPosition(index, layer, version uint64) position.Position {
	return Position{index, layer, version}
}

func NewRootPosition(version uint64) position.Position {
	return NewPosition(0, getDepth(version), version)
}

func getDepth(index uint64) uint64 {
	return uint64(math.Ceil(math.Log2(float64(index + 1))))
}

func (p Position) String() string {
	return fmt.Sprintf("index: %b , layer: %d", p.index, p.layer)
}

func (p Position) Key() []byte {
	return p.indexBytes()
}

func (p Position) Height() uint64 {
	return p.layer // TODO it should be renamed to height
}

func (p Position) Id() []byte {
	id := make([]byte, 16) // idLen is the size of layer and height, which is 16 bytes
	copy(id, p.indexBytes())
	copy(id[len(p.indexBytes()):], p.layerBytes())
	return id
}

func (p Position) StringId() string {
	return fmt.Sprintf("%d|%d", p.index, p.layer)
}

func (p Position) Left() position.Position {
	return NewPosition(p.index, p.layer-1, p.version)
}

func (p Position) Right() position.Position {
	return NewPosition(p.index+pow(2, p.layer-1), p.layer-1, p.version)
}

func (p Position) Direction(target []byte) position.Direction {
	version := binary.LittleEndian.Uint64(target)
	var direction position.Direction
	switch {
	case p.layer == 0 || p.index > version:
		direction = position.Halt
	case version < p.index+pow(2, p.layer-1):
		direction = position.Left
	case version >= p.index+pow(2, p.layer-1):
		direction = position.Right
	}
	return direction
}

func (p Position) IsLeaf() bool {
	return p.layer == 0
}

func (p Position) FirstLeaf() position.Position {
	return NewPosition(p.index, 0, p.version)
}

func (p Position) LastLeaf() position.Position {
	return NewPosition(p.index+pow(2, p.layer)-1, 0, p.version)
}

func (p Position) ShouldBeCached() bool {
	if p.version >= p.index+pow(2, p.layer)-1 {
		return true
	}
	return false
}

func (p Position) indexBytes() []byte {
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, p.index)
	return bytes
}

func (p Position) layerBytes() []byte {
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, p.layer)
	return bytes
}

// Utility to calculate arbitraty pow and return an int64
func pow(x, y uint64) uint64 {
	return uint64(math.Pow(float64(x), float64(y)))
}
