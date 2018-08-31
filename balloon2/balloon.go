package balloon2

import (
	"github.com/bbva/qed/balloon2/common"
	"github.com/bbva/qed/balloon2/history"
	"github.com/bbva/qed/balloon2/hyper"
	"github.com/bbva/qed/db"
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

	var historyCache common.Cache
	var hyperCache common.ModifiableCache

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

func (p MembershipProof) Verify(commitment *Commitment, event []byte) bool {
	return false
}

type IncrementalProof struct {
	Start, End uint64
	AuditPath  common.AuditPath
	hasher     common.Hasher
}

func (b *Balloon) Add(event []byte) (*Commitment, error) {
	return nil, nil
}

func (b Balloon) GenMembershipProof(event []byte, version uint16) (*MembershipProof, error) {
	return nil, nil
}

func (b Balloon) GenIncrementalProof(start, end uint64) (*IncrementalProof, error) {
	return nil, nil
}

func (b *Balloon) Close() {

}
