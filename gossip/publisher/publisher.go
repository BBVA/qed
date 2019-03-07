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

package publisher

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/bbva/qed/api/metricshttp"
	"github.com/bbva/qed/gossip/metrics"
	"github.com/bbva/qed/log"
	"github.com/bbva/qed/protocol"

	"github.com/prometheus/client_golang/prometheus"
)

type Config struct {
	PubUrls               []string
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

func NewConfig(PubUrls []string) *Config {
	cfg := DefaultConfig()
	cfg.PubUrls = PubUrls
	return cfg
}

type Publisher struct {
	store *http.Client
	conf  Config

	metricsServer      *http.Server
	prometheusRegistry *prometheus.Registry

	taskCh          chan Task
	quitCh          chan bool
	executionTicker *time.Ticker
}

func NewPublisher(conf Config) (*Publisher, error) {
	metrics.QedPublisherInstancesCount.Inc()
	publisher := Publisher{
		store:  &http.Client{},
		conf:   conf,
		taskCh: make(chan Task, 100),
		quitCh: make(chan bool),
	}

	r := prometheus.NewRegistry()
	metrics.Register(r)
	publisher.prometheusRegistry = r
	metricsMux := metricshttp.NewMetricsHTTP(r)

	addr := strings.Split(conf.MetricsAddr, ":")
	publisher.metricsServer = &http.Server{
		Addr:    ":1" + addr[1],
		Handler: metricsMux,
	}

	go func() {
		log.Debugf("	* Starting metrics HTTP server in addr: %s", conf.MetricsAddr)
		if err := publisher.metricsServer.ListenAndServe(); err != http.ErrServerClosed {
			log.Errorf("Can't start metrics HTTP server: %s", err)
		}
	}()

	publisher.executionTicker = time.NewTicker(conf.TaskExecutionInterval)
	go publisher.runTaskDispatcher()

	return &publisher, nil
}

func (p *Publisher) Process(b protocol.BatchSnapshots) {
	// Metrics
	metrics.QedPublisherBatchesReceivedTotal.Inc()
	timer := prometheus.NewTimer(metrics.QedPublisherBatchesProcessSeconds)
	defer timer.ObserveDuration()

	task := &PublishTask{
		store:  p.store,
		pubUrl: p.conf.PubUrls[0],
		taskCh: p.taskCh,
		batch:  b,
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
	// Metrics
	metrics.QedPublisherInstancesCount.Dec()

	log.Debugf("Metrics enabled: stopping server...")
	if err := p.metricsServer.Shutdown(context.Background()); err != nil { // TODO include timeout instead nil
		log.Error(err)
	}
	log.Debugf("Done.\n")

	p.executionTicker.Stop()
	p.quitCh <- true
	close(p.quitCh)
	close(p.taskCh)
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

type Task interface {
	Do()
}

type PublishTask struct {
	store  *http.Client
	pubUrl string
	batch  protocol.BatchSnapshots
	taskCh chan Task
}

func (t PublishTask) Do() {
	log.Debugf("Executing task: %+v", t)
	buf, err := t.batch.Encode()
	if err != nil {
		log.Debug("Publisher: Error marshalling: %s\n", err.Error())
		return
	}
	resp, err := t.store.Post(t.pubUrl+"/batch", "application/json", bytes.NewBuffer(buf))
	if err != nil {
		log.Infof("Error saving batch in snapStore: %v\n", err)
		t.taskCh <- t
		return
	}
	defer resp.Body.Close()
	_, err = io.Copy(ioutil.Discard, resp.Body)
	if err != nil {
		log.Infof("Error getting response from snapStore saving a batch: %v", err)
	}
}
