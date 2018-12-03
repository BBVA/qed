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

package main

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

	"github.com/bbva/qed/protocol"
)

type kv struct {
	sync.Mutex
	d map[uint64]*protocol.SignedSnapshot
}

func (kv *kv) Put(b *protocol.BatchSnapshots) {
	kv.Lock()
	defer kv.Unlock()

	for _, s := range b.Snapshots {
		fmt.Println("Storing ", s.Snapshot.Version, " snapshot")
		kv.d[s.Snapshot.Version] = s
	}

}

func (kv *kv) Get(version uint64) (v *protocol.SignedSnapshot, ok bool) {
	kv.Lock()
	defer kv.Unlock()
	v, ok = kv.d[version]
	return v, ok
}

func (s *stats) Add(nodeType string, id, v int) {
	s.Lock()
	defer s.Unlock()
	if s.batch[nodeType] == nil {
		s.batch[nodeType] = make([]int, 10)
	}
	s.batch[nodeType][id] += v
}

type stats struct {
	sync.Mutex
	batch map[string][]int
}

func (s *stats) Get(nodeType string, id int) int {
	s.Lock()
	defer s.Unlock()
	return s.batch[nodeType][id]
}

func (s *stats) Print() {
	s.Lock()
	defer s.Unlock()
	b, err := json.Marshal(s.batch)
	//b, err := json.MarshalIndent(s.batch, "", "  ")
	if err == nil {
		fmt.Println(string(b))
	}
	return
}

var count uint64
var s stats

func statHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	nodeType := q.Get("nodeType")
	id, _ := strconv.Atoi(q.Get("id"))

	s.Add(nodeType, id, 1)

	atomic.AddUint64(&count, 1)
}

func postBatchHandler(kv *kv) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
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

			kv.Put(&b)
			return
		}
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

func getSnapshotHandler(kv *kv) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			q := r.URL.Query()
			version, err := strconv.ParseInt(q.Get("v"), 10, 64)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			b, ok := kv.Get(uint64(version))
			if !ok {
				http.Error(w, fmt.Sprintf("Version not found: %v", version), http.StatusMethodNotAllowed)
				return
			}
			buf, err := b.Encode()
			_, err = w.Write([]byte(base64.StdEncoding.EncodeToString(buf)))
			if err != nil {
				fmt.Println("ERROR: %v", err)
			}
			return
		}
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

func main() {
	var store kv

	store.d = make(map[uint64]*protocol.SignedSnapshot, 0)
	s.batch = make(map[string][]int, 0)

	go func() {
		ticker := time.NewTicker(2 * time.Second)
		for {
			select {
			case <-ticker.C:
				c := atomic.LoadUint64(&count)
				fmt.Println("Request per second: ", c/2)
				s.Print()
				atomic.StoreUint64(&count, 0)
			}
		}
	}()

	http.HandleFunc("/stat", statHandler)
	http.HandleFunc("/batch", postBatchHandler(&store))
	http.HandleFunc("/snapshot", getSnapshotHandler(&store))
	log.Fatal(http.ListenAndServe("127.0.0.1:8888", nil))
}
