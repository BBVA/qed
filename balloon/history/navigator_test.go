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

	"github.com/bbva/qed/balloon/navigator"
	"github.com/stretchr/testify/require"
)

func TestRoot(t *testing.T) {

	testCases := []struct {
		version     uint64
		expectedPos navigator.Position
	}{
		{0, NewPosition(0, 0)},
		{1, NewPosition(0, 1)},
		{2, NewPosition(0, 2)},
		{3, NewPosition(0, 2)},
		{4, NewPosition(0, 3)},
		{5, NewPosition(0, 3)},
		{6, NewPosition(0, 3)},
		{7, NewPosition(0, 3)},
		{8, NewPosition(0, 4)},
	}

	for i, c := range testCases {
		navigator := NewHistoryTreeNavigator(c.version)
		rootPos := navigator.Root()
		require.Equalf(t, c.expectedPos, rootPos, "The root position should match in test case %d", i)
	}

}

func TestIsLeaf(t *testing.T) {

	testCases := []struct {
		position navigator.Position
		ok       bool
	}{
		{NewPosition(0, 0), true},
		{NewPosition(0, 1), false},
		{NewPosition(1, 1), false},
		{NewPosition(2, 0), true},
	}

	navigator := NewHistoryTreeNavigator(7)
	for i, c := range testCases {
		result := navigator.IsLeaf(c.position)
		require.Equalf(t, c.ok, result, "The leaf checking should match for test case %d", i)
	}

}

func TestIsRoot(t *testing.T) {

	testCases := []struct {
		version  uint64
		position navigator.Position
		ok       bool
	}{
		{0, NewPosition(0, 0), true},
		{0, NewPosition(0, 1), false},
		{0, NewPosition(1, 1), false},
		{0, NewPosition(2, 0), false},
		{1, NewPosition(0, 0), false},
		{1, NewPosition(0, 1), true},
		{1, NewPosition(2, 0), false},
		{2, NewPosition(0, 0), false},
		{2, NewPosition(0, 1), false},
		{2, NewPosition(2, 0), false},
		{2, NewPosition(0, 2), true},
	}

	for i, c := range testCases {
		navigator := NewHistoryTreeNavigator(c.version)
		result := navigator.IsRoot(c.position)
		require.Equalf(t, c.ok, result, "The root checking should match for test case %d", i)
	}
}

func TestGoToLeft(t *testing.T) {

	testCases := []struct {
		version      uint64
		position     navigator.Position
		expectedLeft navigator.Position
	}{
		{0, NewPosition(0, 0), nil},
		{1, NewPosition(0, 0), nil},
		{1, NewPosition(1, 0), nil},
		{1, NewPosition(0, 1), NewPosition(0, 0)},
		{2, NewPosition(0, 0), nil},
		{2, NewPosition(1, 0), nil},
		{2, NewPosition(2, 0), nil},
		{2, NewPosition(0, 1), NewPosition(0, 0)},
		{2, NewPosition(2, 1), NewPosition(2, 0)}, // TODO check invalid positions like (1,1)?
		{3, NewPosition(0, 0), nil},
		{3, NewPosition(1, 0), nil},
		{3, NewPosition(2, 0), nil},
		{3, NewPosition(0, 1), NewPosition(0, 0)},
		{3, NewPosition(2, 1), NewPosition(2, 0)},
		{3, NewPosition(0, 2), NewPosition(0, 1)},
	}

	for i, c := range testCases {
		navigator := NewHistoryTreeNavigator(c.version)
		left := navigator.GoToLeft(c.position)
		require.Equalf(t, c.expectedLeft, left, "The left positions should match for test case %d", i)
	}
}

func TestGoToRight(t *testing.T) {

	testCases := []struct {
		version       uint64
		position      navigator.Position
		expectedRight navigator.Position
	}{
		{0, NewPosition(0, 0), nil},
		{1, NewPosition(0, 0), nil},
		{1, NewPosition(1, 0), nil},
		{1, NewPosition(0, 1), NewPosition(1, 0)},
		{2, NewPosition(0, 0), nil},
		{2, NewPosition(1, 0), nil},
		{2, NewPosition(2, 0), nil},
		{2, NewPosition(0, 1), NewPosition(1, 0)},
		{2, NewPosition(2, 1), nil},
		{3, NewPosition(0, 0), nil},
		{3, NewPosition(1, 0), nil},
		{3, NewPosition(2, 0), nil},
		{3, NewPosition(0, 1), NewPosition(1, 0)},
		{3, NewPosition(2, 1), NewPosition(3, 0)},
		{3, NewPosition(0, 2), NewPosition(2, 1)},
	}

	for i, c := range testCases {
		navigator := NewHistoryTreeNavigator(c.version)
		right := navigator.GoToRight(c.position)
		require.Equalf(t, c.expectedRight, right, "The right positions should match for test case %d", i)
	}
}

func TestDescendToFirst(t *testing.T) {

	testCases := []struct {
		version       uint64
		position      navigator.Position
		expectedFirst navigator.Position
	}{
		{0, NewPosition(0, 0), nil},
		{1, NewPosition(0, 0), nil},
		{1, NewPosition(1, 0), nil},
		{1, NewPosition(0, 1), NewPosition(0, 0)},
		{2, NewPosition(0, 0), nil},
		{2, NewPosition(1, 0), nil},
		{2, NewPosition(2, 0), nil},
		{2, NewPosition(0, 1), NewPosition(0, 0)},
		{2, NewPosition(2, 1), NewPosition(2, 0)},
		{3, NewPosition(0, 0), nil},
		{3, NewPosition(1, 0), nil},
		{3, NewPosition(2, 0), nil},
		{3, NewPosition(0, 1), NewPosition(0, 0)},
		{3, NewPosition(2, 1), NewPosition(2, 0)},
		{3, NewPosition(0, 2), NewPosition(0, 0)},
	}

	for i, c := range testCases {
		navigator := NewHistoryTreeNavigator(c.version)
		first := navigator.DescendToFirst(c.position)
		require.Equalf(t, c.expectedFirst, first, "The first positions should match for test case %d", i)
	}
}

func TestDescendToLast(t *testing.T) {

	testCases := []struct {
		version      uint64
		position     navigator.Position
		expectedLast navigator.Position
	}{
		{0, NewPosition(0, 0), nil},
		{1, NewPosition(0, 0), nil},
		{1, NewPosition(1, 0), nil},
		{1, NewPosition(0, 1), NewPosition(1, 0)},
		{2, NewPosition(0, 0), nil},
		{2, NewPosition(1, 0), nil},
		{2, NewPosition(2, 0), nil},
		{2, NewPosition(0, 1), NewPosition(1, 0)},
		{2, NewPosition(2, 1), nil},
		{3, NewPosition(0, 0), nil},
		{3, NewPosition(1, 0), nil},
		{3, NewPosition(2, 0), nil},
		{3, NewPosition(0, 1), NewPosition(1, 0)},
		{3, NewPosition(2, 1), NewPosition(3, 0)},
		{3, NewPosition(0, 2), NewPosition(3, 0)},
	}

	for i, c := range testCases {
		navigator := NewHistoryTreeNavigator(c.version)
		last := navigator.DescendToLast(c.position)
		require.Equalf(t, c.expectedLast, last, "The last positions should match for test case %d", i)
	}
}
