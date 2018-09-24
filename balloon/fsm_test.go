package balloon

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"testing"

	"github.com/hashicorp/raft"
	assert "github.com/stretchr/testify/require"

	"github.com/bbva/qed/hashing"
	storage_utils "github.com/bbva/qed/testutils/storage"
)

func raftLog(c commandType, index, term uint64) *raft.Log {
	var sub json.RawMessage
	sub, _ = json.Marshal(&insertSubCommand{[]byte("All's right with the world")})
	data, _ := json.Marshal(&command{insert, sub})

	return &raft.Log{index, term, raft.LogCommand, data}
}

func TestApply(t *testing.T) {
	store, closeF := storage_utils.OpenBadgerStore(t, "/var/tmp/balloon.test.db")
	defer closeF()

	fsm, err := NewBalloonFSM(store, hashing.NewSha256Hasher)
	assert.NoError(t, err)

	// happy path
	r := fsm.Apply(raftLog(insert, 1, 1)).(*fsmAddResponse)
	assert.Nil(t, r.error)

	// Error: Command already applied
	r = fsm.Apply(raftLog(insert, 1, 1)).(*fsmAddResponse)
	assert.Error(t, r.error)

	// happy path
	r = fsm.Apply(raftLog(insert, 2, 1)).(*fsmAddResponse)
	assert.Nil(t, r.error)

	// Error: Command out of order
	r = fsm.Apply(raftLog(insert, 1, 1)).(*fsmAddResponse)
	assert.Error(t, r.error)

	// Error: Unknown command
	j := fsm.Apply(raftLog(insert-42, 3)).(*fsmGenericResponse)
	assert.Error(t, j.error)

}

func TestSnapshot(t *testing.T) {
	store, closeF := storage_utils.OpenBadgerStore(t, "/var/tmp/balloon.test.db")
	defer closeF()

	fsm, err := NewBalloonFSM(store, hashing.NewSha256Hasher)
	assert.NoError(t, err)

	// _ = fsm.Apply(raftLog(insert, 0)).(*fsmAddResponse)
	fsm.Apply(raftLog(insert, 0))

	// happy path
	_, err = fsm.Snapshot()
	assert.NoError(t, err)
}

type fakeRC struct {
	p []byte
}

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

	// _ = fsm.Apply(raftLog(insert, 0)).(*fsmAddResponse)
	fsm.Apply(raftLog(insert, 0))

	// happy path
	snap, err := fsm.Snapshot()

	assert.NoError(t, err)
	snap.Persist(sink)
	sink.List()

	buf := bytes.NewBuffer(b)

	assert.NoError(t, fsm.Restore(ioutil.NopCloser(buf)))
}
