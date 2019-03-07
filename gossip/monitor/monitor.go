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

	taskCh          chan Task
	quitCh          chan bool
	executionTicker *time.Ticker
}

func NewMonitor(conf Config) (*Monitor, error) {
	// Metrics
	metrics.QedMonitorInstancesCount.Inc()

	monitor := Monitor{
		client: client.NewHTTPClient(client.Config{
			Endpoints: conf.QEDUrls,
			APIKey:    conf.APIKey,
			Insecure:  false,
		}),
		conf:   conf,
		taskCh: make(chan Task, 100),
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

type Task interface {
	Do()
}

func (m Monitor) Process(b protocol.BatchSnapshots) {
	// Metrics
	metrics.QedMonitorBatchesReceivedTotal.Inc()
	timer := prometheus.NewTimer(metrics.QedMonitorBatchesProcessSeconds)
	defer timer.ObserveDuration()

	first := b.Snapshots[0].Snapshot
	last := b.Snapshots[len(b.Snapshots)-1].Snapshot

	log.Debugf("Processing batch from versions %d to %d", first.Version, last.Version)

	task := QueryTask{
		client:        m.client,
		pubUrl:        m.conf.PubUrls[0],
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
			go m.dispatchTasks()
		case <-m.quitCh:
			m.executionTicker.Stop()
			return
		}
	}
}

func (m *Monitor) Shutdown() {
	// Metrics
	metrics.QedMonitorInstancesCount.Dec()

	log.Debugf("Metrics enabled: stopping server...")
	// TODO include timeout instead nil
	if err := m.metricsServer.Shutdown(context.Background()); err != nil {
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
	var task Task
	var ok bool

	for {
		select {
		case task, ok = <-m.taskCh:
			if !ok {
				return
			}
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

type QueryTask struct {
	client                     *client.HTTPClient
	pubUrl                     string
	taskCh                     chan Task
	Start, End                 uint64
	StartSnapshot, EndSnapshot protocol.Snapshot
}

func (q QueryTask) sendAlert(msg string) {
	resp, err := http.Post(fmt.Sprintf("%s/alert", q.pubUrl), "application/json", bytes.NewBufferString(msg))
	if err != nil {
		log.Infof("Error saving batch in alertStore (task re-enqueued): %v", err)
		q.taskCh <- q
		return
	}
	defer resp.Body.Close()
	_, err = io.Copy(ioutil.Discard, resp.Body)
	if err != nil {
		log.Infof("Error getting response from alertStore saving a batch: %v", err)
	}
}

func (q QueryTask) Do() {
	log.Debug("Executing task: %+v", q)
	resp, err := q.client.Incremental(q.Start, q.End)
	if err != nil {
		metrics.QedMonitorGetIncrementalProofErrTotal.Inc()
		log.Infof("Unable to get incremental proof from QED server: %s", err.Error())
		return
	}
	ok := q.client.VerifyIncremental(resp, &q.StartSnapshot, &q.EndSnapshot, hashing.NewSha256Hasher())
	if !ok {
		q.sendAlert(fmt.Sprintf("Unable to verify incremental proof from %d to %d", q.StartSnapshot.Version, q.EndSnapshot.Version))
		log.Infof("Unable to verify incremental proof from %d to %d", q.StartSnapshot.Version, q.EndSnapshot.Version)
	}
	log.Debugf("Consistency between versions %d and %d: %v\n", q.Start, q.End, ok)
}
