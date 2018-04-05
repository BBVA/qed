package ballon

import (
	"fmt"
	"testing"
	"verifiabledata/balloon/hashing"
	"verifiabledata/balloon/storage"
)

func TestAdd(t *testing.T) {
	hasher, n := hashing.Sha256Hasher()
	store := storage.NewBadgerStorage("testadd.db")
	balloon := NewBalloon(store, hasher)

	var testCases = []struct {
		index         uint64
		historyDigest string
		hyperDigest   string
		event         string
	}{
		{0, " ", "", "Hello World1"},
		{1, " ", "", "Hello World2"},
		{2, " ", "", "Hello World3"},
		{3, " ", "", "Hello World4"},
		{4, " ", "", "Hello World5"},
	}

	for _, e := range testCases {
		t.Log("Testing event: ", e.event)
		commitment, err := balloon.Add([]byte(e.event))
		if err != nil {
			t.Fatal("Error in Add call: ", err)
		}

		if e.index != commitment.Version {
			t.Fatal("Incorrect index: ", e.index, " != ", commitment.Version)
		}
		hd := fmt.Sprintf("%x", commitment.HistoryDigest)
		hyd := fmt.Sprintf("%x", commitment.HyperDigest)
		if e.historyDigest != hd {
			t.Fatal("Incorrect history commitment: ", e.HistoryDigest, " != ", hd)
		}

		if e.hyperDigest != hyd {
			t.Fatal("Incorrect hyper commitment: ", e.HyperDigest, " != ", hyd)
		}
	}
}
