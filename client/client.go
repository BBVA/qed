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

// Package client implements the client to interact with QED servers.
package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/bbva/qed/balloon"
	"github.com/bbva/qed/crypto/hashing"
	"github.com/bbva/qed/log"
	"github.com/bbva/qed/protocol"
)

// HTTPClient is an HTTP QED client.
type HTTPClient struct {
	httpClient          *http.Client
	retrier             RequestRetrier
	topology            *topology
	snapshotStore       *endpoint
	apiKey              string
	readPreference      ReadPref
	maxRetries          int
	healthCheckEnabled  bool
	healthCheckTimeout  time.Duration
	healthCheckInterval time.Duration
	discoveryEnabled    bool
	hasherF             func() hashing.Hasher

	mu                sync.RWMutex // guards the next block
	running           bool
	healthCheckStopCh chan bool // notify healthchecker to stop, and notify back
	discoveryStopCh   chan bool // notify sniffer to stop, and notify back
}

// NewSimpleHTTPClient creates a new short-lived client thath can be
// used in use cases where you need one client per request.
//
// All checks are disabled, including timeouts and periodic checks.
// The number of retries is set to 0.
func NewSimpleHTTPClient(httpClient *http.Client, urls []string, snapshotStoreURL string) (*HTTPClient, error) {

	// defaultTransport := http.DefaultTransport.(*http.Transport)
	// // Create new Transport that ignores self-signed SSL
	// transportWithSelfSignedTLS := &http.Transport{
	// 	Proxy:                 defaultTransport.Proxy,
	// 	DialContext:           defaultTransport.DialContext,
	// 	MaxIdleConns:          defaultTransport.MaxIdleConns,
	// 	IdleConnTimeout:       defaultTransport.IdleConnTimeout,
	// 	ExpectContinueTimeout: defaultTransport.ExpectContinueTimeout,
	// 	TLSHandshakeTimeout:   defaultTransport.TLSHandshakeTimeout,
	// 	TLSClientConfig:       &tls.Config{InsecureSkipVerify: insecure},
	// }

	// httpClient := &http.Client{
	// 	Timeout:   DefaultTimeout,
	// 	Transport: transportWithSelfSignedTLS,
	// }

	if len(urls) == 0 {
		return nil, errors.New("Invalid urls")
	}

	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	client := &HTTPClient{
		httpClient:          httpClient,
		topology:            newTopology(false),
		snapshotStore:       newEndpoint(snapshotStoreURL, store),
		healthCheckEnabled:  false,
		healthCheckTimeout:  off,
		healthCheckInterval: off,
		discoveryEnabled:    false,
		readPreference:      Primary,
		maxRetries:          0,
		retrier:             NewNoRequestRetrier(httpClient),
		hasherF:             hashing.NewSha256Hasher,
	}

	client.topology.Update(urls[0], urls[1:]...)

	return client, nil
}

// NewHTTPClientFromConfig initializes a client from a configuration.
func NewHTTPClientFromConfig(conf *Config) (*HTTPClient, error) {
	options, err := configToOptions(conf)
	if err != nil {
		return nil, err
	}
	return NewHTTPClient(options...)
}

