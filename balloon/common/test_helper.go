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

package common

import (
	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/storage"
)

type FakeCache struct {
	FixedDigest hashing.Digest
}

func NewFakeCache(fixedDigest hashing.Digest) *FakeCache {
	return &FakeCache{fixedDigest}
}

func (c FakeCache) Get(Position) (hashing.Digest, bool) {
	return hashing.Digest{0x0}, true
}

func (c *FakeCache) Put(pos Position, value hashing.Digest) {}

func (c *FakeCache) Fill(r storage.KVPairReader) error { return nil }

func (c FakeCache) Size() int { return 1 }
