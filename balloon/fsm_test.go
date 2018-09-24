package balloon

import (
	"encoding/json"
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
}

func TestSnapshot(t *testing.T) {
	assert.True(t, true)
}

func TestRestore(t *testing.T) {
	assert.True(t, true)
}

func TestClose(t *testing.T) {
	assert.True(t, true)
}

func TestAddApply(t *testing.T) {
	assert.True(t, true)
}
