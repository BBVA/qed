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

package e2e

import (
	"fmt"
	"os"
	"os/user"
	"testing"
	"time"

	"github.com/bbva/qed/client"
	"github.com/bbva/qed/gossip"
	"github.com/bbva/qed/gossip/auditor"
	"github.com/bbva/qed/gossip/member"
	"github.com/bbva/qed/gossip/monitor"
	"github.com/bbva/qed/gossip/publisher"
	"github.com/bbva/qed/log"
	"github.com/bbva/qed/server"
	"github.com/bbva/qed/testutils/scope"
)

var apiKey, storageType, keyFile string
var cacheSize uint64

const (
	QEDUrl       = "http://127.0.0.1:8080"
	QEDGossip    = "127.0.0.1:9010"
	QEDTamperURL = "http://localhost:8081/tamper"
	StoreUrl     = "http://127.0.0.1:8888"
	APIKey       = "my-key"
)

func init() {
	apiKey = APIKey
	cacheSize = 50000
	storageType = "badger"

	usr, _ := user.Current()
	keyFile = fmt.Sprintf("%s/.ssh/id_ed25519", usr.HomeDir)

	log.SetLogger("", log.SILENT)
}

// merge function is a helper function that execute all the variadic parameters
// inside a score.TestF function
func merge(list ...scope.TestF) scope.TestF {
	return func(t *testing.T) {
		for _, elem := range list {
			elem(t)
			// time.Sleep(2 * time.Second)
		}
	}
}

func newAgent(id int, name string, role member.Type, p gossip.Processor, t *testing.T) *gossip.Agent {
	agentConf := gossip.DefaultConfig()
	agentConf.NodeName = fmt.Sprintf("%s%d", name, id)

	switch role {
	case member.Auditor:
		agentConf.BindAddr = fmt.Sprintf("127.0.0.1:910%d", id)
	case member.Monitor:
		agentConf.BindAddr = fmt.Sprintf("127.0.0.1:920%d", id)
	case member.Publisher:
		agentConf.BindAddr = fmt.Sprintf("127.0.0.1:930%d", id)
	}

	agentConf.StartJoin = []string{QEDGossip}
	agentConf.EnableCompression = true
	agentConf.AlertsUrls = []string{StoreUrl}
	agentConf.Role = role

	agent, err := gossip.NewAgent(agentConf, []gossip.Processor{p})
	if err != nil {
		t.Fatalf("Failed to start AGENT %s: %v", name, err)
	}
	_, _ = agent.Join([]string{QEDGossip})
	return agent
}

func setupAuditor(id int, t *testing.T) (scope.TestF, scope.TestF) {
	var au *auditor.Auditor
	var agent *gossip.Agent
	var err error

	before := func(t *testing.T) {
		auditorConf := auditor.DefaultConfig()
		auditorConf.QEDUrls = []string{QEDUrl}
		auditorConf.PubUrls = []string{StoreUrl}
		auditorConf.APIKey = APIKey

		au, err = auditor.NewAuditor(*auditorConf)
		if err != nil {
			t.Fatalf("Unable to create a new auditor: %v", err)
		}

		agent = newAgent(id, "auditor", member.Auditor, au, t)
	}

	after := func(t *testing.T) {
		if au != nil {
			au.Shutdown()
			_ = agent.Shutdown()
		} else {
			t.Fatalf("Unable to shutdown the auditor!")
		}
	}
	return before, after
}

func setupMonitor(id int, t *testing.T) (scope.TestF, scope.TestF) {
	var mn *monitor.Monitor
	var agent *gossip.Agent
	var err error

	before := func(t *testing.T) {
		monitorConf := monitor.DefaultConfig()
		monitorConf.QedUrls = []string{QEDUrl}
		monitorConf.PubUrls = []string{StoreUrl}
		monitorConf.APIKey = APIKey

		mn, err = monitor.NewMonitor(*monitorConf)
		if err != nil {
			t.Fatalf("Unable to create a new monitor: %v", err)
		}

		agent = newAgent(id, "monitor", member.Monitor, mn, t)
	}

	after := func(t *testing.T) {
		if mn != nil {
			mn.Shutdown()
			_ = agent.Shutdown()
		} else {
			t.Fatalf("Unable to shutdown the monitor!")
		}
	}
	return before, after
}

func setupPublisher(id int, t *testing.T) (scope.TestF, scope.TestF) {
	var pu *publisher.Publisher
	var agent *gossip.Agent
	var err error

	before := func(t *testing.T) {
		conf := publisher.DefaultConfig()
		conf.PubUrls = []string{StoreUrl}

		pu, err = publisher.NewPublisher(*conf)
		if err != nil {
			t.Fatalf("Unable to create a new publisher: %v", err)
		}

		agent = newAgent(id, "publisher", member.Publisher, pu, t)
	}

	after := func(t *testing.T) {
		if pu != nil {
			pu.Shutdown()
			_ = agent.Shutdown()
		} else {
			t.Fatalf("Unable to shutdown the publisher!")
		}
	}
	return before, after
}

func setupStore(t *testing.T) (scope.TestF, scope.TestF) {
	var s *Service
	before := func(t *testing.T) {
		s = NewService()
		s.Start()
	}

	after := func(t *testing.T) {
		s.Shutdown()
	}
	return before, after
}

func setupServer(id int, joinAddr string, t *testing.T) (scope.TestF, scope.TestF) {
	var srv *server.Server
	var err error
	path := fmt.Sprintf("/var/tmp/e2e-qed%d/", id)

	before := func(t *testing.T) {
		os.RemoveAll(path)
		_ = os.MkdirAll(path, os.FileMode(0755))

		hostname, _ := os.Hostname()
		conf := server.DefaultConfig()
		conf.NodeID = fmt.Sprintf("%s-%d", hostname, id)
		conf.HttpAddr = fmt.Sprintf("127.0.0.1:808%d", id)
		conf.RaftAddr = fmt.Sprintf("127.0.0.1:900%d", id)
		conf.MgmtAddr = fmt.Sprintf("127.0.0.1:809%d", id)
		conf.GossipAddr = fmt.Sprintf("127.0.0.1:901%d", id)
		conf.DBPath = path + "data"
		conf.RaftPath = path + "raft"
		conf.PrivateKeyPath = keyFile
		conf.EnableProfiling = true
		conf.EnableTampering = true

		fmt.Printf("Server config: %+v\n", conf)

		srv, err = server.NewServer(conf)
		if err != nil {
			t.Fatalf("Unable to create a new server: %v", err)
		}

		go (func() {
			err := srv.Start()
			if err != nil {
				t.Log(err)
			}
		})()
		time.Sleep(2 * time.Second)
	}

	after := func(t *testing.T) {
		if srv != nil {
			_ = srv.Stop()
		} else {
			t.Fatalf("Unable to shutdown the server!")
		}
	}
	return before, after
}

func endPoint(id int) string {
	return fmt.Sprintf("http://127.0.0.1:808%d", id)
}

func getClient(id int) *client.HttpClient {
	return client.NewHttpClient(endPoint(id), apiKey)
}
