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

package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/bbva/qed/balloon"
	"github.com/bbva/qed/crypto/hashing"
	"github.com/bbva/qed/log"
	"github.com/bbva/qed/testutils/spec"
	"github.com/pkg/errors"

	"github.com/bbva/qed/protocol"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	log.SetDefault(log.New(&log.LoggerOptions{
		IncludeLocation: true,
		Level:           log.Off,
	}))
	os.Exit(m.Run())
}

func setupServer(input []byte) (string, func()) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)

	mux.HandleFunc("/info/shards", infoHandler(server.URL))
	mux.HandleFunc("/events", defaultHandler(input))
	mux.HandleFunc("/events/bulk", defaultHandler(input))
	mux.HandleFunc("/proofs/membership", defaultHandler(input))
	mux.HandleFunc("/proofs/incremental", defaultHandler(input))
	mux.HandleFunc("/proofs/digest-membership", defaultHandler(input))
	mux.HandleFunc("/healthcheck", defaultHandler(nil))

	return server.URL, func() {
		server.Close()
	}
}

func setupClient(t *testing.T, urls []string) *HTTPClient {
	httpClient := http.DefaultClient
	client, err := NewHTTPClient(
		SetHttpClient(httpClient),
		SetAPIKey("my-awesome-api-key"),
		SetURLs(urls[0], urls[1:]...),
		SetRequestRetrier(NewNoRequestRetrier(httpClient)),
		SetReadPreference(Primary),
		SetMaxRetries(0),
		SetTopologyDiscovery(false),
		SetHealthChecks(false),
		SetHasherFunction(hashing.NewSha256Hasher),
	)
	if err != nil {
		t.Fatal(errors.Wrap(err, "Cannot create http client: "))
	}
	return client
}

func TestCallPrimaryWorking(t *testing.T) {

	var numRequests int
	httpClient := NewTestHttpClient(func(req *http.Request) (*http.Response, error) {
		numRequests++
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewBufferString("OK")),
			Header:     make(http.Header),
		}, nil
	})

	client, err := NewHTTPClient(
		SetHttpClient(httpClient),
		SetAPIKey("my-awesome-api-key"),
		SetURLs("http://primary.foo"),
		SetReadPreference(PrimaryPreferred),
		SetMaxRetries(1),
		SetTopologyDiscovery(false),
		SetHealthChecks(false),
	)
	require.NoError(t, err)

	resp, err := client.callPrimary("GET", "/test", nil)
	require.NoError(t, err, "The requests should not fail")
	require.True(t, len(resp) > 0, "The response should not be empty")
	require.Equal(t, 1, numRequests, "The number of requests should match")
}

// Having 1 single node returning "Internal Server error" make the request fails.
func TestCallPrimaryFails(t *testing.T) {

	var numRequests int
	httpClient := NewTestHttpClient(func(req *http.Request) (*http.Response, error) {
		numRequests++
		return &http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       ioutil.NopCloser(bytes.NewBufferString("Internal server error")),
			Header:     make(http.Header),
		}, nil
	})

	client, err := NewHTTPClient(
		SetHttpClient(httpClient),
		SetAPIKey("my-awesome-api-key"),
		SetURLs("http://primary.foo", "http://secondary1.foo"),
		SetReadPreference(PrimaryPreferred),
		SetMaxRetries(1),
		SetTopologyDiscovery(false),
		SetHealthChecks(false),
	)
	require.NoError(t, err)

	resp, err := client.callPrimary("GET", "/test", nil)
	require.Error(t, err, "The requests should fail")
	require.True(t, len(resp) == 0, "The response should be empty")
	require.Equal(t, 2, numRequests, "The number of requests should match")
}

// Having 1 single and unreachable node make the request fails.
func TestCallPrimaryUnreachable(t *testing.T) {

	var numRequests int
	httpClient := NewTestHttpClient(func(req *http.Request) (*http.Response, error) {
		numRequests++
		return nil, errors.New("Unreachable")
	})

	client, err := NewHTTPClient(
		SetHttpClient(httpClient),
		SetAPIKey("my-awesome-api-key"),
		SetURLs("http://primary.foo", "http://secondary1.foo"),
		SetReadPreference(PrimaryPreferred),
		SetMaxRetries(0),
		SetTopologyDiscovery(false),
		SetHealthChecks(false),
	)
	require.NoError(t, err)

	resp, err := client.callPrimary("GET", "/test", nil)
	require.Error(t, err, "The requests should fail")
	require.True(t, len(resp) == 0, "The response should be empty")
	require.Equal(t, 1, numRequests, "The number of requests should match")
}

