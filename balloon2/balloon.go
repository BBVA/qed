package balloon2

import (
	"fmt"
	"sync"

	"github.com/bbva/qed/balloon2/common"
	"github.com/bbva/qed/balloon2/history"
	"github.com/bbva/qed/balloon2/hyper"
	"github.com/bbva/qed/db"
	"github.com/bbva/qed/util"
)

var (
	BalloonVersionKey = []byte("version")
)

type Balloon struct {
	version uint64
	hasher  common.Hasher
	store   db.Store

	historyTree *history.HistoryTree
	hyperTree   *hyper.HyperTree
}

func NewBalloon(initialVersion uint64, store db.Store, hasherF func() common.Hasher) *Balloon {

	historyCache := common.NewPassThroughCache(db.HistoryCachePrefix, store)
	hyperCache := common.NewSimpleCache(1 << 2)

	// TODO warm up hyper cache

	historyTree := history.NewHistoryTree(hasherF(), historyCache)
	hyperTree := hyper.NewHyperTree(hasherF(), store, hyperCache)

	return &Balloon{
		version:     initialVersion,
		hasher:      hasherF(),
		store:       store,
		historyTree: historyTree,
		hyperTree:   hyperTree,
	}
}

// Commitment is the struct that has both history and hyper digest and the
// current version for that rootNode digests.
type Commitment struct {
	HistoryDigest common.Digest
	HyperDigest   common.Digest
	Version       uint64
}

// MembershipProof is the struct that is required to make a Exisitance Proof.
// It has both Hyper and History AuditPaths, if it exists in first place and
// Current, Actual and Query Versions.
type MembershipProof struct {
	Exists         bool
	HyperProof     common.Verifiable
	HistoryProof   common.Verifiable
	CurrentVersion uint64
	QueryVersion   uint64
	ActualVersion  uint64 //required for consistency proof
	KeyDigest      common.Digest
	hasher         common.Hasher
}

func NewMembershipProof(
	exists bool,
	hyperProof, historyProof common.Verifiable,
	currentVersion, queryVersion, actualVersion uint64,
	keyDigest common.Digest,
	hasher common.Hasher) *MembershipProof {
	return &MembershipProof{
		exists,
		hyperProof,
		historyProof,
		currentVersion,
		queryVersion,
		actualVersion,
		keyDigest,
		hasher,
	}
}

// Verify verifies a proof and answer from QueryMembership. Returns true if the
// answer and proof are correct and consistent, otherwise false.
// Run by a client on input that should be verified.
func (p MembershipProof) Verify(event []byte, commitment *Commitment) bool {
	if p.HyperProof == nil || p.HistoryProof == nil {
		return false
	}

	digest := p.hasher.Do(event)
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
	AuditPath  common.AuditPath
	Hasher     common.Hasher
}

func NewIncrementalProof(
	start, end uint64,
	auditPath common.AuditPath,
	hasher common.Hasher,
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

func (b *Balloon) Add(event []byte) (*Commitment, []db.Mutation, error) {

	// Get version
	version := b.version
	b.version++

	// Update trees
	var historyDigest, hyperDigest common.Digest
	var historyMutations, hyperMutations []db.Mutation
	var historyErr, hyperErr error
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		historyDigest, historyMutations, historyErr = b.historyTree.Add(event, version)
		wg.Done()
	}()
	go func() {
		hyperDigest, hyperMutations, hyperErr = b.hyperTree.Add(event, version)
		wg.Done()
	}()

	// Append version mutation
	mutations := make([]db.Mutation, 0)
	mutations = append(mutations, *db.NewMutation(db.VersionPrefix, BalloonVersionKey, util.Uint64AsBytes(version)))

	wg.Wait()

	if historyErr != nil {
		return nil, nil, historyErr
	}
	if hyperErr != nil {
		return nil, nil, hyperErr
	}

	// Append trees mutations
	mutations = append(mutations, append(historyMutations, hyperMutations...)...)

	commitment := &Commitment{
		HistoryDigest: historyDigest,
		HyperDigest:   hyperDigest,
		Version:       version,
	}

	return commitment, mutations, nil
}

func (b Balloon) QueryMembership(event []byte, version uint64) (*MembershipProof, error) {

	var proof MembershipProof

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
	proof.Hasher = b.hasher

	historyProof, err := b.historyTree.ProveConsistency(start, end)
	if err != nil {
		return nil, fmt.Errorf("Unable to get proof from history tree: %v", err)
	}
	proof.AuditPath = historyProof.AuditPath

	return &proof, nil
}

func (b *Balloon) Close() {

}
