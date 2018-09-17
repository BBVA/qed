package balloon

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/bbva/qed/balloon/common"
	"github.com/bbva/qed/log"
	raftbadger "github.com/bbva/raft-badger"
	"github.com/dgraph-io/badger"
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

	dbPath string // Path to underlying Badger files, if not in-memory.
	dbOpts badger.Options
	db     *badger.DB // The underlying Badger database.

	lock     sync.RWMutex
	closedMu sync.Mutex
	closed   bool // Has the RaftBalloon been closed?

	wg   sync.WaitGroup
	done chan struct{}

	raftLog         raft.LogStore           // Persistent log store.
	raftStable      raft.StableStore        // Persistent k-v store.
	raftBadgerStore *raftbadger.BadgerStore // Physical store.

	fsm *BalloonFSM // balloon's finite state machine
}

// New returns a new RaftBalloon.
func New(dbPath, raftDir, raftBindAddr, raftID string) *RaftBalloon {
	return &RaftBalloon{
		dbPath:       dbPath,
		raftDir:      raftDir,
		raftBindAddr: raftBindAddr,
		raftID:       raftID,
	}
}

// Open opens the Balloon. If enableSingle is set, and there are no existing peers,
// then this node becomes the first node, and therefore, leader of the cluster.
func (b *RaftBalloon) Open(enableSingle bool) error {

	b.closedMu.Lock()
	defer b.closedMu.Unlock()
	if b.closed {
		return ErrBalloonInvalidState
	}

	log.Infof("opening balloon with node ID %s", b.raftID)

	log.Infof("ensuring directory at %s exists", b.raftDir)
	if err := os.MkdirAll(b.raftDir, 0755); err != nil {
		return err
	}

	// Create underlying database.
	if err := b.createDatabase(); err != nil {
		return err
	}

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

	// Create the log store and stable store
	logStore, err := raftbadger.NewBadgerStore(b.raftDir + "/logs")
	if err != nil {
		return fmt.Errorf("new badger store: %s", err)
	}
	stableStore, err := raftbadger.NewBadgerStore(b.raftDir + "/config")
	if err != nil {
		return fmt.Errorf("new badger store: %s", err)
	}

	// Instantiate balloon FSM
	b.fsm = NewBalloonFSM(b.dbPath, common.NewSha256Hasher)

	// Instantiate the Raft system
	ra, err := raft.NewRaft(config, b.fsm, logStore, stableStore, snapshots, transport)
	if err != nil {
		return fmt.Errorf("new raft: %s", err)
	}
	b.raft = ra

	if enableSingle {
		log.Info("bootstrap needed")
		configuration := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      config.LocalID,
					Address: transport.LocalAddr(),
				},
			},
		}
		ra.BootstrapCluster(configuration)
	} else {
		log.Info("no bootstrap needed")
	}

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
	if err := b.db.Close(); err != nil {
		return err
	}

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
	if b.raftBadgerStore != nil {
		if err := b.raftBadgerStore.Close(); err != nil {
			return err
		}
		b.raftBadgerStore = nil
	}
	b.raftLog = nil
	b.raftStable = nil

	return nil
}

// createDatabase creates the file-based database.
func (b *RaftBalloon) createDatabase() error {
	// as it will be rebuilt from (possibly) a snapshot and committed log entries.
	if err := os.Remove(b.dbPath); err != nil && !os.IsNotExist(err) { // TODO not sure of this
		return err
	}
	db, err := badger.Open(b.dbOpts)
	if err != nil {
		return err
	}
	log.Infof("Badger database opened at %s", b.dbPath)
	b.db = db
	return nil
}

func (b *RaftBalloon) Add(event []byte) (*Commitment, error) {
	cmd, err := newCommand(insert, newInsertSubCommand(event))
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
	resp := f.Response().(fsmAddResponse)
	return resp.commitment, nil
}

func (b RaftBalloon) QueryMembership(event []byte, version uint64) (*MembershipProof, error) {
	return b.fsm.QueryMembership(event, version)
}

func (b RaftBalloon) QueryConsistency(start, end uint64) (*IncrementalProof, error) {
	return b.fsm.QueryConsistency(start, end)
}
