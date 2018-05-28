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

// Package balloon implements the tree interface to interact with both hyper
// and history trees.
package balloon

import (
	"encoding/binary"
	"fmt"

	"github.com/bbva/qed/balloon/hashing"
	"github.com/bbva/qed/balloon/history"
	"github.com/bbva/qed/balloon/hyper"
	"github.com/bbva/qed/log"
)

// Balloon is the public set of operations that can be done to a
// HyperBalloon struct.
type Balloon interface {
	Add(event []byte) chan *Commitment
	GenMembershipProof(event []byte, version uint64) chan *MembershipProof
	Close() chan bool
}

// HyperBallon is the struct that links together both hyper and history trees
// the balloon version and the ops channel for the balloon operations
// serializer.
type HyperBalloon struct {
	history *history.Tree
	hyper   *hyper.Tree
	hasher  hashing.Hasher
	version uint64
	ops     chan interface{} // serialize operations
}

// Commitment is the struct that has both history and hyper digest and the
// current version for that rootNode digests.
type Commitment struct {
	HistoryDigest []byte
	HyperDigest   []byte
	Version       uint64
}

// MembershipProof is the struct that is required to make a Exisitance Proof.
// It has both Hyper and History AuditPaths, if it exists in first place and
// Current, Actual and Query Versions.
type MembershipProof struct {
	Exists         bool
	HyperProof     [][]byte
	HistoryProof   []history.Node
	CurrentVersion uint64
	QueryVersion   uint64
	ActualVersion  uint64
	KeyDigest      []byte
}

// NewHyperBallon returns a HyperBalloon struct.
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

// Add is the public balloon interface to add a event in the balloon tree.
func (b HyperBalloon) Add(event []byte) chan *Commitment {
	result := make(chan *Commitment)
	b.ops <- &add{
		event,
		result,
	}
	return result
}

// GenMembership is the public balloon interface to get a MembershipProof to
// do a Existance Proof.
func (b HyperBalloon) GenMembershipProof(event []byte, version uint64) chan *MembershipProof {
	result := make(chan *MembershipProof)
	b.ops <- &membership{
		event,
		version,
		result,
	}
	return result
}

// Close will close the balloon operations serializer channel.
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
						log.Error("Operations error: ", err)
					}
					msg.result <- digest

				case *membership:
					proof, err := b.genMembershipProof(msg.event, msg.version)
					if err != nil {
						log.Debug("Operations error: ", err)
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
	binary.LittleEndian.PutUint64(index, version)
	b.version++
	return &Commitment{
		<-b.history.Add(digest, index),
		<-b.hyper.Add(digest, index),
		version,
	}, nil
}

func (b HyperBalloon) genMembershipProof(event []byte, version uint64) (*MembershipProof, error) {
	digest := b.hasher(event)

	var hyperProof *hyper.MembershipProof
	var historyProof *history.MembershipProof

	hyperProof = <-b.hyper.Prove(digest)

	var exists bool
	var actualVersion uint64

	if len(hyperProof.ActualValue) > 0 {
		exists = true
		actualVersion = binary.LittleEndian.Uint64(hyperProof.ActualValue)
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
		b.version - 1, // notice, this is the current version of balloon version
		digest,
	}, nil

}
