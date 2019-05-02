/*
   Copyright 2018-2019 Banco Bilbao Vizcaya Argentaria, S.A.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package e2e

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"runtime/debug"
	"testing"

	"github.com/bbva/qed/client"
	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/protocol"
	"github.com/bbva/qed/server"
	"github.com/bbva/qed/testutils/keys"
	"github.com/bbva/qed/testutils/rand"
	"github.com/bbva/qed/testutils/scenario"
)

func config_server(id int, pathDB, signPath, tlsPath, addr string, tls bool) *server.Config {
	hostname, _ := os.Hostname()
	conf := server.DefaultConfig()
	conf.APIKey = APIKey
	conf.NodeID = fmt.Sprintf("%s-%d", hostname, id)
	conf.HTTPAddr = fmt.Sprintf("127.0.0.1:880%d", id)
	conf.MgmtAddr = fmt.Sprintf("127.0.0.1:870%d", id)
	conf.MetricsAddr = fmt.Sprintf("127.0.0.1:860%d", id)
	conf.RaftAddr = fmt.Sprintf("127.0.0.1:850%d", id)
	conf.GossipAddr = fmt.Sprintf("127.0.0.1:840%d", id)
	if id > 0 {
		conf.RaftJoinAddr = []string{"127.0.0.1:8700"}
		conf.GossipJoinAddr = []string{"127.0.0.1:8400"}
	}
	conf.DBPath = pathDB + "data"
	conf.RaftPath = pathDB + "raft"
	conf.PrivateKeyPath = signPath
	if tls {
		conf.SSLCertificate = tlsPath + "/cert.pem"
		conf.SSLCertificateKey = tlsPath + "/key.pem"
	}
	conf.EnableTLS = tls

	return conf
}

// This function creates a new server each time a scenario is called
// Each server instance is completely new and blank.
func prepare_new_server(id int, addr string, tls bool) (func() error, func() error) {
	var srv *server.Server
	var path string
	var err error

	before := func() error {
		var tlsPath string

		path, err = ioutil.TempDir("", "e2e-qed-")
		if err != nil {
			return err
		}
		signKeyPath, err := keys.GenerateSignKey(path)
		if err != nil {
			return err
		}
		if tls {
			tlsPath, err = keys.GenerateTlsCert(path)
			if err != nil {
				return err
			}
		}
		conf := config_server(id, path, signKeyPath, tlsPath, addr, false)
		srv, err = server.NewServer(conf)
		if err != nil {
			return err
		}
		return srv.Start()
	}

	after := func() error {
		srv.Stop()
		debug.FreeOSMemory()
		os.RemoveAll(path)
		return nil
	}
	return before, after
}

func new_qed_client(id int) (*client.HTTPClient, error) {
	// QED client
	transport := http.DefaultTransport.(*http.Transport)
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: false}
	httpClient := http.DefaultClient
	httpClient.Transport = transport
	client, err := client.NewHTTPClient(
		client.SetHttpClient(httpClient),
		client.SetURLs(fmt.Sprintf("http://127.0.0.1:880%d", id)),
		client.SetAPIKey(APIKey),
		client.SetTopologyDiscovery(false),
		client.SetHealthChecks(false),
		client.SetMaxRetries(3),
	)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func TestAddVerify(t *testing.T) {
	before, after := prepare_new_server(0, "", false)
	let, report := scenario.New()
	defer func() { t.Logf(report()) }()
	// log.SetLogger("", log.DEBUG)
	event := rand.RandomString(10)
	before()
	let(t, "Add one event and get its membership proof", func(t *testing.T) {
		var snapshot *protocol.Snapshot
		var err error

		client, err := new_qed_client(0)
		scenario.NoError(t, err, "Error creating qed client")

		let(t, "Add event", func(t *testing.T) {
			snapshot, err = client.Add(event)
			scenario.NoError(t, err, "Error adding event")

			scenario.Equal(t, snapshot.EventDigest, hashing.NewSha256Hasher().Do([]byte(event)), "The snapshot's event doesn't match")
			scenario.False(t, snapshot.Version < 0, "The snapshot's version must be greater or equal to 0")
			scenario.False(t, len(snapshot.HyperDigest) == 0, "The snapshot's hyperDigest cannot be empty")
			scenario.False(t, len(snapshot.HistoryDigest) == 0, "The snapshot's hyperDigest cannot be empt")
		})

		let(t, "Get membership proof for first inserted event", func(t *testing.T) {
			result, err := client.Membership([]byte(event), snapshot.Version)
			scenario.NoError(t, err, "Error getting membership proof")

			scenario.True(t, result.Exists, "The queried key should be a member")
			scenario.Equal(t, result.QueryVersion, snapshot.Version, "The query version doest't match the queried one")
			scenario.Equal(t, result.ActualVersion, snapshot.Version, "The actual version should match the queried one")
			scenario.Equal(t, result.CurrentVersion, snapshot.Version, "The current version should match the queried one")
			scenario.Equal(t, []byte(event), result.Key, "The returned event doesn't math the original one")
			scenario.False(t, len(result.KeyDigest) == 0, "The key digest cannot be empty")
			scenario.False(t, len(result.Hyper) == 0, "The hyper proof cannot be empty")
			scenario.False(t, result.ActualVersion > 0 && len(result.History) == 0, "The history proof cannot be empty when version is greater than 0")

		})
	})
	after()
	before()
	let(t, "Add two events, verify the first one", func(t *testing.T) {
		var resultFirst, resultLast *protocol.MembershipResult
		var err error
		var first, last *protocol.Snapshot

		client, err := new_qed_client(0)
		scenario.NoError(t, err, "Error creating a new qed client")

		first, err = client.Add("Test event 1")
		scenario.NoError(t, err, "Error adding event 1")
		last, err = client.Add("Test event 2")
		scenario.NoError(t, err, "Error adding event 2")

		let(t, "Get membership proof for inserted events", func(t *testing.T) {
			resultFirst, err = client.MembershipDigest(first.EventDigest, first.Version)
			scenario.NoError(t, err, "Error getting membership digest")
			resultLast, err = client.MembershipDigest(last.EventDigest, last.Version)
			scenario.NoError(t, err, "Error getting membership digest")
		})

		let(t, "Verify events", func(t *testing.T) {
			first.HyperDigest = last.HyperDigest
			scenario.True(t, client.DigestVerify(resultFirst, first, hashing.NewSha256Hasher), "The first proof should be valid")
			scenario.True(t, client.DigestVerify(resultLast, last, hashing.NewSha256Hasher), "The last proof should be valid")
		})

	})

	after()
	before()
	let(t, "Add 10 events, verify event with index i", func(t *testing.T) {
		var p1, p2 *protocol.MembershipResult
		var err error
		const size int = 10

		var s [size]*protocol.Snapshot

		client, err := new_qed_client(0)
		scenario.NoError(t, err, "Error getting new qed client")

		for i := 0; i < size; i++ {
			s[i], _ = client.Add(fmt.Sprintf("Test Event %d", i))
		}

		i := 3
		j := 6
		k := 9

		let(t, "Get proofs p1, p2 for event with index i in versions j and k", func(t *testing.T) {
			p1, err = client.MembershipDigest(s[i].EventDigest, s[j].Version)
			scenario.NoError(t, err, "Error getting membership digest")
			p2, err = client.MembershipDigest(s[i].EventDigest, s[k].Version)
			scenario.NoError(t, err, "Error getting membership digest")
		})

		let(t, "Verify both proofs against index i event", func(t *testing.T) {
			snap := &protocol.Snapshot{
				HistoryDigest: s[j].HistoryDigest,
				HyperDigest:   s[9].HyperDigest,
				Version:       s[j].Version,
				EventDigest:   s[i].EventDigest,
			}
			scenario.True(t, client.DigestVerify(p1, snap, hashing.NewSha256Hasher), "p1 should be valid")

			snap = &protocol.Snapshot{
				HistoryDigest: s[k].HistoryDigest,
				HyperDigest:   s[9].HyperDigest,
				Version:       s[k].Version,
				EventDigest:   s[i].EventDigest,
			}
			scenario.True(t, client.DigestVerify(p2, snap, hashing.NewSha256Hasher), "p2 should be valid")

		})

	})
	after()

}
