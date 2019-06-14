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
	"context"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/bbva/qed/protocol"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func TestBatchProcessorLoop(t *testing.T) {
	var wg sync.WaitGroup

	ts := &testSubscriber{}
	conf := DefaultConfig()
	conf.NodeName = "testNode"
	conf.Role = "auditor"
	conf.BindAddr = "127.0.0.1:12345"

	a, err := NewAgentFromConfig(conf)
	require.NoError(t, err, "Error creating agent!")

	p := NewBatchProcessor(a, nil)
	a.In.Subscribe(BatchMessageType, p, 1)
	defer p.Stop()

	a.Out.Subscribe(BatchMessageType, ts, 5)
	batch := &protocol.BatchSnapshots{}
	buf, _ := batch.Encode()
	m1 := &Message{
		Kind:    BatchMessageType,
		From:    nil,
		TTL:     0,
		Payload: buf,
	}

	wg.Add(1)
	go func() {
		for {
			select {
			case m2 := <-ts.ch:
				require.Equal(t, m1, m2, "Messages must be equal")
				wg.Done()
				return
			}
		}
	}()

	_ = a.In.Publish(m1)

	wg.Wait()
}

func TestBatchProcessorWasProcessed(t *testing.T) {

	ts := &testSubscriber{}

	conf := DefaultConfig()
	conf.NodeName = "testNode"
	conf.Role = "auditor"
	conf.BindAddr = "127.0.0.1:12345"

	a, err := NewAgentFromConfig(conf)
	require.NoError(t, err, "Error creating agent!")

	p := NewBatchProcessor(a, nil)
	a.In.Subscribe(BatchMessageType, p, 0)
	defer p.Stop()

	a.Out.Subscribe(BatchMessageType, ts, 5)
	batch := &protocol.BatchSnapshots{}
	buf, _ := batch.Encode()
	m1 := &Message{
		Kind:    BatchMessageType,
		From:    nil,
		TTL:     0,
		Payload: buf,
	}

	_ = a.In.Publish(m1)
	_ = a.In.Publish(m1)
	// give time for the scheduler to route all the messages
	time.Sleep(1 * time.Second)

	// only one message must be in the output channel as one must be
	// dropped by the wasProcessed function
	require.Equal(t, 1, len(ts.ch), "Output queue must be 1, duplicate event must be dropped by processor")
}

type fakeTaskFactory struct{}

func (f fakeTaskFactory) Metrics() []prometheus.Collector {
	return []prometheus.Collector{
		prometheus.NewCounter(prometheus.CounterOpts{Name: "fakeCounterMetric"}),
	}
}

func (f fakeTaskFactory) New(c context.Context) Task {
	return func() error {
		return nil
	}
}

func TestBatchProcessorRegisterMetrics(t *testing.T) {

	conf := DefaultConfig()
	conf.NodeName = "testNode"
	conf.Role = "auditor"
	conf.BindAddr = "127.0.0.1:12345"
	conf.MetricsAddr = "127.0.0.1:12346"

	a, err := NewAgentFromConfig(conf)
	require.NoError(t, err, "Error creating agent!")
	a.Start()
	defer a.Shutdown()
	// wait for agent to start
	// all services
	time.Sleep(3 * time.Second)

	p := NewBatchProcessor(a, []TaskFactory{&fakeTaskFactory{}})
	a.In.Subscribe(BatchMessageType, p, 0)
	defer p.Stop()

	resp, err := http.Get("http://" + conf.MetricsAddr + "/metrics")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	found := strings.Index(string(body), "fakeCounterMetric")

	require.True(t, found > 0, "Metric not found!")
}
