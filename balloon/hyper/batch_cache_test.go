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

package hyper

import (
	"testing"

	"github.com/bbva/qed/crypto/hashing"
	"github.com/bbva/qed/util"
	"github.com/stretchr/testify/require"
)

func TestBatchCache(t *testing.T) {

	hasher := hashing.NewSha256Hasher()
	cache := NewBatchCache(1)

	key := hasher.Do([]byte("this should exist"))
	value := util.Uint64AsPaddedBytes(uint64(0), 32)
	batch := newEmptyBatchNode(32)
	batch.AddLeafAt(0, hasher.Do(key), key, value)

	pos := pos(0, 256)
	cache.Put(pos.Bytes(), batch.Serialize())

	cached, ok := cache.Get(pos.Bytes())
	require.True(t, ok)
	require.Equal(t, batch, parseBatchNode(32, cached))

}
