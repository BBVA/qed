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
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/bbva/qed/api/metricshttp"
	"github.com/bbva/qed/gossip/metrics"
	"github.com/bbva/qed/log"
	"github.com/bbva/qed/protocol"
	"github.com/valyala/fasthttp"

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
	client *fasthttp.Client
	conf   Config

	metricsServer      *http.Server
	prometheusRegistry *prometheus.Registry

	taskCh          chan PublishTask
	quitCh          chan bool
	executionTicker *time.Ticker
}

func NewPublisher(conf Config) (*Publisher, error) {
	metrics.Qed_publisher_instances_count.Inc()
	publisher := Publisher{
		client: &fasthttp.Client{},
		conf:   conf,
		taskCh: make(chan PublishTask, 100),
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

type PublishTask struct {
	Batch protocol.BatchSnapshots
}

func (p *Publisher) Process(b protocol.BatchSnapshots) {
	// Metrics
	metrics.Qed_publisher_batches_received_total.Inc()
	timer := prometheus.NewTimer(metrics.Qed_publisher_batches_process_seconds)
	defer timer.ObserveDuration()

	task := &PublishTask{
		Batch: b,
	}
	p.taskCh <- *task
}

func (p Publisher) runTaskDispatcher() {
	for {
		select {
		case <-p.executionTicker.C:
			log.Debug("Dispatching tasks...")
			go p.dispatchTasks()
		case <-p.quitCh:
			p.executionTicker.Stop()
			return
		}
	}
}

func (p *Publisher) Shutdown() {
	// Metrics
	metrics.Qed_publisher_instances_count.Dec()

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
	var task PublishTask
	defer log.Debugf("%d tasks dispatched", count)
	for {
		select {
		case task = <-p.taskCh:
			go p.executeTask(task)
			count++
		default:
			return
		}
		if count >= p.conf.MaxInFlightTasks {
			return
		}
	}
}

func (p Publisher) executeTask(task PublishTask) {
	log.Debugf("Executing task: %+v", task)
	buf, err := task.Batch.Encode()
	if err != nil {
		log.Debug("Publisher: Error marshalling: %s\n", err.Error())
		return
	}
	resp, err := http.Post(fmt.Sprintf("%s/batch", p.conf.PubUrls[0]),
		"application/json", bytes.NewBuffer(buf))
	if err != nil {
		log.Infof("Error saving batch in snapStore: %v\n", err)
		return
	}
	defer resp.Body.Close()
	_, err = io.Copy(ioutil.Discard, resp.Body)
	if err != nil {
		log.Infof("Error getting response from snapStore saving a batch: %v", err)
	}
}