// Having 2 nodes with the primary being unreachable, and there is no
// "discovery" option enabled, the request fails.
// Healthchecks (enabled here) does not change primary.
func TestCallPrimaryUnreachableWithHealthChecks(t *testing.T) {

	var alreadyRetried bool

	httpClient := NewTestHttpClient(func(req *http.Request) (*http.Response, error) {
		if req.Host == "primary1.foo" {
			if !alreadyRetried {
				if req.URL.Path == "/healthcheck" {
					return buildResponse(http.StatusOK, ""), nil
				}
				alreadyRetried = true
			}
			return nil, errors.New("Unreachable")
		}
		if req.Host == "secondary1.foo" {
			if req.URL.Path == "/healthcheck" {
				return buildResponse(http.StatusOK, ""), nil
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(bytes.NewBufferString("OK")),
				Header:     make(http.Header),
			}, nil
		}
		return nil, errors.New("Unreachable")
	})

	client, err := NewHTTPClient(
		SetHttpClient(httpClient),
		SetAPIKey("my-awesome-api-key"),
		SetURLs("http://primary1.foo", "http://secondary1.foo"),
		SetReadPreference(PrimaryPreferred),
		SetMaxRetries(0),
		SetTopologyDiscovery(false),
		SetHealthChecks(true),
	)
	require.NoError(t, err)

	resp, err := client.callPrimary("GET", "/test", nil)
	require.Error(t, err, "The requests should fail")
	require.True(t, len(resp) == 0, "The response should be empty")
}

// Having 2 nodes with the primary being unreachable, and with
// "discovery" option enabled, there should be a leader election and the
// request should go to the new primary.
func TestCallPrimaryUnreachableWithDiscovery(t *testing.T) {

	var numRequests int
	var alreadyRetried bool

	httpClient := NewTestHttpClient(func(req *http.Request) (*http.Response, error) {
		if req.Host == "primary1.foo" {
			numRequests++
			if !alreadyRetried {
				if req.URL.Path == "/info/shards" {
					info := protocol.Shards{
						NodeId:    "primary1",
						LeaderId:  "primary1",
						URIScheme: "http",
						Shards: map[string]protocol.ShardDetail{
							"primary1": protocol.ShardDetail{
								NodeId:   "primary1",
								HTTPAddr: "primary1.foo",
							},
							"secondary1": protocol.ShardDetail{
								NodeId:   "secondary1",
								HTTPAddr: "secondary1.foo",
							},
						},
					}
					body, _ := json.Marshal(info)
					return buildResponse(http.StatusOK, string(body)), nil
				}
				alreadyRetried = true
			}
			return nil, errors.New("Unreachable")
		}
		if req.Host == "secondary1.foo" {
			numRequests++
			if req.URL.Path == "/info/shards" {
				info := protocol.Shards{
					NodeId:    "secondary1",
					LeaderId:  "secondary1",
					URIScheme: "http",
					Shards: map[string]protocol.ShardDetail{
						"secondary1": protocol.ShardDetail{
							NodeId:   "secondary1",
							HTTPAddr: "secondary1.foo",
						},
					},
				}
				body, _ := json.Marshal(info)
				return buildResponse(http.StatusOK, string(body)), nil
			}

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(bytes.NewBufferString("OK")),
				Header:     make(http.Header),
			}, nil
		}
		numRequests++
		return nil, errors.New("Unreachable")
	})

	client, err := NewHTTPClient(
		SetHttpClient(httpClient),
		SetAPIKey("my-awesome-api-key"),
		SetURLs("http://primary1.foo", "http://secondary1.foo"),
		SetReadPreference(PrimaryPreferred),
		SetMaxRetries(0),
		SetTopologyDiscovery(true),
		SetHealthChecks(false),
	)
	require.NoError(t, err)

	// Mark node as dead after NewHTTPClient to simulate a primary failure.
	client.topology.primary.MarkAsDead()

	resp, err := client.callPrimary("GET", "/test", nil)
	require.NoError(t, err, "The requests should not fail")
	require.True(t, len(resp) > 0, "The response should not be empty")
	require.Equal(t, 3, numRequests, "The number of requests should match")
}

