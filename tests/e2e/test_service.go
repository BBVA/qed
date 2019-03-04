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

package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/bbva/qed/api/metricshttp"
	"github.com/bbva/qed/log"
	"github.com/bbva/qed/protocol"
)

var (

	// Prometheus

	QedStoreInstancesCount = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "qed_store_instances_count",
			Help: "Number of store services running.",
		},
	)

	QedStoreBatchesStoredTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "qed_store_batches_stored_total",
			Help: "Number of batches received (POST from publishers).",
		},
	)

	QedStoreSnapshotsRetrievedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "qed_store_snapshots_retrieved_total",
			Help: "Number of snapshots retrieved (GET from auditors).",
		},
	)

	QedStoreAlertsGeneratedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "qed_store_alerts_generated_total",
			Help: "Number of alerts generated.",
		},
	)

	metricsList = []prometheus.Collector{
		QedStoreInstancesCount,
		QedStoreBatchesStoredTotal,
		QedStoreSnapshotsRetrievedTotal,
		QedStoreAlertsGeneratedTotal,
	}

	registerMetrics sync.Once
)

// Register all metrics.
func Register(r *prometheus.Registry) {
	registerMetrics.Do(
		func() {
			for _, metric := range metricsList {
				r.MustRegister(metric)
			}
		},
	)
}

type alertStore struct {
	sync.Mutex
	d []string
}

func (a *alertStore) Append(msg string) {
	a.Lock()
	defer a.Unlock()
	a.d = append(a.d, msg)
}

func (a *alertStore) GetAll() []string {
	a.Lock()
	defer a.Unlock()
	n := make([]string, len(a.d))
	copy(n, a.d)
	return n
}

type snapStore struct {
	sync.Mutex
	d map[uint64]*protocol.SignedSnapshot
}

func (s *snapStore) Put(b *protocol.BatchSnapshots) {
	s.Lock()
	defer s.Unlock()

	for _, snap := range b.Snapshots {
		s.d[snap.Snapshot.Version] = snap
	}
}

func (s *snapStore) Get(version uint64) (v *protocol.SignedSnapshot, ok bool) {
	s.Lock()
	defer s.Unlock()
	v, ok = s.d[version]
	return v, ok
}

const (
	STAT int = iota
	SNAP
	ALERT
	RPS
)

type statStore struct {
	sync.Mutex
	count [4]uint64
	batch map[string][]int
}

func (s *statStore) Add(key string, index, v int) {
	s.Lock()
	defer s.Unlock()
	if s.batch[key] == nil {
		s.batch[key] = make([]int, 10)
	}
	s.batch[key][index] += v
}

func (s *statStore) Get(key string, index int) int {
	s.Lock()
	defer s.Unlock()
	return s.batch[key][index]
}

func (s *statStore) Print() {
	s.Lock()
	defer s.Unlock()
	b, err := json.Marshal(s.batch)
	//b, err := json.MarshalIndent(s.batch, "", "  ")
	if err == nil {
		fmt.Println(string(b))
	}
}

func (s *Service) statHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		key := q.Get("key")
		index, _ := strconv.Atoi(q.Get("index"))
		s.stats.Add(key, index, 1)
		atomic.AddUint64(&s.stats.count[STAT], 1)
	}
}

func (s *Service) postBatchHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		atomic.AddUint64(&s.stats.count[RPS], 1)
		atomic.AddUint64(&s.stats.count[SNAP], 1)
		if r.Method == "POST" {
			QedStoreBatchesStoredTotal.Inc()
			// Decode batch to get signed snapshots and batch version.
			var b protocol.BatchSnapshots
			buf, err := ioutil.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			err = b.Decode(buf)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			s.snaps.Put(&b)
			return
		}
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