// NewHTTPClient creates a new HTTP client to work with QED.
//
// The client, by default, is meant to be long-lived and shared across
// your application. If you need a short-lived client, e.g. for request-scope,
// consider using NewSimpleHttpClient instead.
//
func NewHTTPClient(options ...HTTPClientOptionF) (*HTTPClient, error) {

	client := &HTTPClient{
		httpClient:          http.DefaultClient,
		topology:            newTopology(false),
		snapshotStore:       newEndpoint("", store),
		healthCheckEnabled:  DefaultHealthCheckEnabled,
		healthCheckTimeout:  DefaultHealthCheckTimeout,
		healthCheckInterval: DefaultHealthCheckInterval,
		discoveryEnabled:    DefaultTopologyDiscoveryEnabled,
		readPreference:      Primary,
		maxRetries:          DefaultMaxRetries,
		healthCheckStopCh:   make(chan bool),
		discoveryStopCh:     make(chan bool),
	}

	// Run the options on the client
	for _, option := range options {
		if err := option(client); err != nil {
			return nil, err
		}
	}

	// configure retrier
	_ = client.setRetrier(client.maxRetries)

	// Initial topology assignment
	if client.discoveryEnabled {
		// try to discover the cluster topology initially
		if err := client.discover(); err != nil {
			log.Infof("Unable to get QED topology, we will try it later: %v", err)
		}
	}

	if client.healthCheckEnabled {
		// perform an initial healthcheck
		client.clusterHealthCheck(client.healthCheckTimeout)
	}

	// Ensure that we have at least one endpoint, the primary, available
	if !client.topology.HasActivePrimary() {
		log.Infof("QED does not have a primary node or it is down, we will try it later.")
	}

	// if t.discoveryEnabled {
	// 	go t.startDiscoverer() // periodically update cluster information
	// }
	if client.healthCheckEnabled {
		go client.startHealthChecker() // periodically ping all nodes of the cluster
	}

	client.mu.Lock()
	client.running = true
	client.mu.Unlock()

	return client, nil
}

// Close stops the background processes that the client is running,
// i.e. sniffing the cluster periodically and running health checks
// on the nodes.
//
// If the background processes are not running, this is a no-op.
func (c *HTTPClient) Close() {
	c.mu.RLock()
	if !c.running {
		c.mu.RUnlock()
		return
	}
	c.mu.RUnlock()

	log.Info("Closing QED client...")

	close(c.healthCheckStopCh)
	close(c.discoveryStopCh)

	c.mu.Lock()
	if c.topology != nil {
		c.topology = nil
	}
	c.running = false
	c.mu.Unlock()

	log.Info("QED client closed")

}

func (c *HTTPClient) setRetrier(maxRetries int) error {
	if maxRetries < 0 {
		return errors.New("MaxRetries must be greater than or equal to 0")
	}
	if maxRetries == 0 {
		c.retrier = NewNoRequestRetrier(c.httpClient)
	} else {
		// Create a Retrier that will wait for 100ms between requests.
		ticks := make([]int, maxRetries)
		for i := 0; i < len(ticks); i++ {
			ticks[i] = 1000
		}
		backoff := NewSimpleBackoff(ticks...)
		c.retrier = NewBackoffRequestRetrier(c.httpClient, c.maxRetries, backoff)
	}
	return nil
}

// startDiscoverer periodically runs discover.
func (c *HTTPClient) startDiscoverer() {
	c.mu.RLock()

}

func (c *HTTPClient) callPrimary(method, path string, data []byte) ([]byte, error) {

	var endpoint *endpoint
	var err error
	var discoveryRetried, healthRetried bool
	for {
		// we always send POST requests to the primary endpoint
		endpoint, err = c.topology.Primary()

		if err == ErrPrimaryDead {
			if c.healthCheckEnabled && !healthRetried {
				c.clusterHealthCheck(c.healthCheckTimeout)
				healthRetried = true
				continue
			}
			healthRetried = true
		}

		if err == ErrNoPrimary || (err == ErrPrimaryDead && healthRetried) {
			if c.discoveryEnabled && !discoveryRetried {
				err = c.discover()
				discoveryRetried = true
				if err != nil {
					return nil, err
				}
				continue
			}
			return nil, err
		}

		break
	}
	return c.doReq(method, endpoint, path, data)
}

