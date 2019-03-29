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

package publisher

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/bbva/qed/log"
	"github.com/bbva/qed/metrics"
	"github.com/bbva/qed/protocol"
	"github.com/coocood/freecache"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	QedPublisherInstancesCount = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "qed_publisher_instances_count",
			Help: "Number of publisher agents running.",
		},
	)

	QedPublisherBatchesReceivedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "qed_publisher_batches_received_total",
			Help: "Number of batches received by publishers.",
		},
	)

	QedPublisherBatchesProcessSeconds = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Name: "qed_publisher_batches_process_seconds",
			Help: "Duration of Publisher batch processing",
		},
	)
)

type Config struct {
	PubUrls               []string
	AlertsUrls            []string
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

func NewConfig(urls []string) *Config {
	cfg := DefaultConfig()
	cfg.PubUrls = urls
	return cfg
}

type Publisher struct {
	store *http.Client
	conf  Config

	processed *freecache.Cache

	taskCh          chan Task
	quitCh          chan bool
	executionTicker *time.Ticker
}

type Task interface {
	Do()
}

func NewPublisher(conf Config) (*Publisher, error) {
	QedPublisherInstancesCount.Inc()
	publisher := Publisher{
		store:     &http.Client{},
		conf:      conf,
		processed: freecache.NewCache(1 << 20),
		taskCh:    make(chan Task, 100),
		quitCh:    make(chan bool),
	}

	publisher.executionTicker = time.NewTicker(conf.TaskExecutionInterval)
	go publisher.runTaskDispatcher()

	return &publisher, nil
}

func (p Publisher) RegisterMetrics(srv *metrics.Server) {
	metrics := []prometheus.Collector{
		QedPublisherInstancesCount,
		QedPublisherBatchesReceivedTotal,
		QedPublisherBatchesProcessSeconds,
	}
	srv.Register(metrics)
}

func (p *Publisher) Process(b *protocol.BatchSnapshots) {
	QedPublisherBatchesReceivedTotal.Inc()
	timer := prometheus.NewTimer(QedPublisherBatchesProcessSeconds)
	defer timer.ObserveDuration()
	var batch protocol.BatchSnapshots

	for _, signedSnap := range b.Snapshots {
		_, err := p.processed.Get(signedSnap.Signature)
		if err != nil {
			_ = p.processed.Set(signedSnap.Signature, []byte{0x0}, 0)
			batch.Snapshots = append(batch.Snapshots, signedSnap)
		}
	}

	if len(batch.Snapshots) < 1 {
		return
	}

	batch.From = b.From
	batch.TTL = b.TTL

	task := &PublishTask{
		store:  p.store,
		pubUrl: p.conf.PubUrls[0],
		taskCh: p.taskCh,
		batch:  &batch,
	}
	p.taskCh <- task
}

func (p Publisher) runTaskDispatcher() {
	for {
		select {
		case <-p.executionTicker.C:
			go p.dispatchTasks()
		case <-p.quitCh:
			p.executionTicker.Stop()
			return
		}
	}
}

func (p *Publisher) Shutdown() {
	QedPublisherInstancesCount.Dec()

	p.executionTicker.Stop()
	p.quitCh <- true
	close(p.quitCh)
	close(p.taskCh)
	log.Debugf("Publisher stopped.")
}

func (p Publisher) dispatchTasks() {
	count := 0
	var task Task

	for {
		select {
		case task = <-p.taskCh:
			go task.Do()
			count++
		default:
			return
		}
		if count >= p.conf.MaxInFlightTasks {
			return
		}
	}
}

type PublishTask struct {
	store  *http.Client
	pubUrl string
	batch  *protocol.BatchSnapshots
	taskCh chan Task
}

func (t PublishTask) Do() {
	log.Debugf("Publisher is going to execute task: %+v", t)
	buf, err := t.batch.Encode()
	if err != nil {
		log.Debug("Publisher had an error marshalling: %s\n", err.Error())
		return
	}
	resp, err := t.store.Post(t.pubUrl+"/batch", "application/json", bytes.NewBuffer(buf))
	if err != nil {
		log.Infof("Publisher had an error saving batch in snapStore: %v", err)
		t.taskCh <- t
		return
	}
	defer resp.Body.Close()
	_, err = io.Copy(ioutil.Discard, resp.Body)
	if err != nil {
		log.Infof("Publisher had an error getting response from snapStore saving a batch: %v", err)
	}
}
