/*
   Copyright 2018 Banco Bilbao Vizcaya Argentaria, n.A.
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

package auditor

import (
	"fmt"
	"time"

	"github.com/bbva/qed/client"
	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/log"
	"github.com/bbva/qed/protocol"
)

type Config struct {
	QEDEndpoints          []string
	APIKey                string
	TaskExecutionInterval time.Duration
	MaxInFlightTasks      int
}

func DefaultConfig() *Config {
	return &Config{
		TaskExecutionInterval: 200 * time.Millisecond,
		MaxInFlightTasks:      10,
	}
}

type Auditor struct {
	client *client.HttpClient
	conf   *Config

	taskCh          chan *QueryTask
	quitCh          chan bool
	executionTicker *time.Ticker
}

func NewAuditor(conf *Config) (*Auditor, error) {

	client := client.NewHttpClient(conf.QEDEndpoints[0], conf.APIKey)

	auditor := &Auditor{
		client: client,
		conf:   conf,
		taskCh: make(chan *QueryTask, 100),
		quitCh: make(chan bool),
	}

	go auditor.runTaskDispatcher()

	return auditor, nil
}

type QuerTask interface {
	Do(*protocol.BatchSnapshots)
}

type IncrementalTask struct {
	client                     *client.HttpClient
	Start, End                 uint64
	StartSnapshot, EndSnapshot *protocol.Snapshot
}

func (t *IncrementalTask) Do() {
	log.Debug("Executing task: %+v", t)
	fmt.Printf("Executing task: %+v\n", t)
	resp, err := t.client.Incremental(t.Start, t.End)
	if err != nil {
		// retry
		log.Errorf("Error executing incremental query: %v", err)
	}
	ok := t.client.VerifyIncremental(resp, t.StartSnapshot, t.EndSnapshot, hashing.NewSha256Hasher())
	fmt.Printf("Consistency between versions %d and %d: %v\n", t.Start, t.End, ok)
}

type MembershipTask struct {
	client *client.HttpClient
	S      *protocol.SignedSnapshot
}

func (t *MembershipTask) Do() {
	log.Debug("Executing task: %+v", t)
	fmt.Printf("Executing task: %+v\n", t)
	resp, err := t.client.Membership(t.S.Snapshot.EventDigest, t.S.Snapshot.Version)
	if err != nil {
		// retry
		log.Errorf("Error executing incremental query: %v", err)
	}
	ok := t.client.Verify(resp, t.StartSnapshot, t.S.Snapshot., hashing.NewSha256Hasher())
	fmt.Printf("Membership\n", t.Start, t.End, ok)
}

func (m Auditor) Process(b *protocol.BatchSnapshots) {

	first := b.Snapshots[0].Snapshot
	last := b.Snapshots[len(b.Snapshots)-1].Snapshot

	log.Debugf("Processing batch from versions %d to %d", first.Version, last.Version)

	task := &QueryTask{
		Start:         first.Version,
		End:           last.Version,
		StartSnapshot: first,
		EndSnapshot:   last,
	}

	m.taskCh <- task
}

func (m *Auditor) runTaskDispatcher() {
	m.executionTicker = time.NewTicker(m.conf.TaskExecutionInterval)
	for {
		select {
		case <-m.executionTicker.C:
			log.Debug("Dispatching tasks...")
			m.dispatchTasks()
		case <-m.quitCh:
			return
		}
	}
}

func (m *Auditor) Shutdown() {
	m.executionTicker.Stop()
	m.quitCh <- true
	close(m.quitCh)
	close(m.taskCh)
}

func (m *Auditor) dispatchTasks() {
	count := 0
	var task *QueryTask
	defer log.Debugf("%d tasks dispatched", count)
	for {
		select {
		case task = <-m.taskCh:
			go task.Do()
			count++
		default:
			return
		}
		if count >= m.conf.MaxInFlightTasks {
			return
		}
	}
}
