package publish

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/bbva/qed/balloon"
)

type publisher interface {
	Publish()
	Sign()
}

type httpPublisher struct {
	client   *http.Client
	members  []string
	endpoint string
}

type gossipPublisher struct {
}

func newHttpPublisher() *httpPublisher {
	return &httpPublisher{
		client: &http.Client{},
		members: []string{
			"http://127.0.0.1:4001",
			// "http://127.0.0.1:4002",
		},
		endpoint: "/gossip/add",
	}
}

func newGossipPublisher() *gossipPublisher {
	return nil
}

func SpawnPublishers(ch <-chan *balloon.Commitment) {
	numGoRoutines := 100

	pub := newHttpPublisher()

	for i := 0; i < numGoRoutines; i++ {
		go func() {
			for {
				// <- ch
				commitment := <-ch
				pub.Publish(commitment)
			}
		}()
	}
}

func (p httpPublisher) Publish(commitment *balloon.Commitment) {
	key := strconv.FormatUint(commitment.Version, 10)
	v := sha256.Sum256(commitment.HistoryDigest)
	value := string(v[:])

	// Random request distribution among publisher members
	req, _ := http.NewRequest("GET", p.members[rand.Int()%len(p.members)]+p.endpoint, nil)
	q := req.URL.Query()
	q.Add("key", key)
	q.Add("val", value)
	req.URL.RawQuery = q.Encode()

	rs, err := p.client.Do(req)
	if err != nil {
		fmt.Println("Errored when sending request to the server")
		return
	}
	io.Copy(ioutil.Discard, rs.Body)
	defer rs.Body.Close()
}

func (p gossipPublisher) Publish(commitment *balloon.Commitment) {

}
