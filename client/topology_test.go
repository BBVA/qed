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
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTopologyUpdate(t *testing.T) {
	topology := newTopology(false)
	require.Empty(t, topology.Endpoints(), "The list of endpoints should be empty")

	topology.Update(
		"http://primary:8080",
		"http://secondary1:8080",
		"http://secondary2:8080",
	)

	endpoints := topology.Endpoints()
	expectedEndpoints := []*endpoint{
		newEndpoint("http://primary:8080", primary),
		newEndpoint("http://secondary1:8080", secondary),
		newEndpoint("http://secondary2:8080", secondary),
	}
	require.ElementsMatch(t, expectedEndpoints, endpoints, "The endpoints should match")
}

func TestTopologyPrimary(t *testing.T) {

	topology := newTopology(false)
	endpoint, err := topology.Primary()
	require.Nil(t, endpoint)
	require.Error(t, err)

	topology.Update("http://primary:8080")
	endpoint, err = topology.Primary()
	require.NoError(t, err)
	require.Equalf(t, primary, endpoint.nodeType, "The type of node should match")
	require.Equalf(t, "http://primary:8080", endpoint.URL(), "The URL should match")

}

func TestTopologyNextReadEndpoint(t *testing.T) {

	testCases := []struct {
		primary           string
		secondaries       []string
		readPref          ReadPref
		expectError       bool
		rounds            int
		expectedEndpoints []*endpoint
	}{
		{
			// Preference=Primary with existent primary node
			primary: "http://primary:8080",
			secondaries: []string{
				"http://secondary1:8080",
				"http://secondary2:8080",
			},
			readPref:    Primary,
			expectError: false,
			rounds:      4,
			expectedEndpoints: []*endpoint{
				newEndpoint("http://primary:8080", primary),
				newEndpoint("http://primary:8080", primary),
				newEndpoint("http://primary:8080", primary),
				newEndpoint("http://primary:8080", primary),
			},
		},
		{
			// Preference=Primary with non-existent primary node
			primary: "",
			secondaries: []string{
				"http://secondary1:8080",
				"http://secondary2:8080",
			},
			readPref:          Primary,
			expectError:       true,
			rounds:            4,
			expectedEndpoints: []*endpoint{},
		},
		{
			// Preference=PrimaryPreferred with existent primary node
			primary: "http://primary:8080",
			secondaries: []string{
				"http://secondary1:8080",
				"http://secondary2:8080",
			},
			readPref:    PrimaryPreferred,
			expectError: false,
			rounds:      4,
			expectedEndpoints: []*endpoint{
				newEndpoint("http://primary:8080", primary),
				newEndpoint("http://primary:8080", primary),
				newEndpoint("http://primary:8080", primary),
				newEndpoint("http://primary:8080", primary),
			},
		},
		{
			// Preference=PrimaryPreferred with non-existent primary node
			primary: "",
			secondaries: []string{
				"http://secondary1:8080",
				"http://secondary2:8080",
			},
			readPref:    PrimaryPreferred,
			expectError: false,
			rounds:      4,
			expectedEndpoints: []*endpoint{
				newEndpoint("http://secondary1:8080", secondary),
				newEndpoint("http://secondary2:8080", secondary),
				newEndpoint("http://secondary1:8080", secondary),
				newEndpoint("http://secondary2:8080", secondary),
			},
		},
		{
			// Preference=Secondary with existent secondary nodes
			primary: "http://primary:8080",
			secondaries: []string{
				"http://secondary1:8080",
				"http://secondary2:8080",
			},
			readPref:    Secondary,
			expectError: false,
			rounds:      4,
			expectedEndpoints: []*endpoint{
				newEndpoint("http://secondary1:8080", secondary),
				newEndpoint("http://secondary2:8080", secondary),
				newEndpoint("http://secondary1:8080", secondary),
				newEndpoint("http://secondary2:8080", secondary),
			},
		},
		{
			// Preference=Secondary with non-existent secondary nodes
			primary:           "http://primary:8080",
			secondaries:       []string{},
			readPref:          Secondary,
			expectError:       true,
			rounds:            4,
			expectedEndpoints: []*endpoint{},
		},
		{
			// Preference=SecondaryPreferred with existent secondary nodes
			primary: "http://primary:8080",
			secondaries: []string{
				"http://secondary1:8080",
				"http://secondary2:8080",
			},
			readPref:    SecondaryPreferred,
			expectError: false,
			rounds:      4,
			expectedEndpoints: []*endpoint{
				newEndpoint("http://secondary1:8080", secondary),
				newEndpoint("http://secondary2:8080", secondary),
				newEndpoint("http://secondary1:8080", secondary),
				newEndpoint("http://secondary2:8080", secondary),
			},
		},
		{
			// Preference=SecondaryPreferred with non-existent secondary nodes
			primary:     "http://primary:8080",
			secondaries: []string{},
			readPref:    SecondaryPreferred,
			expectError: false,
			rounds:      4,
			expectedEndpoints: []*endpoint{
				newEndpoint("http://primary:8080", primary),
				newEndpoint("http://primary:8080", primary),
				newEndpoint("http://primary:8080", primary),
				newEndpoint("http://primary:8080", primary),
			},
		},
		{
			// Preference=SecondaryPreferred with nor existent primary neither secondary nodes
			primary:           "",
			secondaries:       []string{},
			readPref:          SecondaryPreferred,
			expectError:       true,
			rounds:            4,
			expectedEndpoints: []*endpoint{},
		},
		{
			// Preference=Any with both existent primary and secondary nodes
			primary: "http://primary:8080",
			secondaries: []string{
				"http://secondary1:8080",
				"http://secondary2:8080",
			},
			readPref:    Any,
			expectError: false,
			rounds:      4,
			expectedEndpoints: []*endpoint{
				newEndpoint("http://primary:8080", primary),
				newEndpoint("http://secondary1:8080", secondary),
				newEndpoint("http://secondary2:8080", secondary),
				newEndpoint("http://primary:8080", primary),
			},
		},
		{
			// Preference=Any with non-existent nodes
			primary:           "",
			secondaries:       []string{},
			readPref:          Any,
			expectError:       true,
			rounds:            4,
			expectedEndpoints: []*endpoint{},
		},
	}

	for i, c := range testCases {
		topology := newTopology(false)
		topology.Update(c.primary, c.secondaries...)
		collectedEndpoints := make([]*endpoint, 0)
		round := 0
		for {
			if round >= c.rounds {
				break
			}
			round++
			endpoint, err := topology.NextReadEndpoint(c.readPref)
			if c.expectError {
				require.NotNil(t, err, "Should return error for test case %d", i)
				break
			}
			collectedEndpoints = append(collectedEndpoints, endpoint)
		}
		require.ElementsMatch(t, c.expectedEndpoints, collectedEndpoints, "The endpoints should match for test case %d", i)
	}

}

func TestTopologyHasActivePrimary(t *testing.T) {
	topology := newTopology(false)
	require.False(t, topology.HasActivePrimary())

	topology.Update("http://primary:8080")
	require.True(t, topology.HasActivePrimary())
}

func TestTopologyHashActiveEndpoint(t *testing.T) {
	topology := newTopology(false)
	require.False(t, topology.HasActiveEndpoint())

	topology.Update("http://primary:8080", "http://secondary1:8080")
	require.True(t, topology.HasActiveEndpoint())
}