// Having 2 nodes with the primary being a follower, and with
// there should be an http redirect and the client should try
// the leader automatically
func TestCallPrimaryRedirect(t *testing.T) {

	var priReqs, secReqs int

	var info1, info2 protocol.Shards

	info1 = protocol.Shards{
		NodeId:    "primary1",
		LeaderId:  "secondary1",
		URIScheme: "http",
		Shards: map[string]protocol.ShardDetail{
			"primary1": protocol.ShardDetail{
				NodeId:   "primary1",
				HTTPAddr: "primary1.foo",
			},
			"secondary1": protocol.ShardDetail{
				NodeId:   "secondary1",
				HTTPAddr: "secondary1.foo",
			},
		},
	}

	info2 = protocol.Shards{
		NodeId:    "secondary1",
		LeaderId:  "secondary1",
		URIScheme: "http",
		Shards: map[string]protocol.ShardDetail{
			"secondary1": protocol.ShardDetail{
				NodeId:   "secondary1",
				HTTPAddr: "secondary1.foo",
			},
		},
	}

	httpClient := NewTestHttpClient(func(req *http.Request) (*http.Response, error) {
		if req.Host == "primary1.foo" {
			body, _ := json.Marshal(info1)
			if req.URL.Path == "/info/shards" {
				priReqs++
				return buildResponse(http.StatusOK, string(body)), nil
			}

			h := req.Header
			u, _ := url.Parse("http://secondary1.foo")
			h.Set("Location", u.String())

			return &http.Response{
				StatusCode: http.StatusMovedPermanently,
				Body:       ioutil.NopCloser(bytes.NewBuffer(body)),
				Header:     h,
			}, nil
		}
		// on a redirect, all headers are empty, somehow, Host is also empty
		//...will this work with virtual hosts?
		if req.Host == "secondary1.foo" {
			secReqs++
			if req.URL.Path == "/info/shards" {
				body, _ := json.Marshal(info2)
				return buildResponse(http.StatusOK, string(body)), nil
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(bytes.NewBufferString("OK")),
				Header:     make(http.Header),
			}, nil
		}
		return nil, errors.New("Unreachable")
	})

	client, err := NewHTTPClient(
		SetHttpClient(httpClient),
		SetAPIKey("my-awesome-api-key"),
		SetURLs("http://primary1.foo", "http://secondary1.foo"),
		SetReadPreference(PrimaryPreferred),
		SetMaxRetries(0),
		SetTopologyDiscovery(false),
		SetHealthChecks(false),
	)
	require.NoError(t, err)

	// Mark node as dead after NewHTTPClient to simulate a primary failure.
	// client.topology.primary.MarkAsDead()

	resp, err := client.callPrimary("GET", "/test", nil)
	require.NoError(t, err, "The requests should not fail")
	require.True(t, len(resp) > 0, "The response should not be empty")
	require.Equal(t, 0, priReqs, "The number of requests should match to primary node")
	require.Equal(t, 1, secReqs, "The number of requests should match to secondary node")
}

// Having 2 nodes with the primary being unreachable, and with
// "discovery" and "healthcheck" options enabled, there should be a leader election and the
// request should go to the new primary.
// Actually HealthCheck does nothing, since the nodes never get recovered in this test.
func TestCallPrimaryUnreachableWithDiscoveryAndHealtChecks(t *testing.T) {

	var alreadyRetried bool

	httpClient := NewTestHttpClient(func(req *http.Request) (*http.Response, error) {
		if req.Host == "primary1.foo" {
			if !alreadyRetried {
				if req.URL.Path == "/healthcheck" {
					alreadyRetried = true // Only 1st healt-check is allowed (httpClient creation)
					return buildResponse(http.StatusOK, ""), nil
				}
				if req.URL.Path == "/info/shards" {
					info := protocol.Shards{
						NodeId:    "primary1",
						LeaderId:  "primary1",
						URIScheme: "http",
						Shards: map[string]protocol.ShardDetail{
							"primary1": protocol.ShardDetail{
								NodeId:   "primary1",
								HTTPAddr: "primary1.foo",
							},
							"secondary1": protocol.ShardDetail{
								NodeId:   "secondary1",
								HTTPAddr: "secondary1.foo",
							},
						},
					}
					body, _ := json.Marshal(info)
					return buildResponse(http.StatusOK, string(body)), nil
				}
				// alreadyRetried = true
			}
			return nil, errors.New("Unreachable")
		}
		if req.Host == "secondary1.foo" {
			if req.URL.Path == "/healthcheck" {
				return buildResponse(http.StatusOK, ""), nil
			}
			if req.URL.Path == "/info/shards" {
				info := protocol.Shards{
					NodeId:    "secondary1",
					LeaderId:  "secondary1",
					URIScheme: "http",
					Shards: map[string]protocol.ShardDetail{
						"secondary1": protocol.ShardDetail{
							NodeId:   "secondary1",
							HTTPAddr: "secondary1.foo",
						},
					},
				}
				body, _ := json.Marshal(info)
				return buildResponse(http.StatusOK, string(body)), nil
			}

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(bytes.NewBufferString("OK")),
				Header:     make(http.Header),
			}, nil
		}
		return nil, errors.New("Unreachable")
	})

	client, err := NewHTTPClient(
		SetHttpClient(httpClient),
		SetAPIKey("my-awesome-api-key"),
		SetURLs("http://primary1.foo", "http://secondary1.foo"),
		SetReadPreference(PrimaryPreferred),
		SetMaxRetries(0),
		SetTopologyDiscovery(true),
		SetHealthChecks(true),
	)
	require.NoError(t, err)

	// Mark node as dead after NewHTTPClient to simulate a primary failure.
	client.topology.primary.MarkAsDead()

	resp, err := client.callPrimary("GET", "/test", nil)
	require.NoError(t, err, "The requests should not fail")
	require.True(t, len(resp) > 0, "The response should not be empty")
}

