package balloon2

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bbva/qed/balloon2/common"
	"github.com/bbva/qed/db/bplus"
	"github.com/bbva/qed/testutils/rand"
)

func TestAdd(t *testing.T) {

	store := bplus.NewBPlusTreeStore()
	defer store.Close()

	balloon := NewBalloon(0, store, common.NewSha256Hasher)

	for i := uint64(0); i < 9; i++ {
		commitment, mutations, err := balloon.Add(rand.Bytes(128))
		store.Mutate(mutations...)

		require.NoError(t, err)
		require.Truef(t, len(mutations) > 0, "There should be some mutations in test %d", i)
		require.Equalf(t, i, commitment.Version, "Wrong version in test %d", i)
		require.NotNil(t, commitment.HyperDigest, "The HyperDigest shouldn't be nil in test %d", i)
		require.NotNil(t, commitment.HistoryDigest, "The HistoryDigest shouldn't be nil in test %d", i)
	}

}

func TestQueryMembership(t *testing.T) {

	store := bplus.NewBPlusTreeStore()
	defer store.Close()

	balloon := NewBalloon(0, store, common.NewFakeXorHasher)

	key := []byte{0x5a}
	version := uint64(0)
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
	proof, err := balloon.QueryMembership(key, version)

	require.NoError(t, err)
	require.True(t, proof.Exists, "The event should exist")
	require.Equalf(t, version, proof.QueryVersion, "The query version does not match: expected %d, actual %d", version, proof.QueryVersion)
	require.Equalf(t, version, proof.ActualVersion, "The actual version does not match: expected %d, actual %d", version, proof.ActualVersion)
	require.Equalf(t, proof.HyperProof.AuditPath(), expectedHyperAuditPath, "Wrong hyper audit path: expected %v, actual %v", expectedHyperAuditPath, proof.HyperProof.AuditPath())
	require.Equalf(t, proof.HistoryProof.AuditPath(), expectedHistoryAuditPath, "Wrong history audit path: expected %v, actual %v", expectedHistoryAuditPath, proof.HistoryProof.AuditPath())

}

func TestMembershipProofVerify(t *testing.T) {

	testCases := []struct {
		exists         bool
		hyperOK        bool
		historyOK      bool
		currentVersion uint64
		queryVersion   uint64
		actualVersion  uint64
		expectedResult bool
	}{
		// Event exists, queryVersion <= actualVersion, and both trees verify it
		{true, true, true, uint64(0), uint64(0), uint64(0), true},
		// Event exists, queryVersion <= actualVersion, but HyperTree does not verify it
		{true, false, true, uint64(0), uint64(0), uint64(0), false},
		// Event exists, queryVersion <= actualVersion, but HistoryTree does not verify it
		{true, true, false, uint64(0), uint64(0), uint64(0), false},

		// Event exists, queryVersion > actualVersion, and both trees verify it
		{true, true, true, uint64(1), uint64(1), uint64(0), true},
		// Event exists, queryVersion > actualVersion, but HyperTree does not verify it
		{true, false, true, uint64(1), uint64(1), uint64(0), false},

		// Event does not exist, HyperTree verifies it
		{false, true, false, uint64(0), uint64(0), uint64(0), true},
		// Event does not exist, HyperTree does not verify it
		{false, false, false, uint64(0), uint64(0), uint64(0), false},
	}

	for i, c := range testCases {
		event := []byte("Yadda yadda")
		commitment := &Commitment{
			common.Digest("Some hyperDigest"),
			common.Digest("Some historyDigest"),
			c.actualVersion,
		}
		proof := NewMembershipProof(
			c.exists,
			common.NewFakeVerifiable(c.hyperOK),
			common.NewFakeVerifiable(c.historyOK),
			c.currentVersion,
			c.queryVersion,
			c.actualVersion,
			event,
			common.NewSha256Hasher(),
		)

		result := proof.Verify(event, commitment)

		require.Equalf(t, c.expectedResult, result, "Unexpected result '%v' in test case '%d'", result, i)
	}
}

func BenchmarkAddBadger(b *testing.B) {
	store, closeF := common.OpenBadgerStore("/var/tmp/ballon_bench.db")
	defer closeF()

	balloon := NewBalloon(0, store, common.NewSha256Hasher)

	b.ResetTimer()
	b.N = 10000
	for i := 0; i < b.N; i++ {
		event := rand.Bytes(128)
		_, mutations, _ := balloon.Add(event)
		store.Mutate(mutations...)
	}

}
