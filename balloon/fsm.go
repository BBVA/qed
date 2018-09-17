package balloon

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/storage"
	bdb "github.com/bbva/qed/storage/badger"
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

	restoreMu sync.RWMutex // Restore needs exclusive access to database.
}

func NewBalloonFSM(dbPath string, hasherF func() hashing.Hasher) *BalloonFSM {
	return &BalloonFSM{
		hasherF: hasherF,
		store:   bdb.NewBadgerStore(dbPath),
	}
}

func (fsm BalloonFSM) QueryMembership(event []byte, version uint64) (*MembershipProof, error) {
	return fsm.balloon.QueryMembership(event, version)
}

func (fsm BalloonFSM) QueryConsistency(start, end uint64) (*IncrementalProof, error) {
	return fsm.balloon.QueryConsistency(start, end)
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
			return &fsmGenericResponse{error: err}
		}
		return fsm.applyAdd(sub.Event)
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
	var err error
	// Set the state from the snapshot, no lock required according to
	// Hashicorp docs.
	if err = fsm.store.Load(rc); err != nil {
		return err
	}
	fsm.balloon, err = NewBalloon(fsm.store, fsm.hasherF)
	return err
}

func (fsm *BalloonFSM) Close() error {
	return fsm.store.Close()
}

func (fsm *BalloonFSM) applyAdd(event []byte) *fsmAddResponse {
	commitment, mutations, err := fsm.balloon.Add(event)
	if err != nil {
		return &fsmAddResponse{error: err}
	}
	fsm.store.Mutate(mutations...)
	return &fsmAddResponse{commitment: commitment}
}
