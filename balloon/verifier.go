// Copyright © 2018 Banco Bilbao Vizcaya Argentaria S.A.  All rights reserved.
// Use of this source code is governed by an Apache 2 License
// that can be found in the LICENSE file

package balloon

import (
	"verifiabledata/balloon/hashing"
	"verifiabledata/balloon/history"
	"verifiabledata/balloon/hyper"
)

type Verifier struct {
	historyVerifier *history.Verifier
	hyperVerifier   *hyper.Verifier
}

func NewDefaultVerifier() *Verifier {
	hasher := hashing.Sha256Hasher

	return NewVerifier(
		history.LeafHasherF(hasher),
		history.InteriorHasherF(hasher),

		"auditor",
		hasher,
		hyper.LeafHasherF(hasher),
		hyper.InteriorHasherF(hasher),
	)
}

func NewVerifier(
	historyLeafHasher history.LeafHasher,
	historyInteriorHasher history.InteriorHasher,

	hyperId string,
	hyperHasher hashing.Hasher,
	hyperLeafHasher hyper.LeafHasher,
	hyperInteriorHasher hyper.InteriorHasher,

) *Verifier {

	return &Verifier{
		history.NewVerifier(historyLeafHasher, historyInteriorHasher),
		hyper.NewVerifier(hyperId, hyperHasher, hyperLeafHasher, hyperInteriorHasher),
	}
}

func (v *Verifier) Verify(proof *MembershipProof, commitment *Commitment, event []byte) (bool, error) {

	historyCorrect, _ := v.historyVerifier.Verify(
		commitment.HistoryDigest, // expectedDigest []byte,
		proof.HistoryProof,       // auditPath []Node,
		event,                    // key []byte,
		proof.ActualVersion,      // version uint
	)
	if !historyCorrect {
		return false, nil
	}

	hyperCorrect, _ := v.hyperVerifier.Verify(
		commitment.IndexDigest, // expectedCommitment []byte,
		proof.HyperProof,       // auditPath [][]byte,
		event,                  // key,
		event,                  // value []byte
	)

	if !hyperCorrect {
		return false, nil
	}

	return true, nil
}
