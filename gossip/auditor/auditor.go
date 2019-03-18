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

package auditor

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

	"github.com/prometheus/client_golang/prometheus"
)

var (
	QedAuditorInstancesCount = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "qed_auditor_instances_count",
			Help: "Number of auditor agents running.",
		},
	)

	QedAuditorBatchesProcessSeconds = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Name: "qed_auditor_batches_process_seconds",
			Help: "Duration of Auditor batch processing",
		},
	)

	QedAuditorBatchesReceivedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "qed_auditor_batches_received_total",
			Help: "Number of batches received by auditors.",
		},
	)

	QedAuditorGetMembershipProofErrTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "qed_auditor_get_membership_proof_err_total",
			Help: "Number of errors trying to get membership proofs by auditors.",
		},
	)
)

type Config struct {
	QEDUrls               []string
	PubUrls               []string
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

type Auditor struct {
	qed  *client.HTTPClient
	conf Config

	taskCh          chan Task
	quitCh          chan bool
	executionTicker *time.Ticker
}

type Task interface {
	Do()
}

func NewAuditor(conf Config) (*Auditor, error) {
	QedAuditorInstancesCount.Inc()
	auditor := Auditor{
		qed: client.NewHTTPClient(client.Config{
			Endpoints: conf.QEDUrls,
			APIKey:    conf.APIKey,
			Insecure:  false,
		}),
		conf:   conf,
		taskCh: make(chan Task, 100),
		quitCh: make(chan bool),
	}

	auditor.executionTicker = time.NewTicker(conf.TaskExecutionInterval)
	go auditor.runTaskDispatcher()

	return &auditor, nil
}

func (a Auditor) RegisterMetrics(r *prometheus.Registry) {
	metrics := []prometheus.Collector{
		QedAuditorInstancesCount,
		QedAuditorBatchesProcessSeconds,
		QedAuditorBatchesReceivedTotal,
		QedAuditorGetMembershipProofErrTotal,
	}

	for _, m := range metrics {
		r.Register(m)
	}
}

func (a Auditor) runTaskDispatcher() {
	for {
		select {
		case <-a.executionTicker.C:
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

func (a Auditor) Process(b *protocol.BatchSnapshots) {
	QedAuditorBatchesReceivedTotal.Inc()
	timer := prometheus.NewTimer(QedAuditorBatchesProcessSeconds)
	defer timer.ObserveDuration()

	task := &MembershipTask{
		qed:       a.qed,
		pubUrl:    a.conf.PubUrls[0],
		alertsUrl: a.conf.AlertsUrls[0],
		taskCh:    a.taskCh,
		retries:   2,
		s:         b.Snapshots[0],
	}

	a.taskCh <- task
}

func (a *Auditor) Shutdown() {
	QedAuditorInstancesCount.Dec()
	a.executionTicker.Stop()
	a.quitCh <- true
	close(a.quitCh)
	close(a.taskCh)
	log.Debugf("Auditor stopped.")
}

type MembershipTask struct {
	qed       *client.HTTPClient
	pubUrl    string
	alertsUrl string
	taskCh    chan Task
	retries   int
	s         *protocol.SignedSnapshot
}

func (t *MembershipTask) Do() {

	proof, err := t.qed.MembershipDigest(t.s.Snapshot.EventDigest, t.s.Snapshot.Version)
	if err != nil {
		log.Infof("Auditor is unable to get membership proof from QED server: %s", err.Error())

		switch fmt.Sprintf("%T", err) {
		case "*errors.errorString":
			t.sendAlert(fmt.Sprintf("Auditor is unable to get membership proof from QED server: %s", err.Error()))
		default:
			QedAuditorGetMembershipProofErrTotal.Inc()
		}

		return
	}

	snap, err := t.getSnapshot(proof.CurrentVersion)
	if err != nil {
		log.Infof("Unable to get snapshot from storage: %v", err)
		if t.retries > 0 {
			log.Infof("Enqueue another try to grt snapshot from storage")
			t.retries -= 1
			t.taskCh <- t
		}
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
	resp, err := http.Post(t.alertsUrl+"/alert", "application/json",
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
