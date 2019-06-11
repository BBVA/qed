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

package balloon

import (
	"errors"
	"fmt"
	"sync"

	"github.com/bbva/qed/balloon/cache"
	"github.com/bbva/qed/balloon/history"
	"github.com/bbva/qed/balloon/hyper"
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

	// create trees
	historyTree := history.NewHistoryTree(hasherF, store, 300)
	hyperTree := hyper.NewHyperTree(hasherF, store, cache.NewFreeCache(hyper.CacheSize))

	balloon := &Balloon{
		version:     0,
		hasherF:     hasherF,
		store:       store,
		historyTree: historyTree,
		hyperTree:   hyperTree,
		hasher:      hasherF(),
	}

	// update version
	err := balloon.RefreshVersion()
	if err != nil {
		return nil, err
	}

	return balloon, nil
}

// Snapshot is the struct that has both history and hyper digest and the
// current version for that rootNode digests.
type Snapshot struct {
	EventDigest   hashing.Digest
	HistoryDigest hashing.Digest
	HyperDigest   hashing.Digest
	Version       uint64
}

type Verifiable interface {
	Verify(key []byte, expectedDigest hashing.Digest) bool
}

// MembershipProof is the struct that is required to make a Exisitance Proof.
// It has both Hyper and History AuditPaths, if it exists in first place and
// burrent balloon version, event actual version and query version.
type MembershipProof struct {
	Exists         bool
	HyperProof     *hyper.QueryProof
	HistoryProof   *history.MembershipProof
	CurrentVersion uint64
	QueryVersion   uint64
	ActualVersion  uint64 //required for consistency proof
	KeyDigest      hashing.Digest
	Hasher         hashing.Hasher
}

func NewMembershipProof(exists bool, hyperProof *hyper.QueryProof, historyProof *history.MembershipProof, currentVersion, queryVersion, actualVersion uint64, keyDigest hashing.Digest, Hasher hashing.Hasher) *MembershipProof {

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

// DigestVerify verifies a proof and answer from QueryMembership. Returns true if the
// answer and proof are correct and consistent, otherwise false.
// Run by a client on input that should be verified.
func (p MembershipProof) DigestVerify(digest hashing.Digest, snapshot *Snapshot) bool {
	if p.HyperProof == nil || p.HistoryProof == nil {
		return false
	}

	hyperCorrect := p.HyperProof.Verify(digest, snapshot.HyperDigest)

	if p.Exists {
		if p.ActualVersion <= p.QueryVersion {
			historyCorrect := p.HistoryProof.Verify(digest, snapshot.HistoryDigest)
			return hyperCorrect && historyCorrect
		}
	}

	return hyperCorrect
}

// Verify verifies a proof and answer from QueryMembership. Returns true if the
// answer and proof are correct and consistent, otherwise false.
// Run by a client on input that should be verified.
func (p MembershipProof) Verify(event []byte, snapshot *Snapshot) bool {
	return p.DigestVerify(p.Hasher.Do(event), snapshot)
}

type IncrementalProof struct {
	Start, End uint64
	AuditPath  history.AuditPath
	Hasher     hashing.Hasher
}

func NewIncrementalProof(start, end uint64, auditPath history.AuditPath, hasher hashing.Hasher) *IncrementalProof {
	return &IncrementalProof{
		start,
		end,
		auditPath,
		hasher,
	}
}

func (p IncrementalProof) Verify(snapshotStart, snapshotEnd *Snapshot) bool {
	ip := history.NewIncrementalProof(p.Start, p.End, p.AuditPath, p.Hasher)
	return ip.Verify(snapshotStart.HistoryDigest, snapshotEnd.HistoryDigest)
}

func (b Balloon) Version() uint64 {
	return b.version
}

func (b *Balloon) RefreshVersion() error {
	// get last stored version
	kv, err := b.store.GetLast(storage.HistoryTable)
	if err != nil {
		if err != storage.ErrKeyNotFound {
			return err
		}
	} else {
		b.version = util.BytesAsUint64(kv.Key[:8]) + 1
	}
	return nil
}

func (b *Balloon) Add(eventDigest hashing.Digest) (*Snapshot, []*storage.Mutation, error) {

	// Get version
	version := b.version
	b.version++

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

	snapshot := &Snapshot{
		EventDigest:   eventDigest,
		HistoryDigest: historyDigest,
		HyperDigest:   hyperDigest,
		Version:       version,
	}

	return snapshot, mutations, nil
}

func (b *Balloon) AddBulk(eventBulkDigest []hashing.Digest) ([]*Snapshot, []*storage.Mutation, error) {

	// Get version
	initialVersion := b.version
	b.version += uint64(len(eventBulkDigest))

	// Update trees
	var historyDigests []hashing.Digest
	var historyMutations []*storage.Mutation
	var historyErr error
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		historyDigests, historyMutations, historyErr = b.historyTree.AddBulk(eventBulkDigest, initialVersion)
		wg.Done()
	}()

	hyperDigest, mutations, hyperErr := b.hyperTree.AddBulk(eventBulkDigest, initialVersion)

	wg.Wait()

	if historyErr != nil {
		return nil, nil, historyErr
	}
	if hyperErr != nil {
		return nil, nil, hyperErr
	}

	// Append trees mutations
	mutations = append(mutations, historyMutations...)

	snapshotBulk := make([]*Snapshot, 0)
	for i, _ := range eventBulkDigest {
		snapshotBulk = append(snapshotBulk, &Snapshot{
			EventDigest:   eventBulkDigest[i],
			HistoryDigest: historyDigests[i],
			HyperDigest:   hyperDigest,
			Version:       initialVersion + uint64(i),
		})
	}

	return snapshotBulk, mutations, nil
}

