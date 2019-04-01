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

package monitor

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/bbva/qed/client"
	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/log"
	"github.com/bbva/qed/metrics"
	"github.com/bbva/qed/protocol"
	"github.com/pkg/errors"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	QedMonitorInstancesCount = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "qed_monitor_instances_count",
			Help: "Number of monitor agents running.",
		},
	)

	QedMonitorBatchesReceivedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "qed_monitor_batches_received_total",
			Help: "Number of batches received by monitors.",
		},
	)

	QedMonitorBatchesProcessSeconds = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Name: "qed_monitor_batches_process_seconds",
			Help: "Duration of Monitor batch processing",
		},
	)

	QedMonitorGetIncrementalProofErrTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "qed_monitor_get_incremental_proof_err_total",
			Help: "Number of errors trying to get incremental proofs by monitors.",
		},
	)
)

type Config struct {
	QEDUrls               []string
	AlertsUrls            []string
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
	client          *client.HTTPClient
	conf            *Config
	taskCh          chan Task
	quitCh          chan bool
	executionTicker *time.Ticker
}

type Task interface {
	Do()
}

func NewMonitor(conf *Config) (*Monitor, error) {
	QedMonitorInstancesCount.Inc()
	// QED client
	transport := http.DefaultTransport.(*http.Transport)
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: false}
	httpClient := http.DefaultClient
	httpClient.Transport = transport
	qed, err := client.NewHTTPClient(
		client.SetHttpClient(httpClient),
		client.SetURLs(conf.QEDUrls[0], conf.QEDUrls[1:]...),
		client.SetAPIKey(conf.APIKey),
		client.SetReadPreference(client.Any),
	)
	if err != nil {
		return nil, errors.Wrap(err, "Cannot start http client: ")
	}
	monitor := Monitor{
		client: qed,
		conf:   conf,
		taskCh: make(chan Task, 100),
		quitCh: make(chan bool),
	}

	monitor.executionTicker = time.NewTicker(conf.TaskExecutionInterval)
	go monitor.runTaskDispatcher()

	return &monitor, nil
}

func (m Monitor) RegisterMetrics(srv *metrics.Server) {
	metrics := []prometheus.Collector{
		QedMonitorInstancesCount,
		QedMonitorBatchesReceivedTotal,
		QedMonitorBatchesProcessSeconds,
		QedMonitorGetIncrementalProofErrTotal,
	}
	srv.MustRegister(metrics...)
}

func (m Monitor) Process(b *protocol.BatchSnapshots) {
	QedMonitorBatchesReceivedTotal.Inc()
	timer := prometheus.NewTimer(QedMonitorBatchesProcessSeconds)
	defer timer.ObserveDuration()

	first := b.Snapshots[0].Snapshot
	last := b.Snapshots[len(b.Snapshots)-1].Snapshot

	log.Debugf("Monitor processing batch from versions %d to %d", first.Version, last.Version)

	task := QueryTask{
		client:        m.client,
		alertsUrl:     m.conf.AlertsUrls[0],
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
	QedMonitorInstancesCount.Dec()

	m.executionTicker.Stop()
	m.quitCh <- true
	close(m.quitCh)
	close(m.taskCh)
	log.Debugf("Monitor stopped.")
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
	alertsUrl                  string
	taskCh                     chan Task
	Start, End                 uint64
	StartSnapshot, EndSnapshot protocol.Snapshot
}

func (q QueryTask) sendAlert(msg string) {
	resp, err := http.Post(q.alertsUrl+"/alert", "application/json", bytes.NewBufferString(msg))
	if err != nil {
		log.Infof("Monitor had an error saving batch in alertStore (task re-enqueued): %v", err)
		q.taskCh <- q
		return
	}
	defer resp.Body.Close()
	_, err = io.Copy(ioutil.Discard, resp.Body)
	if err != nil {
		log.Infof("Monitor had an error from alertStore saving a batch: %v", err)
	}
}

func (q QueryTask) Do() {
	log.Debugf("Executing task: %+v", q)
	resp, err := q.client.Incremental(q.Start, q.End)
	if err != nil {
		QedMonitorGetIncrementalProofErrTotal.Inc()
		log.Infof("Monitor is unable to get incremental proof from QED server: %s", err.Error())
		return
	}
	ok := q.client.VerifyIncremental(resp, &q.StartSnapshot, &q.EndSnapshot, hashing.NewSha256Hasher())
	if !ok {
		q.sendAlert(fmt.Sprintf("Monitor is unable to verify incremental proof from %d to %d", q.StartSnapshot.Version, q.EndSnapshot.Version))
		log.Infof("Monitor is unable to verify incremental proof from %d to %d", q.StartSnapshot.Version, q.EndSnapshot.Version)
	}
	log.Debugf("Monitor verified a consistency proof between versions %d and %d: %v\n", q.Start, q.End, ok)
}
