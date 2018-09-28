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

import "github.com/bbva/qed/balloon/navigator"

type HyperTreeNavigator struct {
	numBits uint16
}

func NewHyperTreeNavigator(numBits uint16) *HyperTreeNavigator {
	return &HyperTreeNavigator{numBits}
}

func (n HyperTreeNavigator) Root() navigator.Position {
	index := make([]byte, n.numBits/8)
	return NewPosition(index, n.numBits)
}

func (n HyperTreeNavigator) IsLeaf(pos navigator.Position) bool {
	return pos.Height() == 0
}

func (n HyperTreeNavigator) IsRoot(pos navigator.Position) bool {
	return pos.Height() == n.numBits
}

func (n HyperTreeNavigator) GoToLeft(pos navigator.Position) navigator.Position {
	if pos.Height() == 0 {
		return nil
	}
	return NewPosition(pos.Index(), pos.Height()-1)
}

func (n HyperTreeNavigator) GoToRight(pos navigator.Position) navigator.Position {
	if pos.Height() == 0 {
		return nil
	}
	return NewPosition(n.splitBase(pos), pos.Height()-1)
}

func (n HyperTreeNavigator) DescendToFirst(pos navigator.Position) navigator.Position {
	return NewPosition(pos.Index(), 0)
}

func (n HyperTreeNavigator) DescendToLast(pos navigator.Position) navigator.Position {
	layer := n.numBits - pos.Height()
	base := make([]byte, n.numBits/8)
	copy(base, pos.Index())
	for bit := layer; bit < n.numBits; bit++ {
		bitSet(base, bit)
	}
	return NewPosition(base, 0)
}

func (n HyperTreeNavigator) splitBase(pos navigator.Position) []byte {
	splitBit := n.numBits - pos.Height()
	split := make([]byte, n.numBits/8)
	copy(split, pos.Index())
	if splitBit < n.numBits {
		bitSet(split, splitBit)
	}
	return split
}

func bitSet(bits []byte, i uint16) {
	bits[i/8] |= 1 << uint(7-i%8)
}
