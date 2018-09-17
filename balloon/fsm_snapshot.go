package balloon

import (
	"github.com/bbva/qed/db"
	"github.com/bbva/qed/log"
	"github.com/hashicorp/raft"
)

type fsmSnapshot struct {
	store db.ManagedStore
}

// Persist writes the snapshot to the given sink.
func (f *fsmSnapshot) Persist(sink raft.SnapshotSink) error {
	log.Debug("Persisting snapshot...")
	err := func() error {
		if err := f.store.Backup(sink, 0); err != nil {
			return err
		}
		return sink.Close()
	}()
	if err != nil {
		sink.Cancel()
	}
	return err
}

// Release is invoked when we are finished with the snapshot.
func (f *fsmSnapshot) Release() {
	log.Debug("Snapshot created.")
}
