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

package publisher

import (
	"encoding/json"

	"github.com/bbva/qed/gossip"
	"github.com/bbva/qed/log"
	"github.com/bbva/qed/protocol"
	"github.com/valyala/fasthttp"
)

type Config struct {
	Client *fasthttp.Client
	SendTo []string
}

func DefaultConfig() *Config {
	return &Config{}
}

func NewConfig(c *fasthttp.Client, to []string) *Config {
	return &Config{
		Client: c,
		SendTo: to,
	}
}

type Publisher struct {
	Agent  *gossip.Agent
	Config *Config
	quit   chan bool
}

func NewPublisher(conf *Config) *Publisher {
	return &Publisher{
		Config: conf,
	}
}

func (p Publisher) Process(b *protocol.BatchSnapshots) {
	body, err := json.Marshal(&b)
	if err != nil {
		log.Debug("\nPublisher: Error marshalling: %s", err.Error())
		return
	}

	req := fasthttp.AcquireRequest()
	// TODO: Implement send to different endpoints
	req.SetRequestURI(p.Config.SendTo[0])
	req.Header.SetMethodBytes([]byte("POST"))
	req.Header.Add("Content-Type", "application/json")
	req.SetBody(body)

	res := fasthttp.AcquireResponse()

	err = p.Config.Client.Do(req, res)
	if err != nil {
		log.Debug("\nPublisher: Error sending request to publishers: %s", err.Error())
		return
	}

	fasthttp.ReleaseRequest(req)
	fasthttp.ReleaseResponse(res)
}
