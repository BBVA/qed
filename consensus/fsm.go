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
	"encoding/json"
	"fmt"
	"io"
	"sync"

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

type balloonFSM struct {
	hasherF func() hashing.Hasher

	store   storage.ManagedStore
	balloon *balloon.Balloon
	state   *fsmState

	clusterInfoMu sync.RWMutex
	clusterInfo   *ClusterInfo

	restoreMu sync.RWMutex // Restore needs exclusive access to database.
}

func (fsm *balloonFSM) loadState() error {
	kvstate, err := fsm.store.Get(storage.FSMStateTable, storage.FSMStateTableKey)
	if err == storage.ErrKeyNotFound {
		log.Infof("Unable to find previous state: assuming a clean instance")
		fsm.state = new(fsmState) // &fsmState{0, 0, 0}
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
	fsm.state = &state
	return nil
}

// newBalloonFSM function creates a balloon with stored in a given storage, and tries to recover
// the FSM state from disk.
func newBalloonFSM(store storage.ManagedStore, hasherF func() hashing.Hasher) (*balloonFSM, error) {

	b, err := balloon.NewBalloon(store, hasherF)
	if err != nil {
		return nil, err
	}

	fsm := &balloonFSM{
		hasherF: hasherF,
		store:   store,
		balloon: b,
	}

	err = fsm.loadState()
	if err != nil {
		log.Infof("There was an error recovering the FSM state!!")
		return nil, err
	}

	return fsm, nil
}

// QueryDigestMembershipConsistency acts as a passthrough when an event digest is given to
// request a membership proof against a certain balloon version.
func (fsm *balloonFSM) QueryDigestMembershipConsistency(keyDigest hashing.Digest, version uint64) (*balloon.MembershipProof, error) {
	return fsm.balloon.QueryDigestMembershipConsistency(keyDigest, version)
}

// QueryMembershipConsistency acts as a passthrough when an event is given to request a
// membership proof against a certain balloon version.
func (fsm *balloonFSM) QueryMembershipConsistency(event []byte, version uint64) (*balloon.MembershipProof, error) {
	return fsm.balloon.QueryMembershipConsistency(event, version)
}

// QueryDigestMembership acts as a passthrough when an event digest is given to request a
// membership proof against the last balloon version.
func (fsm *balloonFSM) QueryDigestMembership(keyDigest hashing.Digest) (*balloon.MembershipProof, error) {
	return fsm.balloon.QueryDigestMembership(keyDigest)
}

// QueryMembership acts as a passthrough when an event is given to request a membership proof
// against the last balloon version.
func (fsm *balloonFSM) QueryMembership(event []byte) (*balloon.MembershipProof, error) {
	return fsm.balloon.QueryMembership(event)
}

// QueryConsistency acts as a passthrough when requesting an incremental proof.
func (fsm *balloonFSM) QueryConsistency(start, end uint64) (*balloon.IncrementalProof, error) {
	return fsm.balloon.QueryConsistency(start, end)
}

// Apply applies a Raft log entry to the database.
func (fsm *balloonFSM) Apply(l *raft.Log) interface{} {
	// TODO should i use a restore mutex?

	cmd := NewCommandFromRaft(l.Data)

	switch cmd.id {
	case addEventCommandType:
		var eventDigests []hashing.Digest
		if err := cmd.Decode(&eventDigests); err != nil {
			return &fsmResponse{err: err}
		}

		// INFO: after applying a bulk there will be a jump in term version due to balloon version mapping.
		newState := &fsmState{l.Index, l.Term, fsm.balloon.Version() + uint64(len(eventDigests)-1)}
		if fsm.state.shouldApply(newState) {
			return fsm.applyAdd(eventDigests, newState)
		}
		return &fsmResponse{fmt.Errorf("state already applied!: %+v -> %+v", fsm.state, newState)}

	default:
		return &fsmResponse{fmt.Errorf("unknown command: %v", cmd.id)}

	}
}

// Snapshot returns a snapshot of the key-value store. The caller must ensure that
// no Raft transaction is taking place during this call. Hashicorp Raft
// guarantees that this function will not be called concurrently with Apply.
func (fsm *balloonFSM) Snapshot() (raft.FSMSnapshot, error) {
	fsm.restoreMu.Lock()
	defer fsm.restoreMu.Unlock()

	lastSeqNum, err := fsm.store.Snapshot()
	if err != nil {
		return nil, err
	}
	log.Debugf("Generating snapshot until seqNum: %d (balloon version %d)", lastSeqNum, fsm.balloon.Version())

	// Copy the node metadata.
	clusterInfo, err := json.Marshal(fsm.clusterInfo)
	if err != nil {
		log.Debugf("failed to encode meta for snapshot: %s", err.Error())
		return nil, err
	}
	// change lastVersion by checkpoint structure
	return &fsmSnapshot{
		LastSeqNum:     lastSeqNum,
		BalloonVersion: fsm.balloon.Version(),
		ClusterInfo:    clusterInfo}, nil
}

// Restore restores the node to a previous state.
func (fsm *balloonFSM) Restore(rc io.ReadCloser) error {

	log.Debug("Restoring Balloon...")

	var err error
	// Set the state from the snapshot, no lock required according to
	// Hashicorp docs.
	if err = fsm.store.Load(rc); err != nil {
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
	// 	fsm.metaMu.Lock()
	// 	defer fsm.metaMu.Unlock()
	// 	return json.Unmarshal(meta, &fsm.meta)
	// }()

	return fsm.balloon.RefreshVersion()
}

// Backup ...
func (fsm *balloonFSM) Backup() error {
	fsm.restoreMu.Lock()
	defer fsm.restoreMu.Unlock()

	metadata := fmt.Sprintf("%d", fsm.balloon.Version())
	err := fsm.store.Backup(metadata)
	if err != nil {
		return err
	}
	log.Debugf("Generating backup until version: %d", fsm.balloon.Version())

	return nil
}

// BackupsInfo ...
func (fsm *balloonFSM) BackupsInfo() []*storage.BackupInfo {
	log.Debugf("Retrieving backups information")
	return fsm.store.GetBackupsInfo()
}

// Close function closes
func (fsm *balloonFSM) Close() error {
	fsm.balloon.Close()
	return nil
}

func (fsm *balloonFSM) applyAdd(hashes []hashing.Digest, state *fsmState) *fsmAddResponse {

	resp := new(fsmAddResponse)

	snapshotBulk, mutations, err := fsm.balloon.AddBulk(hashes)
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
	err = fsm.store.Mutate(mutations, nil)
	if err != nil {
		resp.err = err
		return resp
	}
	fsm.state = state
	resp.snapshot = snapshotBulk

	return resp
}

func (fsm *balloonFSM) metaAppend(id string) {}

// // Metadata returns the value for a given key, for a given node ID.
// func (fsm *balloonFSM) Metadata(id, key string) string {
// 	fsm.metaMu.RLock()
// 	defer fsm.metaMu.RUnlock()

// 	if _, ok := fsm.meta[id]; !ok {
// 		return ""
// 	}
// 	v, ok := fsm.meta[id][key]
// 	if ok {
// 		return v
// 	}
// 	return ""
// }

// // setMetadata adds the metadata md to any existing metadata for
// // the given node ID.
// func (fsm *balloonFSM) setMetadata(id string, md map[string]string) *commands.MetadataSetCommand {
// 	// Check local data first.
// 	if func() bool {
// 		fsm.metaMu.RLock()
// 		defer fsm.metaMu.RUnlock()
// 		if _, ok := fsm.meta[id]; ok {
// 			for k, v := range md {
// 				if fsm.meta[id][k] != v {
// 					return false
// 				}
// 			}
// 			return true
// 		}
// 		return false
// 	}() {
// 		// Local data is same as data being pushed in,
// 		// nothing to do.
// 		return nil
// 	}
// 	cmd := &commands.MetadataSetCommand{Id: id, Data: md}

// 	return cmd
// }
