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

// Package raftwall implements the RaftBalloon life cycle and functionality.
// RaftBalloon means the raft layer above a single balloon, where consensus
// and raft-cluster information operations occurs.
// All balloon operations pass throught this layer.
package raftwal

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/bbva/qed/balloon"
	"github.com/bbva/qed/crypto/hashing"
	"github.com/bbva/qed/log"
	"github.com/bbva/qed/metrics"
	"github.com/bbva/qed/protocol"
	"github.com/bbva/qed/raftwal/commands"
	"github.com/bbva/qed/raftwal/raftrocks"
	"github.com/bbva/qed/storage"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/raft"
)

const (
	retainSnapshotCount = 2
	leaderWaitDelay     = 100 * time.Millisecond
	raftLogCacheSize    = 512
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
	Add(event []byte) (*balloon.Snapshot, error)
	AddBulk(bulk [][]byte) ([]*balloon.Snapshot, error)
	QueryDigestMembershipConsistency(keyDigest hashing.Digest, version uint64) (*balloon.MembershipProof, error)
	QueryMembershipConsistency(event []byte, version uint64) (*balloon.MembershipProof, error)
	QueryDigestMembership(keyDigest hashing.Digest) (*balloon.MembershipProof, error)
	QueryMembership(event []byte) (*balloon.MembershipProof, error)
	QueryConsistency(start, end uint64) (*balloon.IncrementalProof, error)
	// Join joins the node, identified by nodeID and reachable at addr, to the cluster
	Join(nodeID, addr string, metadata map[string]string) error
	Info() map[string]interface{}
	Backup() error
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
		db         storage.ManagedStore    // Persistent database
		log        raft.LogStore           // Persistent log store
		rocksStore *raftrocks.RocksDBStore // Underlying rocksdb-backed persistent log store
		//stable    *raftrocks.RocksDBStore // Persistent k-v store
		snapshots *raft.FileSnapshotStore // Persistent snapstop store
	}

	sync.Mutex
	closed bool
	wg     sync.WaitGroup
	done   chan struct{}

	fsm         *BalloonFSM             // balloon's finite state machine
	snapshotsCh chan *protocol.Snapshot // channel to publish snapshots

	metrics *raftBalloonMetrics
	hasherF func() hashing.Hasher
}

// NewRaftBalloon returns a new RaftBalloon.
func NewRaftBalloon(path, addr, id string, store storage.ManagedStore, snapshotsCh chan *protocol.Snapshot) (*RaftBalloon, error) {

	// Create the log store and stable store
	rocksStore, err := raftrocks.New(raftrocks.Options{Path: path + "/wal", NoSync: true, EnableStatistics: true})
	if err != nil {
		return nil, fmt.Errorf("cannot create a new rocksdb log store: %s", err)
	}
	logStore, err := raft.NewLogCache(raftLogCacheSize, rocksStore)
	if err != nil {
		return nil, fmt.Errorf("cannot create a new cached store: %s", err)
	}

	// stableStore, err := raftrocks.New(raftrocks.Options{Path: path + "/config", NoSync: true})
	// if err != nil {
	// 	return nil, fmt.Errorf("cannot create a new rocksdb stable store: %s", err)
	// }

	// Set hashing function
	hasherF := hashing.NewSha256Hasher

	// Instantiate balloon FSM
	fsm, err := NewBalloonFSM(store, hasherF)
	if err != nil {
		return nil, fmt.Errorf("new balloon fsm: %s", err)
	}

	rb := &RaftBalloon{
		path:        path,
		addr:        addr,
		id:          id,
		done:        make(chan struct{}),
		fsm:         fsm,
		snapshotsCh: snapshotsCh,
		hasherF:     hasherF,
	}

	rb.store.db = store
	rb.store.log = logStore
	rb.store.rocksStore = rocksStore
	rb.metrics = newRaftBalloonMetrics(rb)

	return rb, nil
}