func (b Balloon) QueryDigestMembershipConsistency(keyDigest hashing.Digest, version uint64) (*MembershipProof, error) {

	var proof MembershipProof
	var err error
	proof.Hasher = b.hasherF()
	proof.KeyDigest = keyDigest
	proof.QueryVersion = version
	proof.CurrentVersion = b.version - 1

	if version > proof.CurrentVersion {
		version = proof.CurrentVersion
	}

	proof.HyperProof, err = b.hyperTree.QueryMembership(keyDigest)
	if err != nil {
		return nil, fmt.Errorf("unable to get proof from hyper tree: %v", err)
	}

	if len(proof.HyperProof.Value) == 0 {
		proof.Exists = false
		proof.ActualVersion = version
		return &proof, nil
	}

	proof.Exists = true
	if versionLen := len(proof.HyperProof.Value); versionLen < 8 { // TODO GET RID OF THIS: used only to pass tests
		// the version is stored in the hyper tree with the length of the event digest
		// if the length of the value is less than the length of a uint64 in bytes, we have to add padding
		proof.ActualVersion = util.BytesAsUint64(util.AddPaddingToBytes(proof.HyperProof.Value, 8-versionLen))
	} else {
		// if the length of the value is greater or equal than the length of a uint64 in bytes, we have to truncate
		proof.ActualVersion = util.BytesAsUint64(proof.HyperProof.Value[versionLen-8:])
	}

	if proof.ActualVersion <= version {
		proof.HistoryProof, err = b.historyTree.ProveMembership(proof.ActualVersion, version)
		if err != nil {
			return nil, fmt.Errorf("unable to get proof from history tree: %v", err)
		}
	} else {
		return nil, fmt.Errorf("actual version %d is greater than the query version which is %d", proof.ActualVersion, version)
	}

	return &proof, nil
}

func (b Balloon) QueryMembershipConsistency(event []byte, version uint64) (*MembershipProof, error) {
	// We need a new instance of the hasher because the b.hasher cannot be
	// used concurrently, and we support concurrent queries
	hasher := b.hasherF()
	return b.QueryDigestMembershipConsistency(hasher.Do(event), version)
}

func (b Balloon) QueryDigestMembership(keyDigest hashing.Digest) (*MembershipProof, error) {

	var proof MembershipProof
	var err error
	proof.Hasher = b.hasherF()
	proof.KeyDigest = keyDigest
	proof.QueryVersion = b.version - 1
	proof.CurrentVersion = proof.QueryVersion

	proof.HyperProof, err = b.hyperTree.QueryMembership(keyDigest)
	if err != nil {
		return nil, fmt.Errorf("unable to get proof from hyper tree: %v", err)
	}

	if len(proof.HyperProof.Value) == 0 {
		proof.Exists = false
		proof.ActualVersion = proof.QueryVersion
		return &proof, nil
	}

	proof.Exists = true
	if versionLen := len(proof.HyperProof.Value); versionLen < 8 { // TODO GET RID OF THIS: used only to pass tests
		// the version is stored in the hyper tree with the length of the event digest
		// if the length of the value is less than the length of a uint64 in bytes, we have to add padding
		proof.ActualVersion = util.BytesAsUint64(util.AddPaddingToBytes(proof.HyperProof.Value, 8-versionLen))
	} else {
		// if the length of the value is greater or equal than the length of a uint64 in bytes, we have to truncate
		proof.ActualVersion = util.BytesAsUint64(proof.HyperProof.Value[versionLen-8:])
	}

	if proof.ActualVersion <= proof.QueryVersion {
		proof.HistoryProof, err = b.historyTree.ProveMembership(proof.ActualVersion, proof.QueryVersion)
		if err != nil {
			return nil, fmt.Errorf("unable to get proof from history tree: %v", err)
		}
	} else {
		panic("This cannot happen unless QED was tampered")
	}

	return &proof, nil
}

func (b Balloon) QueryMembership(event []byte) (*MembershipProof, error) {
	// We need a new instance of the hasher because the b.hasher cannot be
	// used concurrently, and we support concurrent queries
	hasher := b.hasherF()
	return b.QueryDigestMembership(hasher.Do(event))
}

func (b Balloon) QueryConsistency(start, end uint64) (*IncrementalProof, error) {

	var proof IncrementalProof

	if start >= b.version || end >= b.version || start > end {
		return nil, errors.New("unable to process proof from history tree: invalid range")
	}

	proof.Start = start
	proof.End = end
	proof.Hasher = b.hasherF()

	historyProof, err := b.historyTree.ProveConsistency(start, end)
	if err != nil {
		return nil, fmt.Errorf("unable to get proof from history tree: %v", err)
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
