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
	"fmt"
	"strconv"
	"strings"

	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/util"
)

type AuditPath map[[KeySize]byte]hashing.Digest

func (p AuditPath) Get(pos []byte) ([]byte, bool) {
	var buff [KeySize]byte
	copy(buff[:], pos)
	digest, ok := p[buff]
	return digest, ok
}

func (p AuditPath) Serialize() map[string]hashing.Digest {
	s := make(map[string]hashing.Digest, len(p))
	for k, v := range p {
		s[fmt.Sprintf("%d|%d", util.BytesAsUint64(k[:8]), util.BytesAsUint16(k[8:]))] = v
	}
	return s
}

func ParseAuditPath(serialized map[string]hashing.Digest) AuditPath {
	parsed := make(AuditPath, len(serialized))
	for k, v := range serialized {
		tokens := strings.Split(k, "|")
		index, _ := strconv.Atoi(tokens[0])
		height, _ := strconv.Atoi(tokens[1])
		var key [KeySize]byte
		copy(key[:8], util.Uint64AsBytes(uint64(index)))
		copy(key[8:], util.Uint16AsBytes(uint16(height)))
		parsed[key] = v
	}
	return parsed
}
