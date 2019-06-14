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

package gossip

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJoin(t *testing.T) {
	conf := DefaultConfig()
	conf.NodeName = "testNode"
	conf.Role = "auditor"
	conf.BindAddr = "127.0.0.1:12345"

	a, _ := NewAgentFromConfig(conf)
	a.Start()

	testCases := []struct {
		agentState             Status
		addrs                  []string
		expectedContactedHosts int
		expectedErr            error
	}{
		{
			AgentStatusAlive,
			[]string{},
			0,
			nil,
		},
		{
			AgentStatusFailed,
			[]string{},
			0,
			fmt.Errorf("Agent can't join after Leave or Shutdown"),
		},
		{
			AgentStatusAlive,
			[]string{"127.0.0.1:12345"},
			1,
			nil,
		},
	}

	for i, c := range testCases {
		a.Self.Status = c.agentState
		result, err := a.Join(c.addrs)
		require.Equal(t, c.expectedContactedHosts, result, "Wrong expected contacted hosts in test %d.", i)
		require.Equal(t, c.expectedErr, err, "Wrong expected error in test %d.", i)
	}
	_ = a.Shutdown()
}

func TestLeave(t *testing.T) {
	conf := DefaultConfig()
	conf.NodeName = "testNode"
	conf.Role = "auditor"
	conf.BindAddr = "127.0.0.1:12346"

	a, _ := NewAgentFromConfig(conf)
	a.Start()
	testCases := []struct {
		agentState  Status
		expectedErr error
		finalStatus Status
	}{
		{
			AgentStatusLeft,
			nil,
			AgentStatusLeft,
		},
		{
			AgentStatusLeaving,
			fmt.Errorf("Leave already in progress"),
			AgentStatusLeaving,
		},
		{
			AgentStatusShutdown,
			fmt.Errorf("Leave called after Shutdown"),
			AgentStatusShutdown,
		},
		{
			AgentStatusAlive,
			nil,
			AgentStatusLeft,
		},
		{
			AgentStatusFailed,
			nil,
			AgentStatusLeft,
		},
	}

	for i, c := range testCases {
		a.Self.Status = c.agentState
		err := a.Leave()
		require.Equal(t, c.expectedErr, err, "Wrong expected error in test %d.", i)
		require.Equal(t, c.finalStatus, a.Self.Status, "Wrong expected status in test %d.", i)
	}
	_ = a.Shutdown()
}
