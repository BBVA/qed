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
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/bbva/qed/crypto/hashing"
	"github.com/bbva/qed/log"
	"github.com/bbva/qed/util"
)

type AuditPath map[[keySize]byte]hashing.Digest

func (p AuditPath) Get(pos []byte) ([]byte, bool) {
	var buff [keySize]byte
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
		var key [keySize]byte
		copy(key[:8], util.Uint64AsBytes(uint64(index)))
		copy(key[8:], util.Uint16AsBytes(uint16(height)))
		parsed[key] = v
	}
	return parsed
}

type MembershipProof struct {
	AuditPath      AuditPath
	Index, Version uint64
	hasher         hashing.Hasher // TODO should we remove this and pass as an argument when verifying?
}

func NewMembershipProof(index, version uint64, auditPath AuditPath, hasher hashing.Hasher) *MembershipProof {
	return &MembershipProof{
		AuditPath: auditPath,
		Index:     index,
		Version:   version,
		hasher:    hasher,
	}
}

// Verify verifies a membership proof
func (p MembershipProof) Verify(eventDigest []byte, expectedRootHash hashing.Digest) (correct bool) {

	log.Debugf("Verifying membership proof for index %d and version %d", p.Index, p.Version)

	// build a visitable pruned tree and then visit it to recompute root hash
	visitor := newComputeHashVisitor(p.hasher, p.AuditPath)
	recomputed := pruneToVerify(p.Index, p.Version, eventDigest).Accept(visitor)

	return bytes.Equal(recomputed, expectedRootHash)
}

type IncrementalProof struct {
	AuditPath                AuditPath
	StartVersion, EndVersion uint64
	hasher                   hashing.Hasher
}

func NewIncrementalProof(start, end uint64, auditPath AuditPath, hasher hashing.Hasher) *IncrementalProof {
	return &IncrementalProof{
		AuditPath:    auditPath,
		StartVersion: start,
		EndVersion:   end,
		hasher:       hasher,
	}
}

func (p IncrementalProof) Verify(startDigest, endDigest hashing.Digest) (correct bool) {

	log.Debugf("Verifying incremental proof between versions %d and %d", p.StartVersion, p.EndVersion)

	// build two visitable pruned trees and then visit them to recompute root hash
	visitor := newComputeHashVisitor(p.hasher, p.AuditPath)
	startRecomputed := pruneToVerifyIncrementalStart(p.StartVersion).Accept(visitor)
	endRecomputed := pruneToVerifyIncrementalEnd(p.StartVersion, p.EndVersion).Accept(visitor)

	return bytes.Equal(startRecomputed, startDigest) && bytes.Equal(endRecomputed, endDigest)

}
