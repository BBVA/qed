package auditor

import (
	"github.com/bbva/qed/gossip"
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

}
