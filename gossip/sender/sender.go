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
		EachN:         2,
	}
}

func NewSender(a *gossip.Agent, c *Config, s sign.Signer) *Sender {
	return &Sender{
		Agent:  a,
		Config: c,
		signer: s,
		quit:   make(chan bool),
	}
}

func (s Sender) Start(ch chan *protocol.Snapshot) {
	ticker := time.NewTicker(1 * time.Second)

	for {
		select {
		case <-ticker.C:
			batch := s.getBatch(ch)
			if batch == nil {
				continue
			}
			log.Debugf("Encoding batch: %+v", batch)
			msg, _ := batch.Encode()

			peers := s.Agent.Topology.Each(s.Config.EachN, nil)

			for _, peer := range peers.L {
				err := s.Agent.Memberlist().SendReliable(peer.Node(), msg)
				if err != nil {
					log.Errorf("Failed send message: %v", err)
				}
			}
		case <-s.quit:
			return
		}
	}
}

func (s Sender) Stop() {
	s.quit <- true
}

func (s *Sender) getBatch(ch chan *protocol.Snapshot) *protocol.BatchSnapshots {

	if len(ch) == 0 {
		return nil
	}

	var snapshot *protocol.Snapshot
	var batch protocol.BatchSnapshots

	var batchSize int = 100
	var counter int = 0

	batch.Snapshots = make([]*protocol.SignedSnapshot, 0)
	batch.TTL = s.Config.TTL
	batch.From = s.Agent.Self

	for {
		select {
		case snapshot = <-ch:
			counter++
		default:
			return &batch
		}

		ss, err := s.doSign(snapshot)
		if err != nil {
			log.Errorf("Failed signing message: %v", err)
		}
		batch.Snapshots = append(batch.Snapshots, ss)

		if counter == batchSize {
			return &batch
		}

	}

}

func (s *Sender) doSign(snapshot *protocol.Snapshot) (*protocol.SignedSnapshot, error) {

	signature, err := s.signer.Sign([]byte(fmt.Sprintf("%v", snapshot)))
	if err != nil {
		fmt.Println("Publisher: error signing commitment")
		return nil, err
	}
	return &protocol.SignedSnapshot{snapshot, signature}, nil
}
