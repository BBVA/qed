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
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRoot(t *testing.T) {

	testCases := []struct {
		numBits     uint16
		expectedPos Position
	}{
		{8, NewPosition(make([]byte, 1), 8)},
		{16, NewPosition(make([]byte, 2), 16)},
		{32, NewPosition(make([]byte, 4), 32)},
		{64, NewPosition(make([]byte, 8), 64)},
		{128, NewPosition(make([]byte, 16), 128)},
		{256, NewPosition(make([]byte, 32), 256)},
	}

	for i, c := range testCases {
		rootPos := NewRootPosition(c.numBits)
		require.Equalf(t, c.expectedPos, rootPos, "The root position should match in test case %d", i)
	}

}

func TestIsLeaf(t *testing.T) {

	testCases := []struct {
		position Position
		ok       bool
	}{
		{pos(0, 0), true},
		{pos(0, 1), false},
		{pos(1, 1), false},
		{pos(2, 0), true},
	}

	for i, c := range testCases {
		result := c.position.IsLeaf()
		require.Equalf(t, c.ok, result, "The leaf checking should match for test case %d", i)
	}

}

func TestLeft(t *testing.T) {

	testCases := []struct {
		position     Position
		expectedLeft Position
	}{
		{pos(0, 0), pos(0, 0)},
		{pos(0, 0), pos(0, 0)},
		{pos(1, 0), pos(1, 0)},
		{pos(0, 1), pos(0, 0)},
		{pos(0, 0), pos(0, 0)},
		{pos(1, 0), pos(1, 0)},
		{pos(2, 0), pos(2, 0)},
		{pos(0, 1), pos(0, 0)},
		{pos(2, 1), pos(2, 0)}, // TODO check invalid positions like (1,1)?
		{pos(0, 0), pos(0, 0)},
		{pos(1, 0), pos(1, 0)},
		{pos(2, 0), pos(2, 0)},
		{pos(0, 1), pos(0, 0)},
		{pos(2, 1), pos(2, 0)},
		{pos(0, 2), pos(0, 1)},
	}

	for i, c := range testCases {
		left := c.position.Left()
		require.Equalf(t, c.expectedLeft, left, "The left positions should match for test case %d", i)
	}
}

func TestRight(t *testing.T) {

	testCases := []struct {
		position      Position
		expectedRight Position
	}{
		{pos(0, 0), pos(0, 0)},
		{pos(0, 0), pos(0, 0)},
		{pos(1, 0), pos(1, 0)},
		{pos(0, 1), pos(1, 0)},
		{pos(0, 0), pos(0, 0)},
		{pos(1, 0), pos(1, 0)},
		{pos(2, 0), pos(2, 0)},
		{pos(0, 1), pos(1, 0)},
		{pos(2, 1), pos(3, 0)},
		{pos(0, 0), pos(0, 0)},
		{pos(1, 0), pos(1, 0)},
		{pos(2, 0), pos(2, 0)},
		{pos(0, 1), pos(1, 0)},
		{pos(2, 1), pos(3, 0)},
		{pos(0, 2), pos(2, 1)},
	}

	for i, c := range testCases {
		right := c.position.Right()
		require.Equalf(t, c.expectedRight, right, "The right positions should match for test case %d", i)
	}
}

func TestFirstDescendant(t *testing.T) {

	testCases := []struct {
		position    Position
		expectedPos Position
	}{
		{pos(0, 0), pos(0, 0)},
		{pos(1, 0), pos(1, 0)},
		{pos(0, 1), pos(0, 0)},
		{pos(2, 0), pos(2, 0)},
		{pos(2, 1), pos(2, 0)},
		{pos(0, 2), pos(0, 0)},
	}

	for i, c := range testCases {
		first := c.position.FirstDescendant()
		require.Equalf(t, c.expectedPos, first, "The first descentant position should match for test case %d", i)
	}

}

func TestLastDescendant(t *testing.T) {

	testCases := []struct {
		position    Position
		expectedPos Position
	}{
		{pos(0, 0), pos(0, 0)},
		{pos(1, 0), pos(1, 0)},
		{pos(0, 1), pos(1, 0)},
		{pos(2, 0), pos(2, 0)},
		{pos(2, 1), pos(3, 0)},
		{pos(0, 2), pos(3, 0)},
	}

	for i, c := range testCases {
		last := c.position.LastDescendant()
		require.Equalf(t, c.expectedPos, last, "The first descentant position should match for test case %d", i)
	}

}