// Open opens the Balloon. If no joinAddr is provided, then there are no existing peers,
// then this node becomes the first node, and therefore, leader of the cluster.
func (b *RaftBalloon) Open(bootstrap bool, metadata map[string]string) error {
	b.Lock()
	defer b.Unlock()

	if b.closed {
		return ErrBalloonInvalidState
	}

	log.Infof("opening balloon with node ID %s", b.id)

	// Setup Raft configuration
	b.raft.config = raft.DefaultConfig()
	b.raft.config.LocalID = raft.ServerID(b.id)
	b.raft.config.Logger = hclog.Default()
	b.raft.applyTimeout = 10 * time.Second

	// Setup Raft communication
	raddr, err := net.ResolveTCPAddr("tcp", b.addr)
	if err != nil {
		return err
	}

	b.raft.transport, err = raft.NewTCPTransportWithLogger(b.addr, raddr, 3, 10*time.Second, log.GetLogger())
	if err != nil {
		return err
	}

	// Create the snapshot store. This allows the Raft to truncate the log. The library creates
	// a folder to store the snapshots in.
	b.store.snapshots, err = raft.NewFileSnapshotStoreWithLogger(b.path, retainSnapshotCount, log.GetLogger())
	if err != nil {
		return fmt.Errorf("file snapshot store: %s", err)
	}

	// Instantiate the Raft system
	b.raft.api, err = raft.NewRaft(b.raft.config, b.fsm, b.store.log, b.store.rocksStore, b.store.snapshots, b.raft.transport)
	if err != nil {
		return fmt.Errorf("new raft: %s", err)
	}

	// If master node...
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

		// Metadata
		if err := b.SetMetadata(b.id, metadata); err != nil {
			return err
		}

	} else {
		log.Info("no bootstrap needed")
	}

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
	if err := b.store.rocksStore.Close(); err != nil {
		return err
	}

	b.store.rocksStore = nil
	b.store.log = nil
	b.metrics = nil

	// Close FSM
	b.fsm.Close()
	b.fsm = nil

	// close database
	if err := b.store.db.Close(); err != nil {
		return err
	}
	b.store.db = nil

	return nil
}

