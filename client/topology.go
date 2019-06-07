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
	"sync"
)

// QED cluster encapsulation
type topology struct {
	endpoints       []*endpoint
	primary         *endpoint
	cIndex          int // index into endpoints (for round-robin)
	attemptToRevive bool
	sync.RWMutex
}

func newTopology(attemptToRevive bool) *topology {
	return &topology{
		endpoints:       make([]*endpoint, 0),
		cIndex:          -1,
		primary:         nil,
		attemptToRevive: attemptToRevive,
	}
}

func (t *topology) Update(primaryNode string, secondaries ...string) {
	t.Lock()
	defer t.Unlock()

	// Build up new endpoints.
	// If we find an existing endpoint, then use it including
	// the previous number of failures, etc.
	var newEndpoints []*endpoint
	if primaryNode != "" {
		t.primary = newEndpoint(primaryNode, primary)
		newEndpoints = append(newEndpoints, t.primary)
	}

	for _, url := range secondaries {
		var found bool
		for _, oldEndpoint := range t.endpoints {
			if oldEndpoint.url == url {
				// Take over the old endpoint
				newEndpoints = append(newEndpoints, oldEndpoint)
				found = true
				break
			}
		}
		if !found && url != "" {
			// New endpoint didn't exist, so add it to our list of new endpoints.
			newEndpoints = append(newEndpoints, newEndpoint(url, secondary))
		}
	}
	t.endpoints = newEndpoints
	t.cIndex = -1
}

func (t *topology) Primary() (*endpoint, error) {
	t.Lock()
	defer t.Unlock()

	if t.primary == nil {
		return nil, ErrNoPrimary
	}
	if t.primary.IsDead() {
		return t.primary, ErrPrimaryDead
	}
	return t.primary, nil
}

func (t *topology) Endpoints() []*endpoint {
	t.Lock()
	defer t.Unlock()
	return t.endpoints
}

// NextReadendpoint returns the next available endpoint to query
// in a round-robin manner, or ErrNoEndpoint
func (t *topology) NextReadEndpoint(pref ReadPref) (*endpoint, error) {

	t.Lock()
	defer t.Unlock()

	switch pref {

	case PrimaryPreferred:
		if t.primary != nil && !t.primary.IsDead() {
			return t.primary, nil
		}
		fallthrough

	case Secondary:
		i := 0
		numEndpoints := len(t.endpoints)
		if numEndpoints > 0 {
			for {
				if i > numEndpoints {
					break // we visited all endpoints and they all seem to be dead
				}
				t.cIndex++
				if t.cIndex >= numEndpoints {
					t.cIndex = 0
				}
				endpoint := t.endpoints[t.cIndex]
				if endpoint.nodeType == secondary && !endpoint.IsDead() {
					return endpoint, nil
				}
				i++
			}
		}
		break

	case SecondaryPreferred:
		i := 0
		numEndpoints := len(t.endpoints)
		if numEndpoints > 0 {
			for {
				if i > numEndpoints {
					break // we visited all endpoints and they all seem to be dead
				}
				t.cIndex++
				if t.cIndex >= numEndpoints {
					t.cIndex = 0
				}
				endpoint := t.endpoints[t.cIndex]
				if endpoint.nodeType == secondary && !endpoint.IsDead() {
					return endpoint, nil
				}
				i++
			}
		}
		fallthrough

	case Primary:
		if t.primary != nil && !t.primary.IsDead() {
			return t.primary, nil
		}
		break

	case Any:
		i := 0
		numEndpoints := len(t.endpoints)
		if numEndpoints > 0 {
			for {
				if i > numEndpoints {
					break // we visited all endpoints and they all seem to be dead
				}
				t.cIndex++
				if t.cIndex >= numEndpoints {
					t.cIndex = 0
				}
				endpoint := t.endpoints[t.cIndex]
				if !endpoint.IsDead() {
					return endpoint, nil
				}
				i++
			}
		}
		break
	}

	// Now all nodes are marked as dead. If attemptToRevive is disabled,
	// endpoints will never be marked alive again, so we need to
	// mark all of them as alive. This way, they will be picked up
	// in the next call to performRequest.
	if t.attemptToRevive {
		for _, endpoint := range t.endpoints {
			endpoint.MarkAsAlive()
		}
	}

	// we tried every endpoint but there is no one available
	return nil, ErrNoEndpoint
}

// HasActivePrimary returns true if there is an active primary endpoint.
func (t *topology) HasActivePrimary() bool {
	t.Lock()
	defer t.Unlock()
	return t.primary != nil && !t.primary.IsDead()
}

// HasActiveEndpoint returns true there is an active endpoint (primary
// or secondary).
func (t *topology) HasActiveEndpoint() bool {
	t.Lock()
	defer t.Unlock()
	for _, e := range t.endpoints {
		if !e.IsDead() {
			return true
		}
	}
	return false
}
