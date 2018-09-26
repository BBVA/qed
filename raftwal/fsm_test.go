package raftwal

import (
	"io"
	"testing"

	"github.com/hashicorp/raft"
	assert "github.com/stretchr/testify/require"

	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/raftwal/commands"
	storage_utils "github.com/bbva/qed/testutils/storage"
)

func TestApply(t *testing.T) {
	store, closeF := storage_utils.OpenBadgerStore(t, "/var/tmp/balloon.test.db")
	defer closeF()

	fsm, err := NewBalloonFSM(store, hashing.NewSha256Hasher)
	assert.NoError(t, err)

	// happy path
	r := fsm.Apply(newRaftLog(1, 1)).(*fsmAddResponse)
	assert.Nil(t, r.error)

	// Error: Command already applied
	r = fsm.Apply(newRaftLog(1, 1)).(*fsmAddResponse)
	assert.Error(t, r.error)

	// happy path
	r = fsm.Apply(newRaftLog(2, 1)).(*fsmAddResponse)
	assert.Nil(t, r.error)

	// Error: Command out of order
	r = fsm.Apply(newRaftLog(1, 1)).(*fsmAddResponse)
	assert.Error(t, r.error)

}

func TestSnapshot(t *testing.T) {
	store, closeF := storage_utils.OpenBadgerStore(t, "/var/tmp/balloon.test.db")
	defer closeF()

	fsm, err := NewBalloonFSM(store, hashing.NewSha256Hasher)
	assert.NoError(t, err)

	fsm.Apply(newRaftLog(0, 0))

	// happy path
	_, err = fsm.Snapshot()
	assert.NoError(t, err)
}

type fakeRC struct{}

func (f *fakeRC) Read(p []byte) (n int, err error) {
	return 0, io.EOF
}

func (f *fakeRC) Close() error {
	return nil
}

func TestRestore(t *testing.T) {
	store, closeF := storage_utils.OpenBadgerStore(t, "/var/tmp/balloon.test.db")
	defer closeF()

	fsm, err := NewBalloonFSM(store, hashing.NewSha256Hasher)
	assert.NoError(t, err)

	assert.NoError(t, fsm.Restore(&fakeRC{}))
}

func TestAddAndRestoreSnapshot(t *testing.T) {
	store, closeF := storage_utils.OpenBadgerStore(t, "/var/tmp/balloon.test.db")
	defer closeF()

	fsm, err := NewBalloonFSM(store, hashing.NewSha256Hasher)
	assert.NoError(t, err)

	fsm.Apply(newRaftLog(0, 0))

	fsmsnap, err := fsm.Snapshot()
	assert.NoError(t, err)

	snap := raft.NewInmemSnapshotStore()

	// Create a new sink
	var configuration raft.Configuration
	configuration.Servers = append(configuration.Servers, raft.Server{
		Suffrage: raft.Voter,
		ID:       raft.ServerID("my id"),
		Address:  raft.ServerAddress("over here"),
	})
	_, trans := raft.NewInmemTransport(raft.NewInmemAddr())
	sink, _ := snap.Create(raft.SnapshotVersionMax, 10, 3, configuration, 2, trans)

	fsmsnap.Persist(sink)
	// fsm.Close()

	// Read the latest snapshot
	snaps, _ := snap.List()
	_, r, _ := snap.Open(snaps[0].ID)

	store2, close2F := storage_utils.OpenBadgerStore(t, "/var/tmp/balloon.test.2.db")
	defer close2F()

	// New FSMStore
	fsm2, err := NewBalloonFSM(store2, hashing.NewSha256Hasher)
	assert.NoError(t, err)

	fsm2.Restore(r)

	// Error: Command already applied
	e := fsm2.Apply(newRaftLog(0, 0)).(*fsmAddResponse)
	assert.Error(t, e.error)
}

func newRaftLog(index, term uint64) *raft.Log {
	event := []byte("All's right with the world")
	data, _ := commands.Encode(commands.AddEventCommandType, &commands.AddEventCommand{event})
	return &raft.Log{index, term, raft.LogCommand, data}
}
