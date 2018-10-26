package publish

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"

	"github.com/bbva/qed/balloon"
	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/sign"
)

type publisher interface {
	Sign(c *balloon.Commitment) (*SignedSnapshot, error)
	Publish(sc *SignedSnapshot)
}

// Snapshot is the public struct that apihttp.Add Handler call returns.
type Snapshot struct {
	HistoryDigest hashing.Digest
	HyperDigest   hashing.Digest
	Version       uint64
}

type SignedSnapshot struct {
	Snapshot  *Snapshot
	Signature []byte
}

type httpPublisher struct {
	client   *http.Client
	members  []string
	endpoint string
	signer   sign.Signer
}

type gossipPublisher struct {
}

func newHttpPublisher(signer sign.Signer) *httpPublisher {
	return &httpPublisher{
		client: &http.Client{},
		// TODO: temporal hardcoding until publisher final decission.
		members: []string{
			"http://127.0.0.1:4001",
			// "http://127.0.0.1:4002", // Add members at will.
		},
		endpoint: "/gossip/add",
		signer:   signer,
	}
}

func SpawnPublishers(signer sign.Signer, numPublishers int, ch <-chan *balloon.Commitment) {
	pub := newHttpPublisher(signer)

	for i := 0; i < numPublishers; i++ {
		go func() {
			for {
				commitment, open := <-ch
				if open {
					sc, _ := pub.Sign(commitment)
					pub.Publish(sc)
				} else {
					return
				}
			}
		}()
	}
}

func (p httpPublisher) Sign(commitment *balloon.Commitment) (*SignedSnapshot, error) {
	snapshot := &Snapshot{
		commitment.HistoryDigest,
		commitment.HyperDigest,
		commitment.Version,
	}

	signature, err := p.signer.Sign([]byte(fmt.Sprintf("%v", snapshot)))
	if err != nil {
		fmt.Println("Publisher: error signing commitment")
		return nil, err
	}
	return &SignedSnapshot{snapshot, signature}, nil
}

func (p httpPublisher) Publish(sc *SignedSnapshot) {
	message, err := json.Marshal(&sc)
	if err != nil {
		fmt.Printf("\nPublisher: Error marshalling: %s", err.Error())
		return
	}

	// Random request distribution among publisher members.
	publisher := p.members[rand.Int()%len(p.members)] + p.endpoint
	rs, err := p.client.Post(
		publisher,
		"application/json",
		bytes.NewBuffer(message))

	if err != nil {
		fmt.Printf("\nPublisher: Error when sending request to publishers: %s", err.Error())
		return
	}
	defer rs.Body.Close()
	io.Copy(ioutil.Discard, rs.Body)
}

// GOSSIP
func newGossipPublisher() *gossipPublisher {
	return nil
}

func (p gossipPublisher) Publish(sc *SignedSnapshot) {

}

func (p gossipPublisher) Sign(commitment *balloon.Commitment) (*SignedSnapshot, error) {
	return nil, nil
}