// Wait until node becomes leader or time is out
func (b *RaftBalloon) WaitForLeader(timeout time.Duration) (string, error) {
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

// Returns whether current node is leader or not.
func (b *RaftBalloon) IsLeader() bool {
	return b.raft.api.State() == raft.Leader
}

// Addr returns the address of the store.
func (b *RaftBalloon) Addr() string {
	return string(b.raft.transport.LocalAddr())
}

// LeaderAddr returns the Raft address of the current leader. Returns a
// blank string if there is no leader.
func (b *RaftBalloon) LeaderAddr() string {
	return string(b.raft.api.Leader())
}

// ID returns the Raft ID of the store.
func (b *RaftBalloon) ID() string {
	return b.id
}

// LeaderID returns the node ID of the Raft leader. Returns a
// blank string if there is no leader, or an error.
func (b *RaftBalloon) LeaderID() (string, error) {
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
func (b *RaftBalloon) Nodes() ([]raft.Server, error) {
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

	cmd := &commands.MetadataDeleteCommand{Id: id}
	_, err := b.raftApply(commands.MetadataDeleteCommandType, cmd)

	return err
}

// applies a command into the Raft.
func (b *RaftBalloon) raftApply(t commands.CommandType, cmd interface{}) (interface{}, error) {
	buf, err := commands.Encode(t, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to encode request: %v", err)
	}
	future := b.raft.api.Apply(buf, b.raft.applyTimeout)
	if err := future.Error(); err != nil {
		return nil, err
	}
	return future.Response(), nil
}

/*
	RaftBalloon API implements the Ballon API in the RAFT system
*/

// Add function applies an add operation into a Raft balloon.
// As a result, it returns a shapshot, but previously it sends the snapshot
// to the agents channel, in order to be published/queried.
func (b *RaftBalloon) Add(event []byte) (*balloon.Snapshot, error) {
	// Hash events
	eventDigest := b.hasherF().Do(event)
	// Create and apply command.
	cmd := &commands.AddEventCommand{EventDigest: eventDigest}
	resp, err := b.raftApply(commands.AddEventCommandType, cmd)
	if err != nil {
		return nil, err
	}
	b.metrics.Adds.Inc()

	snapshot := resp.(*fsmAddResponse).snapshot
	p := protocol.Snapshot(*snapshot)

	//Send snapshot to the snapshot channel
	b.snapshotsCh <- &p // TODO move this to an upper layer (shard manager?)

	return snapshot, nil
}

// AddBulk function applies an add bulk operation into a Raft balloon.
// As a result, it returns a bulk of shapshots, but previously it sends each snapshot
// of the bulk to the agents channel, in order to be published/queried.
func (b *RaftBalloon) AddBulk(bulk [][]byte) ([]*balloon.Snapshot, error) {
	// Hash events
	var eventHashBulk []hashing.Digest
	for _, event := range bulk {
		eventHashBulk = append(eventHashBulk, b.hasherF().Do(event))
	}
	// Create and apply command.
	cmd := &commands.AddEventsBulkCommand{EventDigests: eventHashBulk}
	resp, err := b.raftApply(commands.AddEventsBulkCommandType, cmd)
	if err != nil {
		return nil, err
	}
	b.metrics.Adds.Add(float64(len(bulk)))

	snapshotBulk := resp.(*fsmAddBulkResponse).snapshotBulk

	//Send snapshot to the snapshot channel
	// TODO move this to an upper layer (shard manager?)
	for _, s := range snapshotBulk {
		p := protocol.Snapshot(*s)
		b.snapshotsCh <- &p
	}

	return snapshotBulk, nil
}

// QueryDigestMembershipConsistency acts as a passthrough when an event digest is given to
// request a membership proof against a certain balloon version.
func (b *RaftBalloon) QueryDigestMembershipConsistency(keyDigest hashing.Digest, version uint64) (*balloon.MembershipProof, error) {
	b.metrics.DigestMembershipQueries.Inc()
	return b.fsm.QueryDigestMembershipConsistency(keyDigest, version)
}

// QueryMembershipConsistency acts as a passthrough when an event is given to request a
// membership proof against a certain balloon version.
func (b *RaftBalloon) QueryMembershipConsistency(event []byte, version uint64) (*balloon.MembershipProof, error) {
	b.metrics.MembershipQueries.Inc()
	return b.fsm.QueryMembershipConsistency(event, version)
}

// QueryDigestMembership acts as a passthrough when an event digest is given to request a
// membership proof against the last balloon version.
func (b *RaftBalloon) QueryDigestMembership(keyDigest hashing.Digest) (*balloon.MembershipProof, error) {
	b.metrics.DigestMembershipQueries.Inc()
	return b.fsm.QueryDigestMembership(keyDigest)
}

// QueryMembership acts as a passthrough when an event is given to request a membership proof
// against the last balloon version.
func (b *RaftBalloon) QueryMembership(event []byte) (*balloon.MembershipProof, error) {
	b.metrics.MembershipQueries.Inc()
	return b.fsm.QueryMembership(event)
}

// QueryConsistency acts as a passthrough when requesting an incremental proof.
func (b *RaftBalloon) QueryConsistency(start, end uint64) (*balloon.IncrementalProof, error) {
	b.metrics.IncrementalQueries.Inc()
	return b.fsm.QueryConsistency(start, end)
}

// Join joins a node, identified by id and located at addr, to this store.
// The node must be ready to respond to Raft communications at that address.
// This must be called from the Leader or it will fail.
func (b *RaftBalloon) Join(nodeID, addr string, metadata map[string]string) error {

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

	// Metadata
	if err := b.SetMetadata(nodeID, metadata); err != nil {
		return err
	}

	log.Infof("node %s at %s joined successfully", nodeID, addr)
	return nil
}

// SetMetadata adds the metadata md to any existing metadata for
// this node.
func (b *RaftBalloon) SetMetadata(nodeInvolved string, md map[string]string) error {
	cmd := b.fsm.setMetadata(nodeInvolved, md)
	_, err := b.WaitForLeader(5 * time.Second)
	if err != nil {
		return err
	}

	resp, err := b.raftApply(commands.MetadataSetCommandType, cmd)
	if err != nil {
		return err
	}

	return resp.(*fsmGenericResponse).error
}

// Info function returns Raft current node info plus certain raft cluster
// info. Used in /info/shard.
func (b *RaftBalloon) Info() map[string]interface{} {
	m := make(map[string]interface{})
	m["nodeID"] = b.ID()
	m["leaderID"], _ = b.LeaderID()
	m["meta"] = b.fsm.meta
	return m
}

// RegisterMetrics register raft metrics: prometheus collectors and storage metrics.
func (b *RaftBalloon) RegisterMetrics(registry metrics.Registry) {
	if registry != nil {
		b.store.rocksStore.RegisterMetrics(registry)
	}
	registry.MustRegister(b.metrics.collectors()...)
}

func (b *RaftBalloon) Backup() error {
	return b.fsm.Backup()
}
