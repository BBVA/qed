package balloon

import (
	"fmt"

	"qed/balloon/hashing"
)

type Verifiable interface {
	Verify([]byte, []byte, uint64) bool
}

type Proof struct {
	Exists        bool
	HyperProof    Verifiable
	HistoryProof  Verifiable
	QueryVersion  uint64
	ActualVersion uint64
	hasher        hashing.Hasher
}

func NewProof(
	exists bool,
	hyperProof Verifiable,
	historyProof Verifiable,
	queryVersion uint64,
	actualVersion uint64,
	hasher hashing.Hasher,
) *Proof {
	return &Proof{
		exists,
		hyperProof,
		historyProof,
		queryVersion,
		actualVersion,
		hasher,
	}
}

func (p *Proof) String() string {
	return fmt.Sprintf(`{"Exists":%v, "HyperProof": "%v", "HistoryProof": "%v", "QueryVersion": "%d", "ActualVersion": "%d"}`,
		p.Exists, p.HyperProof, p.HistoryProof, p.QueryVersion, p.ActualVersion)
}

func (p *Proof) Verify(commitment *Commitment, event []byte) bool {
	digest := p.hasher(event)
	hyperCorrect := p.HyperProof.Verify(
		commitment.HyperDigest,
		digest,
		p.QueryVersion,
	)

	if p.Exists {
		if p.QueryVersion <= p.ActualVersion {
			historyCorrect := p.HistoryProof.Verify(
				commitment.HistoryDigest,
				digest,
				p.QueryVersion,
			)
			return hyperCorrect && historyCorrect
		}
	}

	return hyperCorrect

}
