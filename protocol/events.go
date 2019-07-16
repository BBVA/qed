/*
   Copyright 2018-2019 Banco Bilbao Vizcaya Argentaria, S.A.

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

// Package protocol defines the information types required and expected when
// interacting with QED.
package protocol

import (
	"encoding/json"

	"github.com/bbva/qed/balloon"
	"github.com/bbva/qed/balloon/history"
	"github.com/bbva/qed/balloon/hyper"
	"github.com/bbva/qed/crypto/hashing"
	"github.com/bbva/qed/util"
)

// Event is the public struct that Add handler function uses to
// parse the post params.
type Event struct {
	Event []byte
}

// EventBulk is the public struct that AddBulk handler function uses to
// parse the post params.
type EventsBulk struct {
	Events [][]byte
}

// MembershipQuery is the public struct that apihttp.Membership
// Handler uses to parse the post params.
type MembershipQuery struct {
	Key     []byte
	Version *uint64
}

// MembershipDigest is the public struct that apihttp.DigestMembership
// Handler uses to parse the post params.
type MembershipDigest struct {
	KeyDigest hashing.Digest
	Version   *uint64
}

// Snapshot is the public struct that apihttp.Add Handler call returns.
type Snapshot struct {
	EventDigest   hashing.Digest
	HistoryDigest hashing.Digest
	HyperDigest   hashing.Digest
	Version       uint64
}

func (b *Snapshot) Encode() ([]byte, error) {
	return json.Marshal(b)
}

func (b *Snapshot) Decode(msg []byte) error {
	err := json.Unmarshal(msg, b)
	return err
}

// SignedSnapshot is the public struct that apihttp.Add Handler call returns.
// It is comprised of a Snapshot and a signature.
type SignedSnapshot struct {
	Snapshot  *Snapshot
	Signature []byte
}

func (b *SignedSnapshot) Encode() ([]byte, error) {
	return json.Marshal(b)
}

func (b *SignedSnapshot) Decode(msg []byte) error {
	err := json.Unmarshal(msg, b)
	return err
}

// BatchSnapshots is information structure that QED sends to Agents, and
// Agents to alerts/snapshot store.
// It is comprised of an array of Signed Snapshots.
type BatchSnapshots struct {
	Snapshots []*SignedSnapshot
}

func (b *BatchSnapshots) Encode() ([]byte, error) {
	return json.Marshal(b)
}

func (b *BatchSnapshots) Decode(msg []byte) error {
	err := json.Unmarshal(msg, b)
	return err
}

// MembershipResult is the information structure needed or a Membership proof.
type MembershipResult struct {
	Exists         bool
	Hyper          map[string]hashing.Digest
	History        map[string]hashing.Digest
	CurrentVersion uint64
	QueryVersion   uint64
	ActualVersion  uint64
	KeyDigest      hashing.Digest
	Key            []byte
}

// IncrementalRequest is the information structure needed to ask for an incremental request.
type IncrementalRequest struct {
	Start uint64
	End   uint64
}

// IncrementalResponse is the information structure expected from an incremental proof request.
type IncrementalResponse struct {
	Start     uint64
	End       uint64
	AuditPath map[string]hashing.Digest
}

// ToMembershipProof translates internal api balloon.MembershipProof to the
// public struct protocol.MembershipResult.
func ToMembershipResult(key []byte, mp *balloon.MembershipProof) *MembershipResult {

	var serialized map[string]hashing.Digest
	if mp.HistoryProof != nil && mp.HistoryProof.AuditPath != nil {
		serialized = mp.HistoryProof.AuditPath.Serialize()
	}

	return &MembershipResult{
		mp.Exists,
		mp.HyperProof.AuditPath,
		serialized,
		mp.CurrentVersion,
		mp.QueryVersion,
		mp.ActualVersion,
		mp.KeyDigest,
		key,
	}
}

// ToBaloonProof translate public protocol.MembershipResult to internal
// balloon.MembershipProof.
func ToBalloonProof(mr *MembershipResult, hasherF func() hashing.Hasher) *balloon.MembershipProof {

	historyProof := history.NewMembershipProof(
		mr.ActualVersion,
		mr.QueryVersion,
		history.ParseAuditPath(mr.History),
		hasherF(),
	)

	hasher := hasherF()
	hyperProof := hyper.NewQueryProof(
		mr.KeyDigest,
		util.Uint64AsPaddedBytes(mr.ActualVersion, int(hasher.Len())),
		mr.Hyper,
		hasher,
	)

	return balloon.NewMembershipProof(
		mr.Exists,
		hyperProof,
		historyProof,
		mr.CurrentVersion,
		mr.QueryVersion,
		mr.ActualVersion,
		mr.KeyDigest,
		hasherF(),
	)

}

// ToIncrementalResponse translates internal api balloon.IncrementalProof to the
// public struct protocol.IncrementalResponse.
func ToIncrementalResponse(proof *balloon.IncrementalProof) *IncrementalResponse {
	return &IncrementalResponse{
		proof.Start,
		proof.End,
		proof.AuditPath.Serialize(),
	}
}

// ToIncrementalProof translate public protocol.IncrementalResponse to internal
// balloon.IncrementalProof.
func ToIncrementalProof(ir *IncrementalResponse, hasherF func() hashing.Hasher) *balloon.IncrementalProof {
	return balloon.NewIncrementalProof(ir.Start, ir.End, history.ParseAuditPath(ir.AuditPath), hasherF())
}

// BackupInfo is the public struct used to parse the backup information.
// Metadata is suposed to contains only the balloon version.
type BackupInfo struct {
	ID        int64
	Timestamp int64
	Size      int64
	NumFiles  int32
	Metadata  string
}
