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
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/bbva/qed/api/metricshttp"
	"github.com/bbva/qed/client"
	"github.com/bbva/qed/gossip/metrics"
	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/log"
	"github.com/bbva/qed/protocol"

	"github.com/prometheus/client_golang/prometheus"
)

type Config struct {
	QEDUrls               []string
	PubUrls               []string
	APIKey                string
	TaskExecutionInterval time.Duration
	MaxInFlightTasks      int
	MetricsAddr           string
}

func DefaultConfig() *Config {
	return &Config{
		TaskExecutionInterval: 200 * time.Millisecond,
		MaxInFlightTasks:      10,
	}
}

type Monitor struct {
	client *client.HTTPClient
	conf   Config

	metricsServer      *http.Server
	prometheusRegistry *prometheus.Registry

	taskCh          chan QueryTask
	quitCh          chan bool
	executionTicker *time.Ticker
}

func NewMonitor(conf Config) (*Monitor, error) {
	// Metrics
	metrics.Qed_monitor_instances_count.Inc()

	monitor := Monitor{
		client: client.NewHTTPClient(client.Config{
			Cluster:  client.QEDCluster{Endpoints: conf.QEDUrls, Leader: conf.QEDUrls[0]},
			APIKey:   conf.APIKey,
			Insecure: false,
		}),
		conf:   conf,
		taskCh: make(chan QueryTask, 100),
		quitCh: make(chan bool),
	}

	r := prometheus.NewRegistry()
	metrics.Register(r)
	monitor.prometheusRegistry = r
	metricsMux := metricshttp.NewMetricsHTTP(r)

	addr := strings.Split(conf.MetricsAddr, ":")
	monitor.metricsServer = &http.Server{
		Addr:    ":1" + addr[1],
		Handler: metricsMux,
	}

	go func() {
		log.Debugf("	* Starting metrics HTTP server in addr: %s", conf.MetricsAddr)
		if err := monitor.metricsServer.ListenAndServe(); err != http.ErrServerClosed {
			log.Errorf("Can't start metrics HTTP server: %s", err)
		}
	}()

	monitor.executionTicker = time.NewTicker(conf.TaskExecutionInterval)
	go monitor.runTaskDispatcher()

	return &monitor, nil
}

type QueryTask struct {
	Start, End                 uint64
	StartSnapshot, EndSnapshot protocol.Snapshot
}

func (m Monitor) Process(b protocol.BatchSnapshots) {
	// Metrics
	metrics.Qed_monitor_batches_received_total.Inc()
	timer := prometheus.NewTimer(metrics.Qed_monitor_batches_process_seconds)
	defer timer.ObserveDuration()

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
	// Metrics
	metrics.Qed_monitor_instances_count.Dec()

	log.Debugf("Metrics enabled: stopping server...")
	if err := m.metricsServer.Shutdown(context.Background()); err != nil { // TODO include timeout instead nil
		log.Error(err)
	}
	log.Debugf("Done.\n")

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
		return
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
