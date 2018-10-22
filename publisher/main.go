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

// This binary allows client and auditor streaming commands and also manual
// event insertion and validation against a qed server.
package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/bbva/qed/publisher/api"
	"github.com/bbva/qed/publisher/gossip"
)

var (
	// mtx     sync.RWMutex
	members = flag.String("members", "", "comma seperated list of members")
	port    = flag.Int("port", 4001, "http port")

//	items   = map[string]string{}
// broadcasts *memberlist.TransmitLimitedQueue
)

func main() {

	flag.Parse()

	if err := gossip.Start(members); err != nil {
		fmt.Println(err)
	}

	api.HandleFunc("/health-check", api.HealthCheckHandler)
	http.HandleFunc("/gossip/add", api.AddHandler)
	http.HandleFunc("/gossip/del", api.DelHandler)
	http.HandleFunc("/gossip/get", api.GetHandler)
	fmt.Printf("Listening on :%d\n", *port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil); err != nil {
		fmt.Println(err)
	}

}
