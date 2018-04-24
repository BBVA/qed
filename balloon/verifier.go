// Copyright © 2018 Banco Bilbao Vizcaya Argentaria S.A.  All rights reserved.
// Use of this source code is governed by an Apache 2 License
// that can be found in the LICENSE file

package balloon

import (
	"encoding/binary"
	"log"
	"os"
	"verifiabledata/balloon/hashing"
	"verifiabledata/balloon/history"
	"verifiabledata/balloon/hyper"
)

type Verifier struct {
	historyVerifier *history.Verifier
	hyperVerifier   *hyper.Verifier
	log             *log.Logger
}

func NewDefaultVerifier(balloonId string) *Verifier {
	hasher := hashing.Sha256Hasher

	return NewVerifier(
		balloonId,
		hasher,
		history.LeafHasherF(hasher),
		history.InteriorHasherF(hasher),
		hyper.LeafHasherF(hasher),
		hyper.InteriorHasherF(hasher),
	)
}

func NewVerifier(
	balloonId string,
	hasher hashing.Hasher,
	historyLH history.LeafHasher,
	historyIH history.InteriorHasher,
	hyperLH hyper.LeafHasher,
	hyperIH hyper.InteriorHasher,

) *Verifier {

	return &Verifier{
		history.NewVerifier(historyLH, historyIH),
		hyper.NewVerifier(balloonId, hasher, hyperLH, hyperIH),
		log.New(os.Stdout, "HyperBalloon", log.Ldate|log.Ltime|log.Lmicroseconds|log.Llongfile),
	}
}

func (v *Verifier) Verify(proof *MembershipProof, commitment *Commitment, event []byte) (bool, error) {

	queryVersion := make([]byte, 8)
	binary.LittleEndian.PutUint64(queryVersion, uint64(proof.QueryVersion))

	hyperCorrect, _ := v.hyperVerifier.Verify(
		commitment.HyperDigest, // expectedCommitment []byte,
		proof.HyperProof,       // auditPath [][]byte,
		event,                  // key,
		queryVersion,           // value []byte
	)

	if proof.Exists {
		if proof.QueryVersion <= proof.ActualVersion {
			historyCorrect, _ := v.historyVerifier.Verify(
				commitment.HistoryDigest, // expectedDigest []byte,
				proof.HistoryProof,       // auditPath []Node,
				event,                    // key []byte,
				proof.QueryVersion,       // version uint
			)
			return hyperCorrect && historyCorrect, nil
		}
	}

	return hyperCorrect, nil

}
