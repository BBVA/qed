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
	"flag"
	"os"
	"os/signal"
	"testing"
	"time"

	"github.com/bbva/qed/client"
	"github.com/bbva/qed/log"
	"github.com/bbva/qed/server"
)

func MainTest(m *testing.M) {
	os.Exit(RunTests(m))
}

// RunTests runs the tests in a package while gracefully handling interrupts.
func RunTests(m *testing.M) int {
	log.SetLogger("client-test", "info")
	flag.Parse()
	if !testing.Short() {
		stopServer := setupServer()
		go func() {
			// Shut down tests when interrupted (for example CTRL+C).
			sig := make(chan os.Signal, 1)
			signal.Notify(sig, os.Interrupt)
			<-sig
			select {
			default:
				stopServer()
			}
		}()
	}
	return m.Run()
}

func setupServer() func() {
	path := "/var/tmp/balloonE2E"
	clearPath(path)
	server := server.NewServer(":8079", path, "my-awesome-api-key", uint64(50000), "badger", false, true)

	go (func() {
		err := server.Run()
		if err != nil {
			log.Info(err)
		}
	})()

	// Give things a few seconds to tidy up
	time.Sleep(time.Second * 2)

	return func() {
		server.Stop()
		// Give things a few seconds to tidy up
		time.Sleep(time.Second * 2)
	}
}

func clearPath(path string) {
	os.RemoveAll(path)
	os.MkdirAll(path, os.FileMode(0755))
}

func getClient() *client.HttpClient {
	return client.NewHttpClient("http://localhost:8079", "my-awesome-api-key")
}
