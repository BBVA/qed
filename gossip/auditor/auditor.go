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

package auditor

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

type Auditor struct {
	qed  *client.HTTPClient
	conf Config

	metricsServer      *http.Server
	prometheusRegistry *prometheus.Registry

	taskCh          chan Task
	quitCh          chan bool
	executionTicker *time.Ticker
}

func NewAuditor(conf Config) (*Auditor, error) {
	metrics.Qed_auditor_instances_count.Inc()
	auditor := Auditor{
		qed: client.NewHTTPClient(client.Config{
			Endpoint: conf.QEDUrls[0],
			APIKey:   conf.APIKey,
			Insecure: false,
		}),
		conf:   conf,
		taskCh: make(chan Task, 100),
		quitCh: make(chan bool),
	}

	r := prometheus.NewRegistry()
	metrics.Register(r)
	auditor.prometheusRegistry = r
	metricsMux := metricshttp.NewMetricsHTTP(r)

	addr := strings.Split(conf.MetricsAddr, ":")
	auditor.metricsServer = &http.Server{
		Addr:    ":1" + addr[1],
		Handler: metricsMux,
	}

	go func() {
		log.Debugf("	* Starting metrics HTTP server in addr: %s", conf.MetricsAddr)
		if err := auditor.metricsServer.ListenAndServe(); err != http.ErrServerClosed {
			log.Errorf("Can't start metrics HTTP server: %s", err)
		}
	}()

	auditor.executionTicker = time.NewTicker(conf.TaskExecutionInterval)
	go auditor.runTaskDispatcher()

	return &auditor, nil
}

func (a Auditor) runTaskDispatcher() {
	for {
		select {
		case <-a.executionTicker.C:
			log.Debug("Dispatching tasks...")
			go a.dispatchTasks()
		case <-a.quitCh:
			a.executionTicker.Stop()
			return
		}
	}
}

func (a Auditor) dispatchTasks() {
	count := 0
	var task Task
	defer log.Debugf("%d tasks dispatched", count)
	for {
		select {
		case task = <-a.taskCh:
			go task.Do()
			count++
		default:
			return
		}
		if count >= a.conf.MaxInFlightTasks {
			return
		}
	}
}

func (a Auditor) Process(b protocol.BatchSnapshots) {
	// Metrics
	metrics.Qed_auditor_batches_received_total.Inc()
	timer := prometheus.NewTimer(metrics.Qed_auditor_batches_process_seconds)
	defer timer.ObserveDuration()

	task := &MembershipTask{
		qed:    a.qed,
		pubUrl: a.conf.PubUrls[0],
		taskCh: a.taskCh,
		s:      *b.Snapshots[0],
	}

	a.taskCh <- task
}

func (a *Auditor) Shutdown() {
	// Metrics
	metrics.Qed_auditor_instances_count.Dec()

	log.Debugf("Metrics enabled: stopping server...")
	if err := a.metricsServer.Shutdown(context.Background()); err != nil { // TODO include timeout instead nil
		log.Error(err)
	}
	log.Debugf("Done.\n")

	a.executionTicker.Stop()
	a.quitCh <- true
	close(a.quitCh)
	close(a.taskCh)
}

type Task interface {
	Do()
}

type MembershipTask struct {
	qed    *client.HTTPClient
	pubUrl string
	taskCh chan Task
	s      protocol.SignedSnapshot
}

func (t MembershipTask) Do() {
	proof, err := t.qed.MembershipDigest(t.s.Snapshot.EventDigest, t.s.Snapshot.Version)
	if err != nil {
		// retry
		t.sendAlert(fmt.Sprintf("Unable to verify snapshot %v", t.s.Snapshot))
		log.Infof("Error executing membership query: %v", err)
		return
	}

	snap, err := t.getSnapshot(proof.CurrentVersion)
	if err != nil {
		log.Infof("Unable to get snapshot from storage, try later: %v", err)
		t.taskCh <- t
		return
	}
	checkSnap := &protocol.Snapshot{
		HistoryDigest: t.s.Snapshot.HistoryDigest,
		HyperDigest:   snap.Snapshot.HyperDigest,
		Version:       t.s.Snapshot.Version,
		EventDigest:   t.s.Snapshot.EventDigest,
	}
	ok := t.qed.DigestVerify(proof, checkSnap, hashing.NewSha256Hasher)
	if !ok {
		t.sendAlert(fmt.Sprintf("Unable to verify snapshot %v", t.s.Snapshot))
		log.Infof("Unable to verify snapshot %v", t.s.Snapshot)
	}

	log.Infof("MembershipTask.Do(): Snapshot %v has been verified by QED", t.s.Snapshot)
}

func (t MembershipTask) getSnapshot(version uint64) (*protocol.SignedSnapshot, error) {
	resp, err := http.Get(fmt.Sprintf("%s/snapshot?v=%d", t.pubUrl, version))
	if err != nil {
		return nil, fmt.Errorf("Error getting snapshot from the store: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Error getting snapshot from the store. Status: %d", resp.StatusCode)
	}
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Infof("Error reading request body: %v", err)
	}
	var s protocol.SignedSnapshot
	err = s.Decode(buf)
	if err != nil {
		return nil, fmt.Errorf("Error decoding signed snapshot %d codec", t.s.Snapshot.Version)
	}
	return &s, nil
}

func (t MembershipTask) sendAlert(msg string) {
	resp, err := http.Post(fmt.Sprintf("%s/alert", t.pubUrl), "application/json",
		bytes.NewBufferString(msg))
	if err != nil {
		log.Infof("Error saving batch in alertStore: %v", err)
		return
	}
	defer resp.Body.Close()
	_, err = io.Copy(ioutil.Discard, resp.Body)
	if err != nil {
		log.Infof("Error reading request body: %v", err)
	}
}
