// Copyright © 2018 Banco Bilbao Vizcaya Argentaria S.A.  All rights reserved.
// Use of this source code is governed by an Apache 2 License
// that can be found in the LICENSE file

package balloon

import (
	"fmt"
	"log"
	"os"
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
	log           *log.Logger
}

func NewProof(
	exists bool,
	hyperProof Verifiable,
	historyProof Verifiable,
	queryVersion uint,
	actualVersion uint,
) *Proof {
	return &Proof{
		exists,
		hyperProof,
		historyProof,
		queryVersion,
		actualVersion,
		log.New(os.Stdout, "BalloonProof", log.Ldate|log.Ltime|log.Lmicroseconds|log.Llongfile),
	}
}

func (p *Proof) String() string {
	return fmt.Sprintf(`{"Exists":%v, "HyperProof": "%v", "HistoryProof": "%v", "QueryVersion": "%d", "ActualVersion": "%d"}`,
		p.Exists, p.HyperProof, p.HistoryProof, p.QueryVersion, p.ActualVersion)
}

func (p *Proof) Verify(commitment *Commitment, event []byte) bool {
	hyperCorrect := p.HyperProof.Verify(
		commitment.HyperDigest,
		event,
		p.QueryVersion,
	)

	if p.Exists {
		if p.QueryVersion <= p.ActualVersion {
			historyCorrect := p.HistoryProof.Verify(
				commitment.HistoryDigest,
				event,
				p.QueryVersion,
			)
			return hyperCorrect && historyCorrect
		}
	}

	return hyperCorrect

}
