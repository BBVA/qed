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
	"sync"
	"time"

	"github.com/bbva/qed/gossip"
	"github.com/bbva/qed/log"
	"github.com/bbva/qed/metrics"
	"github.com/bbva/qed/protocol"
	"github.com/bbva/qed/sign"
)

type Sender struct {
	Agent  *gossip.Agent
	Config *Config
	signer sign.Signer
	quit   chan bool
}

type Config struct {
	BatchSize     uint
	BatchInterval time.Duration
	TTL           int
	EachN         int
}

func DefaultConfig() *Config {
	return &Config{
		BatchSize:     100,
		BatchInterval: 1 * time.Second,
		TTL:           2,
		EachN:         1,
	}
}

func NewSender(a *gossip.Agent, c *Config, s sign.Signer) *Sender {
	metrics.Qed_sender_instances_count.Inc()
	return &Sender{
		Agent:  a,
		Config: c,
		signer: s,
		quit:   make(chan bool),
	}
}

func (s Sender) batcherSender(id int, ch chan *protocol.Snapshot, quit chan bool) {
	batches := []*protocol.BatchSnapshots{}
	batch := &protocol.BatchSnapshots{
		TTL:       s.Config.TTL,
		From:      s.Agent.Self,
		Snapshots: make([]*protocol.SignedSnapshot, 0),
	}

	ticker := time.NewTicker(500 * time.Millisecond)

	resetBatches := func() {
		batches = append(batches, batch)
		batch = &protocol.BatchSnapshots{
			TTL:       s.Config.TTL,
			From:      s.Agent.Self,
			Snapshots: make([]*protocol.SignedSnapshot, 0),
		}
	}

	for {
		select {
		case snap := <-ch:
			// TODO: batchSize 100 must be configurable
			if len(batch.Snapshots) == 100 {
				resetBatches()
			}
			ss, err := s.doSign(snap)
			if err != nil {
				log.Errorf("Failed signing message: %v", err)
			}
			batch.Snapshots = append(batch.Snapshots, ss)

		case <-ticker.C:
			if len(batch.Snapshots) > 0 {
				resetBatches()
			}
			for _, b := range batches {
				s.sender(*b)
			}
			batches = []*protocol.BatchSnapshots{}

		case <-quit:
			return
			// default:
			// fmt.Println("Doing nothing", id)
		}
	}
}

func (s Sender) sender(batch protocol.BatchSnapshots) {
	var wg sync.WaitGroup
	msg, _ := batch.Encode()

	peers := s.Agent.Topology.Each(s.Config.EachN, nil)
	for _, peer := range peers.L {
		// Metrics
		metrics.Qed_sender_batches_sent_total.Inc()

		dst := peer.Node()
		log.Infof("Sending batch %+v to node %+v\n", batch, dst.Name)
		wg.Add(1)
		go func() {
			err := s.Agent.Memberlist().SendReliable(dst, msg)
			if err != nil {
				log.Errorf("Failed send message: %v", err)
			}
		}()
	}
	wg.Wait()
	log.Infof("Sent batch %+v to nodes %+v\n", batch, peers.L)
}

func (s Sender) Start(ch chan *protocol.Snapshot) {
	ticker := time.NewTicker(1000 * time.Millisecond)

	for i := 0; i < 1; i++ {
		go s.batcherSender(i, ch, s.quit)
	}

	for {
		select {
		case <-ticker.C:
			log.Debug("QUEUE LENGTH: ", len(ch))
		case <-s.quit:
			return
		}
	}
}

func (s Sender) Stop() {
	metrics.Qed_sender_instances_count.Dec()
	s.quit <- true
}

func (s *Sender) doSign(snapshot *protocol.Snapshot) (*protocol.SignedSnapshot, error) {

	signature, err := s.signer.Sign([]byte(fmt.Sprintf("%v", snapshot)))
	if err != nil {
		fmt.Println("Publisher: error signing snapshot")
		return nil, err
	}
	return &protocol.SignedSnapshot{Snapshot: snapshot, Signature: signature}, nil
}