func (s *Service) getSnapshotHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		atomic.AddUint64(&s.stats.count[RPS], 1)
		atomic.AddUint64(&s.stats.count[SNAP], 1)
		if r.Method == "GET" {
			QedStoreSnapshotsRetrievedTotal.Inc()
			q := r.URL.Query()
			version, err := strconv.ParseInt(q.Get("v"), 10, 64)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			b, ok := s.snaps.Get(uint64(version))
			if !ok {
				http.Error(w, fmt.Sprintf("Version not found: %v", version), http.StatusUnprocessableEntity)
				return
			}
			buf, err := b.Encode()
			if err != nil {
				fmt.Printf("ERROR: %v", err)
			}

			_, err = w.Write(buf)
			if err != nil {
				fmt.Printf("ERROR: %v", err)
			}
			return
		}
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

func (s *Service) alertHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		atomic.AddUint64(&s.stats.count[RPS], 1)
		atomic.AddUint64(&s.stats.count[ALERT], 1)
		if r.Method == "GET" {
			b, err := json.Marshal(s.alerts.GetAll())
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			_, err = w.Write(b)
			if err != nil {
				fmt.Printf("ERROR: %v", err)
			}
			return
		} else if r.Method == "POST" {
			QedStoreAlertsGeneratedTotal.Inc()

			buf, err := ioutil.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			s.alerts.Append(string(buf))
			return
		}
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

type Service struct {
	snaps  *snapStore
	alerts *alertStore
	stats  *statStore

	metricsServer      *http.Server
	prometheusRegistry *prometheus.Registry
	httpServer         *http.Server

	quitCh chan bool
}

func NewService() *Service {
	var snaps snapStore
	var alerts alertStore
	var stats statStore
	snaps.d = make(map[uint64]*protocol.SignedSnapshot)
	stats.batch = make(map[string][]int)
	alerts.d = make([]string, 0)

	return &Service{
		snaps:  &snaps,
		alerts: &alerts,
		stats:  &stats,
		quitCh: make(chan bool),
	}
}

func (s *Service) Start(foreground bool) {
	// Metrics server.
	r := prometheus.NewRegistry()
	Register(r)
	s.prometheusRegistry = r
	metricsMux := metricshttp.NewMetricsHTTP(r)
	s.metricsServer = &http.Server{Addr: ":18888", Handler: metricsMux}

	QedStoreInstancesCount.Inc()

	go func() {
		log.Debugf("	* Starting metrics HTTP server ")
		if err := s.metricsServer.ListenAndServe(); err != http.ErrServerClosed {
			log.Errorf("Can't start metrics HTTP server: %s", err)
		}
	}()

	// Snapshot/alert store server.
	router := http.NewServeMux()
	router.HandleFunc("/stat", s.statHandler())
	router.HandleFunc("/batch", s.postBatchHandler())
	router.HandleFunc("/snapshot", s.getSnapshotHandler())
	router.HandleFunc("/alert", s.alertHandler())

	s.httpServer = &http.Server{Addr: ":8888", Handler: router}
	fmt.Println("Starting test service...")
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		for {
			select {
			case <-ticker.C:
				c := atomic.LoadUint64(&s.stats.count[RPS])
				log.Debugf("Request per second: ", c)
				log.Debugf("Counters ", s.stats.count)
				atomic.StoreUint64(&s.stats.count[RPS], 0)
			case <-s.quitCh:
				log.Debugf("\nShutting down the server...")
				_ = s.httpServer.Shutdown(context.Background())
				return
			}
		}
	}()

	if foreground {
		if err := s.httpServer.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal(err)
		}
	} else {
		go (func() {
			if err := s.httpServer.ListenAndServe(); err != http.ErrServerClosed {
				log.Fatal(err)
			}
		})()
	}
}

func (s *Service) Shutdown() {
	// Metrics
	QedStoreInstancesCount.Dec()

	log.Debugf("Metrics enabled: stopping server...")
	if err := s.metricsServer.Shutdown(context.Background()); err != nil { // TODO include timeout instead nil
		log.Error(err)
	}
	log.Debugf("Done.\n")

	s.quitCh <- true
	close(s.quitCh)
}
