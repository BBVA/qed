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
	"github.com/bbva/qed/balloon/history"
	"github.com/bbva/qed/balloon/hyper"
	"github.com/bbva/qed/balloon/proof"
	"github.com/bbva/qed/hashing"
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
	HistoryDigest []byte
	HyperDigest   []byte
	Version       uint64
	Event         []byte
}

type SignedSnapshot struct {
	Snapshot  *Snapshot
	Signature []byte
}

type MembershipResult struct {
	Exists         bool
	Hyper          map[string][]byte
	History        map[string][]byte
	CurrentVersion uint64
	QueryVersion   uint64
	ActualVersion  uint64
	KeyDigest      []byte
	Key            []byte
}

type IncrementalRequest struct {
	Start uint64
	End   uint64
}

type IncrementalResponse struct {
	Start     uint64
	End       uint64
	AuditPath map[string][]byte
}

// ToMembershipProof translates internal api balloon.MembershipProof to the
// public struct apihttp.MembershipResult.
func ToMembershipResult(key []byte, mp *balloon.MembershipProof) *MembershipResult {
	return &MembershipResult{
		mp.Exists,
		mp.HyperProof.AuditPath(),
		mp.HistoryProof.AuditPath(),
		mp.CurrentVersion,
		mp.QueryVersion,
		mp.ActualVersion,
		mp.KeyDigest,
		key,
	}
}

// ToBaloonProof translate public apihttp.MembershipResult:w to internal
// balloon.Proof.

func ToBalloonProof(id []byte, mr *MembershipResult, hasher hashing.Hasher) *balloon.MembershipProof {

	historyPos := history.NewRootPosition(mr.QueryVersion)
	hyperPos := hyper.NewRootPosition(hasher.Len(), 0)

	historyProof := proof.NewProof(historyPos, mr.History, hasher)
	hyperProof := proof.NewProof(hyperPos, mr.Hyper, hasher)

	return balloon.NewMembershipProof(mr.Exists, hyperProof, historyProof, mr.CurrentVersion, mr.ActualVersion, mr.QueryVersion, mr.KeyDigest, hasher)

}

func ToIncrementalResponse(proof *balloon.IncrementalProof) *IncrementalResponse {
	return &IncrementalResponse{
		proof.Start,
		proof.End,
		proof.AuditPath,
	}
}

func ToIncrementalProof(ir *IncrementalResponse, hasher hashing.Hasher) *balloon.IncrementalProof {
	return balloon.NewIncrementalProof(ir.Start, ir.End, ir.AuditPath, hasher)
}
