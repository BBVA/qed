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

package server

import (
	"net/http"

	"github.com/bbva/qed/publisher/api"
)

type Server struct {
	nodeID   string // unique name for node. If not set, fallback to hostname
	httpAddr string // HTTP server bind address

	httpServer *http.Server
}

// NewServer synthesizes a new Server based on the parameters it receives.
func NewServer(
	nodeID string,
	httpAddr string,
) (*Server, error) {

	server := &Server{
		nodeID:   nodeID,
		httpAddr: httpAddr,
	}

	// Create http endpoints
	server.httpServer = newHTTPServer(server.httpAddr)

	return server, nil
}

func newHTTPServer(endpoint string) *http.Server {
	router := api.NewApiHttp()
	return &http.Server{
		Addr:    endpoint,
		Handler: api.LogHandler(router),
	}
}
