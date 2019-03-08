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
	"crypto/tls"
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
	"github.com/pkg/errors"

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

	metrics.QedAuditorInstancesCount.Inc()

	// QED client
	transport := http.DefaultTransport.(*http.Transport)
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: false}
	httpClient := http.DefaultClient
	httpClient.Transport = transport
	client, err := client.NewHTTPClient(
		client.SetHttpClient(httpClient),
		client.SetURLs(conf.QEDUrls[0], conf.QEDUrls[1:]...),
		client.SetAPIKey(conf.APIKey),
	)
	if err != nil {
		return nil, errors.Wrap(err, "Cannot start http client: ")
	}

	auditor := Auditor{
		qed:    client,
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

func (a Auditor) Process(b protocol.BatchSnapshots) {
	// Metrics
	metrics.QedAuditorBatchesReceivedTotal.Inc()
	timer := prometheus.NewTimer(metrics.QedAuditorBatchesProcessSeconds)
	defer timer.ObserveDuration()

	task := &MembershipTask{
		qed:     a.qed,
		pubUrl:  a.conf.PubUrls[0],
		taskCh:  a.taskCh,
		retries: 2,
		s:       *b.Snapshots[0],
	}

	a.taskCh <- task
}

func (a *Auditor) Shutdown() {
	// Metrics
	metrics.QedAuditorInstancesCount.Dec()

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
	qed     *client.HTTPClient
	pubUrl  string
	taskCh  chan Task
	retries int
	s       protocol.SignedSnapshot
}

func (t *MembershipTask) Do() {

	proof, err := t.qed.MembershipDigest(t.s.Snapshot.EventDigest, t.s.Snapshot.Version)
	if err != nil {
		// TODO: retry
		log.Infof("Unable to get membership proof from QED server: %s", err.Error())

		switch fmt.Sprintf("%T", err) {
		case "*errors.errorString":
			t.sendAlert(fmt.Sprintf("Unable to get membership proof from QED server: %s", err.Error()))
		default:
			metrics.QedAuditorGetMembershipProofErrTotal.Inc()
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
