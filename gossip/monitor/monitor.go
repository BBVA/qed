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

package monitor

import (
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

type Monitor struct {
	client *client.HttpClient
	conf   *Config

	taskCh          chan *QueryTask
	quitCh          chan bool
	executionTicker *time.Ticker
}

func NewMonitor(conf *Config) (*Monitor, error) {

	client := client.NewHttpClient(conf.QEDEndpoints[0], conf.APIKey)

	monitor := &Monitor{
		client: client,
		conf:   conf,
		taskCh: make(chan *QueryTask, 100),
		quitCh: make(chan bool),
	}

	go monitor.runTaskDispatcher()

	return monitor, nil
}

type QueryTask struct {
	Start, End             uint64
	StartDigest, EndDigest hashing.Digest
}

func (m Monitor) Process(b *protocol.BatchSnapshots) {

	first := b.Snapshots[0].Snapshot
	last := b.Snapshots[len(b.Snapshots)-1].Snapshot

	log.Debugf("Processing batch from versions %d to %d", first.Version, last.Version)

	task := &QueryTask{
		Start:       first.Version,
		End:         last.Version,
		StartDigest: first.HistoryDigest,
		EndDigest:   last.HistoryDigest,
	}

	m.taskCh <- task
}

func (m *Monitor) runTaskDispatcher() {
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

func (m *Monitor) Shutdown() {
	m.executionTicker.Stop()
	m.quitCh <- true
	close(m.quitCh)
	close(m.taskCh)
}

func (m *Monitor) dispatchTasks() {
	count := 0
	for i := 0; i < m.conf.MaxInFlightTasks; i++ {
		task := <-m.taskCh
		go m.executeTask(task)
		count++
	}
	// var task *QueryTask

	// for {
	// 	select {
	// 	case task = <-m.taskCh:
	// 		go m.executeTask(task)
	// 		count++
	// 	default:
	// 		return
	// 	}
	// }
	log.Debugf("%d tasks dispatched")
}

func (m *Monitor) executeTask(task *QueryTask) {
	log.Debug("Executing task: %v", task)
}
