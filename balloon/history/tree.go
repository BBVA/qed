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

// Package history implements the history tree (a merkel tree, append only structure)
// life cycle, its operations, different visitors to navigate the tree, as well as
// the functionality of request and verify membership and incremental proofs.
package history

import (
	"github.com/bbva/qed/balloon/cache"
	"github.com/bbva/qed/crypto/hashing"
	"github.com/bbva/qed/log"
	"github.com/bbva/qed/storage"
)

type HistoryTree struct {
	hasherF    func() hashing.Hasher
	hasher     hashing.Hasher
	writeCache cache.ModifiableCache
	readCache  cache.Cache

	log log.Logger
}

func NewHistoryTree(hasherF func() hashing.Hasher, store storage.Store, cacheSize uint16) *HistoryTree {
	return NewHistoryTreeWithLogger(hasherF, store, cacheSize, log.L())
}

func NewHistoryTreeWithLogger(hasherF func() hashing.Hasher, store storage.Store, cacheSize uint16, logger log.Logger) *HistoryTree {

	// create cache for Adding
	writeCache := cache.NewLruReadThroughCache(storage.HistoryTable, store, cacheSize)

	// create cache for Membership and Incremental
	readCache := cache.NewPassThroughCache(storage.HistoryTable, store)

	return &HistoryTree{
		hasherF:    hasherF,
		hasher:     hasherF(),
		writeCache: writeCache,
		readCache:  readCache,
		log:        logger,
	}
}

// Add function adds an event digest into the history tree.
// It builds an insert visitor, calculates the expected root hash, and returns it along
// with the storage mutations to be done at balloon level.
func (t *HistoryTree) Add(eventDigest hashing.Digest, version uint64) (hashing.Digest, []*storage.Mutation, error) {

	// t.log.Tracef("Adding new event digest %x with version %d", eventDigest, version)

	// build a visitable pruned tree and then visit it to generate the root hash
	visitor := newInsertVisitor(t.hasher, t.writeCache, storage.HistoryTable)
	rh := pruneToInsert(version, eventDigest).Accept(visitor)

	return rh, visitor.Result(), nil
}

// AddBulk function adds a bulk of event digests (one after another) into the history tree.
// It builds an insert visitor, calculates the expected bulk of root hashes, and returns them along
// with the storage mutations to be done at balloon level.
func (t *HistoryTree) AddBulk(eventDigests []hashing.Digest, initialVersion uint64) ([]hashing.Digest, []*storage.Mutation, error) {

	visitor := newInsertVisitor(t.hasher, t.writeCache, storage.HistoryTable)

	rootHashes := make([]hashing.Digest, 0)
	for i, e := range eventDigests {
		rootHashes = append(rootHashes, pruneToInsert(initialVersion+uint64(i), e).Accept(visitor))
	}

	return rootHashes, visitor.Result(), nil

}

// ProveMembership function builds the membership proof of the given index against the given
// version. It builds an audit-path visitor to build the proof.
func (t *HistoryTree) ProveMembership(index, version uint64) (*MembershipProof, error) {

	//t.log.Tracef("Proving membership for index %d with version %d", index, version)

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

// ProveConsistency function builds the incremental proof between the given event versions.
// It builds an audit-path visitor to build the proof.
func (t *HistoryTree) ProveConsistency(start, end uint64) (*IncrementalProof, error) {

	//t.log.Tracef("Proving consistency between versions %d and %d", start, end)

	// build a visitable pruned tree and then visit it to collect the audit path
	visitor := newAuditPathVisitor(t.hasherF(), t.readCache)
	pruneToCheckConsistency(start, end).Accept(visitor)

	proof := NewIncrementalProof(start, end, visitor.Result(), t.hasherF())

	return proof, nil
}

// Close function resets history tree's write and read caches, and hasher.
func (t *HistoryTree) Close() {
	t.hasher = nil
	t.writeCache = nil
	t.readCache = nil
}