func TestCallAnyPrimaryFails(t *testing.T) {

	var numRequests int
	httpClient := NewTestHttpClient(func(req *http.Request) (*http.Response, error) {
		numRequests++
		if strings.HasPrefix(req.URL.Hostname(), "primary") {
			return &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       ioutil.NopCloser(bytes.NewBufferString("Internal server error")),
				Header:     make(http.Header),
			}, nil
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewBufferString("OK")),
			Header:     make(http.Header),
		}, nil
	})

	client, err := NewHTTPClient(
		SetHttpClient(httpClient),
		SetAPIKey("my-awesome-api-key"),
		SetURLs("http://primary.foo", "http://secondary1.foo"),
		SetReadPreference(PrimaryPreferred),
		SetMaxRetries(1),
		SetTopologyDiscovery(false),
		SetHealthChecks(false),
	)
	require.NoError(t, err)

	resp, err := client.callAny("GET", "/test", nil)
	require.NoError(t, err, "The requests should not fail")
	require.True(t, len(resp) > 0, "The response should not be empty")
	require.Equal(t, 3, numRequests, "The number of requests should match")
}

func TestCallAnyAllFail(t *testing.T) {

	var numRequests int
	httpClient := NewTestHttpClient(func(req *http.Request) (*http.Response, error) {
		numRequests++
		return &http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       ioutil.NopCloser(bytes.NewBufferString("Internal server error")),
			Header:     make(http.Header),
		}, nil
	})

	client, err := NewHTTPClient(
		SetHttpClient(httpClient),
		SetAPIKey("my-awesome-api-key"),
		SetURLs("http://primary.foo", "http://secondary1.foo", "http://secondary2.foo"),
		SetReadPreference(PrimaryPreferred),
		SetMaxRetries(1),
		SetTopologyDiscovery(false),
		SetHealthChecks(false),
	)
	require.NoError(t, err)

	resp, err := client.callAny("GET", "/test", nil)
	require.Error(t, err, "The request should fail")
	require.True(t, len(resp) == 0, "The response should be empty")
	require.Equal(t, 6, numRequests, "The number of requests should match")
}

func TestHealthCheck(t *testing.T) {

	var numRequests int
	httpClient := NewTestHttpClient(func(req *http.Request) (*http.Response, error) {
		if req.Method == "HEAD" {
			numRequests++
			return &http.Response{
				StatusCode: http.StatusNoContent,
				Header:     make(http.Header),
			}, nil
		}
		return &http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       ioutil.NopCloser(bytes.NewBufferString("Internal server error")),
			Header:     make(http.Header),
		}, nil
	})

	client, err := NewHTTPClient(
		SetHttpClient(httpClient),
		SetAPIKey("my-awesome-api-key"),
		SetURLs("http://primary.foo", "http://secondary1.foo", "http://secondary2.foo"),
		SetReadPreference(PrimaryPreferred),
		SetMaxRetries(1),
		SetTopologyDiscovery(false),
		SetHealthChecks(false),
	)
	require.NoError(t, err)

	// force all endpoints to get marked as dead
	_, err = client.callAny("GET", "/events", nil)
	require.Error(t, err)
	require.False(t, client.topology.HasActiveEndpoint())

	// try to revive them
	client.clusterHealthCheck(5 * time.Second)
	time.Sleep(1 * time.Second)
	require.True(t, client.topology.HasActiveEndpoint())
}

