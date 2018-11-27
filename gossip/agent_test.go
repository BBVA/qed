package gossip

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bbva/qed/gossip/member"
)

func TestStart(t *testing.T) {

}

func TestJoin(t *testing.T) {
	conf := DefaultConfig()
	conf.NodeName = "testNode"
	conf.Role = member.Auditor
	a, _ := NewAgent(conf, []Processor{FakeProcessor{}})

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
	conf.BindAddr = "0.0.0.0:12346"
	a, _ := NewAgent(conf, []Processor{FakeProcessor{}})

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
	conf.BindAddr = "0.0.0.0:12347"
	a, _ := NewAgent(conf, []Processor{FakeProcessor{}})

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
