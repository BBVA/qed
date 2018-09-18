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

import (
	"testing"

	"github.com/bbva/qed/balloon/common"
	"github.com/stretchr/testify/require"
)

func TestShouldBeInCache(t *testing.T) {
	testCases := []struct {
		testname       string
		numBits        uint16
		cacheLevel     uint16
		targetKey      []byte
		position       common.Position
		expectedResult bool
	}{
		{"Position on path", 8, 3, []byte{0}, NewPosition([]byte{0}, 2), false},
		{"Height <= cacheLevel", 8, 3, []byte{0}, NewPosition([]byte{8}, 3), false},
		{"All conditions", 8, 3, []byte{0}, NewPosition([]byte{16}, 4), true},
	}

	for _, test := range testCases {
		cacheResolver := NewSingleTargetedCacheResolver(test.numBits, test.cacheLevel, test.targetKey)
		result := cacheResolver.ShouldBeInCache(test.position)
		require.Equalf(t, test.expectedResult, result, "Wrong shouldBeInCache in test case %s", test.testname)
	}
}

func TestShouldCache(t *testing.T) {
	testCases := []struct {
		numBits        uint16
		cacheLevel     uint16
		targetKey      []byte
		position       common.Position
		expectedResult bool
	}{
		{8, 3, []byte{0}, NewPosition([]byte{0}, 0), false},
		{8, 3, []byte{0}, NewPosition([]byte{0}, 3), false},
		{8, 3, []byte{0}, NewPosition([]byte{0}, 4), true},
	}

	for i, test := range testCases {
		cacheResolver := NewSingleTargetedCacheResolver(test.numBits, test.cacheLevel, test.targetKey)
		result := cacheResolver.ShouldCache(test.position)
		require.Equalf(t, test.expectedResult, result, "Wrong shouldCache in test case %d", i)
	}
}

func TestIsOnPath(t *testing.T) {
	testCases := []struct {
		numBits        uint16
		cacheLevel     uint16
		targetKey      []byte
		position       common.Position
		expectedResult bool
	}{
		{8, 3, []byte{0}, NewPosition([]byte{0}, 1), true},
		{8, 3, []byte{0}, NewPosition([]byte{0}, 2), true},
		{8, 3, []byte{0}, NewPosition([]byte{4}, 2), false},
		{8, 3, []byte{0}, NewPosition([]byte{0}, 3), true},
		{8, 3, []byte{0}, NewPosition([]byte{8}, 3), false},
		{8, 3, []byte{0}, NewPosition([]byte{0}, 4), true},
		{8, 3, []byte{0}, NewPosition([]byte{16}, 4), false},
		{8, 3, []byte{0}, NewPosition([]byte{0}, 5), true},
		{8, 3, []byte{0}, NewPosition([]byte{32}, 5), false},
		{8, 3, []byte{0}, NewPosition([]byte{0}, 6), true},
		{8, 3, []byte{0}, NewPosition([]byte{64}, 6), false},
		{8, 3, []byte{0}, NewPosition([]byte{0}, 7), true},
	}
	for i, test := range testCases {
		cacheResolver := NewSingleTargetedCacheResolver(test.numBits, test.cacheLevel, test.targetKey)
		result := cacheResolver.IsOnPath(test.position)
		require.Equalf(t, test.expectedResult, result, "Wrong isOnPath in test case %d", i)
	}
}