func TestPeriodicHealthCheck(t *testing.T) {

	var numChecks int
	httpClient := NewTestHttpClient(func(req *http.Request) (*http.Response, error) {
		fmt.Println(req.URL)
		if req.Method == "HEAD" {
			numChecks++
			if numChecks > 3 {
				return &http.Response{
					StatusCode: http.StatusNoContent,
					Header:     make(http.Header),
				}, nil
			}
			return nil, errors.New("Unreachable")
		}
		return &http.Response{
			StatusCode: http.StatusNoContent,
			Header:     make(http.Header),
		}, nil
	})

	client, err := NewHTTPClient(
		SetHttpClient(httpClient),
		SetAPIKey("my-awesome-api-key"),
		SetURLs("http://primary.foo", "http://secondary1.foo", "http://secondary2.foo"),
		SetReadPreference(PrimaryPreferred),
		SetMaxRetries(0),
		SetTopologyDiscovery(false),
		SetHealthChecks(true),
		SetHealthCheckInterval(2*time.Second),
		SetAttemptToReviveEndpoints(false),
	)
	require.NoError(t, err)

	// wait for all endpoints to get marked as dead
	spec.RetryOnFalse(t, 50, 200*time.Millisecond, func() bool {
		return !client.topology.HasActiveEndpoint()
	}, "The topology still have active endpoints")
	_, err = client.callAny("GET", "/events", nil)
	require.Error(t, err)

	// wait for all endpoints to get marked as alive
	spec.Retry(t, 50, 200*time.Millisecond, func() error {
		_, err = client.callAny("GET", "/events", nil)
		return err
	})

}

func TestManualDiscoveryPrimaryLost(t *testing.T) {

	httpClient := NewTestHttpClient(func(req *http.Request) (*http.Response, error) {
		if req.Host == "primary.foo" {
			return buildResponse(http.StatusInternalServerError, "Internal server error"), nil
		}
		if req.Host == "secondary1.foo" && req.URL.Path == "/info/shards" {
			info := protocol.Shards{
				NodeId:    "secondary1",
				LeaderId:  "primary2",
				URIScheme: "http",
				Shards: map[string]protocol.ShardDetail{
					"primary2": protocol.ShardDetail{
						NodeId:   "primary2",
						HTTPAddr: "primary2.foo",
					},
					"secondary1": protocol.ShardDetail{
						NodeId:   "secondary1",
						HTTPAddr: "secondary1.foo",
					},
				},
			}
			body, _ := json.Marshal(info)
			return buildResponse(http.StatusOK, string(body)), nil
		}
		if req.Host == "primary2.foo" {
			return buildResponse(http.StatusOK, string(req.Host)), nil
		}
		return nil, errors.New("Unreachable")
	})

	client, err := NewHTTPClient(
		SetHttpClient(httpClient),
		SetAPIKey("my-awesome-api-key"),
		SetURLs("http://primary.foo", "http://secondary1.foo", "http://primary2.foo"),
		SetReadPreference(PrimaryPreferred),
		SetMaxRetries(1),
		SetTopologyDiscovery(false),
		SetHealthChecks(false),
	)
	require.NoError(t, err)

	// force all endpoints to get marked as dead
	_, err = client.callPrimary("GET", "/events", nil)
	require.Error(t, err)
	require.False(t, client.topology.HasActivePrimary())

	// try to discovery a new primary endpoint
	_ = client.discover()
	require.True(t, client.topology.HasActivePrimary())
	resp, err := client.callPrimary("GET", "/events", nil)
	require.NoError(t, err)
	require.Equal(t, "primary2.foo", string(resp))
}

func TestAutoDiscoveryPrimaryLost(t *testing.T) {

	httpClient := NewTestHttpClient(func(req *http.Request) (*http.Response, error) {
		if req.Host == "primary.foo" {
			return buildResponse(http.StatusInternalServerError, "Internal server error"), nil
		}
		if req.Host == "secondary1.foo" && req.URL.Path == "/info/shards" {
			info := protocol.Shards{
				NodeId:    "secondary1",
				LeaderId:  "primary2",
				URIScheme: "http",
				Shards: map[string]protocol.ShardDetail{
					"primary2": protocol.ShardDetail{
						NodeId:   "primary2",
						HTTPAddr: "primary2.foo",
					},
					"secondary1": protocol.ShardDetail{
						NodeId:   "secondary1",
						HTTPAddr: "secondary1.foo",
					},
				},
			}
			body, _ := json.Marshal(info)
			return buildResponse(http.StatusOK, string(body)), nil
		}
		if req.Host == "primary2.foo" {
			return buildResponse(http.StatusOK, string(req.Host)), nil
		}
		return nil, errors.New("Unreachable")
	})

	client, err := NewHTTPClient(
		SetHttpClient(httpClient),
		SetAPIKey("my-awesome-api-key"),
		SetURLs("http://primary.foo", "http://secondary1.foo", "http://primary2.foo"),
		SetReadPreference(PrimaryPreferred),
		SetMaxRetries(1),
		SetTopologyDiscovery(true),
		SetHealthChecks(false),
	)
	require.NoError(t, err)

	resp, err := client.callPrimary("GET", "/events", nil)
	require.NoError(t, err)
	require.True(t, client.topology.HasActivePrimary())
	require.Equal(t, "primary2.foo", string(resp))

}

func buildResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       ioutil.NopCloser(bytes.NewBufferString(body)),
		Header:     make(http.Header),
	}
}

func TestAddSuccess(t *testing.T) {

	event := "Hello world!"
	snap := &protocol.Snapshot{
		HistoryDigest: []byte("history"),
		HyperDigest:   []byte("hyper"),
		Version:       0,
		EventDigest:   []byte(event),
	}
	input, _ := json.Marshal(snap)

	serverURL, tearDown := setupServer(input)
	defer tearDown()
	client := setupClient(t, []string{serverURL})

	snapshot, err := client.Add(event)
	assert.NoError(t, err)
	assert.Equal(t, snap, snapshot, "The snapshots should match")
}

func TestAddBulkSuccess(t *testing.T) {

	eventBulk := []string{"This is event 1", "This is event 2"}
	bulk := []*protocol.Snapshot{
		{
			HistoryDigest: []byte("history"),
			HyperDigest:   []byte("hyper"),
			Version:       0,
			EventDigest:   []byte(eventBulk[0]),
		},
		{
			HistoryDigest: []byte("history"),
			HyperDigest:   []byte("hyper"),
			Version:       1,
			EventDigest:   []byte(eventBulk[1]),
		},
	}
	input, _ := json.Marshal(bulk)

	serverURL, tearDown := setupServer(input)
	defer tearDown()
	client := setupClient(t, []string{serverURL})

	snapshotBulk, err := client.AddBulk(eventBulk)
	assert.NoError(t, err)
	assert.Equal(t, bulk, snapshotBulk, "The snapshots should match")
}

func TestAddWithServerFailure(t *testing.T) {

	serverURL, tearDown := setupServer(nil)
	defer tearDown()
	client := setupClient(t, []string{serverURL})

	event := "Hello world!"
	_, err := client.Add(event)
	assert.Error(t, err)
}

func TestMembership(t *testing.T) {

	event := []byte{0x0}
	version := uint64(0)

	fakeHttpClient := NewTestHttpClient(func(req *http.Request) (*http.Response, error) {
		if req.Host == "primary.foo" && req.URL.Path == "/proofs/membership" {
			m := protocol.MembershipResult{} // We dont care about content here.
			body, _ := json.Marshal(m)
			return buildResponse(http.StatusOK, string(body)), nil
		}
		return nil, errors.New("Unreachable")
	})

	client, err := NewHTTPClient(
		SetHttpClient(fakeHttpClient),
		SetAPIKey("my-awesome-api-key"),
		SetURLs("http://primary.foo"),
		SetReadPreference(Primary),
		SetMaxRetries(0),
		SetTopologyDiscovery(false),
		SetHealthChecks(false),
		SetHasherFunction(hashing.NewFakeXorHasher),
	)
	require.NoError(t, err)

	proof, err := client.Membership(event, &version)
	assert.NotNil(t, proof)
	require.NoError(t, err)

	client.Close()
}

func TestMembershipDigest(t *testing.T) {

	eventDigest := hashing.Digest([]byte{0x0})
	version := uint64(0)

	fakeHttpClient := NewTestHttpClient(func(req *http.Request) (*http.Response, error) {
		if req.URL.Path == "/proofs/digest-membership" {
			m := protocol.MembershipResult{} // We dont care about content here.
			body, _ := json.Marshal(m)
			return buildResponse(http.StatusOK, string(body)), nil
		}
		return nil, errors.New("Unreachable")
	})

	client, err := NewHTTPClient(
		SetHttpClient(fakeHttpClient),
		SetAPIKey("my-awesome-api-key"),
		SetURLs("http://primary.foo"),
		SetReadPreference(Primary),
		SetMaxRetries(0),
		SetTopologyDiscovery(false),
		SetHealthChecks(false),
		SetHasherFunction(hashing.NewSha256Hasher),
	)
	require.NoError(t, err)

	proof, err := client.MembershipDigest(eventDigest, &version)
	assert.NotNil(t, proof)
	assert.NoError(t, err)

	client.Close()
}

func TestMembershipWithServerFailure(t *testing.T) {

	serverURL, tearDown := setupServer(nil)
	defer tearDown()
	client := setupClient(t, []string{serverURL})

	event := "Hello world!"
	i := uint64(0)
	_, err := client.Membership(hashing.Digest(event), &i)
	assert.Error(t, err)
}

