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
	"fmt"
	"sync"
	"time"
)

type nodeType int

const (
	primary nodeType = iota
	secondary
)

// endpoint represents status information of a single endpointection to a node in a cluster
type endpoint struct {
	sync.RWMutex
	url       string // [scheme://host:port]
	nodeType  nodeType
	failures  int
	dead      bool
	deadSince *time.Time
}

// newEndpoint creates a new endpoint to the given URL [scheme://host:port].
func newEndpoint(url string, nodeType nodeType) *endpoint {
	return &endpoint{
		url:      url,
		nodeType: nodeType,
	}
}

// String returns a representation of the endpoint status.
func (c *endpoint) String() string {
	c.RLock()
	defer c.RUnlock()
	return fmt.Sprintf("%s [type=%v,dead=%v,failures=%d,deadSince=%v]", c.url, c.nodeType, c.dead, c.failures, c.deadSince)
}

// URL returns the url string of this endpoint.
func (c *endpoint) URL() string {
	c.RLock()
	defer c.RUnlock()
	return c.url
}

// Type returns true if the node type is primary.
func (c *endpoint) IsPrimary() bool {
	c.RLock()
	defer c.RUnlock()
	return c.nodeType == primary
}

// IsDead returns true if this endpoint is marked as dead, i.e. a previous
// request to the url has been unsuccessful.
func (c *endpoint) IsDead() bool {
	c.RLock()
	defer c.RUnlock()
	return c.dead
}

// MarkAsDead marks this endpoint as dead, increments the failures
// counter and stores the current time in dead since.
func (c *endpoint) MarkAsDead() {
	c.Lock()
	c.dead = true
	if c.deadSince == nil {
		utcNow := time.Now().UTC()
		c.deadSince = &utcNow
	}
	c.failures++
	c.Unlock()
}

// MarkAsAlive marks this endpoint as eligible to be returned from the
// pool of endpoint by the selector.
func (c *endpoint) MarkAsAlive() {
	c.Lock()
	c.dead = false
	c.Unlock()
}

// MarkAsHealthy marks this endpoint as healthy, i.e. a request has been
// successfully performed with it.
func (c *endpoint) MarkAsHealthy() {
	c.Lock()
	c.dead = false
	c.deadSince = nil
	c.failures = 0
	c.Unlock()
}