func (c *HTTPClient) callAny(method, path string, data []byte) ([]byte, error) {

	var endpoint *endpoint
	var retried bool
	var errTopology, errRequest error
	var result []byte

	for {
		// check every endpoint available in a round-robin manner
		endpoint, errTopology = c.topology.NextReadEndpoint(c.readPreference)
		if errTopology != nil {
			if !retried && c.discoveryEnabled {
				_ = c.discover()
				retried = true
				continue
			}
			if errRequest != nil {
				return nil, errRequest
			}
			return nil, errTopology
		}
		result, errRequest = c.doReq(method, endpoint, path, data)
		if errRequest == nil {
			break
		}
		endpoint.MarkAsDead()
	}
	if errRequest != nil {
		return nil, errRequest
	}

	return result, errTopology
}

func (c *HTTPClient) doReq(method string, endpoint *endpoint, path string, data []byte) ([]byte, error) {

	url, err := url.Parse(endpoint.URL() + path)
	if err != nil {
		return nil, err
	}

	// Build request
	req, err := NewRetriableRequest(method, url.String(), data)
	if err != nil {
		return nil, err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Api-Key", c.apiKey)

	// Get response
	resp, err := c.retrier.DoReq(req)
	if err != nil {
		log.Infof("Request error: %v\n", err)
		endpoint.MarkAsDead()
		log.Infof("%s is dead\n", endpoint)
		return nil, err
	}

	var bodyBytes []byte
	if resp.Body != nil {
		defer resp.Body.Close()
		bodyBytes, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
	}

	if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		return nil, fmt.Errorf("Invalid request %v", string(bodyBytes))
	}

	// we successfully made a request to this endpoint
	endpoint.MarkAsHealthy()

	return bodyBytes, nil
}

// healthCheck does a health check on all nodes in the cluster.
// Depending on the node state, it marks connections as dead, alive etc.
// The timeout specifies how long to wait for a response from QED.
func (c *HTTPClient) clusterHealthCheck(timeout time.Duration) {

	var wg sync.WaitGroup
	for _, e := range c.topology.Endpoints() {

		wg.Add(1)
		endpoint := e
		// the goroutines execute the health-check HTTP request and sets status
		go func(endpointURL string) {
			defer wg.Done()

			// Run a HEAD request against QED with a timeout
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			req, err := http.NewRequest("HEAD", endpointURL+"/healthcheck", nil)
			if err != nil {
				return
			}
			req.Header.Set("Api-Key", c.apiKey)

			resp, err := c.httpClient.Do(req.WithContext(ctx))
			if err != nil {
				log.Infof("%s is dead", endpoint.URL())
				endpoint.MarkAsDead()
				return
			}
			if resp != nil {
				status := resp.StatusCode
				if resp.Body != nil {
					resp.Body.Close()
				}
				if status >= 200 && status < 300 {
					endpoint.MarkAsAlive()
				} else {
					log.Infof("%s is dead [status=%d]", endpoint.URL(), status)
					endpoint.MarkAsDead()
				}
			}
		}(endpoint.URL())
	}
	wg.Wait()
}

// discover uses the shards info API to return the list of nodes in the cluster.
// It uses the list of URLs passed on startup plus the list of URLs found
// by the preceding discovery process (if discovery is enabled).
func (c *HTTPClient) discover() error {

	for {
		e, err := c.topology.NextReadEndpoint(Any)
		if err != nil {
			return err
		}

		body, err := c.doReq("GET", e, "/info/shards", nil)
		if err == nil {
			var shards protocol.Shards
			err = json.Unmarshal(body, &shards)
			if err != nil {
				return err
			}

			var primary string
			secondaries := make([]string, 0)
			for id, shard := range shards.Shards {
				if id == shards.LeaderId {
					primary = fmt.Sprintf("%s://%s", shards.URIScheme, shard.HTTPAddr)
				} else {
					secondaries = append(secondaries, fmt.Sprintf("%s://%s", shards.URIScheme, shard.HTTPAddr))
				}
			}
			c.topology.Update(primary, secondaries...)
			break
		}
	}

	return nil
}

