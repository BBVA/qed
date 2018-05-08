// Copyright © 2018 Banco Bilbao Vizcaya Argentaria S.A.  All rights reserved.
// Use of this source code is governed by an Apache 2 License
// that can be found in the LICENSE file

package balloon

import (
	"encoding/binary"
	"fmt"

	"verifiabledata/balloon/hashing"
	"verifiabledata/balloon/history"
	"verifiabledata/balloon/hyper"
	"verifiabledata/log"
)

type Balloon interface {
	Add(event []byte) chan *Commitment
	GenMembershipProof(event []byte, version uint64) chan *MembershipProof
	Close() chan bool
}

type HyperBalloon struct {
	history *history.Tree
	hyper   *hyper.Tree
	hasher  hashing.Hasher
	version uint64
	ops     chan interface{} // serialize operations
}

type Commitment struct {
	HistoryDigest []byte
	HyperDigest   []byte
	Version       uint64
}

type MembershipProof struct {
	Exists        bool
	HyperProof    [][]byte
	HistoryProof  []history.Node
	QueryVersion  uint64
	ActualVersion uint64
	KeyDigest     []byte
}

func NewHyperBalloon(hasher hashing.Hasher, history *history.Tree, hyper *hyper.Tree) *HyperBalloon {

	b := HyperBalloon{
		history,
		hyper,
		hasher,
		0,
		nil,
	}
	b.ops = b.operations()
	return &b

}

func (b HyperBalloon) Add(event []byte) chan *Commitment {
	result := make(chan *Commitment)
	b.ops <- &add{
		event,
		result,
	}
	return result
}

func (b HyperBalloon) GenMembershipProof(event []byte, version uint64) chan *MembershipProof {
	result := make(chan *MembershipProof)
	b.ops <- &membership{
		event,
		version,
		result,
	}
	return result
}

func (b HyperBalloon) Close() chan bool {
	result := make(chan bool)

	b.history.Close()
	b.hyper.Close()

	b.ops <- &close{true, result}
	return result
}

// INTERNALS

type add struct {
	event  []byte
	result chan *Commitment
}

type membership struct {
	event   []byte
	version uint64
	result  chan *MembershipProof
}

type close struct {
	stop   bool
	result chan bool
}

// Run listens in channel operations to execute in the tree
func (b *HyperBalloon) operations() chan interface{} {
	operations := make(chan interface{}, 1000)
	go func() {
		for {
			select {
			case op := <-operations:
				switch msg := op.(type) {
				case *close:
					msg.result <- true
					return
				case *add:
					digest, err := b.add(msg.event)
					if err != nil {
						log.Error("Operations error: %v", err)
					}
					msg.result <- digest
				case *membership:
					proof, err := b.genMembershipProof(msg.event, msg.version)
					if err != nil {
						log.Error("Operations error: %v", err)
					}
					msg.result <- proof
				default:
					log.Error("Hyper tree Run() message not implemented!!")
				}

			}
		}
	}()
	return operations
}

func (b *HyperBalloon) add(event []byte) (*Commitment, error) {
	digest := b.hasher(event)
	version := b.version
	index := make([]byte, 8)
	binary.LittleEndian.PutUint64(index, uint64(version))
	b.version++
	return &Commitment{
		<-b.history.Add(digest, index),
		<-b.hyper.Add(digest, index),
		version,
	}, nil
}

func (b *HyperBalloon) genMembershipProof(event []byte, version uint64) (*MembershipProof, error) {
	digest := b.hasher(event)

	var hyperProof *hyper.MembershipProof
	var historyProof *history.MembershipProof

	hyperProof = <-b.hyper.Prove(digest)

	var exists bool
	var actualVersion uint64

	if len(hyperProof.ActualValue) > 0 {
		exists = true
		actualVersion = uint64(binary.LittleEndian.Uint64(hyperProof.ActualValue))
	}

	if exists && actualVersion <= version {
		historyProof = <-b.history.Prove(hyperProof.ActualValue, actualVersion, version)
	} else {
		return &MembershipProof{}, fmt.Errorf("Unable to get proof from history tree")
	}

	return &MembershipProof{
		exists,
		hyperProof.AuditPath,
		historyProof.Nodes,
		version,
		actualVersion,
		digest,
	}, nil

}
