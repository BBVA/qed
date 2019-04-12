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
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/bbva/qed/hashing"
	"github.com/pkg/errors"

	"github.com/bbva/qed/log"
	"github.com/bbva/qed/protocol"
	"github.com/stretchr/testify/assert"
)

func setupServer(input []byte) (string, func()) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)

	mux.HandleFunc("/info/shards", infoHandler(server.URL))
	mux.HandleFunc("/events", defaultHandler(input))
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
	)
	if err != nil {
		t.Fatal(errors.Wrap(err, "Cannot create http client"))
	}
	return client
}

func TestCallPrimaryWorking(t *testing.T) {

	log.SetLogger("TestCallPrimaryWorking", log.SILENT)

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

func TestCallPrimaryFails(t *testing.T) {

	log.SetLogger("TestCallPrimaryFails", log.SILENT)

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

func TestCallAnyPrimaryFails(t *testing.T) {

	log.SetLogger("TestCallAnyPrimaryFails", log.SILENT)

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

	log.SetLogger("TestCallAnyAllFail", log.SILENT)

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

	log.SetLogger("TestHealthCheck", log.SILENT)

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
	client.healthCheck(5 * time.Second)
	time.Sleep(1 * time.Second)
	require.True(t, client.topology.HasActiveEndpoint())
}

func TestPeriodicHealthCheck(t *testing.T) {

	log.SetLogger("TestPeriodicHealthCheck", log.SILENT)

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
	time.Sleep(2 * time.Second)
	require.False(t, client.topology.HasActiveEndpoint())
	_, err = client.callAny("GET", "/events", nil)
	require.Error(t, err)

	// wait for all endpoints to get marked as alive
	time.Sleep(1 * time.Second)
	_, err = client.callAny("GET", "/events", nil)
	require.NoError(t, err)

}

func TestManualDiscoveryPrimaryLost(t *testing.T) {

	log.SetLogger("TestDiscoverPrimaryLost", log.SILENT)

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
	client.discover()
	require.True(t, client.topology.HasActivePrimary())
	resp, err := client.callPrimary("GET", "/events", nil)
	require.NoError(t, err)
	require.Equal(t, "primary2.foo", string(resp))
}

func TestAutoDiscoveryPrimaryLost(t *testing.T) {

	log.SetLogger("TestDiscoverPrimaryLost", log.SILENT)

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

	log.SetLogger("TestAddSuccess", log.SILENT)

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

func TestAddWithServerFailure(t *testing.T) {

	log.SetLogger("TestAddWithServerFailure", log.SILENT)

	serverURL, tearDown := setupServer(nil)
	defer tearDown()
	client := setupClient(t, []string{serverURL})

	event := "Hello world!"
	_, err := client.Add(event)
	assert.Error(t, err)
}

func TestMembership(t *testing.T) {

	log.SetLogger("TestMembership", log.SILENT)

	event := "Hello world!"
	version := uint64(0)
	fakeResult := &protocol.MembershipResult{
		Key:            []byte(event),
		KeyDigest:      []byte("digest"),
		Exists:         true,
		Hyper:          make(map[string]hashing.Digest),
		History:        make(map[string]hashing.Digest),
		CurrentVersion: version,
		QueryVersion:   version,
		ActualVersion:  version,
	}
	inputJSON, _ := json.Marshal(fakeResult)

	serverURL, tearDown := setupServer(inputJSON)
	defer tearDown()
	client := setupClient(t, []string{serverURL})

	result, err := client.Membership([]byte(event), version)
	assert.NoError(t, err)
	assert.Equal(t, fakeResult, result, "The inputs should match")
}

func TestDigestMembership(t *testing.T) {

	log.SetLogger("TestDigestMembership", log.SILENT)

	event := "Hello world!"
	version := uint64(0)
	fakeResult := &protocol.MembershipResult{
		Key:            []byte(event),
		KeyDigest:      []byte("digest"),
		Exists:         true,
		Hyper:          make(map[string]hashing.Digest),
		History:        make(map[string]hashing.Digest),
		CurrentVersion: version,
		QueryVersion:   version,
		ActualVersion:  version,
	}
	inputJSON, _ := json.Marshal(fakeResult)

	serverURL, tearDown := setupServer(inputJSON)
	defer tearDown()
	client := setupClient(t, []string{serverURL})

	result, err := client.MembershipDigest([]byte("digest"), version)
	assert.NoError(t, err)
	assert.Equal(t, fakeResult, result, "The results should match")
}

func TestMembershipWithServerFailure(t *testing.T) {

	log.SetLogger("TestMembershipWithServerFailure", log.SILENT)

	serverURL, tearDown := setupServer(nil)
	defer tearDown()
	client := setupClient(t, []string{serverURL})

	event := "Hello world!"

	_, err := client.Membership([]byte(event), 0)
	assert.Error(t, err)
}

func TestIncremental(t *testing.T) {

	log.SetLogger("TestIncremental", log.SILENT)

	start := uint64(2)
	end := uint64(8)
	fakeResult := &protocol.IncrementalResponse{
		Start:     start,
		End:       end,
		AuditPath: map[string]hashing.Digest{"0|0": []uint8{0x0}},
	}

	inputJSON, _ := json.Marshal(fakeResult)

	serverURL, tearDown := setupServer(inputJSON)
	defer tearDown()
	client := setupClient(t, []string{serverURL})

	result, err := client.Incremental(start, end)
	assert.NoError(t, err)
	assert.Equal(t, fakeResult, result, "The inputs should match")
}

func TestIncrementalWithServerFailure(t *testing.T) {

	log.SetLogger("TestIncrementalWithServerFailure", log.SILENT)

	serverURL, tearDown := setupServer(nil)
	defer tearDown()
	client := setupClient(t, []string{serverURL})

	_, err := client.Incremental(uint64(2), uint64(8))
	assert.Error(t, err)
}

// TODO implement a test to verify proofs using fake hash function

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
