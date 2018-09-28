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

package cache

import (
	"testing"

	"github.com/bbva/qed/balloon/navigator"
	"github.com/bbva/qed/hashing"
	"github.com/stretchr/testify/require"

	"github.com/bbva/qed/storage"
	storage_utils "github.com/bbva/qed/testutils/storage"
)

func TestPassThroughCache(t *testing.T) {

	testCases := []struct {
		pos    navigator.Position
		value  hashing.Digest
		cached bool
	}{
		{&navigator.FakePosition{[]byte{0x0}, 0}, hashing.Digest{0x1}, true},
		{&navigator.FakePosition{[]byte{0x1}, 0}, hashing.Digest{0x2}, true},
		{&navigator.FakePosition{[]byte{0x2}, 0}, hashing.Digest{0x3}, false},
	}

	store, closeF := storage_utils.OpenBPlusTreeStore()
	defer closeF()
	prefix := byte(0x0)
	cache := NewPassThroughCache(prefix, store)

	for i, c := range testCases {
		if c.cached {
			err := store.Mutate([]*storage.Mutation{
				{prefix, c.pos.Bytes(), c.value},
			})
			require.NoError(t, err)
		}

		cachedValue, ok := cache.Get(c.pos)

		if c.cached {
			require.Truef(t, ok, "The key should exists in cache in test case %d", i)
			require.Equalf(t, c.value, cachedValue, "The cached value should be equal to stored value in test case %d", i)
		} else {
			require.Falsef(t, ok, "The key should not exist in cache in test case %d", i)
		}
	}

}
