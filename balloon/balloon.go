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
	"github.com/bbva/qed/metrics"
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
	balloon.RefreshVersion()

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

func NewMembershipProof(
	exists bool,
	hyperProof *hyper.QueryProof,
	historyProof *history.MembershipProof,
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

func NewIncrementalProof(
	start, end uint64,
	auditPath history.AuditPath,
	hasher hashing.Hasher,
) *IncrementalProof {
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

func (b *Balloon) TamperHyper(eventDigest []byte, versionValue uint64) (*Snapshot, []*storage.Mutation, error) {
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

	hyperDigest, mutations, hyperErr := b.hyperTree.Add(eventDigest, versionValue)

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

func (b *Balloon) Add(event []byte) (*Snapshot, []*storage.Mutation, error) {

	// Metrics
	metrics.QedBalloonAddTotal.Inc()
	//timer := prometheus.NewTimer(metrics.QedBalloonAddDurationSeconds)
	//defer timer.ObserveDuration()

	// Activate metrics gathering
	stats := metrics.Balloon

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

	snapshot := &Snapshot{
		EventDigest:   eventDigest,
		HistoryDigest: historyDigest,
		HyperDigest:   hyperDigest,
		Version:       version,
	}

	// Increment add hits and version
	stats.AddFloat("add_hits", 1)
	stats.Set("version", metrics.Uint64ToVar(version))

	return snapshot, mutations, nil
}

func (b Balloon) QueryDigestMembership(keyDigest hashing.Digest, version uint64) (*MembershipProof, error) {
	// Metrics
	metrics.QedBalloonDigestMembershipTotal.Inc()
	//timer := prometheus.NewTimer(metrics.QedBalloonDigestMembershipDurationSeconds)
	//defer timer.ObserveDuration()

	stats := metrics.Balloon
	stats.AddFloat("QueryMembership", 1)

	var proof MembershipProof
	var wg sync.WaitGroup
	var hyperErr, historyErr error
	var hyperProof *hyper.QueryProof
	var historyProof *history.MembershipProof
	var leaf *storage.KVPair
	var err error

	proof.Hasher = b.hasherF()
	proof.KeyDigest = keyDigest
	proof.QueryVersion = version
	proof.CurrentVersion = b.version - 1

	if version > proof.CurrentVersion {
		version = proof.CurrentVersion
	}

	leaf, err = b.store.Get(storage.IndexPrefix, proof.KeyDigest)
	switch {
	case err != nil && err != storage.ErrKeyNotFound:
		return nil, fmt.Errorf("error reading leaf %v data: %v", proof.KeyDigest, err)

	case err != nil && err == storage.ErrKeyNotFound:
		proof.Exists = false
		proof.ActualVersion = version
		leaf = &storage.KVPair{Key: keyDigest, Value: util.Uint64AsBytes(version)}

	case err == nil:
		proof.Exists = true
		proof.ActualVersion = util.BytesAsUint64(leaf.Value)

		if proof.ActualVersion <= version {
			wg.Add(1)
			go func() {
				defer wg.Done()
				historyProof, historyErr = b.historyTree.ProveMembership(proof.ActualVersion, version)
			}()
		} else {
			return nil, fmt.Errorf("The actual version of the entry is %d, but the query version is %d,  we're unable to build the proof becasue the entry was not inserted in version we are asking for.", proof.ActualVersion, version)
		}

	}

	hyperProof, hyperErr = b.hyperTree.QueryMembership(leaf.Key, leaf.Value)

	wg.Wait()
	if hyperErr != nil {
		return nil, fmt.Errorf("unable to get proof from hyper tree: %v", err)
	}

	if historyErr != nil {
		return nil, fmt.Errorf("unable to get proof from history tree: %v", err)
	}

	proof.HyperProof = hyperProof
	proof.HistoryProof = historyProof
	return &proof, nil
}

func (b Balloon) QueryMembership(event []byte, version uint64) (*MembershipProof, error) {
	hasher := b.hasherF()
	return b.QueryDigestMembership(hasher.Do(event), version)
}

func (b Balloon) QueryConsistency(start, end uint64) (*IncrementalProof, error) {

	// Metrics
	metrics.QedBalloonIncrementalTotal.Inc()

	stats := metrics.Balloon
	stats.AddFloat("QueryConsistency", 1)
	var proof IncrementalProof

	if start >= b.version ||
		end >= b.version ||
		start >= end {

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
