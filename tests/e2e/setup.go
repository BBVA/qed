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
	"github.com/bbva/qed/server"
	"github.com/bbva/qed/testutils/scope"
)

var apiKey, storageType, listenAddr, keyFile string
var cacheSize uint64

func init() {
	apiKey = "my-key"
	cacheSize = 50000
	storageType = "badger"

	usr, _ := user.Current()
	keyFile = fmt.Sprintf("%s/.ssh/id_ed25519", usr.HomeDir)
}

func setup(id int, joinAddr string, t *testing.T) (scope.TestF, scope.TestF) {
	var srv *server.Server
	var err error
	path := fmt.Sprintf("/var/tmp/e2e-qed%d/", id)

	before := func(t *testing.T) {
		os.RemoveAll(path)
		os.MkdirAll(path, os.FileMode(0755))

		hostname, _ := os.Hostname()
		conf := server.DefaultConfig()
		conf.NodeID = fmt.Sprintf("%s-%d", hostname, id)
		conf.HttpAddr = fmt.Sprintf("127.0.0.1:850%d", id)
		conf.RaftAddr = fmt.Sprintf("127.0.0.1:830%d", id)
		conf.MgmtAddr = fmt.Sprintf("127.0.0.1:840%d", id)
		conf.GossipAddr = fmt.Sprintf("127.0.0.1:860%d", id)
		conf.DBPath = path + "data"
		conf.RaftPath = path + "raft"
		conf.PrivateKeyPath = keyFile
		conf.EnableProfiling = true
		conf.EnableTampering = true

		fmt.Printf("%+v", conf)

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
			srv.Stop()
		} else {
			t.Fatalf("Unable to shutdown the server!")
		}
	}
	return before, after
}

func endPoint(id int) string {
	return fmt.Sprintf("http://127.0.0.1:850%d", id)
}

func getClient(id int) *client.HttpClient {
	return client.NewHttpClient(endPoint(id), apiKey)
}
