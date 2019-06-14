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
	"github.com/bbva/qed/balloon/cache"
	"github.com/bbva/qed/crypto/hashing"
	"github.com/bbva/qed/storage"
)

type HistoryTree struct {
	hasherF    func() hashing.Hasher
	hasher     hashing.Hasher
	writeCache cache.ModifiableCache
	readCache  cache.Cache
}

func NewHistoryTree(hasherF func() hashing.Hasher, store storage.Store, cacheSize uint16) *HistoryTree {

	// create cache for Adding
	writeCache := cache.NewLruReadThroughCache(storage.HistoryTable, store, cacheSize)

	// create cache for Membership and Incremental
	readCache := cache.NewPassThroughCache(storage.HistoryTable, store)

	return &HistoryTree{
		hasherF:    hasherF,
		hasher:     hasherF(),
		writeCache: writeCache,
		readCache:  readCache,
	}
}

func (t *HistoryTree) Add(eventDigest hashing.Digest, version uint64) (hashing.Digest, []*storage.Mutation, error) {

	// log.Debugf("Adding new event digest %x with version %d", eventDigest, version)

	// build a visitable pruned tree and then visit it to generate the root hash
	visitor := newInsertVisitor(t.hasher, t.writeCache, storage.HistoryTable)
	rh := pruneToInsert(version, eventDigest).Accept(visitor)

	return rh, visitor.Result(), nil
}

func (t *HistoryTree) AddBulk(eventDigests []hashing.Digest, initialVersion uint64) ([]hashing.Digest, []*storage.Mutation, error) {

	visitor := newInsertVisitor(t.hasher, t.writeCache, storage.HistoryTable)

	rootHashes := make([]hashing.Digest, 0)
	for i, e := range eventDigests {
		rootHashes = append(rootHashes, pruneToInsert(initialVersion+uint64(i), e).Accept(visitor))
	}

	return rootHashes, visitor.Result(), nil

}

func (t *HistoryTree) ProveMembership(index, version uint64) (*MembershipProof, error) {

	//log.Debugf("Proving membership for index %d with version %d", index, version)

	// build a visitable pruned tree and then visit it to collect the audit path
	visitor := newAuditPathVisitor(t.hasherF(), t.readCache)
	if index == version {
		pruneToFind(index).Accept(visitor) // faster pruning
	} else {
		pruneToFindConsistent(index, version).Accept(visitor)
	}

	proof := NewMembershipProof(index, version, visitor.Result(), t.hasherF())
	return proof, nil
}

func (t *HistoryTree) ProveConsistency(start, end uint64) (*IncrementalProof, error) {

	//log.Debugf("Proving consistency between versions %d and %d", start, end)

	// build a visitable pruned tree and then visit it to collect the audit path
	visitor := newAuditPathVisitor(t.hasherF(), t.readCache)
	pruneToCheckConsistency(start, end).Accept(visitor)

	proof := NewIncrementalProof(start, end, visitor.Result(), t.hasherF())

	return proof, nil
}

func (t *HistoryTree) Close() {
	t.hasher = nil
	t.writeCache = nil
	t.readCache = nil
}
