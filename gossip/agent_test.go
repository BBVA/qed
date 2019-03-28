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

	"github.com/bbva/qed/gossip/member"
	"github.com/bbva/qed/metrics"
)

func TestJoin(t *testing.T) {
	conf := DefaultConfig()
	conf.NodeName = "testNode"
	conf.Role = member.Auditor
	conf.BindAddr = "127.0.0.1:12345"
	metricsServer := metrics.NewServer("127.0.0.2:23464")
	a, _ := NewAgent(conf, []Processor{FakeProcessor{}}, metricsServer, make(chan string, 1))

	testCases := []struct {
		agentState             member.Status
		addrs                  []string
		expectedContactedHosts int
		expectedErr            error
	}{
		{
			member.Alive,
			[]string{},
			0,
			nil,
		},
		{
			member.Failed,
			[]string{},
			0,
			fmt.Errorf("Agent can't join after Leave or Shutdown"),
		},
		{
			member.Alive,
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
}

func TestLeave(t *testing.T) {
	conf := DefaultConfig()
	conf.NodeName = "testNode"
	conf.Role = member.Auditor
	conf.BindAddr = "127.0.0.1:12346"
	metricsServer := metrics.NewServer("127.0.0.2:13445")
	a, _ := NewAgent(conf, []Processor{FakeProcessor{}}, metricsServer, make(chan string, 1))

	testCases := []struct {
		agentState  member.Status
		expectedErr error
		finalStatus member.Status
	}{
		{
			member.Left,
			nil,
			member.Left,
		},
		{
			member.Leaving,
			fmt.Errorf("Leave already in progress"),
			member.Leaving,
		},
		{
			member.Shutdown,
			fmt.Errorf("Leave called after Shutdown"),
			member.Shutdown,
		},
		{
			member.Alive,
			nil,
			member.Left,
		},
		{
			member.Failed,
			nil,
			member.Left,
		},
	}

	for i, c := range testCases {
		a.Self.Status = c.agentState
		err := a.Leave()
		require.Equal(t, c.expectedErr, err, "Wrong expected error in test %d.", i)
		require.Equal(t, c.finalStatus, a.Self.Status, "Wrong expected status in test %d.", i)
	}
}

func TestShutdown(t *testing.T) {

	conf := DefaultConfig()
	conf.NodeName = "testNode"
	conf.Role = member.Auditor
	conf.BindAddr = "127.0.0.1:12347"
	metricsServer := metrics.NewServer("127.0.0.2:43512")
	a, _ := NewAgent(conf, []Processor{FakeProcessor{}}, metricsServer, make(chan string, 1))

	testCases := []struct {
		agentState  member.Status
		expectedErr error
		finalStatus member.Status
	}{
		{
			member.Shutdown,
			nil,
			member.Shutdown,
		},
		{
			member.Left,
			nil,
			member.Shutdown,
		},
		{
			member.Alive,
			nil,
			member.Shutdown,
		},
		{
			member.Failed,
			nil,
			member.Shutdown,
		},
		{
			member.Leaving,
			nil,
			member.Shutdown,
		},
	}

	for i, c := range testCases {
		a.Self.Status = c.agentState
		err := a.Shutdown()
		require.Equal(t, c.expectedErr, err, "Wrong expected error in test %d.", i)
		require.Equal(t, c.finalStatus, a.Self.Status, "Wrong expected status in test %d.", i)
	}
}
