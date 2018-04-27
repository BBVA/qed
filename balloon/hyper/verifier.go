// Copyright © 2018 Banco Bilbao Vizcaya Argentaria S.A.  All rights reserved.
// Use of this source code is governed by an Apache 2 License
// that can be found in the LICENSE file

package hyper

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"verifiabledata/log"
)

type Proof struct {
	id             []byte
	auditPath      [][]byte
	digestLength   int
	leafHasher     LeafHasher
	interiorHasher InteriorHasher
	log            log.Logger
}

func NewProof(id string, auditPath [][]byte, leafHasher LeafHasher, interiorHasher InteriorHasher) *Proof {
	digestLength := len(leafHasher([]byte{0x0}, []byte{0x0}, []byte{0x0})) * 8
	return &Proof{
		[]byte(id),
		auditPath,
		digestLength,
		leafHasher,
		interiorHasher,
		log.NewError(os.Stdout, "HyperProof", log.Ldate|log.Ltime|log.Lmicroseconds|log.Llongfile),
	}
}

func (p Proof) String() string {
	return fmt.Sprintf(`{"id": "%s", "auditPathLen": "%d"}`, p.id, len(p.auditPath))
}

func (p *Proof) Verify(expectedDigest []byte, key []byte, value uint) bool {
	p.log.Infof("\nVerifying commitment %v with auditpath %v, key %v and value %v\n", expectedDigest, p.auditPath, key, value)
	valueBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(valueBytes, uint64(value))

	recomputed := p.rootHash(p.auditPath, rootPosition(p.digestLength), key, valueBytes)
	fmt.Printf("Recomputed: %x\n", recomputed)
	fmt.Printf("Expected:   %x\n", expectedDigest)
	return bytes.Equal(expectedDigest, recomputed)
}

func (p *Proof) rootHash(auditPath [][]byte, pos *Position, key, value []byte) []byte {
	p.log.Infof("Calling rootHash with auditpath %v, position %v, key %v, and value %v\n", auditPath, pos, key, value)
	if pos.height == 0 {
		return p.leafHasher(p.id, value, pos.base)
	}
	if !bitIsSet(key, p.digestLength-pos.height) { // if k_j == 0
		left := p.rootHash(auditPath, pos.left(), key, value)
		right := auditPath[pos.height-1]
		return p.interiorHasher(left, right, pos.base, pos.heightBytes())
	}
	left := auditPath[pos.height-1]
	right := p.rootHash(auditPath, pos.right(), key, value)
	return p.interiorHasher(left, right, pos.base, pos.heightBytes())
}
