// Copyright © 2018 Banco Bilbao Vizcaya Argentaria S.A.  All rights reserved.
// Use of this source code is governed by an Apache 2 License
// that can be found in the LICENSE file

package hyper

import (
	"bytes"
	"encoding/binary"
	"log"
	"os"
	"verifiabledata/balloon/hashing"
)

type Proof struct {
	id             []byte
	auditPath      [][]byte
	digestLength   int
	leafHasher     LeafHasher
	interiorHasher InteriorHasher
	log            *log.Logger
}

func NewProof(id string, auditPath [][]byte, hasher hashing.Hasher, leafHasher LeafHasher, interiorHasher InteriorHasher) *Proof {
	digestLength := len(hasher([]byte("x"))) * 8
	return &Proof{
		[]byte(id),
		auditPath,
		digestLength,
		leafHasher,
		interiorHasher,
		log.New(os.Stdout, "HyperProof", log.Ldate|log.Ltime|log.Lmicroseconds|log.Llongfile),
	}
}

func (p *Proof) Verify(expectedDigest []byte, key []byte, value uint) bool {
	p.log.Printf("\nVerifying commitment %v with auditpath %v, key %v and value %v\n", expectedDigest, p.auditPath, key, value)
	valueBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(valueBytes, uint64(value))
	recomputed := p.rootHash(p.auditPath, rootPosition(p.digestLength), key, valueBytes)
	return bytes.Equal(expectedDigest, recomputed)
}

func (p *Proof) rootHash(auditPath [][]byte, pos *Position, key, value []byte) []byte {
	p.log.Printf("Calling rootHash with auditpath %v, position %v, key %v, and value %v\n", auditPath, pos, key, value)
	if pos.height == 0 {
		return p.leafHasher(p.id, value, pos.base)
	}
	if !bitIsSet(key, p.digestLength-pos.height) { // if k_j == 0
		left := p.rootHash(auditPath, pos.left(), key, value)
		right := auditPath[pos.height]
		next := pos.right()
		return p.interiorHasher(left, right, next.base, next.heightBytes())
	}
	left := auditPath[pos.height]
	right := p.rootHash(auditPath, pos.right(), key, value)
	next := pos.left()
	return p.interiorHasher(left, right, next.base, next.heightBytes())
}
