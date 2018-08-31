package balloon2

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bbva/qed/balloon2/common"
	"github.com/bbva/qed/testutils/rand"
	"github.com/bbva/qed/testutils/storage"
)

func TestAdd(t *testing.T) {

	store, closeF := storage.NewBPlusTreeStore()
	defer closeF()

	balloon := NewBalloon(0, store, common.NewFakeXorHasher)

	var testCases = []struct {
		event         string
		hyperDigest   common.Digest
		historyDigest common.Digest
		version       uint64
	}{
		{"test event 0", common.Digest{0x4a}, common.Digest{0x4a}, 0},
		{"test event 1", common.Digest{0x01}, common.Digest{0x01}, 1},
		{"test event 2", common.Digest{0x4b}, common.Digest{0x4a}, 2},
		{"test event 3", common.Digest{0x00}, common.Digest{0x00}, 3},
		{"test event 4", common.Digest{0x4a}, common.Digest{0x4a}, 4},
		{"test event 5", common.Digest{0x01}, common.Digest{0x00}, 5},
		{"test event 6", common.Digest{0x4b}, common.Digest{0x4d}, 6},
		{"test event 7", common.Digest{0x00}, common.Digest{0x07}, 7},
		{"test event 8", common.Digest{0x4a}, common.Digest{0x41}, 8},
		{"test event 9", common.Digest{0x01}, common.Digest{0x0b}, 9},
	}

	for i, e := range testCases {
		commitment, err := balloon.Add([]byte(e.event))
		require.NoError(t, err)
		require.Equalf(t, commitment.Version, e.version, "Wrong version for test %d: expected %d, actual %d", i, e.version, commitment.Version)
		require.Equalf(t, commitment.HyperDigest, e.hyperDigest, "Wrong index digest for test %d: expected: %x, Actual: %x", i, e.hyperDigest, commitment.HyperDigest)
		require.Equalf(t, commitment.HistoryDigest, e.historyDigest, "Wrong history digest for test %d: expected: %x, Actual: %x", i, e.historyDigest, commitment.HistoryDigest)
	}

}

func TestGenMembershipProof(t *testing.T) {

	store, closeF := storage.NewBPlusTreeStore()
	defer closeF()

	balloon := NewBalloon(0, store, common.NewFakeXorHasher)

	key := []byte{0x5a}
	version := uint16(0)
	expectedHyperAuditPath := common.AuditPath{
		"50|3": common.Digest{0x00},
		"40|4": common.Digest{0x00},
		"5a|0": common.Digest{0x00},
		"80|7": common.Digest{0x00},
		"00|6": common.Digest{0x00},
		"5c|2": common.Digest{0x00},
		"58|1": common.Digest{0x00},
		"5b|0": common.Digest{0x00},
	}
	expectedHistoryAuditPath := map[string][]byte{
		"0|0": common.Digest{0x5a},
	}

	balloon.Add(key)
	proof, err := balloon.GenMembershipProof(key, version)

	require.NoError(t, err)
	require.True(t, proof.Exists, "The event should exist")
	require.Equalf(t, version, proof.QueryVersion, "The query version does not match: expected %d, actual %d", version, proof.QueryVersion)
	require.Equalf(t, version, proof.ActualVersion, "The actual version does not match: expected %d, actual %d", version, proof.ActualVersion)
	require.Equalf(t, proof.HyperAuditPath, expectedHyperAuditPath, "Wrong hyper audit path: expected %v, actual %v", expectedHyperAuditPath, proof.HyperAuditPath)
	require.Equalf(t, proof.HistoryAuditPath, expectedHistoryAuditPath, "Wrong history audit path: expected %v, actual %v", expectedHistoryAuditPath, proof.HistoryAuditPath)

}

func BenchmarkAddBadger(b *testing.B) {
	store, closeF := storage.NewBadgerStore("/var/tmp/ballon_bench.db")
	defer closeF()

	balloon := NewBalloon(0, store, common.NewFakeSha256Hasher)

	b.ResetTimer()
	b.N = 10000
	for i := 0; i < b.N; i++ {
		event := rand.Bytes(128)
		balloon.Add(event)
	}

}