func TestIncremental(t *testing.T) {

	start := uint64(2)
	end := uint64(8)

	fakeHttpClient := NewTestHttpClient(func(req *http.Request) (*http.Response, error) {
		if req.Host == "primary.foo" && req.URL.Path == "/proofs/incremental" {
			m := protocol.IncrementalResponse{} // We dont care about content here.
			body, _ := json.Marshal(m)
			return buildResponse(http.StatusOK, string(body)), nil
		}
		return nil, errors.New("Unreachable")
	})

	client, err := NewHTTPClient(
		SetHttpClient(fakeHttpClient),
		SetAPIKey("my-awesome-api-key"),
		SetURLs("http://primary.foo"),
		SetReadPreference(Primary),
		SetMaxRetries(0),
		SetTopologyDiscovery(false),
		SetHealthChecks(false),
		SetHasherFunction(hashing.NewFakeXorHasher),
	)
	require.NoError(t, err)

	proof, err := client.Incremental(start, end)
	assert.NotNil(t, proof)
	assert.NoError(t, err)

	client.Close()
}

func TestIncrementalWithServerFailure(t *testing.T) {

	serverURL, tearDown := setupServer(nil)
	defer tearDown()
	client := setupClient(t, []string{serverURL})

	_, err := client.Incremental(uint64(2), uint64(8))
	assert.Error(t, err)
}

func TestMembershipVerify(t *testing.T) {

	eventDigest := hashing.Digest([]byte{0x0})
	m := &protocol.MembershipResult{
		Exists: true,
		Hyper: map[string]hashing.Digest{
			"0x80|7": hashing.Digest{0x0},
			"0x40|6": hashing.Digest{0x0},
			"0x20|5": hashing.Digest{0x0},
			"0x10|4": hashing.Digest{0x0},
		},
		History:        map[string]hashing.Digest{}, // Dont care about this value in this test
		CurrentVersion: uint64(0),
		QueryVersion:   uint64(0),
		ActualVersion:  uint64(0),
		KeyDigest:      eventDigest,
		Key:            []byte{0x0},
	}
	proof := protocol.ToBalloonProof(m, hashing.NewFakeXorHasher)
	snapshot := &balloon.Snapshot{
		EventDigest:   eventDigest,
		HyperDigest:   hashing.Digest([]byte{0x0}),
		HistoryDigest: hashing.Digest([]byte{0x0}),
		Version:       uint64(0),
	}

	client, err := NewHTTPClient(
		SetAPIKey("my-awesome-api-key"),
		SetReadPreference(Primary),
		SetMaxRetries(0),
		SetTopologyDiscovery(false),
		SetHealthChecks(false),
	)
	require.NoError(t, err)

	ok, err := client.MembershipVerify(eventDigest, proof, snapshot)
	require.True(t, ok)
	require.NoError(t, err)

	client.Close()
}

func TestIncrementalVerify(t *testing.T) {

	eventDigest := hashing.Digest([]byte{0x0})
	m := &protocol.IncrementalResponse{
		Start: 0,
		End:   2,
		AuditPath: map[string]hashing.Digest{
			"0|0": []byte{0x0},
			"1|0": []byte{0x1},
			"2|0": []byte{0x2},
		},
	}
	proof := protocol.ToIncrementalProof(m, hashing.NewFakeXorHasher)

	startSnapshot := &balloon.Snapshot{
		EventDigest:   eventDigest,
		HyperDigest:   hashing.Digest([]byte{0x0}),
		HistoryDigest: hashing.Digest([]byte{0x0}),
		Version:       uint64(0),
	}
	endSnapshot := &balloon.Snapshot{
		EventDigest:   hashing.Digest{},
		HyperDigest:   hashing.Digest{},
		HistoryDigest: hashing.Digest([]byte{0x3}),
		Version:       2,
	}

	client, err := NewHTTPClient(
		SetAPIKey("my-awesome-api-key"),
		SetReadPreference(Primary),
		SetMaxRetries(0),
		SetTopologyDiscovery(false),
		SetHealthChecks(false),
	)
	require.NoError(t, err)

	ok, err := client.IncrementalVerify(proof, startSnapshot, endSnapshot)
	require.True(t, ok)
	require.NoError(t, err)

	client.Close()
}

