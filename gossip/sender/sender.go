/*
   Copyright 2018 Banco Bilbao Vizcaya Argentaria, S.A.

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

package sender

import (
	"fmt"
	"time"

	"github.com/bbva/qed/gossip"
	"github.com/bbva/qed/log"
	"github.com/bbva/qed/metrics"
	"github.com/bbva/qed/protocol"
	"github.com/bbva/qed/sign"
)

const (
	NumSenders = 10
)

type Sender struct {
	Agent  *gossip.Agent
	Config *Config
	signer sign.Signer
	quit   chan bool
}

type Config struct {
	BatchSize     int
	BatchInterval time.Duration
	TTL           int
	EachN         int
	SendTimer     time.Duration
}

func DefaultConfig() *Config {
	return &Config{
		BatchSize:     100,
		BatchInterval: 1 * time.Second,
		TTL:           2,
		EachN:         1,
		SendTimer:     500 * time.Millisecond,
	}
}

func NewSender(a *gossip.Agent, c *Config, s sign.Signer) *Sender {
	metrics.QedSenderInstancesCount.Inc()
	return &Sender{
		Agent:  a,
		Config: c,
		signer: s,
		quit:   make(chan bool),
	}
}

func (s Sender) Start(ch chan *protocol.Snapshot) {

	for i := 0; i < NumSenders; i++ {
		go s.batcherSender(i, ch, s.quit)
	}

	for {
		select {
		case <-time.After(s.Config.BatchInterval * 60):
			log.Debug("Messages in sender queue: ", len(ch))
		case <-s.quit:
			return
		}
	}
}

func (s Sender) batcherSender(id int, ch chan *protocol.Snapshot, quit chan bool) {
	batches := []*protocol.BatchSnapshots{}
	batch := &protocol.BatchSnapshots{
		TTL:       s.Config.TTL,
		From:      s.Agent.Self,
		Snapshots: make([]*protocol.SignedSnapshot, 0),
	}

	nextBatch := func() {
		batches = append(batches, batch)
		batch = &protocol.BatchSnapshots{
			TTL:       s.Config.TTL,
			From:      s.Agent.Self,
			Snapshots: make([]*protocol.SignedSnapshot, 0),
		}
	}

	ticker := time.NewTicker(s.Config.SendTimer)

	for {
		select {
		case snap := <-ch:
			if len(batch.Snapshots) == s.Config.BatchSize {
				nextBatch()
			}
			ss, err := s.doSign(snap)
			if err != nil {
				log.Errorf("Failed signing message: %v", err)
			}
			batch.Snapshots = append(batch.Snapshots, ss)

		case <-ticker.C:
			if len(batch.Snapshots) > 0 {
				nextBatch()
			}
			for _, b := range batches {
				go s.sender(*b)
			}
			batches = []*protocol.BatchSnapshots{}

		case <-quit:
			return
		}
	}
}

func (s Sender) sender(batch protocol.BatchSnapshots) {
	msg, _ := batch.Encode()

	peers := s.Agent.Topology.Each(s.Config.EachN, nil)
	for _, peer := range peers.L {
		// Metrics
		metrics.QedSenderBatchesSentTotal.Inc()

		dst := peer.Node()
		log.Debugf("Sending batch %+v to node %+v\n", batch, dst.Name)

		retries := uint(5)
		for {
			err := s.Agent.Memberlist().SendReliable(dst, msg)
			if err == nil {
				break
			} else {
				if retries == 0 {
					log.Infof("Failed send message: %v", err)
					break
				}
				delay := (10 << retries) * time.Millisecond
				time.Sleep(delay)
				retries -= 1
				continue
			}
		}
	}
}

func (s Sender) Stop() {
	metrics.QedSenderInstancesCount.Dec()

	for i := 0; i < NumSenders+1; i++ {
		// INFO: we need NumSenders+1 for the debug ticker in Start function
		s.quit <- true
	}
}

func (s *Sender) doSign(snapshot *protocol.Snapshot) (*protocol.SignedSnapshot, error) {

	signature, err := s.signer.Sign([]byte(fmt.Sprintf("%v", snapshot)))
	if err != nil {
		log.Info("Publisher: error signing snapshot")
		return nil, err
	}
	return &protocol.SignedSnapshot{Snapshot: snapshot, Signature: signature}, nil
}
