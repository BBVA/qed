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

	"github.com/bbva/qed/hashing"
	"github.com/stretchr/testify/assert"
)

func TestAuditPathSerialization(t *testing.T) {

	testCases := []struct {
		path     AuditPath
		expected map[string]hashing.Digest
	}{
		{
			AuditPath{},
			map[string]hashing.Digest{},
		},
		{
			AuditPath{
				NewPosition(0, 0).FixedBytes(): []byte{0x0},
			},
			map[string]hashing.Digest{
				"0|0": []byte{0x0},
			},
		},
		{
			AuditPath{
				NewPosition(0, 0).FixedBytes(): []byte{0x0},
				NewPosition(2, 1).FixedBytes(): []byte{0x1},
				NewPosition(4, 2).FixedBytes(): []byte{0x2},
			},
			map[string]hashing.Digest{
				"0|0": []byte{0x0},
				"2|1": []byte{0x1},
				"4|2": []byte{0x2},
			},
		},
	}

	for i, c := range testCases {
		serialized := c.path.Serialize()
		assert.Equalf(t, c.expected, serialized, "The serialized paths should match for test case %d", i)
	}

}

func TestParseAuditPath(t *testing.T) {

	testCases := []struct {
		path     map[string]hashing.Digest
		expected AuditPath
	}{
		{
			map[string]hashing.Digest{},
			AuditPath{},
		},
		{
			map[string]hashing.Digest{
				"0|0": []byte{0x0},
			},
			AuditPath{
				NewPosition(0, 0).FixedBytes(): []byte{0x0},
			},
		},
		{
			map[string]hashing.Digest{
				"0|0": []byte{0x0},
				"2|1": []byte{0x1},
				"4|2": []byte{0x2},
			},
			AuditPath{
				NewPosition(0, 0).FixedBytes(): []byte{0x0},
				NewPosition(2, 1).FixedBytes(): []byte{0x1},
				NewPosition(4, 2).FixedBytes(): []byte{0x2},
			},
		},
	}

	for i, c := range testCases {
		parsed := ParseAuditPath(c.path)
		assert.Equalf(t, c.expected, parsed, "The parsed paths should match for test case %d", i)
	}

}
