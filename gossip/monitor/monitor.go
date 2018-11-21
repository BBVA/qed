package monitor

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/bbva/qed/gossip"
	"github.com/bbva/qed/log"
	"github.com/bbva/qed/protocol"
)

type Monitor struct {
	Agent  *gossip.Agent
	Config *Config
	quit   chan bool
}

type Config struct {
}

func DefaultConfig() *Config {
	return &Config{}
}

func NewMonitorHandlerBuilder(c *Config) gossip.MessageHandlerBuilder {
	monitor := &Monitor{
		Config: c,
		quit:   make(chan bool),
	}
	return func(a *gossip.Agent) gossip.MessageHandler {
		monitor.Agent = a
		return monitor
	}
}

func (m *Monitor) HandleMsg(msg []byte) {
	var batch protocol.BatchSnapshots
	err := batch.Decode(msg)
	if err != nil {
		log.Errorf("Unable to decode message: %v", err)
		return
	}

	log.Infof("Batch received, TTL: %d: %v", batch.TTL, batch)

	m.Process(&batch)

	if batch.TTL <= 0 {
		return
	}

	// Exclude origin and myself from peer list to send message.
	var excludedPeers []*protocol.Source
	excludedPeers = append(excludedPeers, batch.From)
	addr, port := m.Agent.GetAddrPort()
	myself := &protocol.Source{
		Addr: addr,
		Port: port,
		Role: m.Agent.Metadata().Role.String(),
	}
	excludedPeers = append(excludedPeers, myself)

	peers := m.Agent.GetPeers(2, gossip.AuditorType, excludedPeers)
	peers = append(peers, m.Agent.GetPeers(2, gossip.MonitorType, excludedPeers)...)
	peers = append(peers, m.Agent.GetPeers(2, gossip.PublisherType, excludedPeers)...)

	batch.TTL--
	batch.From = myself
	newMsg, _ := batch.Encode()

	for _, peer := range peers {
		err := m.Agent.Memberlist().SendReliable(peer, newMsg)
		if err != nil {
			log.Errorf("Failed send message: %v", err)
		}
	}

}

func (m *Monitor) Process(b *protocol.BatchSnapshots) {
	for i := 0; i < len(b.Snapshots); i++ {
		res, err := http.Get(fmt.Sprintf("http://127.0.0.1:8888/?nodeType=monitor&id=%d", b.Snapshots[0].Snapshot.Version))
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
