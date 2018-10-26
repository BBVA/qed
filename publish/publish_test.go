package publish

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bbva/qed/balloon"
	"github.com/bbva/qed/hashing"

	"github.com/bbva/qed/sign"
)

type ClientMock struct {
}

func (c *ClientMock) Do(req *http.Request) (*http.Response, error) {
	return &http.Response{}, nil
}

func TestSign(t *testing.T) {
	signer := sign.NewEd25519Signer()
	p := newGossipPublisher(signer) // The type of publisher does not matter for signing.

	commitment := &balloon.Commitment{
		HistoryDigest: hashing.Digest([]byte{0x0}),
		HyperDigest:   hashing.Digest([]byte{0x0}),
		Version:       0,
	}

	signedSnapshot, err := p.Sign(commitment)
	require.Nil(t, err, "Error signing")

	if !bytes.Equal(signedSnapshot.Snapshot.HyperDigest, []byte{0x0}) {
		t.Errorf("HyperDigest is not consistent: %s", signedSnapshot.Snapshot.HyperDigest)
	}

	if !bytes.Equal(signedSnapshot.Snapshot.HistoryDigest, []byte{0x0}) {
		t.Errorf("HistoryDigest is not consistent %s", signedSnapshot.Snapshot.HistoryDigest)
	}

	if signedSnapshot.Snapshot.Version != 0 {
		t.Errorf("Version is not consistent")
	}

	signature, _ := signer.Sign([]byte(fmt.Sprintf("%v", signedSnapshot.Snapshot)))

	if !bytes.Equal(signedSnapshot.Signature, signature) {
		t.Errorf("Signature is not consistent")
	}
}

func TestPublish(t *testing.T) {

	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer fakeServer.Close()

	signedSnapshot := SignedSnapshot{
		Snapshot: &Snapshot{
			HistoryDigest: []byte{0x0},
			HyperDigest:   []byte{0x0},
			Version:       0,
		},
		Signature: []byte{0x0},
	}

	signer := sign.NewEd25519Signer()
	members := []string{fakeServer.URL}
	endpoint := "/test-endpoint"

	p := newHttpPublisher(http.Client{}, signer, members, endpoint)
	err := p.Publish(&signedSnapshot)

	assert.Nil(t, err)
}
