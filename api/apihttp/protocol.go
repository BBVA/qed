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

package apihttp

import (
	"github.com/bbva/qed/balloon"
	"github.com/bbva/qed/balloon/hashing"
	"github.com/bbva/qed/balloon/history"
	"github.com/bbva/qed/balloon/hyper"
)

// Event is the public struct that Add handler function uses to
// parse the post params.
type Event struct {
	Event []byte
}

// MembershipQuery is the public struct that apihttp.Membership
// Handler uses to parse the post params.
type MembershipQuery struct {
	Key     []byte
	Version uint64
}

// Snapshot is the public struct that apihttp.Add Handler call returns.
type Snapshot struct {
	HyperDigest   []byte
	HistoryDigest []byte
	Version       uint64
	Event         []byte
	//TODO: implement this
	// EventDigest   string
}

// HistoryNode is part of the apihttp.MembershipResult used to parse the
// result of apihttp.Membership handler.
type HistoryNode struct {
	Digest       []byte
	Index, Layer uint64
}

// Proofs is part of the apihttp.MembershipResult used to parse the
// result if Membership Handler.
type Proofs struct {
	HyperAuditPath   [][]byte
	HistoryAuditPath []HistoryNode
}

// MembershipResult is the public struct that returns the Membership
// handler
type MembershipResult struct {
	Key                                         []byte
	KeyDigest                                   []byte
	IsMember                                    bool
	Proofs                                      *Proofs
	CurrentVersion, QueryVersion, ActualVersion uint64
}

// ToSnapshot translates the internal struct used in balloon
// (balloon.Commitment and original event) to the public protocol struct
// apihttp.Snapshot.
func ToSnapshot(commitment *balloon.Commitment, event []byte) *Snapshot {
	return &Snapshot{
		commitment.HyperDigest,
		commitment.HistoryDigest,
		commitment.Version,
		event,
	}
}

// ToHistoryAuditPath translates the internal api balloon.history.Node to
// public struct apihttp.HistoryNode array.
func ToHistoryAuditPath(path []history.Node) []HistoryNode {
	result := make([]HistoryNode, 0)
	for _, elem := range path {
		result = append(result, HistoryNode{elem.Digest, elem.Index, elem.Layer})
	}
	return result
}

// ToMembershipProof translates internal api balloon.MembershipProof to the
// public struct apihttp.MembershipResult.
func ToMembershipProof(event []byte, proof *balloon.MembershipProof) *MembershipResult {
	return &MembershipResult{
		event,
		proof.KeyDigest,
		proof.Exists,
		&Proofs{
			proof.HyperProof,
			ToHistoryAuditPath(proof.HistoryProof),
		},
		proof.CurrentVersion,
		proof.QueryVersion,
		proof.ActualVersion,
	}
}

// ToHistoryNode translates public apihttp.HistoryNode to internal
// balloon.history.Node struct array.
func ToHistoryNode(path []HistoryNode) []history.Node {
	result := make([]history.Node, 0)
	for _, elem := range path {
		result = append(result, history.Node{elem.Digest, elem.Index, elem.Layer})
	}
	return result
}

// ToBaloonProof translate public apihttp.MembershipResult:w to internal
// balloon.Proof.
func ToBalloonProof(id string, p *MembershipResult, hasher hashing.Hasher) *balloon.Proof {

	historyProof := history.NewProof(ToHistoryNode(p.Proofs.HistoryAuditPath), p.QueryVersion, hasher)
	hyperProof := hyper.NewProof(id, p.Proofs.HyperAuditPath, hasher)

	return balloon.NewProof(p.IsMember, hyperProof, historyProof, p.CurrentVersion, p.QueryVersion, p.ActualVersion, hasher)

}
