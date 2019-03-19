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
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRoot(t *testing.T) {

	testCases := []struct {
		version     uint64
		expectedPos *position
	}{
		{0, newPosition(0, 0)},
		{1, newPosition(0, 1)},
		{2, newPosition(0, 2)},
		{3, newPosition(0, 2)},
		{4, newPosition(0, 3)},
		{5, newPosition(0, 3)},
		{6, newPosition(0, 3)},
		{7, newPosition(0, 3)},
		{8, newPosition(0, 4)},
	}

	for i, c := range testCases {
		rootPos := newRootPosition(c.version)
		require.Equalf(t, c.expectedPos, rootPos, "The root position should match in test case %d", i)
	}

}

func TestIsLeaf(t *testing.T) {

	testCases := []struct {
		pos *position
		ok  bool
	}{
		{newPosition(0, 0), true},
		{newPosition(0, 1), false},
		{newPosition(1, 1), false},
		{newPosition(2, 0), true},
	}

	for i, c := range testCases {
		result := c.pos.IsLeaf()
		require.Equalf(t, c.ok, result, "The leaf checking should match for test case %d", i)
	}

}

func TestLeft(t *testing.T) {

	testCases := []struct {
		pos          *position
		expectedLeft *position
	}{
		{newPosition(0, 0), nil},
		{newPosition(0, 0), nil},
		{newPosition(1, 0), nil},
		{newPosition(0, 1), newPosition(0, 0)},
		{newPosition(0, 0), nil},
		{newPosition(1, 0), nil},
		{newPosition(2, 0), nil},
		{newPosition(0, 1), newPosition(0, 0)},
		{newPosition(2, 1), newPosition(2, 0)}, // TODO check invalid positions like (1,1)?
		{newPosition(0, 0), nil},
		{newPosition(1, 0), nil},
		{newPosition(2, 0), nil},
		{newPosition(0, 1), newPosition(0, 0)},
		{newPosition(2, 1), newPosition(2, 0)},
		{newPosition(0, 2), newPosition(0, 1)},
	}

	for i, c := range testCases {
		left := c.pos.Left()
		require.Equalf(t, c.expectedLeft, left, "The left positions should match for test case %d", i)
	}
}

func TestRight(t *testing.T) {

	testCases := []struct {
		pos           *position
		expectedRight *position
	}{
		{newPosition(0, 0), nil},
		{newPosition(0, 0), nil},
		{newPosition(1, 0), nil},
		{newPosition(0, 1), newPosition(1, 0)},
		{newPosition(0, 0), nil},
		{newPosition(1, 0), nil},
		{newPosition(2, 0), nil},
		{newPosition(0, 1), newPosition(1, 0)},
		{newPosition(2, 1), newPosition(3, 0)},
		{newPosition(0, 0), nil},
		{newPosition(1, 0), nil},
		{newPosition(2, 0), nil},
		{newPosition(0, 1), newPosition(1, 0)},
		{newPosition(2, 1), newPosition(3, 0)},
		{newPosition(0, 2), newPosition(2, 1)},
	}

	for i, c := range testCases {
		right := c.pos.Right()
		require.Equalf(t, c.expectedRight, right, "The right positions should match for test case %d", i)
	}
}

func TestFirstDescendant(t *testing.T) {

	testCases := []struct {
		pos         *position
		expectedPos *position
	}{
		{newPosition(0, 0), newPosition(0, 0)},
		{newPosition(1, 0), newPosition(1, 0)},
		{newPosition(0, 1), newPosition(0, 0)},
		{newPosition(2, 0), newPosition(2, 0)},
		{newPosition(2, 1), newPosition(2, 0)},
		{newPosition(0, 2), newPosition(0, 0)},
	}

	for i, c := range testCases {
		first := c.pos.FirstDescendant()
		require.Equalf(t, c.expectedPos, first, "The first descentant position should match for test case %d", i)
	}

}

func TestLastDescendant(t *testing.T) {

	testCases := []struct {
		pos         *position
		expectedPos *position
	}{
		{newPosition(0, 0), newPosition(0, 0)},
		{newPosition(1, 0), newPosition(1, 0)},
		{newPosition(0, 1), newPosition(1, 0)},
		{newPosition(2, 0), newPosition(2, 0)},
		{newPosition(2, 1), newPosition(3, 0)},
		{newPosition(0, 2), newPosition(3, 0)},
	}

	for i, c := range testCases {
		last := c.pos.LastDescendant()
		require.Equalf(t, c.expectedPos, last, "The first descentant position should match for test case %d", i)
	}

}
