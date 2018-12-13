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

package raftwal

import (
	"bytes"
	"fmt"
	"io"
	"sync"

	"github.com/bbva/qed/protocol"

	"github.com/bbva/qed/balloon"
	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/log"
	"github.com/bbva/qed/raftwal/commands"
	"github.com/bbva/qed/storage"
	"github.com/hashicorp/go-msgpack/codec"
	"github.com/hashicorp/raft"
)

type fsmGenericResponse struct {
	error error
}

type fsmAddResponse struct {
	snapshot *balloon.Snapshot
	error    error
}

type BalloonFSM struct {
	hasherF func() hashing.Hasher

	store   storage.ManagedStore
	balloon *balloon.Balloon
	state   *fsmState

	agentsQueue chan *protocol.Snapshot

	restoreMu sync.RWMutex // Restore needs exclusive access to database.
}

func loadState(s storage.ManagedStore) (*fsmState, error) {
	var state fsmState
	kvstate, err := s.Get(storage.FSMStatePrefix, []byte{0xab})
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

func NewBalloonFSM(store storage.ManagedStore, hasherF func() hashing.Hasher, agentsQueue chan *protocol.Snapshot) (*BalloonFSM, error) {

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
		hasherF:     hasherF,
		store:       store,
		balloon:     b,
		state:       state,
		agentsQueue: agentsQueue,
	}, nil
}

func (fsm *BalloonFSM) QueryDigestMembership(keyDigest hashing.Digest, version uint64) (*balloon.MembershipProof, error) {
	return fsm.balloon.QueryDigestMembership(keyDigest, version)
}

func (fsm *BalloonFSM) QueryMembership(event []byte, version uint64) (*balloon.MembershipProof, error) {
	return fsm.balloon.QueryMembership(event, version)
}

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

	if f.BalloonVersion > 0 && s.BalloonVersion != (f.BalloonVersion-1) {
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
			return fsm.applyAdd(cmd.Event, newState)
		}
		return &fsmAddResponse{error: fmt.Errorf("state already applied!: %+v -> %+v", fsm.state, newState)}
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
	version, err := fsm.store.GetLastVersion()
	if err != nil {
		return nil, err
	}
	log.Debugf("Generating snapshot until version: %d (balloon version %d)", version, fsm.balloon.Version())
	return &fsmSnapshot{lastVersion: version, store: fsm.store}, nil
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
	return fsm.balloon.RefreshVersion()
}

func (fsm *BalloonFSM) Close() error {
	return fsm.store.Close()
}

func (fsm *BalloonFSM) applyAdd(event []byte, state *fsmState) *fsmAddResponse {

	snapshot, mutations, err := fsm.balloon.Add(event)
	if err != nil {
		return &fsmAddResponse{error: err}
	}

	stateBuff, err := encodeMsgPack(state)
	if err != nil {
		return &fsmAddResponse{error: err}
	}

	mutations = append(mutations, storage.NewMutation(storage.FSMStatePrefix, []byte{0xab}, stateBuff.Bytes()))
	err = fsm.store.Mutate(mutations)
	if err != nil {
		return &fsmAddResponse{error: err}
	}
	fsm.state = state

	//Send snapshot to gossip agents
	fsm.agentsQueue <- &protocol.Snapshot{
		HistoryDigest: snapshot.HistoryDigest,
		HyperDigest:   snapshot.HyperDigest,
		Version:       snapshot.Version,
		EventDigest:   snapshot.EventDigest,
	}

	return &fsmAddResponse{snapshot: snapshot}
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
