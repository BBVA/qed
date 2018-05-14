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

	"github.com/bbva/qed/balloon/hashing"
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

	if p.HyperProof == nil || p.HistoryProof == nil {
		return false
	}

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
