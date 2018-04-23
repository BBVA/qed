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
	return NewVerifier(
		history.LeafHasherF,
		history.InteriorHasherF,

		"auditor",
		hashing.Sha256Hasher,
		hyper.LeafHasherF,
		hyper.InteriorHasherF,
	)
}

func NewVerifier(
	historyLeafHasher history.LeafHaser,
	historyInteriorHasher history.InteriorHaser,

	hyperId string,
	hyperhasher hashing.Hasher,
	hyperLeafHasher hyper.LeafHaser,
	hyperInteriorHasher hyper.InteriorHaser,

) *Verifier {

	return &Verifier{
		history.NewVerifier(historyLeafHasher, historyInteriorHasher),
		hyper.NewVerifier(hyperId, hyperHasher, hyperLeafHasher, hyperInteriorHasher),
	}
}

func (v *Verifier) Verify(proof *MembershipProof, key []byte, version uint, value []byte) (bool, error) {

	correct, recomputed := historyVerifier.Verify(
		// expectedDigest []byte,
		proof.HistoryProof, // auditPath []Node,
		// key []byte,
		proof.ActualVersion, // version uint
	)
	if !correct {
		return false, nil
	}

	correct, recomputed := hyperVerifier.Verify(
		// expectedCommitment []byte,
		proof.HyperProof, // auditPath [][]byte,
		// key,
		proof.QueryVersion, // value []byte
	)
	if !correct {
		return false, nil
	}

	return true, nil
}
