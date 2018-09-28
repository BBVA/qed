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

package balloon

import (
	"fmt"
	"sync"

	"github.com/bbva/qed/balloon/cache"
	"github.com/bbva/qed/balloon/history"
	"github.com/bbva/qed/balloon/hyper"
	"github.com/bbva/qed/balloon/visitor"
	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/storage"
	"github.com/bbva/qed/util"
)

var (
	BalloonVersionKey = []byte("version")
)

type Balloon struct {
	version uint64
	hasherF func() hashing.Hasher
	store   storage.Store

	historyTree *history.HistoryTree
	hyperTree   *hyper.HyperTree
	hasher      hashing.Hasher
}

func NewBalloon(store storage.Store, hasherF func() hashing.Hasher) (*Balloon, error) {

	// create caches
	historyCache := cache.NewFIFOReadThroughCache(storage.HistoryCachePrefix, store, 30)
	hyperCache := cache.NewSimpleCache(1 << 2)

	// create trees
	historyTree := history.NewHistoryTree(hasherF, historyCache)
	hyperTree := hyper.NewHyperTree(hasherF, store, hyperCache)

	balloon := &Balloon{
		version:     0,
		hasherF:     hasherF,
		store:       store,
		historyTree: historyTree,
		hyperTree:   hyperTree,
		hasher:      hasherF(),
	}

	// update version
	balloon.RefreshVersion()

	return balloon, nil
}

// Commitment is the struct that has both history and hyper digest and the
// current version for that rootNode digests.
type Commitment struct {
	HistoryDigest hashing.Digest
	HyperDigest   hashing.Digest
	Version       uint64
}

// MembershipProof is the struct that is required to make a Exisitance Proof.
// It has both Hyper and History AuditPaths, if it exists in first place and
// Current, Actual and Query Versions.
type MembershipProof struct {
	Exists         bool
	HyperProof     visitor.Verifiable
	HistoryProof   visitor.Verifiable
	CurrentVersion uint64
	QueryVersion   uint64
	ActualVersion  uint64 //required for consistency proof
	KeyDigest      hashing.Digest
	Hasher         hashing.Hasher
}

func NewMembershipProof(
	exists bool,
	hyperProof, historyProof visitor.Verifiable,
	currentVersion, queryVersion, actualVersion uint64,
	keyDigest hashing.Digest,
	Hasher hashing.Hasher) *MembershipProof {
	return &MembershipProof{
		exists,
		hyperProof,
		historyProof,
		currentVersion,
		queryVersion,
		actualVersion,
		keyDigest,
		Hasher,
	}
}

// Verify verifies a proof and answer from QueryMembership. Returns true if the
// answer and proof are correct and consistent, otherwise false.
// Run by a client on input that should be verified.
func (p MembershipProof) Verify(event []byte, commitment *Commitment) bool {
	if p.HyperProof == nil || p.HistoryProof == nil {
		return false
	}

	digest := p.Hasher.Do(event)
	hyperCorrect := p.HyperProof.Verify(digest, commitment.HyperDigest)

	if p.Exists {
		if p.QueryVersion <= p.ActualVersion {
			historyCorrect := p.HistoryProof.Verify(digest, commitment.HistoryDigest)
			return hyperCorrect && historyCorrect
		}
	}

	return hyperCorrect
}

type IncrementalProof struct {
	Start, End uint64
	AuditPath  visitor.AuditPath
	Hasher     hashing.Hasher
}

func NewIncrementalProof(
	start, end uint64,
	auditPath visitor.AuditPath,
	hasher hashing.Hasher,
) *IncrementalProof {
	return &IncrementalProof{
		start,
		end,
		auditPath,
		hasher,
	}
}

func (p IncrementalProof) Verify(commitmentStart, commitmentEnd *Commitment) bool {
	ip := history.NewIncrementalProof(p.Start, p.End, p.AuditPath, p.Hasher)
	return ip.Verify(commitmentStart.HistoryDigest, commitmentEnd.HistoryDigest)
}

func (b Balloon) Version() uint64 {
	return b.version
}

func (b *Balloon) RefreshVersion() error {
	// get last stored version
	kv, err := b.store.GetLast(storage.HistoryCachePrefix)
	if err != nil {
		if err != storage.ErrKeyNotFound {
			return err
		}
	} else {
		b.version = util.BytesAsUint64(kv.Key[:8]) + 1
	}
	return nil
}

func (b *Balloon) Add(event []byte) (*Commitment, []*storage.Mutation, error) {

	// Get version
	version := b.version
	b.version++

	// Hash event
	eventDigest := b.hasher.Do(event)

	// Update trees
	var historyDigest hashing.Digest
	var historyMutations []*storage.Mutation
	var historyErr error
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		historyDigest, historyMutations, historyErr = b.historyTree.Add(eventDigest, version)
		wg.Done()
	}()

	hyperDigest, mutations, hyperErr := b.hyperTree.Add(eventDigest, version)

	wg.Wait()

	if historyErr != nil {
		return nil, nil, historyErr
	}
	if hyperErr != nil {
		return nil, nil, hyperErr
	}

	// Append trees mutations
	mutations = append(mutations, historyMutations...)

	commitment := &Commitment{
		HistoryDigest: historyDigest,
		HyperDigest:   hyperDigest,
		Version:       version,
	}

	return commitment, mutations, nil
}

func (b Balloon) QueryMembership(event []byte, version uint64) (*MembershipProof, error) {

	var proof MembershipProof

	proof.Hasher = b.hasherF()
	proof.KeyDigest = b.hasher.Do(event)
	proof.QueryVersion = version
	proof.CurrentVersion = b.version - 1

	hyperProof, err := b.hyperTree.QueryMembership(proof.KeyDigest)
	if err != nil {
		return nil, fmt.Errorf("Unable to get proof from hyper tree: %v", err)
	}
	proof.HyperProof = hyperProof

	if len(hyperProof.Value) > 0 {
		proof.Exists = true
		proof.ActualVersion = util.BytesAsUint64(hyperProof.Value)
	}

	if proof.Exists && proof.ActualVersion <= proof.QueryVersion {
		historyProof, err := b.historyTree.ProveMembership(proof.ActualVersion, proof.QueryVersion)
		if err != nil {
			return nil, fmt.Errorf("Unable to get proof from history tree: %v", err)
		}
		proof.HistoryProof = historyProof
	} else {
		proof.Exists = false
	}

	return &proof, nil
}

func (b Balloon) QueryConsistency(start, end uint64) (*IncrementalProof, error) {

	var proof IncrementalProof

	proof.Start = start
	proof.End = end
	proof.Hasher = b.hasherF()

	historyProof, err := b.historyTree.ProveConsistency(start, end)
	if err != nil {
		return nil, fmt.Errorf("Unable to get proof from history tree: %v", err)
	}
	proof.AuditPath = historyProof.AuditPath

	return &proof, nil
}

func (b *Balloon) Close() {
	b.historyTree.Close()
	b.hyperTree.Close()
	b.historyTree = nil
	b.hyperTree = nil
	b.version = 0
}
