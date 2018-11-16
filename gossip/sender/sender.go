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
	"bytes"
	"fmt"
	"time"

	"github.com/bbva/qed/gossip"
	"github.com/bbva/qed/log"
	"github.com/bbva/qed/protocol"
	"github.com/bbva/qed/sign"
	"github.com/hashicorp/go-msgpack/codec"
)

type Config struct {
	BatchSize     uint
	BatchInterval time.Duration
	TTL           int
}

func DefaultConfig() *Config {
	return &Config{
		100,
		1 * time.Second,
		2,
	}
}

func Start(n *gossip.Node, ch chan *protocol.Snapshot) {
	ticker := time.NewTicker(1 * time.Second)

	for {
		select {
		case <-ticker.C:
			msg, _ := encode(getBatch(ch))

			peers := n.GetPeers(1, gossip.AuditorType)
			peers = append(peers, n.GetPeers(1, gossip.MonitorType)...)
			peers = append(peers, n.GetPeers(1, gossip.PublisherType)...)

			for _, peer := range peers {
				err := n.Memberlist().SendReliable(peer.Node, msg)
				if err != nil {
					log.Errorf("Failed send message: %v", err)
				}
			}
			// TODO: Implement graceful shutdown.
		}
	}
}

func encode(msg protocol.BatchSnapshots) ([]byte, error) {
	var buf bytes.Buffer
	encoder := codec.NewEncoder(&buf, &codec.MsgpackHandle{})
	if err := encoder.Encode(msg); err != nil {
		log.Errorf("Failed to encode message: %v", err)
		return nil, err
	}
	return buf.Bytes(), nil
}

func getBatch(ch chan *protocol.Snapshot) protocol.BatchSnapshots {

	var snapshot *protocol.Snapshot
	var batch protocol.BatchSnapshots
	var batchSize int = 100
	var counter int = 0
	batch.Snapshots = make([]*protocol.SignedSnapshot, 0)
	batch.TTL = 3

	for {
		select {
		case snapshot = <-ch:
			counter++
		default:
			return batch
		}

		ss, err := doSign(sign.NewEd25519Signer(), snapshot)
		if err != nil {
			log.Errorf("Failed signing message: %v", err)
		}
		batch.Snapshots = append(batch.Snapshots, ss)

		if counter == batchSize {
			return batch
		}

	}

}

func doSign(signer sign.Signer, snapshot *protocol.Snapshot) (*protocol.SignedSnapshot, error) {

	signature, err := signer.Sign([]byte(fmt.Sprintf("%v", snapshot)))
	if err != nil {
		fmt.Println("Publisher: error signing commitment")
		return nil, err
	}
	return &protocol.SignedSnapshot{snapshot, signature}, nil
}
