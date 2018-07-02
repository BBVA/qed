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

	"github.com/bbva/qed/balloon/history"
	"github.com/bbva/qed/balloon/hyper"
	"github.com/bbva/qed/balloon/proof"
	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/log"
)

// Balloon is the public set of operations that can be done to a
// HyperBalloon struct.
type Balloon interface {
	Add(event []byte) chan *Commitment
	GenMembershipProof(event []byte, version uint64) chan *MembershipProof
	Close() chan bool
}

// HyperBalloon is the struct that links together both hyper and history trees
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
	HyperProof     proof.Verifiable
	HistoryProof   proof.Verifiable
	CurrentVersion uint64
	QueryVersion   uint64
	ActualVersion  uint64
	KeyDigest      []byte
	hasher         hashing.Hasher
}

func NewMembershipProof(
	exists bool,
	hyperProof proof.Verifiable,
	historyProof proof.Verifiable,
	currentVersion uint64,
	queryVersion uint64,
	actualVersion uint64,
	keyDigest []byte,
	hasher hashing.Hasher,
) *MembershipProof {
	return &MembershipProof{
		exists,
		hyperProof,
		historyProof,
		currentVersion,
		queryVersion,
		actualVersion,
		keyDigest,
		hasher,
	}
}

func (p MembershipProof) Verify(commitment *Commitment, event []byte) bool {
	if p.HyperProof == nil || p.HistoryProof == nil {
		return false
	}

	digest := p.hasher.Do(event)
	hyperCorrect := p.HyperProof.Verify(
		commitment.HyperDigest,
		digest,
		uint2bytes(p.QueryVersion),
	)

	if p.Exists {
		if p.QueryVersion <= p.ActualVersion {
			historyCorrect := p.HistoryProof.Verify(
				commitment.HistoryDigest,
				uint2bytes(p.QueryVersion),
				digest,
			)
			return hyperCorrect && historyCorrect
		}
	}

	return hyperCorrect
}

type IncrementalProof struct {
	history.IncrementalProof
}

func uint2bytes(i uint64) []byte {
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, i)
	return bytes
}

// NewHyperBalloon returns a HyperBalloon struct.
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
// do a Existence Proof.
func (b HyperBalloon) GenMembershipProof(event []byte, version uint64) chan *MembershipProof {
	result := make(chan *MembershipProof)
	b.ops <- &membership{
		event,
		version,
		result,
	}
	return result
}

// GenIncrementalProof is the public balloon interface to get an IncrementalProof to
// generate a consistency proof
func (b HyperBalloon) GenIncrementalProof(start, end uint64) chan *IncrementalProof {
	result := make(chan *IncrementalProof)
	b.ops <- &incremental{
		start,
		end,
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

type incremental struct {
	start  uint64
	end    uint64
	result chan *IncrementalProof
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

				case *incremental:
					proof, err := b.genIncrementalProof(msg.start, msg.end)
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
	digest := b.hasher.Do(event)
	version := b.version
	index := make([]byte, 8)
	binary.LittleEndian.PutUint64(index, version)
	b.version++

	history, err := b.history.Add(digest, index)
	if err != nil {
		return nil, err
	}
	hyper, err := b.hyper.Add(digest, index)
	if err != nil {
		return nil, err
	}
	return &Commitment{
		history,
		hyper,
		version,
	}, nil
}

func (b HyperBalloon) genMembershipProof(event []byte, version uint64) (*MembershipProof, error) {
	var err error
	var mp MembershipProof
	var actualValue []byte
	mp.KeyDigest = b.hasher.Do(event)
	mp.QueryVersion = version
	mp.CurrentVersion = b.version - 1

	versionBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(versionBytes, version)

	mp.HyperProof, actualValue, err = b.hyper.ProveMembership(mp.KeyDigest, versionBytes)
	if err != nil {
		return nil, fmt.Errorf("Unable to get proof from hyper tree: %v", err)
	}

	if len(actualValue) > 0 {
		mp.Exists = true
		mp.ActualVersion = binary.LittleEndian.Uint64(actualValue)
	}

	if mp.Exists && mp.ActualVersion <= mp.QueryVersion {
		mp.HistoryProof, err = b.history.ProveMembership(actualValue, mp.ActualVersion, mp.QueryVersion)
	} else {
		mp.Exists = false
		return &mp, fmt.Errorf("Unable to get proof from history tree: %v", err)
	}

	return &mp, nil

}

func (b HyperBalloon) genIncrementalProof(start, end uint64) (*IncrementalProof, error) {

	startKey := uint2bytes(start)
	endKey := uint2bytes(end)

	proof, err := b.history.ProveIncremental(startKey, endKey, start, end)
	if err != nil {
		return nil, err
	}

	return &IncrementalProof{*proof}, nil
}