// startHealthChecker periodically runs healthcheck.
func (c *HTTPClient) startHealthChecker() {
	c.mu.RLock()
	timeout := c.healthCheckTimeout
	interval := c.healthCheckInterval
	c.mu.RUnlock()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-c.healthCheckStopCh:
			return
		case <-ticker.C:
			c.clusterHealthCheck(timeout)
		}
	}
}

// Ping will do a healthcheck request to the primary node
func (c *HTTPClient) Ping() error {
	_, err := c.callPrimary("HEAD", "/healthcheck", nil)
	if err != nil {
		return err
	}
	return nil
}

// Add will do a request to the server with a post data to store a new event.
func (c *HTTPClient) Add(event string) (*protocol.Snapshot, error) {

	data, _ := json.Marshal(&protocol.Event{Event: []byte(event)})
	body, err := c.callPrimary("POST", "/events", data)
	if err != nil {
		return nil, err
	}

	var snapshot protocol.Snapshot
	err = json.Unmarshal(body, &snapshot)
	if err != nil {
		return nil, err
	}

	return &snapshot, nil
}

// AddBulk will do a request to the server with a post data to store a bulk of new events.
func (c *HTTPClient) AddBulk(events []string) ([]*protocol.Snapshot, error) {

	eventBulk := protocol.EventsBulk{}
	for _, e := range events {
		eventBulk.Events = append(eventBulk.Events, []byte(e))
	}

	data, _ := json.Marshal(eventBulk)
	body, err := c.callPrimary("POST", "/events/bulk", data)
	if err != nil {
		return nil, err
	}

	bs := []*protocol.Snapshot{}
	err = json.Unmarshal(body, &bs)
	if err != nil {
		return nil, err
	}

	return bs, nil
}

// Membership will ask for a Proof to the server.
func (c *HTTPClient) Membership(key []byte, version *uint64) (*balloon.MembershipProof, error) {
	var query []byte

	if version == nil {
		query, _ = json.Marshal(&protocol.MembershipQuery{
			Key: key,
		})
	} else {
		query, _ = json.Marshal(&protocol.MembershipQuery{
			Key:     key,
			Version: version,
		})

	}
	body, err := c.callAny("POST", "/proofs/membership", query)
	if err != nil {
		return nil, err
	}

	var result *protocol.MembershipResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	proof := protocol.ToBalloonProof(result, c.hasherF)
	return proof, nil
}

// Membership will ask for a Proof to the server.
func (c *HTTPClient) MembershipDigest(keyDigest hashing.Digest, version *uint64) (*balloon.MembershipProof, error) {
	var query []byte

	if version == nil {
		query, _ = json.Marshal(&protocol.MembershipDigest{
			KeyDigest: keyDigest,
		})
	} else {
		query, _ = json.Marshal(&protocol.MembershipDigest{
			KeyDigest: keyDigest,
			Version:   version,
		})
	}

	body, err := c.callAny("POST", "/proofs/digest-membership", query)
	if err != nil {
		return nil, err
	}

	var result *protocol.MembershipResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	proof := protocol.ToBalloonProof(result, c.hasherF)
	return proof, nil
}

// MembershipVerify will compute the Proof given in Membership and the snapshot from the
// add and returns the verification result.
func (c *HTTPClient) MembershipVerify(
	eventDigest hashing.Digest,
	proof *balloon.MembershipProof,
	snapshot *balloon.Snapshot,
) (bool, error) {

	return proof.DigestVerify(eventDigest, snapshot), nil
}

