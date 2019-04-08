/*
   Copyright 2018-2019 Banco Bilbao Vizcaya Argentaria, S.A.

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
package gossip

import (
	"bytes"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"time"

	"github.com/bbva/qed/log"
)

// Notifies string messages to external services.
// The process of sending the notifications is
// asynchronous, so a start and stop method is
// needed to activate/desactivate the process.
type Notifier interface {
	Alert(msg string) error
	Start()
	Stop()
}

type DefaultNotifierConfig struct {
	Servers      []string      `desc:"Notification service endpoint list http://ip1:port1,http://ip2:port2... "`
	QueueTimeout time.Duration `desc:"Timeout enqueuing elements on a channel"`
	DialTimeout  time.Duration `desc:"Timeout dialing the notification service"`
	ReadTimeout  time.Duration `desc:"Timeout reading the notification service response"`
}

func NewDefaultNotifierFromConfig(c *DefaultNotifierConfig) *DefaultNotifier {
	return NewDefaultNotifier(c.Servers, c.QueueTimeout, c.DialTimeout, c.ReadTimeout)
}

// Implements the default notification service
// client using an HTTP API:
//
// This notifier posts the msg contents to
// the specified servers.
type DefaultNotifier struct {
	client   *http.Client
	servers  []string
	timeout  *time.Ticker
	alertsCh chan string
	quitCh   chan bool
}

// Returns a new default notififier client configured
// to post messages to the servers provided.
// To use the default timeouts of 200ms set them to 0:
//   queueTimeout is the time to wait for the queue to accept a new message
//   dialTimeout is the time to wait for dial to the notifications server
//   readTimeout is the time to wait for the notifications server response
func NewDefaultNotifier(servers []string, queueTimeout, dialTimeout, readTimeout time.Duration) *DefaultNotifier {
	d := DefaultNotifier{
		alertsCh: make(chan string, 10),
		quitCh:   make(chan bool),
		servers:  servers,
	}
	if queueTimeout == 0 {
		queueTimeout = 200 * time.Millisecond
	}
	if dialTimeout == 0 {
		dialTimeout = 200 * time.Millisecond
	}
	if readTimeout == 0 {
		readTimeout = 200 * time.Millisecond
	}
	d.timeout = time.NewTicker(queueTimeout)
	d.client = &http.Client{
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				// timeout calling the server
				conn, err := net.DialTimeout(netw, addr, dialTimeout)
				if err != nil {
					return nil, err
				}
				// timeout reading from the connection
				conn.SetDeadline(time.Now().Add(readTimeout))
				return conn, nil
			},
		}}

	return &d
}

// Alert enqueue a message into the alerts
// queue to be sent
func (n *DefaultNotifier) Alert(msg string) error {
	for {
		select {
		case <-n.timeout.C:
			return ChTimedOut
		case n.alertsCh <- msg:
			return nil
		}
	}
}

// Starts a process which send notifications
//  to a random url selected from the configuration list of urls.
//
func (n *DefaultNotifier) Start() {
	go func() {
		for {
			select {
			case msg := <-n.alertsCh:
				i := len(n.servers)
				server := n.servers[0]
				if i > 1 {
					server = n.servers[rand.Intn(i)]
				}

				resp, err := n.client.Post(server, "application/json", bytes.NewBufferString(msg))
				if err != nil {
					log.Infof("Agent had an error sending the alert %v because %v ", msg, err)
					continue
				}
				defer resp.Body.Close()
				_, err = io.Copy(ioutil.Discard, resp.Body)
				if err != nil {
					log.Infof("Agent had the error %v when reading the response from the alert %v ", err, msg)
				}
			case <-n.quitCh:
				return
			}
		}
	}()
}

// Makes the notifications process to end
func (n *DefaultNotifier) Stop() {
	close(n.quitCh)
}
