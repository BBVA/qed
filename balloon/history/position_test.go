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

	"github.com/bbva/qed/balloon/position"
	assert "github.com/stretchr/testify/require"
)

func TestString(t *testing.T) {

	p := NewPosition(0, 0, 0)

	assert.Equal(t, p.String(), "index: 0 , layer: 0", "Invalid history position")
}

func TestPositionId(t *testing.T) {
	p := NewPosition(0, 0, 0)
	assert.Equal(t, p.Id(), make([]byte, 16), "Invalid history position")
}

func TestPositionStringId(t *testing.T) {
	p := NewPosition(0, 0, 0)
	assert.Equal(t, p.StringId(), "0|0", "Invalid history position")
}

func TestPositionLeft(t *testing.T) {
	p := NewPosition(0, 1, 0)
	l := p.Left()
	assert.Equal(t, make([]byte, 8), l.Key(), "Invalid index")
	assert.Equal(t, uint64(0), l.Height(), "Invalid height")
	assert.Equal(t, l.String(), "index: 0 , layer: 0", "Invalid history position")
}

func TestPositionRight(t *testing.T) {
	p := NewPosition(0, 1, 0)
	r := p.Right()
	assert.Equal(t, []byte{0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}, r.Key(), "Invalid index")
	assert.Equal(t, uint64(0), r.Height(), "Invalid height")
	assert.Equal(t, r.String(), "index: 1 , layer: 0", "Invalid history position")
}

func TestPositionDirection(t *testing.T) {
	p := NewPosition(0, 2, 4)
	assert.Equal(t, p.Direction(p.Left().Key()), position.Left, "Invalid direction")
	assert.Equal(t, p.Direction(p.Right().Key()), position.Right, "Invalid direction")
}

func TestPositionIsLeaf(t *testing.T) {
	p := NewPosition(0, 0, 0)
	assert.True(t, p.IsLeaf(), "Position should be a leaf")
	p = NewPosition(0, 1, 0)
	assert.False(t, p.IsLeaf(), "Position shouldn't be a leaf")
}

func TestPositionShouldBeCached(t *testing.T) {
	p := NewPosition(0, 0, 1)
	assert.True(t, p.ShouldBeCached(), "Position should be cached")
	p = NewPosition(1, 1, 0)
	assert.False(t, p.ShouldBeCached(), "Position shouldn't be cached")
}
