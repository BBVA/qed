package auditor

import (
	"bytes"

	"github.com/bbva/qed/gossip"
	"github.com/bbva/qed/log"
	"github.com/bbva/qed/protocol"
	"github.com/hashicorp/go-msgpack/codec"
	"github.com/hashicorp/memberlist"
)

type Auditor struct {
	Agent  *gossip.Agent
	Config *Config
	quit   chan bool
}

type Config struct {
}

func DefaultConfig() *Config {
	return &Config{}
}

func NewAuditorHandlerBuilder(c *Config) gossip.MessageHandlerBuilder {
	auditor := &Auditor{
		Config: c,
		quit:   make(chan bool),
	}
	return func(a *gossip.Agent) gossip.MessageHandler {
		auditor.Agent = a
		return auditor
	}
}

func (a *Auditor) HandleMsg(msg []byte) {

	batch, err := a.decode(msg)
	if err != nil {
		log.Errorf("Unable to decode message: %v", err)
		return
	}

	log.Infof("Batch received, TTL: %d: %v", batch.TTL, *batch)

	if batch.TTL <= 0 {
		return
	}

	peers := a.Agent.GetPeers(2, gossip.AuditorType)
	peers = append(peers, a.Agent.GetPeers(2, gossip.MonitorType)...)
	peers = append(peers, a.Agent.GetPeers(2, gossip.PublisherType)...)

	batch.TTL--
	newBatch, _ := encode(batch)

	for _, peer := range peers {
		err := a.Agent.Memberlist().SendReliable(&memberlist.Node{Addr: peer.Addr, Port: peer.Port}, newBatch)
		if err != nil {
			log.Errorf("Failed send message: %v", err)
		}
	}

}

func (a *Auditor) decode(buf []byte) (*protocol.BatchSnapshots, error) {
	batch := &protocol.BatchSnapshots{}
	reader := bytes.NewReader(buf)
	decoder := codec.NewDecoder(reader, &codec.MsgpackHandle{})
	if err := decoder.Decode(batch); err != nil {
		log.Errorf("Failed to decode snapshots batch: %v", err)
		return nil, err
	}
	return batch, nil
}

func encode(msg *protocol.BatchSnapshots) ([]byte, error) {
	var buf bytes.Buffer
	encoder := codec.NewEncoder(&buf, &codec.MsgpackHandle{})
	if err := encoder.Encode(msg); err != nil {
		log.Errorf("Failed to encode message: %v", err)
		return nil, err
	}
	return buf.Bytes(), nil
}
