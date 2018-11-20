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
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type stats struct {
	sync.Mutex
	batch map[string][]int
}

func (s *stats) Add(nodeType string, id, v int) {
	s.Lock()
	defer s.Unlock()
	if s.batch[nodeType] == nil {
		s.batch[nodeType] = make([]int, 10)
	}
	s.batch[nodeType][id] += v
}

func (s stats) Get(nodeType string, id int) int {
	s.Lock()
	defer s.Unlock()
	return s.batch[nodeType][id]
}

func (s stats) Print() {
	s.Lock()
	defer s.Unlock()
	b, err := json.MarshalIndent(s.batch, "", "  ")
	if err == nil {
		fmt.Println(string(b))
	}
	return
}

var count uint64
var s stats

func handler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	nodeType := q.Get("nodeType")
	id, _ := strconv.Atoi(q.Get("id"))

	s.Add(nodeType, id, 1)

	atomic.AddUint64(&count, 1)
}

func main() {

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
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe("127.0.0.1:8888", nil))
}
