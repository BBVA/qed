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
package consensus

import (
	"fmt"
	"io"

	"github.com/hashicorp/raft"
	"github.com/pkg/errors"

	"github.com/bbva/qed/balloon"
	"github.com/bbva/qed/crypto/hashing"
	"github.com/bbva/qed/log"
	"github.com/bbva/qed/storage"
)

type fsmResponse struct {
	err error
}

type fsmAddResponse struct {
	fsmResponse
	snapshot []*balloon.Snapshot
}

type fsmState struct {
	Index, Term, BalloonVersion uint64
}

func (s *fsmState) encode() ([]byte, error) {
	return encodeMsgPack(s)
}

func (s *fsmState) shouldApply(f *fsmState) bool {
	if f.Term < s.Term {
		return false
	}
	if f.Term == s.Term && f.Index <= s.Index {
		return false
	}

	if f.BalloonVersion > 0 && s.BalloonVersion >= f.BalloonVersion {
		panic(fmt.Sprintf("balloonVersion panic! old: %+v, new %+v", s, f))
	}

	return true
}

func (n *RaftNode) loadState() error {
	kvstate, err := n.db.Get(storage.FSMStateTable, storage.FSMStateTableKey)
	if err == storage.ErrKeyNotFound {
		log.Infof("Unable to find previous state: assuming a clean instance")
		n.state = new(fsmState) // &fsmState{0, 0, 0}
		return nil
	}
	if err != nil {
		return errors.Wrap(err, "loading state failed")
	}
	var state fsmState
	err = decodeMsgPack(kvstate.Value, &state)
	if err != nil {
		return errors.Wrap(err, "unable to decode state")
	}
	n.state = &state
	return nil
}

// QueryDigestMembershipConsistency acts as a passthrough when an event digest is given to
// request a membership proof against a certain balloon version.
func (n *RaftNode) QueryDigestMembershipConsistency(keyDigest hashing.Digest, version uint64) (*balloon.MembershipProof, error) {
	return n.balloon.QueryDigestMembershipConsistency(keyDigest, version)
}

// QueryMembershipConsistency acts as a passthrough when an event is given to request a
// membership proof against a certain balloon version.
func (n *RaftNode) QueryMembershipConsistency(event []byte, version uint64) (*balloon.MembershipProof, error) {
	return n.balloon.QueryMembershipConsistency(event, version)
}

// QueryDigestMembership acts as a passthrough when an event digest is given to request a
// membership proof against the last balloon version.
func (n *RaftNode) QueryDigestMembership(keyDigest hashing.Digest) (*balloon.MembershipProof, error) {
	return n.balloon.QueryDigestMembership(keyDigest)
}

// QueryMembership acts as a passthrough when an event is given to request a membership proof
// against the last balloon version.
func (n *RaftNode) QueryMembership(event []byte) (*balloon.MembershipProof, error) {
	return n.balloon.QueryMembership(event)
}

// QueryConsistency acts as a passthrough when requesting an incremental proof.
func (n *RaftNode) QueryConsistency(start, end uint64) (*balloon.IncrementalProof, error) {
	return n.balloon.QueryConsistency(start, end)
}

// Apply applies a Raft log entry to the database.
func (n *RaftNode) Apply(l *raft.Log) interface{} {
	// TODO should i use a restore mutex?

	cmd := newCommandFromRaft(l.Data)

	switch cmd.id {
	case addEventCommandType:
		var eventDigests []hashing.Digest
		if err := cmd.decode(&eventDigests); err != nil {
			return &fsmResponse{err: err}
		}
		// INFO: after applying a bulk there will be a jump in term version due to balloon version mapping.
		newState := &fsmState{l.Index, l.Term, n.balloon.Version() + uint64(len(eventDigests)-1)}
		if n.state.shouldApply(newState) {
			return n.applyAdd(eventDigests, newState)
		}
		return &fsmResponse{fmt.Errorf("state already applied!: %+v -> %+v", n.state, newState)}

	case infoSetCommandType:
		var info ClusterInfo
		if err := cmd.decode(&info); err != nil {
			return &fsmResponse{err: err}
		}
		return n.applyClusterInfo(&info)

	default:
		return &fsmResponse{fmt.Errorf("unknown command: %v", cmd.id)}

	}
}

// Snapshot returns a snapshot of the key-value store. The caller must ensure that
// no Raft transaction is taking place during this call. Hashicorp Raft
// guarantees that this function will not be called concurrently with Apply.
func (n *RaftNode) Snapshot() (raft.FSMSnapshot, error) {
	lastSeqNum, err := n.db.Snapshot()
	if err != nil {
		return nil, err
	}
	log.Debugf("Generating snapshot until seqNum: %d (balloon version %d)", lastSeqNum, n.balloon.Version())

	// change lastVersion by checkpoint structure
	return &fsmSnapshot{
		LastSeqNum:     lastSeqNum,
		BalloonVersion: n.balloon.Version(),
		Info:           n.clusterInfo}, nil
}

// Restore restores the node to a previous state.
func (n *RaftNode) Restore(rc io.ReadCloser) error {

	log.Debug("Restoring Balloon...")

	var err error
	// Set the state from the snapshot, no lock required according to
	// Hashicorp docs.
	if err = n.db.Load(rc); err != nil {
		return err
	}

	// TODO: Restore metadata??

	// log.Debug("Restoring Metadata...")
	// var sz uint64

	// // Get size of meta, read those bytes, and set to meta.
	// if err := binary.Read(rc, binary.LittleEndian, &sz); err != nil {
	// 	return err
	// }
	// meta := make([]byte, sz)
	// if _, err := io.ReadFull(rc, meta); err != nil {
	// 	return err
	// }
	// err = func() error {
	// 	n.metaMu.Lock()
	// 	defer n.metaMu.Unlock()
	// 	return json.Unmarshal(meta, &n.meta)
	// }()

	return n.balloon.RefreshVersion()
}

// Backup ...
func (n *RaftNode) Backup() error {
	metadata := fmt.Sprintf("%d", n.balloon.Version())
	err := n.db.Backup(metadata)
	if err != nil {
		return err
	}
	log.Debugf("Generating backup until version: %d", n.balloon.Version())

	return nil
}

// BackupsInfo ...
func (n *RaftNode) BackupsInfo() []*storage.BackupInfo {
	log.Debugf("Retrieving backups information")
	return n.db.GetBackupsInfo()
}

func (n *RaftNode) applyAdd(hashes []hashing.Digest, state *fsmState) *fsmAddResponse {

	resp := new(fsmAddResponse)

	snapshotBulk, mutations, err := n.balloon.AddBulk(hashes)
	if err != nil {
		resp.err = err
		return resp
	}

	stateBuff, err := state.encode()
	if err != nil {
		resp.err = err
		return resp
	}
	mutations = append(mutations, storage.NewMutation(storage.FSMStateTable, storage.FSMStateTableKey, stateBuff))

	err = n.db.Mutate(mutations, nil)
	if err != nil {
		resp.err = err
		return resp
	}
	n.state = state
	resp.snapshot = snapshotBulk

	return resp
}

func (n *RaftNode) applyClusterInfo(info *ClusterInfo) *fsmResponse {
	n.infoMu.Lock()
	for id, data := range info.Nodes {
		if id != n.info.NodeId {
			n.clusterInfo.Nodes[id] = data
		}
	}
	n.infoMu.Unlock()
	return &fsmResponse{}
}
