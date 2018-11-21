package auditor

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/bbva/qed/gossip"
	"github.com/bbva/qed/log"
	"github.com/bbva/qed/protocol"
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
	var batch protocol.BatchSnapshots
	err := batch.Decode(msg)
	if err != nil {
		log.Errorf("Unable to decode message: %v", err)
		return
	}

	log.Infof("Batch received, TTL: %d: %v", batch.TTL, batch)

	a.Process(&batch)

	if batch.TTL <= 0 {
		return
	}

	// Exclude origin and myself from peer list to send message.
	var excludedPeers []*protocol.Source
	excludedPeers = append(excludedPeers, batch.From)
	addr, port := a.Agent.GetAddrPort()
	myself := &protocol.Source{
		Addr: addr,
		Port: port,
		Role: a.Agent.Metadata().Role.String(),
	}
	excludedPeers = append(excludedPeers, myself)

	peers := a.Agent.GetPeers(2, gossip.AuditorType, excludedPeers)
	peers = append(peers, a.Agent.GetPeers(2, gossip.MonitorType, excludedPeers)...)
	peers = append(peers, a.Agent.GetPeers(2, gossip.PublisherType, excludedPeers)...)

	batch.TTL--
	batch.From = myself
	newMsg, _ := batch.Encode()

	for _, peer := range peers {
		err := a.Agent.Memberlist().SendReliable(peer, newMsg)
		if err != nil {
			log.Errorf("Failed send message: %v", err)
		}
	}

}
func (a *Auditor) Process(b *protocol.BatchSnapshots) {
	for i := 0; i < len(b.Snapshots); i++ {
		res, err := http.Get(fmt.Sprintf("http://127.0.0.1:8888/?nodeType=auditor&id=%d", b.Snapshots[0].Snapshot.Version))
		if err != nil || res == nil {
			log.Debugf("Error contacting service with error %v", err)
		}
		// to reuse connections we need to do this
		io.Copy(ioutil.Discard, res.Body)
		res.Body.Close()

		// time.Sleep(1 * time.Second)
	}

	log.Debugf("process(): Processed %v elements of batch id %v", len(b.Snapshots), b.Snapshots[0].Snapshot.Version)
}
