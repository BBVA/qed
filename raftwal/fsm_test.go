package raftwal

import (
	"io"
	"testing"

	"github.com/hashicorp/raft"
	"github.com/stretchr/testify/require"

	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/log"
	"github.com/bbva/qed/raftwal/commands"
	"github.com/bbva/qed/testutils/rand"
	storage_utils "github.com/bbva/qed/testutils/storage"
)

func TestApplyAdd(t *testing.T) {

	log.SetLogger("TestApplyAdd", log.SILENT)

	store, closeF := storage_utils.OpenRocksDBStore(t, "/var/tmp/balloon.test.db")
	defer closeF()

	fsm, err := NewBalloonFSM(store, hashing.NewSha256Hasher)
	require.NoError(t, err)

	tests := []struct {
		log           *raft.Log
		expectedError bool
	}{
		{newRaftAddLog(1, 1), false}, // happy path
		{newRaftAddLog(1, 1), true},  // Error: Command already applied
		{newRaftAddLog(2, 1), false}, // happy path
		{newRaftAddLog(1, 1), true},  // Error: Command out of order
	}

	for _, test := range tests {
		r := fsm.Apply(test.log).(*fsmAddResponse)
		require.Equal(t, test.expectedError, r.error != nil)
	}
}

func TestApplyAddBulk(t *testing.T) {

	log.SetLogger("TestApplyAddBulk", log.SILENT)

	store, closeF := storage_utils.OpenRocksDBStore(t, "/var/tmp/balloon.test.db")
	defer closeF()

	fsm, err := NewBalloonFSM(store, hashing.NewSha256Hasher)
	require.NoError(t, err)

	tests := []struct {
		log           *raft.Log
		expectedError bool
	}{
		{newRaftAddBulkLog(1, 1), false}, // happy path
		{newRaftAddBulkLog(1, 1), true},  // Error: Command already applied
		{newRaftAddBulkLog(2, 1), false}, // happy path
		{newRaftAddBulkLog(1, 1), true},  // Error: Command out of order
	}

	for _, test := range tests {
		r := fsm.Apply(test.log).(*fsmAddBulkResponse)
		require.Equal(t, test.expectedError, r.error != nil)
	}
}

func TestSnapshot(t *testing.T) {

	log.SetLogger("TestSnapshot", log.SILENT)

	store, closeF := storage_utils.OpenRocksDBStore(t, "/var/tmp/balloon.test.db")
	defer closeF()

	fsm, err := NewBalloonFSM(store, hashing.NewSha256Hasher)
	require.NoError(t, err)

	fsm.Apply(newRaftAddLog(0, 0))

	// happy path
	_, err = fsm.Snapshot()
	require.NoError(t, err)
}

type fakeRC struct{}

func (f *fakeRC) Read(p []byte) (n int, err error) {
	return 0, io.EOF
}

func (f *fakeRC) Close() error {
	return nil
}

func TestRestore(t *testing.T) {

	log.SetLogger("TestRestore", log.SILENT)

	store, closeF := storage_utils.OpenRocksDBStore(t, "/var/tmp/balloon.test.db")
	defer closeF()

	fsm, err := NewBalloonFSM(store, hashing.NewSha256Hasher)
	require.NoError(t, err)

	require.NoError(t, fsm.Restore(&fakeRC{}))
}

func TestAddAndRestoreSnapshot(t *testing.T) {

	log.SetLogger("TestAddAndRestoreSnapshot", log.SILENT)

	store, closeF := storage_utils.OpenRocksDBStore(t, "/var/tmp/balloon.test.db")
	defer closeF()

	fsm, err := NewBalloonFSM(store, hashing.NewSha256Hasher)
	require.NoError(t, err)

	fsm.Apply(newRaftAddLog(0, 0))

	fsmsnap, err := fsm.Snapshot()
	require.NoError(t, err)

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

	err = fsmsnap.Persist(sink)
	require.NoError(t, err)
	// fsm.Close()

	// Read the latest snapshot
	snaps, _ := snap.List()
	_, r, _ := snap.Open(snaps[0].ID)

	store2, close2F := storage_utils.OpenRocksDBStore(t, "/var/tmp/balloon.test.2.db")
	defer close2F()

	// New FSMStore
	fsm2, err := NewBalloonFSM(store2, hashing.NewSha256Hasher)
	require.NoError(t, err)

	err = fsm2.Restore(r)
	require.NoError(t, err)

	// Error: Command already applied
	e := fsm2.Apply(newRaftAddLog(0, 0)).(*fsmAddResponse)
	require.Error(t, e.error)
}

func BenchmarkApplyAdd(b *testing.B) {

	log.SetLogger("BenchmarkApplyAdd", log.SILENT)

	store, closeF := storage_utils.OpenRocksDBStore(b, "/var/tmp/fsm_bench.db")
	defer closeF()

	fsm, err := NewBalloonFSM(store, hashing.NewSha256Hasher)
	defer fsm.Close()
	require.NoError(b, err)

	b.ResetTimer()
	b.N = 2000000
	for i := 0; i < b.N; i++ {
		log := newRandomRaftLog(uint64(i), uint64(1))
		resp := fsm.Apply(log)
		require.NoError(b, resp.(*fsmAddResponse).error)
	}

}

func newRaftAddLog(index, term uint64) *raft.Log {
	event := []byte("All's right with the world")
	data, _ := commands.Encode(commands.AddEventCommandType, &commands.AddEventCommand{Event: event})
	return &raft.Log{Index: index, Term: term, Type: raft.LogCommand, Data: data}
}

func newRaftAddBulkLog(index, term uint64) *raft.Log {
	events := [][]byte{
		[]byte("The year’s at the spring,"),
		[]byte("And day's at the morn;"),
		[]byte("Morning's at seven;"),
		[]byte("The hill-side’s dew-pearled;"),
		[]byte("The lark's on the wing;"),
		[]byte("The snail's on the thorn;"),
		[]byte("God's in his heaven—"),
		[]byte("All's right with the world!"),
	}
	data, _ := commands.Encode(commands.AddBulkEventCommandType, &commands.AddBulkEventCommand{Events: events})
	return &raft.Log{Index: index, Term: term, Type: raft.LogCommand, Data: data}
}

func newRandomRaftLog(index, term uint64) *raft.Log {
	event := rand.Bytes(128)
	data, _ := commands.Encode(commands.AddEventCommandType, &commands.AddEventCommand{Event: event})
	return &raft.Log{Index: index, Term: term, Type: raft.LogCommand, Data: data}
}
