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

package raftwal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/hashicorp/go-msgpack/codec"
	"github.com/hashicorp/raft"

	"github.com/bbva/qed/balloon"
	"github.com/bbva/qed/crypto/hashing"
	"github.com/bbva/qed/log"
	"github.com/bbva/qed/raftwal/commands"
	"github.com/bbva/qed/storage"
)

// fsmGenericResponse is used when an unexpected output is received from
// any operation.
type fsmGenericResponse struct {
	error error
}

// fsmAddResponse is the output data structure from an add operation.
type fsmAddResponse struct {
	snapshot *balloon.Snapshot
	error    error
}

// fsmAddBulkResponse is the output data structure from an addBulk operation.
type fsmAddBulkResponse struct {
	snapshotBulk []*balloon.Snapshot
	error        error
}

type BalloonFSM struct {
	hasherF func() hashing.Hasher

	store   storage.ManagedStore
	balloon *balloon.Balloon
	state   *fsmState

	metaMu sync.RWMutex
	meta   map[string]map[string]string

	restoreMu sync.RWMutex // Restore needs exclusive access to database.
}

func loadState(s storage.ManagedStore) (*fsmState, error) {
	var state fsmState
	kvstate, err := s.Get(storage.FSMStateTable, storage.FSMStateTableKey)
	if err == storage.ErrKeyNotFound {
		log.Infof("Unable to find previous state: assuming a clean instance")
		return &fsmState{0, 0, 0}, nil
	}
	if err != nil {
		return nil, err
	}
	err = decodeMsgPack(kvstate.Value, &state)

	return &state, err
}

// NewBalloonFSM function creates a balloon with stored in a given storage, and tries to recover
// the FSM state from disk.
func NewBalloonFSM(store storage.ManagedStore, hasherF func() hashing.Hasher) (*BalloonFSM, error) {

	b, err := balloon.NewBalloon(store, hasherF)
	if err != nil {
		return nil, err
	}
	state, err := loadState(store)
	if err != nil {
		log.Infof("There was an error recovering the FSM state!!")
		return nil, err
	}

	return &BalloonFSM{
		hasherF: hasherF,
		store:   store,
		balloon: b,
		state:   state,
		meta:    make(map[string]map[string]string),
	}, nil
}

// QueryDigestMembershipConsistency acts as a passthrough when an event digest is given to
// request a membership proof against a certain balloon version.
func (fsm *BalloonFSM) QueryDigestMembershipConsistency(keyDigest hashing.Digest, version uint64) (*balloon.MembershipProof, error) {
	return fsm.balloon.QueryDigestMembershipConsistency(keyDigest, version)
}

// QueryMembershipConsistency acts as a passthrough when an event is given to request a
// membership proof against a certain balloon version.
func (fsm *BalloonFSM) QueryMembershipConsistency(event []byte, version uint64) (*balloon.MembershipProof, error) {
	return fsm.balloon.QueryMembershipConsistency(event, version)
}

// QueryDigestMembership acts as a passthrough when an event digest is given to request a
// membership proof against the last balloon version.
func (fsm *BalloonFSM) QueryDigestMembership(keyDigest hashing.Digest) (*balloon.MembershipProof, error) {
	return fsm.balloon.QueryDigestMembership(keyDigest)
}

// QueryMembership acts as a passthrough when an event is given to request a membership proof
// against the last balloon version.
func (fsm *BalloonFSM) QueryMembership(event []byte) (*balloon.MembershipProof, error) {
	return fsm.balloon.QueryMembership(event)
}

// QueryConsistency acts as a passthrough when requesting an incremental proof.
func (fsm *BalloonFSM) QueryConsistency(start, end uint64) (*balloon.IncrementalProof, error) {
	return fsm.balloon.QueryConsistency(start, end)
}

type fsmState struct {
	Index, Term, BalloonVersion uint64
}

