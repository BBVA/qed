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

package monitor

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/bbva/qed/client"
	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/log"
	"github.com/bbva/qed/protocol"
)

type Config struct {
	QedUrls               []string
	PubUrls               []string
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
	conf   Config

	taskCh          chan QueryTask
	quitCh          chan bool
	executionTicker *time.Ticker
}

func NewMonitor(conf Config) (*Monitor, error) {
	client := client.NewHttpClient(conf.QedUrls[0], conf.APIKey)
	monitor := Monitor{
		client: client,
		conf:   conf,
		taskCh: make(chan QueryTask, 100),
		quitCh: make(chan bool),
	}

	monitor.executionTicker = time.NewTicker(conf.TaskExecutionInterval)
	go monitor.runTaskDispatcher()

	return &monitor, nil
}

type QueryTask struct {
	Start, End                 uint64
	StartSnapshot, EndSnapshot protocol.Snapshot
}

func (m Monitor) Process(b protocol.BatchSnapshots) {
	first := b.Snapshots[0].Snapshot
	last := b.Snapshots[len(b.Snapshots)-1].Snapshot

	log.Debugf("Processing batch from versions %d to %d", first.Version, last.Version)

	task := QueryTask{
		Start:         first.Version,
		End:           last.Version,
		StartSnapshot: *first,
		EndSnapshot:   *last,
	}

	m.taskCh <- task
}

func (m Monitor) runTaskDispatcher() {
	for {
		select {
		case <-m.executionTicker.C:
			log.Debug("Dispatching tasks...")
			go m.dispatchTasks()
		case <-m.quitCh:
			m.executionTicker.Stop()
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

func (m Monitor) dispatchTasks() {
	count := 0
	var task QueryTask
	var ok bool
	defer log.Debugf("%d tasks dispatched", count)
	for {
		select {
		case task, ok = <-m.taskCh:
			if !ok {
				return
			}
			go m.executeTask(task)
			count++
		default:
			return
		}
		if count >= m.conf.MaxInFlightTasks {
			return
		}
	}
}

func (m Monitor) sendAlert(msg string) {
	resp, err := http.Post(fmt.Sprintf("%s/alert", m.conf.PubUrls[0]), "application/json",
		bytes.NewBufferString(msg))
	if err != nil {
		log.Infof("Error saving batch in alertStore: %v", err)
	}
	defer resp.Body.Close()
	_, err = io.Copy(ioutil.Discard, resp.Body)
	if err != nil {
		log.Infof("Error getting response from alertStore saving a batch: %v", err)
	}
}

func (m Monitor) executeTask(task QueryTask) {
	log.Debug("Executing task: %+v", task)
	resp, err := m.client.Incremental(task.Start, task.End)
	if err != nil {
		// TODO: retry
		m.sendAlert(fmt.Sprintf("Unable to verify incremental proof from %d to %d", task.Start, task.End))
		log.Infof("Unable to verify incremental proof from %d to %d", task.Start, task.End)
		return
	}
	ok := m.client.VerifyIncremental(resp, &task.StartSnapshot, &task.EndSnapshot, hashing.NewSha256Hasher())
	if !ok {
		m.sendAlert(fmt.Sprintf("Unable to verify incremental proof from %d to %d",
			task.StartSnapshot.Version, task.EndSnapshot.Version))
		log.Infof("Unable to verify incremental proof from %d to %d",
			task.StartSnapshot.Version, task.EndSnapshot.Version)
	}
	log.Debugf("Consistency between versions %d and %d: %v\n", task.Start, task.End, ok)
}
