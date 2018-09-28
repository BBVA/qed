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

func TestSingleTargetedCacheResolver(t *testing.T) {

	testCases := []struct {
		version  uint64
		position navigator.Position
		ok       bool
	}{
		{0, NewPosition(0, 0), false},

		{1, NewPosition(0, 0), true},
		{1, NewPosition(1, 0), false},
		{1, NewPosition(0, 1), false},

		{2, NewPosition(0, 0), true},
		{2, NewPosition(1, 0), true},
		{2, NewPosition(2, 0), false},
		{2, NewPosition(0, 1), true},
		{2, NewPosition(0, 2), false},

		{3, NewPosition(0, 0), true},
		{3, NewPosition(1, 0), true},
		{3, NewPosition(2, 0), true},
		{3, NewPosition(3, 0), false},
		{3, NewPosition(0, 1), true},
		{3, NewPosition(0, 2), false},
		{3, NewPosition(2, 1), false},

		{4, NewPosition(0, 0), true},
		{4, NewPosition(1, 0), true},
		{4, NewPosition(2, 0), true},
		{4, NewPosition(3, 0), true},
		{4, NewPosition(4, 0), false},
		{4, NewPosition(0, 1), true},
		{4, NewPosition(0, 2), true},
		{4, NewPosition(0, 3), false},
	}

	for i, c := range testCases {
		resolver := NewSingleTargetedCacheResolver(c.version)
		result := resolver.ShouldGetFromCache(c.position)
		require.Equalf(t, c.ok, result, "The result should match for test case %d", i)
	}

}

func TestNewDoubleTargetedCacheResolver(t *testing.T) {

	testCases := []struct {
		start, end uint64
		position   navigator.Position
		ok         bool
	}{
		{0, 0, NewPosition(0, 0), false},

		{0, 1, NewPosition(0, 0), false},
		{0, 1, NewPosition(1, 0), true},
		{0, 1, NewPosition(0, 1), false},

		{0, 2, NewPosition(0, 0), false},
		{0, 2, NewPosition(1, 0), true},
		{0, 2, NewPosition(0, 1), false},
		{0, 2, NewPosition(2, 0), true},

		{0, 3, NewPosition(0, 0), false},
		{0, 3, NewPosition(1, 0), true},
		{0, 3, NewPosition(0, 1), false},
		{0, 3, NewPosition(2, 0), true},
		{0, 3, NewPosition(3, 0), true},

		{0, 4, NewPosition(0, 0), false},
		{0, 4, NewPosition(1, 0), true},
		{0, 4, NewPosition(2, 1), true},
		{0, 4, NewPosition(4, 0), true},
	}

	for i, c := range testCases {
		resolver := NewDoubleTargetedCacheResolver(c.start, c.end)
		result := resolver.ShouldGetFromCache(c.position)
		require.Equalf(t, c.ok, result, "The result should match for test case %d", i)
	}

}

func TestIncrementalCacheResolver(t *testing.T) {

	testCases := []struct {
		start, end uint64
		position   navigator.Position
		ok         bool
	}{
		{0, 0, NewPosition(0, 0), true},

		{0, 1, NewPosition(0, 0), true},
		{0, 1, NewPosition(1, 0), true},
		{0, 1, NewPosition(0, 1), false},

		{0, 2, NewPosition(0, 0), true},
		{0, 2, NewPosition(1, 0), true},
		{0, 2, NewPosition(0, 1), false},
		{0, 2, NewPosition(2, 0), true},

		{0, 3, NewPosition(0, 0), true},
		{0, 3, NewPosition(1, 0), true},
		{0, 3, NewPosition(0, 1), false},
		{0, 3, NewPosition(2, 0), true},
		{0, 3, NewPosition(3, 0), true},

		{0, 4, NewPosition(0, 0), true},
		{0, 4, NewPosition(1, 0), true},
		{0, 4, NewPosition(2, 1), true},
		{0, 4, NewPosition(4, 0), true},
	}

	for i, c := range testCases {
		resolver := NewIncrementalCacheResolver(c.start, c.end)
		result := resolver.ShouldGetFromCache(c.position)
		require.Equalf(t, c.ok, result, "The result should match for test case %d", i)
	}

}
