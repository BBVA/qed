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
	raftDir      string
	raftBindAddr string
	raftID       string     // Node ID.
	raft         *raft.Raft // The consensus mechanism.

	lock     sync.RWMutex
	closedMu sync.Mutex
	closed   bool // Has the RaftBalloon been closed?

	wg   sync.WaitGroup
	done chan struct{}

	store       storage.ManagedStore    // Persistent database
	logStore    *raftbadger.BadgerStore // Persistent log store.
	stableStore *raftbadger.BadgerStore // Persistent k-v store.

	fsm *BalloonFSM // balloon's finite state machine
}

// New returns a new RaftBalloon.
func NewRaftBalloon(raftDir, raftBindAddr, raftID string, store storage.ManagedStore) (*RaftBalloon, error) {

	// Create the log store and stable store
	logStore, err := raftbadger.NewBadgerStore(raftDir + "/logs")
	if err != nil {
		return nil, fmt.Errorf("new badger store: %s", err)
	}
	stableStore, err := raftbadger.NewBadgerStore(raftDir + "/config")
	if err != nil {
		return nil, fmt.Errorf("new badger store: %s", err)
	}

	// Instantiate balloon FSM
	fsm, err := NewBalloonFSM(store, hashing.NewSha256Hasher)
	if err != nil {
		return nil, fmt.Errorf("new balloon fsm: %s", err)
	}

	return &RaftBalloon{
		raftDir:      raftDir,
		raftBindAddr: raftBindAddr,
		raftID:       raftID,
		done:         make(chan struct{}),
		store:        store,
		logStore:     logStore,
		stableStore:  stableStore,
		fsm:          fsm,
	}, nil
}

// Open opens the Balloon. If no joinAddr is provided, then there are no existing peers,
// then this node becomes the first node, and therefore, leader of the cluster.
// Otherwise, it will try to join the cluster via the aforementioned joinAddr.
func (b *RaftBalloon) Open(joinAddr, raftAddr, nodeID string) error {

	b.closedMu.Lock()
	defer b.closedMu.Unlock()
	if b.closed {
		return ErrBalloonInvalidState
	}

	log.Infof("opening balloon with node ID %s", b.raftID)

	// Setup Raft configuration
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(b.raftID)

	// Setup Raft communication
	addr, err := net.ResolveTCPAddr("tcp", b.raftBindAddr)
	if err != nil {
		return err
	}
	transport, err := raft.NewTCPTransport(b.raftBindAddr, addr, 3, 10*time.Second, os.Stderr)
	if err != nil {
		return err
	}

	// Create the snapshot store. This allows the Raft to truncate the log.
	snapshots, err := raft.NewFileSnapshotStore(b.raftDir, retainSnapshotCount, os.Stderr)
	if err != nil {
		return fmt.Errorf("file snapshot store: %s", err)
	}

	// Instantiate the Raft system
	b.raft, err = raft.NewRaft(config, b.fsm, b.logStore, b.stableStore, snapshots, transport)
	if err != nil {
		return fmt.Errorf("new raft: %s", err)
	}

	if joinAddr == "" {
		log.Info("bootstrap needed")
		configuration := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      config.LocalID,
					Address: transport.LocalAddr(),
				},
			},
		}
		b.raft.BootstrapCluster(configuration)
	} else {
		log.Info("no bootstrap needed")
		// If join was specified, make the join request.
		if err := join(joinAddr, raftAddr, nodeID); err != nil {
			log.Fatalf("failed to join node at %s: %s", joinAddr, err.Error())
		}
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

// Join joins a node, identified by id and located ad addr, to this store.
// The node must be ready to respond to Raft communications at that address.
func (b *RaftBalloon) Join(nodeID, addr string) error {

	log.Infof("received join request for remote node %s at %s", nodeID, addr)

	configFuture := b.raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		log.Errorf("failed to get raft configuration: %v", err)
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

			future := b.raft.RemoveServer(srv.ID, 0, 0)
			if err := future.Error(); err != nil {
				return fmt.Errorf("error removing existing node %s at %s: %s", nodeID, addr, err)
			}
		}
	}

	f := b.raft.AddVoter(raft.ServerID(nodeID), raft.ServerAddress(addr), 0, 0)
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
	b.closedMu.Lock()
	defer b.closedMu.Unlock()
	if b.closed {
		return nil
	}
	defer func() {
		b.closed = true
	}()

	close(b.done)
	b.wg.Wait()

	// close database
	if err := b.store.Close(); err != nil {
		return err
	}
	b.store = nil

	// shutdown raft
	if b.raft != nil {
		f := b.raft.Shutdown()
		if wait {
			if e := f.(raft.Future); e.Error() != nil {
				return e.Error()
			}
		}
		b.raft = nil
	}

	// close raft store
	if err := b.logStore.Close(); err != nil {
		return err
	}
	if err := b.stableStore.Close(); err != nil {
		return err
	}
	b.logStore = nil
	b.stableStore = nil

	return nil
}

func (b *RaftBalloon) Add(event []byte) (*Commitment, error) {
	b.lock.Lock()
	defer b.lock.Unlock()

	cmd, err := newCommand(insert, newInsertSubCommand(event, b.fsm.Version()))
	if err != nil {
		return nil, err
	}
	cmdBytes, err := json.Marshal(cmd)
	if err != nil {
		return nil, err
	}

	f := b.raft.Apply(cmdBytes, raftTimeout)
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