// MembershipAutoVerify will compute the Proof given in Membership,
// get hyper and history digests from the snapshot store,
// and returns the verification result.
func (c *HTTPClient) MembershipAutoVerify(eventDigest hashing.Digest, version *uint64) (bool, error) {

	// Get membership proof
	proof, err := c.MembershipDigest(eventDigest, version)
	if err != nil {
		log.Info("Error getting membership proof: %s", err)
		return false, err
	}

	// Build snapshot info from snapshot store and params.
	snapshot := &balloon.Snapshot{
		HistoryDigest: nil,
		HyperDigest:   nil,
		Version:       uint64(0),
		EventDigest:   eventDigest,
	}

	s, err := c.GetSnapshot(proof.QueryVersion)
	if err != nil {
		log.Info("Error getting snapshot from snapshot store: %s", err)
		return false, err
	}

	snapshot.HistoryDigest = s.HistoryDigest
	snapshot.HyperDigest = s.HyperDigest

	if proof.CurrentVersion != proof.ActualVersion {
		s, err := c.GetSnapshot(proof.CurrentVersion)
		if err != nil {
			log.Info("Error getting snapshot from snapshot store: %s", err)
			return false, err
		}
		snapshot.HyperDigest = s.HyperDigest
	}

	// Verify
	return proof.DigestVerify(eventDigest, snapshot), nil
}

// GetSnapshot will ask for a given snapshot version to the snapshot store
// and returns the required snapshot
func (c *HTTPClient) GetSnapshot(version uint64) (*protocol.Snapshot, error) {
	var ss protocol.SignedSnapshot

	body, err := c.doReq("GET", c.snapshotStore, fmt.Sprintf("/snapshot?v=%d", version), nil)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &ss)
	if err != nil {
		return nil, err
	}

	return ss.Snapshot, nil
}

// Incremental will ask for an IncrementalProof to the server.
func (c *HTTPClient) Incremental(start, end uint64) (*balloon.IncrementalProof, error) {

	query, _ := json.Marshal(&protocol.IncrementalRequest{
		Start: start,
		End:   end,
	})

	body, err := c.callAny("POST", "/proofs/incremental", query)
	if err != nil {
		return nil, err
	}

	var response *protocol.IncrementalResponse
	_ = json.Unmarshal(body, &response)

	proof := protocol.ToIncrementalProof(response, c.hasherF)
	return proof, nil
}

// IncrementalVerify will verify a proof against two snapshots, given these 3 elements.
// It returns the verification result.
func (c *HTTPClient) IncrementalVerify(
	proof *balloon.IncrementalProof,
	startSnapshot, endSnapshot *balloon.Snapshot,
) (bool, error) {

	return proof.Verify(startSnapshot, endSnapshot), nil
}

// IncrementalAutoVerify will ask for an Incremental proof to the server, given both a
// start and end versions. With these versions, it will ask to the snapshot store to get
// both snapshots, and finally it will verify the proof.
// It returns the verification result.
func (c *HTTPClient) IncrementalAutoVerify(
	start, end uint64,
) (bool, error) {

	// Get incrementral proof
	proof, err := c.Incremental(start, end)
	if err != nil {
		return false, err
	}

	// Build start and end snapshots info from snapshot snapshot store and params.
	// Start snapshot
	startSnapshot := balloon.Snapshot{
		EventDigest: hashing.Digest{},
		HyperDigest: hashing.Digest{},
		Version:     start,
	}
	s, err := c.GetSnapshot(start)
	if err != nil {
		log.Info("Error getting snapshot from snapshot store: %s", err)
		return false, err
	}
	startSnapshot.HistoryDigest = s.HistoryDigest

	// End snapshot
	endSnapshot := balloon.Snapshot{
		EventDigest: hashing.Digest{},
		HyperDigest: hashing.Digest{},
		Version:     end,
	}
	s, err = c.GetSnapshot(end)
	if err != nil {
		log.Info("Error getting snapshot from snapshot store: %s", err)
		return false, err
	}
	endSnapshot.HistoryDigest = s.HistoryDigest

	// Verify
	return c.IncrementalVerify(proof, &startSnapshot, &endSnapshot)
}

// Backup function ...
func (c *HTTPClient) Backup(path string) error {
	return nil
}

// Restore function ...
func (c *HTTPClient) Restore(path string) error {
	return nil
}
