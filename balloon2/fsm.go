package balloon2

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/bbva/qed/balloon2/common"
	"github.com/bbva/qed/db"
	bdb "github.com/bbva/qed/db/badger"
	"github.com/bbva/qed/util"
	"github.com/hashicorp/raft"
)

type fsmGenericResponse struct {
	error error
}

type fsmAddResponse struct {
	commitment Commitment
	error      error
}

type BalloonFSM struct {
	hasher common.Hasher

	store   db.ManagedStore
	version uint64
	balloon *Balloon

	restoreMu sync.RWMutex // Restore needs exclusive access to database.
}

func NewBalloonFSM(hasher common.Hasher, dbPath string) *BalloonFSM {
	return &BalloonFSM{
		hasher:  hasher,
		store:   bdb.NewBadgerStore(dbPath),
		version: 0,
	}
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
		return fsm.applyAdd(sub.Key, sub.Value)
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

// Restore stores the key-value store to a previous state.
func (fsm *BalloonFSM) Restore(rc io.ReadCloser) error {
	// Set the state from the snapshot, no lock required according to
	// Hashicorp docs.
	if err := fsm.store.Load(rc); err != nil {
		return err
	}

	// get stored last version
	kv, err := fsm.store.Get(db.VersionPrefix, BalloonVersionKey)
	if err != nil {
		return err
	}
	fsm.version = util.BytesAsUint64(kv.Value) + 1

	fsm.balloon = NewBalloon(fsm.version, fsm.hasher, fsm.store)

	return nil
}

func (fsm *BalloonFSM) Close() error {
	return fsm.store.Close()
}

func (fsm *BalloonFSM) applyAdd(key, value string) *fsmAddResponse {
	return nil
}
