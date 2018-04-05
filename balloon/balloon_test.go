package balloon

import (
	"crypto/rand"
//	"fmt"
	"testing"
	"verifiabledata/balloon/hashing"
	"verifiabledata/balloon/storage"
)
/*
func TestAdd(t *testing.T) {
	hasher, _ := hashing.Sha256Hasher()
	store := storage.NewBadgerStorage("/tmp/testadd.db")
	balloon := NewBalloon(store, hasher)

	var testCases = []struct {
		index         uint
		historyDigest string
		indexDigest   string
		event         string
	}{
		{1, "fa712069a4f6ece78619c7ab233b42b94e40a7bab384311ee1e16b101a8478ec", "f6ffb919a4222494cf056404fb5615631594d4f25eb28489b06193c752b467a8", "Hello World1"},
		{2, "db1d613425a77f0f129c55af46407f74a804ac1fb9ea6b27694dbc3628bc299b", "05eda02b799506b3efcae013134378a44c0bceee3086c16ce487efd97bb55e39", "Hello World2"},
		{3, "952f3d4d5a242c29192b132a9f10d0dcbd20fb7b8a8b0a92cc5e777c5eee889f", "4e756426816a4fa17c777bb121d180b1cef8ef459541a396e61f44066a04c033", "Hello World3"},
		{4, "2ed6c9bf02b523a9d9e29dbd4ad52242f31b1666503d44e33f3723c20db7bc9b", "3033269ebb98f44bfbac874771f9154f58ea1eaaeb09286d45510d66c6df9ca1", "Hello World4"},
		{5, "887c24c88a1f9cfb006a7ee23d891b3bbaed6842026dfe00df647b8db3e18f7b", "b89ddff96f7dc22cbf233d83c1ca7bec826889530a8e47138dbc1493f05a0dbd", "Hello World5"},
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
		hyd := fmt.Sprintf("%x", commitment.IndexDigest)
		if e.historyDigest != hd {
			t.Fatal("Incorrect history commitment: ", e.historyDigest, " != ", hd)
		}

		if e.indexDigest != hyd {
			t.Fatal("Incorrect hyper commitment: ", e.indexDigest, " != ", hyd)
		}
	}
}
*/
func randomBytes(n int) []byte {
	bytes := make([]byte, n)
	_, err := rand.Read(bytes)
	if err != nil {
		panic(err)
	}

	return bytes
}

func BenchmarkAdd(b *testing.B) {
	hasher, _ := hashing.Sha256Hasher()
	bs := storage.NewBadgerStorage("/tmp/bench_ball.db")
	hs := storage.NewBadgerStorage("/tmp/bench_frozen.db")
	hys:= storage.NewBadgerStorage("/tmp/bench_hyper.db")
	balloon := NewBalloon(bs, hs, hys, hasher)

	for i := 0; i < b.N; i++ {
		event := randomBytes(128)
		balloon.Add(event)
	}

}
