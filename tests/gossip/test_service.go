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
	batch []int
}

func (s *stats) Add(i, v int) {
	s.Lock()
	defer s.Unlock()
	s.batch[i] = s.batch[i] + v
}

func (s stats) Get(i int) int {
	s.Lock()
	defer s.Unlock()
	return s.batch[i]
}

var count uint64
var s stats

func handler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path[1:]
	i, _ := strconv.Atoi(key)

	s.Add(i, 1)

	atomic.AddUint64(&count, 1)
}

func main() {

	s.batch = make([]int, 10000)

	go func() {
		ticker := time.NewTicker(2 * time.Second)
		for {
			select {
			case <-ticker.C:
				c := atomic.LoadUint64(&count)
				fmt.Println("Reuqest per second: ", c/2)
				for i := 0; i < len(s.batch); i++ {
					if s.batch[i] == 0 {
						break
					}
					fmt.Println("Batch ", i, " visited ", s.Get(i))
				}
				atomic.StoreUint64(&count, 0)
			}
		}
	}()
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
