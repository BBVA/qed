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
	"math/bits"

	"github.com/bbva/qed/balloon/common"
)

type HistoryTreeNavigator struct {
	version uint64
	depth   uint16
}

func NewHistoryTreeNavigator(version uint64) *HistoryTreeNavigator {
	depth := uint16(bits.Len64(version))
	return &HistoryTreeNavigator{version, depth}
}

func (n HistoryTreeNavigator) Root() common.Position {
	return NewPosition(0, n.depth)
}

func (n HistoryTreeNavigator) IsLeaf(pos common.Position) bool {
	return pos.Height() == 0
}

func (n HistoryTreeNavigator) IsRoot(pos common.Position) bool {
	return pos.Height() == n.depth && pos.IndexAsUint64() == 0
}

func (n HistoryTreeNavigator) GoToLeft(pos common.Position) common.Position {
	if pos.Height() == 0 {
		return nil
	}
	return NewPosition(pos.IndexAsUint64(), pos.Height()-1)
}
func (n HistoryTreeNavigator) GoToRight(pos common.Position) common.Position {
	rightIndex := pos.IndexAsUint64() + 1<<(pos.Height()-1)
	if pos.Height() == 0 || rightIndex > n.version {
		return nil
	}
	return NewPosition(rightIndex, pos.Height()-1)
}

func (n HistoryTreeNavigator) DescendToFirst(pos common.Position) common.Position {
	if n.IsLeaf(pos) {
		return nil
	}
	return NewPosition(pos.IndexAsUint64(), 0)
}

func (n HistoryTreeNavigator) DescendToLast(pos common.Position) common.Position {
	if n.IsLeaf(pos) {
		return nil
	}
	lastDescendantIndex := pos.IndexAsUint64() + 1<<pos.Height() - 1
	if lastDescendantIndex > n.version {
		return nil
	}
	return NewPosition(lastDescendantIndex, 0)
}
