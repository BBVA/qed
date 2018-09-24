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
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/log"
	"github.com/bbva/qed/storage"
	raftbadger "github.com/bbva/raft-badger"
	"github.com/hashicorp/raft"
)

const (
	retainSnapshotCount = 2
	raftTimeout         = 10 * time.Second
	leaderWaitDelay     = 100 * time.Millisecond
)

var (
	// ErrBalloonInvalidState is returned when a Balloon is in an invalid
	// state for the requested operation.
	ErrBalloonInvalidState = errors.New("balloon not in valid state")

	// ErrNotLeader is returned when a node attempts to execute a leader-only
	// operation.
	ErrNotLeader = errors.New("not leader")
)

// RaftBalloon is the interface Raft-backed balloons must implement.
type RaftBalloonApi interface {
	Add(event []byte) (*Commitment, error)
	QueryMembership(event []byte, version uint64) (*MembershipProof, error)
	QueryConsistency(start, end uint64) (*IncrementalProof, error)
	// Join joins the node, identified by nodeID and reachable at addr, to the cluster
	Join(nodeID, addr string) error
}

// RaftBalloon is a replicated verifiable key-value store, where changes are made via Raft consensus.
type RaftBalloon struct {
	path string // Base path for the node
	addr string // Node addr
	id   string // Node ID

	raft struct {
		api          *raft.Raft             // The consensus mechanism
		transport    *raft.NetworkTransport // Raft network transport
		config       *raft.Config           //Config provides any necessary configuration for the Raft server.
		nodes        *raft.Configuration    //Configuration tracks which servers are in the cluster, and whether they have votes.
		applyTimeout time.Duration
	}

	store struct {
		db        storage.ManagedStore    // Persistent database
		log       *raftbadger.BadgerStore // Persistent log store
		stable    *raftbadger.BadgerStore // Persistent k-v store
		snapshots *raft.FileSnapshotStore // Persisten snapstop store
	}

	sync.Mutex
	closed bool
	wg     sync.WaitGroup
	done   chan struct{}

	fsm *BalloonFSM // balloon's finite state machine

}

// New returns a new RaftBalloon.
func NewRaftBalloon(path, addr, id string, store storage.ManagedStore) (*RaftBalloon, error) {

	// Create the log store and stable store
	logStore, err := raftbadger.NewBadgerStore(path + "/logs")
	if err != nil {
		return nil, fmt.Errorf("new badger store: %s", err)
	}
	stableStore, err := raftbadger.NewBadgerStore(path + "/config")
	if err != nil {
		return nil, fmt.Errorf("new badger store: %s", err)
	}

	// Instantiate balloon FSM
	fsm, err := NewBalloonFSM(store, hashing.NewSha256Hasher)
	if err != nil {
		return nil, fmt.Errorf("new balloon fsm: %s", err)
	}

	rb := &RaftBalloon{
		path: path,
		addr: addr,
		id:   id,
		done: make(chan struct{}),
		fsm:  fsm,
	}

	rb.store.db = store
	rb.store.log = logStore
	rb.store.stable = stableStore

	return rb, nil
}