func (s fsmState) shouldApply(f *fsmState) bool {
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

// Apply applies a Raft log entry to the database.
func (fsm *BalloonFSM) Apply(l *raft.Log) interface{} {
	// TODO should i use a restore mutex?

	buf := l.Data
	cmdType := commands.CommandType(buf[0])

	switch cmdType {
	case commands.AddEventCommandType:
		var cmd commands.AddEventCommand
		if err := commands.Decode(buf[1:], &cmd); err != nil {
			return &fsmAddResponse{error: err}
		}
		newState := &fsmState{l.Index, l.Term, fsm.balloon.Version()}
		if fsm.state.shouldApply(newState) {
			return fsm.applyAdd(cmd.EventDigest, newState)
		}
		return &fsmAddResponse{error: fmt.Errorf("state already applied!: %+v -> %+v", fsm.state, newState)}

	case commands.AddEventsBulkCommandType:
		var cmd commands.AddEventsBulkCommand
		if err := commands.Decode(buf[1:], &cmd); err != nil {
			return &fsmAddBulkResponse{error: err}
		}
		// INFO: after applying a bulk there will be a jump in term version due to balloon version mapping.
		newState := &fsmState{l.Index, l.Term, fsm.balloon.Version() + uint64(len(cmd.EventDigests)-1)}
		if fsm.state.shouldApply(newState) {
			return fsm.applyAddBulk(cmd.EventDigests, newState)
		}
		return &fsmAddBulkResponse{error: fmt.Errorf("state already applied!: %+v -> %+v", fsm.state, newState)}

	case commands.MetadataSetCommandType:
		var cmd commands.MetadataSetCommand
		if err := commands.Decode(buf[1:], &cmd); err != nil {
			return &fsmGenericResponse{error: err}
		}

		fsm.metaMu.Lock()
		defer fsm.metaMu.Unlock()
		if _, ok := fsm.meta[cmd.Id]; !ok {
			fsm.meta[cmd.Id] = make(map[string]string)
		}
		for k, v := range cmd.Data {
			fsm.meta[cmd.Id][k] = v
		}

		return &fsmGenericResponse{}

	case commands.MetadataDeleteCommandType:
		var cmd commands.MetadataDeleteCommand
		if err := commands.Decode(buf[1:], &cmd); err != nil {
			return &fsmGenericResponse{error: err}
		}

		fsm.metaMu.Lock()
		defer fsm.metaMu.Unlock()
		delete(fsm.meta, cmd.Id)

		return &fsmGenericResponse{}

	default:
		return &fsmGenericResponse{error: fmt.Errorf("unknown command: %v", cmdType)}

	}
}

// Snapshot returns a snapshot of the key-value store. The caller must ensure that
// no Raft transaction is taking place during this call. Hashicorp Raft
// guarantees that this function will not be called concurrently with Apply.
func (fsm *BalloonFSM) Snapshot() (raft.FSMSnapshot, error) {
	fsm.restoreMu.Lock()
	defer fsm.restoreMu.Unlock()

	id, err := fsm.store.Snapshot()
	if err != nil {
		return nil, err
	}
	log.Debugf("Generating snapshot until version: %d (balloon version %d)", id, fsm.balloon.Version())

	// Copy the node metadata.
	meta, err := json.Marshal(fsm.meta)
	if err != nil {
		log.Debugf("failed to encode meta for snapshot: %s", err.Error())
		return nil, err
	}
	// change lastVersion by checkpoint structure
	return &fsmSnapshot{id: id, store: fsm.store, meta: meta}, nil
}

// Restore restores the node to a previous state.
func (fsm *BalloonFSM) Restore(rc io.ReadCloser) error {

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

// Backup function calls store's backup function, passing certain metadata.
// Previously, it gets balloon version to build this metadata.
func (fsm *BalloonFSM) Backup() error {
	fsm.restoreMu.Lock()
	defer fsm.restoreMu.Unlock()

	v := fsm.balloon.Version()
	metadata := fmt.Sprintf("%d", v-1)
	err := fsm.store.Backup(metadata)
	if err != nil {
		return err
	}
	log.Debugf("Generating backup until version: %d", v-1)

	return nil
}

// DeleteBackup function is a passthough to store's equivalent funcion.
func (fsm *BalloonFSM) DeleteBackup(backupID uint32) error {
	log.Debugf("Deleting backup %d", backupID)
	return fsm.store.DeleteBackup(backupID)
}

// BackupsInfo function is a passthough to store's equivalent funcion.
func (fsm *BalloonFSM) BackupsInfo() []*storage.BackupInfo {
	log.Debugf("Retrieving backups information")
	return fsm.store.GetBackupsInfo()
}

// Close function closes
func (fsm *BalloonFSM) Close() error {
	fsm.balloon.Close()
	return nil
}

func (fsm *BalloonFSM) applyAdd(eventHash hashing.Digest, state *fsmState) *fsmAddResponse {

	snapshot, mutations, err := fsm.balloon.Add(eventHash)
	if err != nil {
		return &fsmAddResponse{error: err}
	}

	stateBuff, err := encodeMsgPack(state)
	if err != nil {
		return &fsmAddResponse{error: err}
	}

	mutations = append(mutations, storage.NewMutation(storage.FSMStateTable, storage.FSMStateTableKey, stateBuff.Bytes()))
	err = fsm.store.Mutate(mutations)
	if err != nil {
		return &fsmAddResponse{error: err}
	}
	fsm.state = state

	return &fsmAddResponse{snapshot: snapshot}
}

func (fsm *BalloonFSM) applyAddBulk(hashes []hashing.Digest, state *fsmState) *fsmAddBulkResponse {

	snapshotBulk, mutations, err := fsm.balloon.AddBulk(hashes)
	if err != nil {
		return &fsmAddBulkResponse{error: err}
	}

	stateBuff, err := encodeMsgPack(state)
	if err != nil {
		return &fsmAddBulkResponse{error: err}
	}

	mutations = append(mutations, storage.NewMutation(storage.FSMStateTable, storage.FSMStateTableKey, stateBuff.Bytes()))
	err = fsm.store.Mutate(mutations)
	if err != nil {
		return &fsmAddBulkResponse{error: err}
	}
	fsm.state = state

	return &fsmAddBulkResponse{snapshotBulk: snapshotBulk}
}

// Decode reverses the encode operation on a byte slice input
func decodeMsgPack(buf []byte, out interface{}) error {
	r := bytes.NewBuffer(buf)
	hd := codec.MsgpackHandle{}
	dec := codec.NewDecoder(r, &hd)
	return dec.Decode(out)
}

// Encode writes an encoded object to a new bytes buffer
func encodeMsgPack(in interface{}) (*bytes.Buffer, error) {
	buf := bytes.NewBuffer(nil)
	hd := codec.MsgpackHandle{}
	enc := codec.NewEncoder(buf, &hd)
	err := enc.Encode(in)
	return buf, err
}

// Metadata returns the value for a given key, for a given node ID.
func (fsm *BalloonFSM) Metadata(id, key string) string {
	fsm.metaMu.RLock()
	defer fsm.metaMu.RUnlock()

	if _, ok := fsm.meta[id]; !ok {
		return ""
	}
	v, ok := fsm.meta[id][key]
	if ok {
		return v
	}
	return ""
}

// setMetadata adds the metadata md to any existing metadata for
// the given node ID.
func (fsm *BalloonFSM) setMetadata(id string, md map[string]string) *commands.MetadataSetCommand {
	// Check local data first.
	if func() bool {
		fsm.metaMu.RLock()
		defer fsm.metaMu.RUnlock()
		if _, ok := fsm.meta[id]; ok {
			for k, v := range md {
				if fsm.meta[id][k] != v {
					return false
				}
			}
			return true
		}
		return false
	}() {
		// Local data is same as data being pushed in,
		// nothing to do.
		return nil
	}
	cmd := &commands.MetadataSetCommand{Id: id, Data: md}

	return cmd
}
