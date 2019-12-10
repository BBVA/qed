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

package notifierstore

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/coocood/freecache"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/bbva/qed/api/metricshttp"
	"github.com/bbva/qed/log"
	"github.com/bbva/qed/protocol"
	"github.com/bbva/qed/util"
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

	QedStoreEventsStoredTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "qed_store_events_stored_total",
			Help: "Number of events stored.",
		},
	)
	QedStorePutRequest = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "qed_store_put_requests",
			Help: "Number of current Put requests.",
		},
	)
	QedStoreGetRequest = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "qed_store_get_requests",
			Help: "Number of current Get requests.",
		},
	)

	metricsList = []prometheus.Collector{
		QedStoreInstancesCount,
		QedStoreBatchesStoredTotal,
		QedStoreSnapshotsRetrievedTotal,
		QedStoreAlertsGeneratedTotal,
		QedStoreEventsStoredTotal,
		QedStoreGetRequest,
		QedStorePutRequest,
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

func newAlertStore() *alertStore {
	return &alertStore{d: make([]string, 0)}
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
	data  *freecache.Cache
	count *uint64
	log   log.Logger
}

func newSnapStore() *snapStore {
	return &snapStore{
		data:  freecache.NewCache(2 << 30),
		count: new(uint64),
		log:   log.L(),
	}
}

func newSnapStoreWithLogger(l log.Logger) *snapStore {
	return &snapStore{
		data:  freecache.NewCache(2 << 30),
		count: new(uint64),
		log:   l,
	}
}

func (s *snapStore) Put(b *protocol.BatchSnapshots) error {
	for _, snap := range b.Snapshots {
		atomic.AddUint64(s.count, 1)
		key := util.Uint64AsBytes(snap.Snapshot.Version)
		val, err := snap.Encode()
		if err != nil {
			return err
		}
		s.data.Set(key, val, 0)
		s.log.Debugf("snapStore(): saved snapshot with version ", snap.Snapshot.Version)
		QedStoreEventsStoredTotal.Inc()
	}
	return nil
}

func (s *snapStore) Get(version uint64) (*protocol.SignedSnapshot, error) {
	var snap protocol.SignedSnapshot
	key := util.Uint64AsBytes(version)
	val, err := s.data.Get(key)
	if err != nil {
		return nil, err
	}
	err = snap.Decode(val)
	if err != nil {
		return nil, err
	}
	return &snap, nil
}

func (s *snapStore) Count() uint64 {
	return *s.count
}

type Service struct {
	snaps  *snapStore
	alerts *alertStore

	metricsServer      *http.Server
	prometheusRegistry *prometheus.Registry
	httpServer         *http.Server

	log log.Logger

	quitCh chan bool
}

func NewService() *Service {
	return &Service{
		snaps:  newSnapStore(),
		alerts: newAlertStore(),
		quitCh: make(chan bool),
		log:    log.L(),
	}
}

func NewServiceWithLogger(l log.Logger) *Service {
	return &Service{
		snaps:  newSnapStoreWithLogger(l.Named("snapshot-store")),
		alerts: newAlertStore(),
		quitCh: make(chan bool),
		log:    l,
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
		s.log.Debugf("	* Starting metrics HTTP server ")
		if err := s.metricsServer.ListenAndServe(); err != http.ErrServerClosed {
			s.log.Errorf("Can't start metrics HTTP server: %s", err)
		}
	}()

	// Snapshot/alert store server.
	router := http.NewServeMux()
	router.HandleFunc("/batch", s.postBatchHandler())
	router.HandleFunc("/count", s.getSnapshotCountHandler())
	router.HandleFunc("/snapshot", s.getSnapshotHandler())
	router.HandleFunc("/alert", s.alertHandler())

	s.httpServer = newHttpServer(":8888", router, s.log)

	s.log.Info("Starting test service...")

	go func() {
		for {
			select {
			case <-s.quitCh:
				s.log.Debug("\nShutting down the server...")
				_ = s.httpServer.Shutdown(context.Background())
				return
			}
		}
	}()

	if foreground {
		if err := s.httpServer.ListenAndServe(); err != http.ErrServerClosed {
			s.log.Fatalf("Cannot open http server: %v", err)
		}
	} else {
		go (func() {
			if err := s.httpServer.ListenAndServe(); err != http.ErrServerClosed {
				s.log.Fatalf("Cannot open http server: %v", err)
			}
		})()
	}
}

func (s *Service) Shutdown() {

	// Metrics
	QedStoreInstancesCount.Dec()

	s.log.Debug("Metrics enabled: stopping server...")
	if err := s.metricsServer.Shutdown(context.Background()); err != nil { // TODO include timeout instead nil
		s.log.Fatalf("An error ocurred when shutting down metrics server: %v", err)
	}
	s.log.Debug("Done.\n")

	s.quitCh <- true
	close(s.quitCh)
}

func (s *Service) postBatchHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		QedStorePutRequest.Inc()
		defer QedStorePutRequest.Dec()
		if r.Method == "POST" {
			QedStoreBatchesStoredTotal.Inc()
			// Decode batch to get signed snapshots and batch version.
			var b protocol.BatchSnapshots
			buf, err := ioutil.ReadAll(r.Body)
			if err != nil {
				s.log.Infof("test_service(POST /batch): %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			err = b.Decode(buf)
			if err != nil {
				s.log.Infof("test_service(POST /batch): %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if len(b.Snapshots) < 1 {
				s.log.Infof("test_service(POST /batch): Empty batch recevied!")
				http.Error(w, "Empty batch recevied!", http.StatusInternalServerError)
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
		QedStoreGetRequest.Inc()
		defer QedStoreGetRequest.Dec()
		if r.Method == "GET" {
			QedStoreSnapshotsRetrievedTotal.Inc()
			q := r.URL.Query()
			version, err := strconv.ParseInt(q.Get("v"), 10, 64)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			b, err := s.snaps.Get(uint64(version))
			if err != nil {
				s.log.Infof("test_service(GET /snapshots?v=%d): not found because %v", version, err)
				http.Error(w, fmt.Sprintf("Version not found: %v", version), http.StatusNotFound)
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

func (s *Service) getSnapshotCountHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			QedStoreSnapshotsRetrievedTotal.Inc()
			count := s.snaps.Count()

			_, err := w.Write([]byte(fmt.Sprintf("%d", count)))
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
		if r.Method == "GET" {
			b, err := json.Marshal(s.alerts.GetAll())
			if err != nil {
				s.log.Infof("test_service(GET /alert) error decoding alerts because ", err)
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
				s.log.Infof("test_service(GET /alert) error reading alerts because ", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			s.alerts.Append(string(buf))
			return
		}
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}
