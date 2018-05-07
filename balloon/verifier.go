// Copyright © 2018 Banco Bilbao Vizcaya Argentaria S.A.  All rights reserved.
// Use of this source code is governed by an Apache 2 License
// that can be found in the LICENSE file

package balloon

import (
	"fmt"
	"log"
	"os"

	"verifiabledata/balloon/hashing"
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
	log           *log.Logger
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
