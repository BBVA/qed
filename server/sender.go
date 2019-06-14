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

package server

import (
	"fmt"
	"time"

	"github.com/bbva/qed/crypto/sign"
	"github.com/bbva/qed/gossip"
	"github.com/bbva/qed/log"
	"github.com/bbva/qed/metrics"
	"github.com/bbva/qed/protocol"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// SENDER

	QedSenderInstancesCount = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "qed_sender_instances_count",
			Help: "Number of sender agents running",
		},
	)
	QedSenderBatchesSentTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "qed_sender_batches_sent_total",
			Help: "Number of batches sent by the sender.",
		},
	)
)

type Sender struct {
	agent      *gossip.Agent
	Interval   time.Duration
	BatchSize  int
	NumSenders int
	TTL        int
	signer     sign.Signer
	quitCh     chan bool
}

func NewSender(a *gossip.Agent, s sign.Signer, size, ttl, n int) *Sender {
	return &Sender{
		agent:      a,
		Interval:   100 * time.Millisecond,
		BatchSize:  size,
		NumSenders: n,
		TTL:        ttl,
		signer:     s,
		quitCh:     make(chan bool),
	}
}

// Start NumSenders concurrent senders and waits for them
// to finish
func (s Sender) Start(ch chan *protocol.Snapshot) {
	QedSenderInstancesCount.Inc()
	for i := 0; i < s.NumSenders; i++ {
		log.Debugf("Starting sender %d", i)
		go s.batcher(i, ch)
	}
}

func (s Sender) RegisterMetrics(srv *metrics.Server) {
	metrics := []prometheus.Collector{
		QedSenderInstancesCount,
		QedSenderBatchesSentTotal,
	}
	srv.MustRegister(metrics...)
}

func (s Sender) newBatch() *protocol.BatchSnapshots {
	return &protocol.BatchSnapshots{
		Snapshots: make([]*protocol.SignedSnapshot, 0),
	}
}

// Sign snapshots, build batches of signed snapshots and send those batches
// to other members of the gossip network.
// If the out queue is full,  we drop the current batch and pray other sender will
// send the batches to the gossip network.
func (s Sender) batcher(id int, ch chan *protocol.Snapshot) {
	batch := s.newBatch()

	for {
		select {
		case snap := <-ch:
			if len(batch.Snapshots) == s.BatchSize {
				payload, err := batch.Encode()
				if err != nil {
					log.Infof("Error encoding batch, dropping it")
					continue
				}

				s.agent.Out.Publish(&gossip.Message{
					Kind:    gossip.BatchMessageType,
					TTL:     s.TTL,
					Payload: payload,
				})
				QedSenderBatchesSentTotal.Inc()

				batch = s.newBatch()
			}
			ss, err := s.doSign(snap)
			if err != nil {
				log.Errorf("Failed signing message: %v", err)
			}
			batch.Snapshots = append(batch.Snapshots, ss)
		case <-time.After(s.Interval):
			// send whatever we have on each tick, do not wait
			// to have complete batches
			if len(batch.Snapshots) > 0 {
				payload, err := batch.Encode()
				if err != nil {
					log.Infof("Error encoding batch, dropping it")
					continue
				}
				s.agent.Out.Publish(&gossip.Message{
					Kind:    gossip.BatchMessageType,
					TTL:     s.TTL,
					Payload: payload,
				})
				QedSenderBatchesSentTotal.Inc()
				batch = s.newBatch()
			}
		case <-s.quitCh:
			return
		}
	}
}

func (s Sender) Stop() {
	QedSenderInstancesCount.Dec()
	close(s.quitCh)
}

func (s *Sender) doSign(snapshot *protocol.Snapshot) (*protocol.SignedSnapshot, error) {
	signature, err := s.signer.Sign([]byte(fmt.Sprintf("%v", snapshot)))
	if err != nil {
		log.Info("Publisher: error signing snapshot")
		return nil, err
	}
	return &protocol.SignedSnapshot{Snapshot: snapshot, Signature: signature}, nil
}
