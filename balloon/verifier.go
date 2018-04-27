// Copyright © 2018 Banco Bilbao Vizcaya Argentaria S.A.  All rights reserved.
// Use of this source code is governed by an Apache 2 License
// that can be found in the LICENSE file

package balloon

import (
	"fmt"
	"log"
	"os"

	"verifiabledata/balloon/hashing"
	"verifiabledata/balloon/history"
	"verifiabledata/balloon/hyper"
)

type Verifiable interface {
	Verify([]byte, []byte, uint) bool
}

type Proof struct {
	Exists        bool
	HyperProof    Verifiable
	HistoryProof  Verifiable
	QueryVersion  uint
	ActualVersion uint
	hasher        hashing.Hasher
	log           *log.Logger
}

func NewProof(
	exists bool,
	hyperProof Verifiable,
	historyProof Verifiable,
	queryVersion uint,
	actualVersion uint,
	hasher hashing.Hasher,
) *Proof {
	return &Proof{
		exists,
		hyperProof,
		historyProof,
		queryVersion,
		actualVersion,
		hasher,
		log.New(os.Stdout, "BalloonProof", log.Ldate|log.Ltime|log.Lmicroseconds|log.Llongfile),
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
	fmt.Printf("Hyper correct: %v\n", hyperCorrect)
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

func ToBalloonProof(id string, p *MembershipProof, hasher hashing.Hasher) *Proof {
	htlh := history.LeafHasherF(hasher)
	htih := history.InteriorHasherF(hasher)

	hylh := hyper.LeafHasherF(hasher)
	hyih := hyper.InteriorHasherF(hasher)

	historyProof := history.NewProof(p.HistoryProof, htlh, htih)
	hyperProof := hyper.NewProof("", p.HyperProof, hylh, hyih)

	return NewProof(p.Exists, hyperProof, historyProof, p.QueryVersion, p.ActualVersion, hasher)

}
