/*
   Copyright 2018 Banco Bilbao Vizcaya Argentaria, S.A.

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

// Package client implements the command line interface to interact with
// the REST API
package client

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/bbva/qed/balloon"
	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/log"
	"github.com/bbva/qed/metrics"
	"github.com/bbva/qed/protocol"
)

// HTTPClient ist the stuct that has the required information for the cli.
type HTTPClient struct {
	conf *Config
	*http.Client
	topology Topology
}

// NewHTTPClient will return a new instance of HTTPClient.
func NewHTTPClient(conf Config) *HTTPClient {
	var tlsConf *tls.Config

	if conf.Insecure {
		tlsConf = &tls.Config{InsecureSkipVerify: true}
	} else {
		tlsConf = &tls.Config{}
	}

	client := &HTTPClient{
		&conf,
		&http.Client{
			Timeout: time.Second * 10,
			Transport: &http.Transport{
				Dial: (&net.Dialer{
					Timeout: 5 * time.Second,
				}).Dial,
				TLSClientConfig:     tlsConf,
				TLSHandshakeTimeout: 5 * time.Second,
			},
		},
		Topology{},
	}

	// Initial topology assignment
	client.topology.Leader = conf.Endpoints[0]
	client.topology.Endpoints = conf.Endpoints

	var info map[string]interface{}
	var err error

	info, err = client.getClusterInfo()
	if err != nil {
		log.Errorf("Failed to get raft cluster info. Error: %v", err)
		return nil
	}

	client.updateTopology(info)

	return client
}

func (c *HTTPClient) exponentialBackoff(req *http.Request) (*http.Response, error) {

	var retries uint

	for {
		resp, err := c.Do(req)
		if err != nil {
			if retries == 5 {
				return nil, err
			}
			retries = retries + 1
			delay := time.Duration(10 << retries * time.Millisecond)
			time.Sleep(delay)
			continue
		}
		return resp, err
	}
}

func (c HTTPClient) getClusterInfo() (map[string]interface{}, error) {
	var retries uint
	info := make(map[string]interface{})

	for {
		body, err := c.doReq("GET", "/info/shards", []byte{})

		if err != nil {
			log.Debugf("Failed to get raft cluster info through server %s. Error: %v",
				c.topology.Leader, err)
			if retries == 5 {
				return nil, err
			}
			c.topology.Leader = c.topology.Endpoints[rand.Intn(len(c.topology.Endpoints))]
			retries = retries + 1
			delay := time.Duration(10 << retries * time.Millisecond)
			time.Sleep(delay)
			continue
		}

		err = json.Unmarshal(body, &info)
		if err != nil {
			return nil, err
		}

		return info, err
	}
}

func (c *HTTPClient) updateTopology(info map[string]interface{}) {

	clusterMeta := info["meta"].(map[string]interface{})
	leaderID := info["leaderID"].(string)
	scheme := info["URIScheme"].(string)

	var leaderAddr string
	var endpoints []string

	leaderMeta := clusterMeta[leaderID].(map[string]interface{})
	for k, addr := range leaderMeta {
		if k == "HTTPAddr" {
			leaderAddr = scheme + addr.(string)
		}
	}
	c.topology.Leader = leaderAddr

	for _, nodeMeta := range clusterMeta {
		for k, address := range nodeMeta.(map[string]interface{}) {
			if k == "HTTPAddr" {
				url := scheme + address.(string)
				endpoints = append(endpoints, url)
			}
		}
	}
	c.topology.Endpoints = endpoints
}

func (c *HTTPClient) doReq(method, path string, data []byte) ([]byte, error) {

	url, err := url.Parse(c.topology.Leader + path)
	if err != nil {
		return nil, err //panic(err)
	}

	req, err := http.NewRequest(method, fmt.Sprintf("%s", url), bytes.NewBuffer(data))
	if err != nil {
		return nil, err //panic(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Api-Key", c.conf.APIKey)

	resp, err := c.exponentialBackoff(req)
	if err != nil {
		return nil, err
		// NetworkTransport error. Check topology info
	}
	defer resp.Body.Close()

	bodyBytes, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode >= 500 {
		return nil, fmt.Errorf("Server error: %v", string(bodyBytes))
		// Non Leader error. Check topology info.
	}

	if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		return nil, fmt.Errorf("Invalid request %v", string(bodyBytes))
	}

	return bodyBytes, nil
}

// Ping will do a request to the server
func (c HTTPClient) Ping() error {
	_, err := c.doReq("GET", "/health-check", nil)
	if err != nil {
		return err
	}

	return nil
}

// Add will do a request to the server with a post data to store a new event.
func (c *HTTPClient) Add(event string) (*protocol.Snapshot, error) {

	metrics.ClientEventAdd.Inc()
	data, _ := json.Marshal(&protocol.Event{Event: []byte(event)})

	body, err := c.doReq("POST", "/events", data)
	if err != nil {
		return nil, err
	}

	var snapshot protocol.Snapshot
	_ = json.Unmarshal(body, &snapshot)

	return &snapshot, nil

}

// Membership will ask for a Proof to the server.
func (c *HTTPClient) Membership(key []byte, version uint64) (*protocol.MembershipResult, error) {

	metrics.ClientQueryMembership.Inc()

	query, _ := json.Marshal(&protocol.MembershipQuery{
		Key:     key,
		Version: version,
	})

	body, err := c.doReq("POST", "/proofs/membership", query)
	if err != nil {
		return nil, err
	}

	var proof *protocol.MembershipResult
	_ = json.Unmarshal(body, &proof)

	return proof, nil

}

// Membership will ask for a Proof to the server.
func (c *HTTPClient) MembershipDigest(keyDigest hashing.Digest, version uint64) (*protocol.MembershipResult, error) {

	metrics.ClientQueryMembership.Inc()

	query, _ := json.Marshal(&protocol.MembershipDigest{
		KeyDigest: keyDigest,
		Version:   version,
	})

	body, err := c.doReq("POST", "/proofs/digest-membership", query)
	if err != nil {
		return nil, err
	}

	var proof *protocol.MembershipResult
	_ = json.Unmarshal(body, &proof)

	return proof, nil

}

// Incremental will ask for an IncrementalProof to the server.
func (c *HTTPClient) Incremental(start, end uint64) (*protocol.IncrementalResponse, error) {

	metrics.ClientQueryIncremental.Inc()

	query, _ := json.Marshal(&protocol.IncrementalRequest{
		Start: start,
		End:   end,
	})

	body, err := c.doReq("POST", "/proofs/incremental", query)
	if err != nil {
		return nil, err
	}

	var response *protocol.IncrementalResponse
	_ = json.Unmarshal(body, &response)

	return response, nil
}

// Verify will compute the Proof given in Membership and the snapshot from the
// add and returns a proof of existence.
func (c HTTPClient) Verify(
	result *protocol.MembershipResult,
	snap *protocol.Snapshot,
	hasherF func() hashing.Hasher,
) bool {

	proof := protocol.ToBalloonProof(result, hasherF)

	return proof.Verify(snap.EventDigest, &balloon.Snapshot{
		EventDigest:   snap.EventDigest,
		HistoryDigest: snap.HistoryDigest,
		HyperDigest:   snap.HyperDigest,
		Version:       snap.Version,
	})

}

// Verify will compute the Proof given in Membership and the snapshot from the
// add and returns a proof of existence.
func (c HTTPClient) DigestVerify(
	result *protocol.MembershipResult,
	snap *protocol.Snapshot,
	hasherF func() hashing.Hasher,
) bool {

	proof := protocol.ToBalloonProof(result, hasherF)

	return proof.DigestVerify(snap.EventDigest, &balloon.Snapshot{
		EventDigest:   snap.EventDigest,
		HistoryDigest: snap.HistoryDigest,
		HyperDigest:   snap.HyperDigest,
		Version:       snap.Version,
	})

}

func (c HTTPClient) VerifyIncremental(
	result *protocol.IncrementalResponse,
	startSnapshot, endSnapshot *protocol.Snapshot,
	hasher hashing.Hasher,
) bool {

	proof := protocol.ToIncrementalProof(result, hasher)

	start := &balloon.Snapshot{
		EventDigest:   startSnapshot.EventDigest,
		HistoryDigest: startSnapshot.HistoryDigest,
		HyperDigest:   startSnapshot.HyperDigest,
		Version:       startSnapshot.Version,
	}
	end := &balloon.Snapshot{
		EventDigest:   endSnapshot.EventDigest,
		HistoryDigest: endSnapshot.HistoryDigest,
		HyperDigest:   endSnapshot.HyperDigest,
		Version:       endSnapshot.Version,
	}

	return proof.Verify(start, end)
}
