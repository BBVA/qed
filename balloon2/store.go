package balloon2

import (
	"errors"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/bbva/qed/balloon2/common"
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
)

// BalloonStore is a replicated verifiable key-value store, where changes are made via Raft consensus.
type BalloonStore struct {
	raftDir      string
	raftBindAddr string
	raftID       string     // Node ID.
	raft         *raft.Raft // The consensus mechanism.

	dbPath string // Path to underlying Badger files, if not in-memory.
	dbOpts badger.Options
	db     *badger.DB // The underlying Badger database.

	lock     sync.RWMutex
	closedMu sync.Mutex
	closed   bool // Has the BalloonStore been closed?

	wg   sync.WaitGroup
	done chan struct{}

	raftLog         raft.LogStore           // Persistent log store.
	raftStable      raft.StableStore        // Persistent k-v store.
	raftBadgerStore *raftbadger.BadgerStore // Physical store.

}

// New returns a new BalloonStore.
func New(dbPath, raftDir, raftBindAddr, raftID string) *BalloonStore {
	return &BalloonStore{
		dbPath:       dbPath,
		raftDir:      raftDir,
		raftBindAddr: raftBindAddr,
		raftID:       raftID,
	}
}

// Open opens the Balloon. If enableSingle is set, and there are no existing peers,
// then this node becomes the first node, and therefore, leader of the cluster.
func (b *BalloonStore) Open(enableSingle bool) error {

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
	fsm := NewBalloonFSM(b.dbPath, common.NewSha256Hasher)

	// Instantiate the Raft system
	ra, err := raft.NewRaft(config, fsm, logStore, stableStore, snapshots, transport)
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

// Close closes the BalloonStore. If wait is true, waits for a graceful shutdown.
// Once closed, a BalloonStore may not be re-opened.
func (b *BalloonStore) Close(wait bool) error {
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
func (b *BalloonStore) createDatabase() error {
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
