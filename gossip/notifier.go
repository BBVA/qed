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

	"github.com/bbva/qed/log2"
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

//SimpleNotifier configuration object used to parse
//cli options and to build the SimpleNotifier instance
type SimpleNotifierConfig struct {
	Endpoint    []string      `desc:"Notification service endpoint list http://ip1:port1/path1,http://ip2:port2/path2... "`
	QueueSize   int           `desc:"Notifications queue size"`
	DialTimeout time.Duration `desc:"Timeout dialing the notification service"`
	ReadTimeout time.Duration `desc:"Timeout reading the notification service response"`
}

// Returns the default configuration for the SimpleNotifier
func DefaultSimpleNotifierConfig() *SimpleNotifierConfig {
	return &SimpleNotifierConfig{
		QueueSize:   10,
		DialTimeout: 200 * time.Millisecond,
		ReadTimeout: 200 * time.Millisecond,
	}
}

// Returns a SimpleNotifier pointer configured with configuration c.
func NewSimpleNotifierFromConfig(c *SimpleNotifierConfig, logger log2.Logger) *SimpleNotifier {
	return NewSimpleNotifier(c.Endpoint, c.QueueSize, c.DialTimeout, c.ReadTimeout, logger)
}

// Implements the default notification service
// client using an HTTP API:
//
// This notifier posts the msg contents to
// the specified endpoint.
type SimpleNotifier struct {
	client        *http.Client
	endpoint      []string
	notifications chan string
	quitCh        chan bool
	log           log2.Logger
}

// Returns a new default notififier client configured
// to post messages to the endpoint provided.
// To use the default timeouts of 200ms set them to 0:
//   queueTimeout is the time to wait for the queue to accept a new message
//   dialTimeout is the time to wait for dial to the notifications server
//   readTimeout is the time to wait for the notifications server response
func NewSimpleNotifier(endpoint []string, size int, dialTimeout, readTimeout time.Duration, logger log2.Logger) *SimpleNotifier {

	log := logger
	if log == nil {
		log = log2.L()
	}

	d := SimpleNotifier{
		notifications: make(chan string, size),
		quitCh:        make(chan bool),
		endpoint:      endpoint,
		log:           log,
	}

	d.client = &http.Client{
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				// timeout calling the server
				conn, err := net.DialTimeout(netw, addr, dialTimeout)
				if err != nil {
					return nil, err
				}
				// timeout reading from the connection
				_ = conn.SetDeadline(time.Now().Add(readTimeout))
				return conn, nil
			},
		}}

	return &d
}

// Alert enqueue a message into the notifications
// queue to be sent. It will block if the notifications
// queue is full.
func (n *SimpleNotifier) Alert(msg string) error {
	n.notifications <- msg
	return nil
}

// Starts a process which send notifications
//  to a random url selected from the configuration list of urls.
func (n *SimpleNotifier) Start() {
	go func() {
		for {
			select {
			case msg := <-n.notifications:
				i := len(n.endpoint)
				url := n.endpoint[0]
				if i > 1 {
					url = n.endpoint[rand.Intn(i)]
				}

				resp, err := n.client.Post(url, "application/json", bytes.NewBufferString(msg))
				if err != nil {
					n.log.Infof("Agent had an error sending the alert %v because %v ", msg, err)
					continue
				}
				defer resp.Body.Close()
				_, err = io.Copy(ioutil.Discard, resp.Body)
				if err != nil {
					n.log.Infof("Agent had the error %v when reading the response from the alert %v ", err, msg)
				}
			case <-n.quitCh:
				return
			}
		}
	}()
}

// Makes the notifications process to end
func (n *SimpleNotifier) Stop() {
	close(n.quitCh)
}
