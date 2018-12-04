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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bbva/qed/gossip"
	"github.com/bbva/qed/protocol"
)

type alertStore struct {
	sync.Mutex
	d []*gossip.Alert
}

func (a *alertStore) Append(n *gossip.Alert) {
	a.Lock()
	defer a.Unlock()
	a.d = append(a.d, n)
}

func (a *alertStore) GetAll() []*gossip.Alert {
	a.Lock()
	defer a.Unlock()
	n := make([]*gossip.Alert, len(a.d))
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
		fmt.Println("Storing ", snap.Snapshot.Version, " snapshot")
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
)

type statStore struct {
	sync.Mutex
	count [3]uint64
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
	return
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
		atomic.AddUint64(&s.stats.count[SNAP], 1)
		if r.Method == "POST" {
			// Decode batch to get signed snapshots and batch version.
			var b protocol.BatchSnapshots
			buf, err := ioutil.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			sDec, err := base64.StdEncoding.DecodeString(string(buf))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			err = b.Decode(sDec)
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
		atomic.AddUint64(&s.stats.count[SNAP], 1)
		if r.Method == "GET" {
			q := r.URL.Query()
			version, err := strconv.ParseInt(q.Get("v"), 10, 64)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			b, ok := s.snaps.Get(uint64(version))
			if !ok {
				http.Error(w, fmt.Sprintf("Version not found: %v", version), http.StatusMethodNotAllowed)
				return
			}
			buf, err := b.Encode()
			_, err = w.Write([]byte(base64.StdEncoding.EncodeToString(buf)))
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
			var b gossip.Alert
			buf, err := ioutil.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			sDec, err := base64.StdEncoding.DecodeString(string(buf))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			err = b.Decode(sDec)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			s.alerts.Append(&b)
			return
		}
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

type Service struct {
	snaps  *snapStore
	alerts *alertStore
	stats  *statStore
}

func NewService() *Service {
	var snaps snapStore
	var alerts alertStore
	var stats statStore
	snaps.d = make(map[uint64]*protocol.SignedSnapshot, 0)
	stats.batch = make(map[string][]int, 0)
	alerts.d = make([]*gossip.Alert, 0)
	return &Service{
		snaps:  &snaps,
		alerts: &alerts,
		stats:  &stats,
	}
}

func (s *Service) Start() {
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		for {
			select {
			case <-ticker.C:
				c := atomic.LoadUint64(&s.stats.count[STAT])
				fmt.Println("Request per second: ", c/2)
				atomic.StoreUint64(&s.stats.count[STAT], 0)
			}
		}
	}()

	http.HandleFunc("/stat", s.statHandler())
	http.HandleFunc("/batch", s.postBatchHandler())
	http.HandleFunc("/snapshot", s.getSnapshotHandler())
	http.HandleFunc("/alert", s.alertHandler())
	log.Fatal(http.ListenAndServe("127.0.0.1:8888", nil))
}
