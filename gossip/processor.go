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
	"bytes"
	"context"

	"github.com/bbva/qed/crypto/hashing"
	"github.com/bbva/qed/log2"
	"github.com/bbva/qed/protocol"
	"github.com/hashicorp/go-msgpack/codec"
	"github.com/prometheus/client_golang/prometheus"
)

// A processor mission is to translate from
// and to the gossip network []byte type to
// whatever has semantic sense.
//
// Also it should enqueue tasks in the agent task
// manager.
type Processor interface {
	Start()
	Stop()
	Metrics() []prometheus.Collector
}

// Reads agents in queue, and generates a
// *protocol.BatchSnapshots queue.
// It also calls the tasks factories and enqueue
// the generated tasks in the agent task manager.
type BatchProcessor struct {
	mh      *codec.MsgpackHandle
	a       *Agent
	tf      []TaskFactory
	metrics []prometheus.Collector
	quitCh  chan bool
	ctx     context.Context
	id      int
	log     log2.Logger
}

func NewBatchProcessor(a *Agent, tf []TaskFactory, l log2.Logger) *BatchProcessor {

	logger := l
	if logger == nil {
		logger = log2.L()
	}

	b := &BatchProcessor{
		mh:     &codec.MsgpackHandle{},
		a:      a,
		tf:     tf,
		quitCh: make(chan bool),
		ctx:    context.WithValue(context.Background(), "agent", a),
		log:    logger,
	}

	// register all tasks metrics
	for _, t := range tf {
		b.metrics = append(b.metrics, t.Metrics()...)
	}

	return b
}

func (d *BatchProcessor) Stop() {
	close(d.quitCh)
}

func (d *BatchProcessor) Metrics() []prometheus.Collector {
	return d.metrics
}

// This function requires the cache of the agent to be defined, and will return
// false if the cache is not present in the agent
func (d *BatchProcessor) wasProcessed(b *protocol.BatchSnapshots) bool {
	if d.a.Cache == nil {
		return false
	}

	var buf bytes.Buffer
	err := codec.NewEncoder(&buf, d.mh).Encode(b.Snapshots)
	if err != nil {
		d.log.Infof("Error encoding batchsnapshots to calculate its digest. Dropping batch.")
		return false
	}
	bb := buf.Bytes()
	digest := hashing.NewSha256Hasher().Do(bb)
	// batch already processed, discard it
	_, err = d.a.Cache.Get(digest)
	if err == nil {
		return true
	}
	_ = d.a.Cache.Set(digest, []byte{0x1}, 0)
	return false
}

func (d *BatchProcessor) Subscribe(id int, ch <-chan *Message) {
	d.id = id

	if d.a.metrics != nil {
		d.a.metrics.MustRegister(d.metrics...)
	}

	go func() {
		for {
			select {
			case msg := <-ch:
				// if the message is not a batch, ignore it
				if msg.Kind != BatchMessageType {
					d.log.Debug("BatchProcessor got an unknown message from agent")
					continue
				}

				batch := new(protocol.BatchSnapshots)
				err := batch.Decode(msg.Payload)
				if err != nil {
					d.log.Info("BatchProcessor unable to decode batch!. Dropping message.")
					continue
				}

				if d.wasProcessed(batch) {
					d.log.Debug("BatchProcessor got an already processed message from agent")
					continue
				}

				ctx := context.WithValue(d.ctx, "batch", batch)
				for _, t := range d.tf {
					d.log.Debug("Batch processor creating a new task")
					err := d.a.Tasks.Add(t.New(ctx))
					if err != nil {
						d.log.Infof("BatchProcessor was unable to enqueue new task becasue %v", err)
					}
				}

				_ = d.a.Out.Publish(msg)
			case <-d.quitCh:
				return
			}
		}
	}()
}