// Open opens the Balloon. If no joinAddr is provided, then there are no existing peers,
// then this node becomes the first node, and therefore, leader of the cluster.
func (b *RaftBalloon) Open(bootstrap bool) error {
	b.Lock()
	defer b.Unlock()

	if b.closed {
		return ErrBalloonInvalidState
	}

	log.Infof("opening balloon with node ID %s", b.id)

	// Setup Raft configuration
	b.raft.config = raft.DefaultConfig()
	b.raft.config.LocalID = raft.ServerID(b.id)
	b.raft.applyTimeout = 10 * time.Second

	// Setup Raft communication
	raddr, err := net.ResolveTCPAddr("tcp", b.addr)
	if err != nil {
		return err
	}

	b.raft.transport, err = raft.NewTCPTransport(b.addr, raddr, 3, 10*time.Second, os.Stderr)
	if err != nil {
		return err
	}

	// Create the snapshot store. This allows the Raft to truncate the log. The library creates
	// a folder to store the snapshots in.
	b.store.snapshots, err = raft.NewFileSnapshotStore(b.path, retainSnapshotCount, os.Stderr)
	if err != nil {
		return fmt.Errorf("file snapshot store: %s", err)
	}

	// Instantiate the Raft system
	b.raft.api, err = raft.NewRaft(b.raft.config, b.fsm, b.store.log, b.store.stable, b.store.snapshots, b.raft.transport)
	if err != nil {
		return fmt.Errorf("new raft: %s", err)
	}

	if bootstrap {
		log.Info("bootstrap needed")
		b.raft.nodes = &raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      b.raft.config.LocalID,
					Address: b.raft.transport.LocalAddr(),
				},
			},
		}
		b.raft.api.BootstrapCluster(*b.raft.nodes)
	} else {
		log.Info("no bootstrap needed")
	}

	return nil
}

func join(joinAddr, raftAddr, nodeID string) error {
	b, err := json.Marshal(map[string]string{"addr": raftAddr, "id": nodeID})
	if err != nil {
		return err
	}
	resp, err := http.Post(fmt.Sprintf("http://%s/join", joinAddr), "application-type/json", bytes.NewReader(b))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// Join joins a node, identified by id and located at addr, to this store.
// The node must be ready to respond to Raft communications at that address.
// This must be called from the Leader or it will fail.
func (b *RaftBalloon) Join(nodeID, addr string) error {

	log.Infof("received join request for remote node %s at %s", nodeID, addr)

	configFuture := b.raft.api.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		log.Errorf("failed to get raft servers configuration: %v", err)
		return err
	}

	for _, srv := range configFuture.Configuration().Servers {
		// If a node already exists with either the joining node's ID or address,
		// that node may need to be removed from the config first.
		if srv.ID == raft.ServerID(nodeID) || srv.Address == raft.ServerAddress(addr) {
			// However if *both* the ID and the address are the same, then nothing -- not even
			// a join operation -- is needed.
			if srv.Address == raft.ServerAddress(addr) && srv.ID == raft.ServerID(nodeID) {
				log.Infof("node %s at %s already member of cluster, ignoring join request", nodeID, addr)
				return nil
			}

			future := b.raft.api.RemoveServer(srv.ID, 0, 0)
			if err := future.Error(); err != nil {
				return fmt.Errorf("error removing existing node %s at %s: %s", nodeID, addr, err)
			}
		}
	}

	f := b.raft.api.AddVoter(raft.ServerID(nodeID), raft.ServerAddress(addr), 0, 0)
	if e := f.(raft.Future); e.Error() != nil {
		if e.Error() == raft.ErrNotLeader {
			return ErrNotLeader
		}
		return e.Error()
	}

	log.Infof("node %s at %s joined successfully", nodeID, addr)
	return nil
}

// Close closes the RaftBalloon. If wait is true, waits for a graceful shutdown.
// Once closed, a RaftBalloon may not be re-opened.
func (b *RaftBalloon) Close(wait bool) error {
	b.Lock()
	defer b.Unlock()
	if b.closed {
		return nil
	}
	defer func() {
		b.closed = true
	}()

	close(b.done)
	b.wg.Wait()

	// close database
	if err := b.store.db.Close(); err != nil {
		return err
	}
	b.store.db = nil

	// shutdown raft
	if b.raft.api != nil {
		f := b.raft.api.Shutdown()
		if wait {
			if e := f.(raft.Future); e.Error() != nil {
				return e.Error()
			}
		}
		b.raft.api = nil
	}

	// close raft store
	if err := b.store.log.Close(); err != nil {
		return err
	}
	if err := b.store.stable.Close(); err != nil {
		return err
	}
	b.store.log = nil
	b.store.stable = nil

	return nil
}