func TestMembershipAutoVerify(t *testing.T) {

	eventDigest := hashing.Digest([]byte{0x0})
	version := uint64(0)

	fakeHttpClient := NewTestHttpClient(func(req *http.Request) (*http.Response, error) {
		if req.Host == "primary.foo" && req.URL.Path == "/proofs/digest-membership" {
			m := protocol.MembershipResult{
				Exists: true,
				Hyper: map[string]hashing.Digest{
					"0x80|7": hashing.Digest{0x0},
					"0x40|6": hashing.Digest{0x0},
					"0x20|5": hashing.Digest{0x0},
					"0x10|4": hashing.Digest{0x0},
				},
				History:        map[string]hashing.Digest{}, // Dont care about this value in this test
				CurrentVersion: uint64(0),
				QueryVersion:   uint64(0),
				ActualVersion:  uint64(0),
				KeyDigest:      eventDigest,
				Key:            []byte{0x0},
			}
			body, _ := json.Marshal(m)
			return buildResponse(http.StatusOK, string(body)), nil
		}
		if req.Host == "snapshotStore.foo" && req.URL.Path == "/snapshot" {
			ss := protocol.SignedSnapshot{
				Snapshot: &protocol.Snapshot{
					EventDigest:   eventDigest,
					HyperDigest:   hashing.Digest([]byte{0x0}),
					HistoryDigest: hashing.Digest([]byte{0x0}),
					Version:       uint64(0),
				},
				Signature: nil,
			}
			body, _ := json.Marshal(ss)
			return buildResponse(http.StatusOK, string(body)), nil
		}
		return nil, errors.New("Unreachable")
	})

	client, err := NewHTTPClient(
		SetHttpClient(fakeHttpClient),
		SetAPIKey("my-awesome-api-key"),
		SetURLs("http://primary.foo"),
		SetSnapshotStoreURL("http://snapshotStore.foo"),
		SetMaxRetries(0),
		SetTopologyDiscovery(false),
		SetHealthChecks(false),
		SetHasherFunction(hashing.NewFakeXorHasher),
	)
	require.NoError(t, err)

	ok, err := client.MembershipAutoVerify(eventDigest, &version)
	require.True(t, ok)
	require.NoError(t, err)

	client.Close()
}

func TestIncrementalAutoVerify(t *testing.T) {

	start := uint64(0)
	end := uint64(2)

	fakeHttpClient := NewTestHttpClient(func(req *http.Request) (*http.Response, error) {
		if req.Host == "primary.foo" && req.URL.Path == "/proofs/incremental" {
			m := protocol.IncrementalResponse{
				Start: start,
				End:   end,
				AuditPath: map[string]hashing.Digest{
					"0|0": []byte{0x0},
					"1|0": []byte{0x1},
					"2|0": []byte{0x2},
				},
			}
			body, _ := json.Marshal(m)
			return buildResponse(http.StatusOK, string(body)), nil
		}
		if req.Host == "snapshotStore.foo" && req.URL.Path == "/snapshot" {
			params := req.URL.Query()
			version := params.Get("v")

			ss := protocol.SignedSnapshot{
				Signature: nil,
			}

			if version == "0" {
				ss.Snapshot = &protocol.Snapshot{
					EventDigest:   hashing.Digest{},
					HyperDigest:   hashing.Digest{}, // hashing.Digest([]byte{0x0}),
					HistoryDigest: hashing.Digest([]byte{0x0}),
					Version:       start,
				}
			} else if version == "2" {
				ss.Snapshot = &protocol.Snapshot{
					EventDigest:   hashing.Digest{},
					HyperDigest:   hashing.Digest{}, // hashing.Digest([]byte{0x3}),
					HistoryDigest: hashing.Digest([]byte{0x3}),
					Version:       end,
				}
			} else {
				return nil, errors.New("Snapshot version not found in snapshot store")
			}

			body, _ := json.Marshal(ss)
			return buildResponse(http.StatusOK, string(body)), nil
		}
		return nil, errors.New("Unreachable")
	})

	client, err := NewHTTPClient(
		SetHttpClient(fakeHttpClient),
		SetAPIKey("my-awesome-api-key"),
		SetURLs("http://primary.foo"),
		SetSnapshotStoreURL("http://snapshotStore.foo"),
		SetReadPreference(Primary),
		SetMaxRetries(0),
		SetTopologyDiscovery(false),
		SetHealthChecks(false),
		SetHasherFunction(hashing.NewFakeXorHasher),
	)
	require.NoError(t, err)

	ok, err := client.IncrementalAutoVerify(start, end)
	require.True(t, ok)
	require.NoError(t, err)

	client.Close()
}

func defaultHandler(input []byte) func(http.ResponseWriter, *http.Request) {
	statusOK := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		out := new(bytes.Buffer)
		_ = json.Compact(out, input)
		_, _ = w.Write(out.Bytes())
	}

	statusInternalServerError := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
	}

	if input == nil {
		return statusInternalServerError
	} else {
		return statusOK
	}
}

func infoHandler(serverURL string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		var md = make(map[string]interface{})
		md["nodeID"] = "node01"
		md["leaderID"] = "node01"
		md["URIScheme"] = "http://"
		md["meta"] = map[string]map[string]string{
			"node01": map[string]string{
				"HTTPAddr": strings.Trim(serverURL, "http://"),
			},
		}

		out, _ := json.Marshal(md)
		_, _ = w.Write(out)
	}
}
