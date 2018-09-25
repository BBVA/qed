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

package balloon

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/log"
	"github.com/bbva/qed/storage"
	"github.com/hashicorp/go-msgpack/codec"
	"github.com/hashicorp/raft"
)

type fsmGenericResponse struct {
	error error
}

type fsmAddResponse struct {
	commitment *Commitment
	error      error
}

type BalloonFSM struct {
	hasherF func() hashing.Hasher

	store   storage.ManagedStore
	balloon *Balloon
	state   *fsmState

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

func NewBalloonFSM(store storage.ManagedStore, hasherF func() hashing.Hasher) (*BalloonFSM, error) {

	balloon, err := NewBalloon(store, hasherF)
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
		balloon: balloon,
		state:   state,
	}, nil
}

func (fsm BalloonFSM) QueryMembership(event []byte, version uint64) (*MembershipProof, error) {
	return fsm.balloon.QueryMembership(event, version)
}

func (fsm BalloonFSM) QueryConsistency(start, end uint64) (*IncrementalProof, error) {
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
		panic(fmt.Sprintf("balloonVersion panic! old: %d new %d", s.BalloonVersion, f.BalloonVersion))
	}

	return true
}

// Apply applies a Raft log entry to the database.
func (fsm *BalloonFSM) Apply(l *raft.Log) interface{} {
	// TODO should i use a restore mutex

	var cmd command
	if err := json.Unmarshal(l.Data, &cmd); err != nil {
		panic(fmt.Sprintf("failed to unmarshal cluster command: %s", err.Error()))
	}

	switch cmd.Type {
	case insert:
		var sub insertSubCommand
		if err := json.Unmarshal(cmd.Sub, &sub); err != nil {
			return &fsmAddResponse{error: err}
		}
		newState := &fsmState{l.Index, l.Term, fsm.balloon.Version()}
		if fsm.state.shouldApply(newState) {
			return fsm.applyAdd(sub.Event, newState)
		}
		return &fsmAddResponse{error: fmt.Errorf("state already applied!: %+v -> %+v", fsm.state, newState)}
	default:
		return &fsmGenericResponse{error: fmt.Errorf("unknown command: %v", cmd.Type)}

	}
}

// Snapshot returns a snapshot of the key-value store. The caller must ensure that
// no Raft transaction is taking place during this call. Hashicorp Raft
// guarantees that this function will not be called concurrently with Apply.
func (fsm *BalloonFSM) Snapshot() (raft.FSMSnapshot, error) {
	fsm.restoreMu.Lock()
	defer fsm.restoreMu.Unlock()
	return &fsmSnapshot{store: fsm.store}, nil
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
	fsm.balloon.RefreshVersion()
	return err
}

func (fsm *BalloonFSM) Close() error {
	return fsm.store.Close()
}

func (fsm *BalloonFSM) applyAdd(event []byte, state *fsmState) *fsmAddResponse {

	commitment, mutations, err := fsm.balloon.Add(event)
	if err != nil {
		return &fsmAddResponse{error: err}
	}

	stateBuff, err := encodeMsgPack(state)
	if err != nil {
		return &fsmAddResponse{error: err}
	}

	mutations = append(mutations, storage.NewMutation(storage.FSMStatePrefix, []byte{0xab}, stateBuff.Bytes()))
	fsm.store.Mutate(mutations)
	fsm.state = state
	return &fsmAddResponse{commitment: commitment}
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