// Wait until node becomes leader or time is out
func (b RaftBalloon) WaitForLeader(timeout time.Duration) (string, error) {
	tck := time.NewTicker(leaderWaitDelay)
	defer tck.Stop()
	tmr := time.NewTimer(timeout)
	defer tmr.Stop()

	for {
		select {
		case <-tck.C:
			l := string(b.raft.api.Leader())
			if l != "" {
				return l, nil
			}
		case <-tmr.C:
			return "", fmt.Errorf("timeout expired")
		}
	}
}

func (b RaftBalloon) IsLeader() bool {
	return b.raft.api.State() == raft.Leader
}

// Addr returns the address of the store.
func (b RaftBalloon) Addr() string {
	return string(b.raft.transport.LocalAddr())
}

// LeaderAddr returns the Raft address of the current leader. Returns a
// blank string if there is no leader.
func (b RaftBalloon) LeaderAddr() string {
	return string(b.raft.api.Leader())
}

// ID returns the Raft ID of the store.
func (b RaftBalloon) ID() string {
	return b.id
}

// LeaderID returns the node ID of the Raft leader. Returns a
// blank string if there is no leader, or an error.
func (b RaftBalloon) LeaderID() (string, error) {
	addr := b.LeaderAddr()
	configFuture := b.raft.api.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		log.Infof("failed to get raft configuration: %v", err)
		return "", err
	}

	for _, srv := range configFuture.Configuration().Servers {
		if srv.Address == raft.ServerAddress(addr) {
			return string(srv.ID), nil
		}
	}
	return "", nil
}

// Nodes returns the slice of nodes in the cluster, sorted by ID ascending.
func (b RaftBalloon) Nodes() ([]raft.Server, error) {
	f := b.raft.api.GetConfiguration()
	if f.Error() != nil {
		return nil, f.Error()
	}

	return f.Configuration().Servers, nil

}

// Remove removes a node from the store, specified by ID.
func (b *RaftBalloon) Remove(id string) error {
	log.Infof("received request to remove node %s", id)
	if err := b.remove(id); err != nil {
		log.Infof("failed to remove node %s: %s", id, err.Error())
		return err
	}

	log.Infof("node %s removed successfully", id)
	return nil
}

// remove removes the node, with the given ID, from the cluster.
func (b *RaftBalloon) remove(id string) error {
	if b.raft.api.State() != raft.Leader {
		return ErrNotLeader
	}

	f := b.raft.api.RemoveServer(raft.ServerID(id), 0, 0)
	if f.Error() != nil {
		if f.Error() == raft.ErrNotLeader {
			return ErrNotLeader
		}
		return f.Error()
	}

	// TODO: implement metadataDelete
	c, err := newCommand(metadataDelete, id)
	cb, err := json.Marshal(c)
	if err != nil {
		return err
	}
	f = b.raft.api.Apply(cb, b.raft.applyTimeout)
	if e := f.(raft.Future); e.Error() != nil {
		if e.Error() == raft.ErrNotLeader {
			return ErrNotLeader
		}
		e.Error()
	}

	return nil
}

/*
	RaftBalloon API implemnts the Ballon API in the RAFT system

*/

func (b *RaftBalloon) Add(event []byte) (*Commitment, error) {
	b.Lock()
	defer b.Unlock()

	cmd, err := newCommand(insert, newInsertSubCommand(event))
	if err != nil {
		return nil, err
	}
	cmdBytes, err := json.Marshal(cmd)
	if err != nil {
		return nil, err
	}

	f := b.raft.api.Apply(cmdBytes, raftTimeout)
	if e := f.(raft.Future); e.Error() != nil {
		if e.Error() == raft.ErrNotLeader {
			return nil, ErrNotLeader
		}
		return nil, e.Error()
	}
	resp := f.Response().(*fsmAddResponse)
	return resp.commitment, nil
}

func (b RaftBalloon) QueryMembership(event []byte, version uint64) (*MembershipProof, error) {
	return b.fsm.QueryMembership(event, version)
}

func (b RaftBalloon) QueryConsistency(start, end uint64) (*IncrementalProof, error) {
	return b.fsm.QueryConsistency(start, end)
}