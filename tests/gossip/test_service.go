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
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bbva/qed/gossip/member"
	"github.com/go-redis/redis"
	"github.com/hashicorp/go-msgpack/codec"
)

type stats struct {
	sync.Mutex
	batch map[string][]int
}

type Digest []byte

type Snapshot struct {
	HistoryDigest Digest
	HyperDigest   Digest
	Version       uint64
	EventDigest   Digest
}

type SignedSnapshot struct {
	Snapshot  *Snapshot
	Signature []byte
}

type BatchSnapshots struct {
	Snapshots []*SignedSnapshot
	TTL       int
	From      *member.Peer
}

type RedisCli struct {
	rcli *redis.Client
}

func NewRedisClient() *RedisCli {
	c := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	pong, err := c.Ping().Result()
	if err == nil {
		fmt.Printf("%s: Redis successfully connected in %s \n", pong, c.Options().Addr)
	}
	return &RedisCli{rcli: c}
}

func (c *RedisCli) QueueCommands(key string, value []byte) {
	err := c.rcli.Pipeline().Set(key, value, 0).Err()
	if err != nil {
		panic(err)
	}
}

func (c *RedisCli) Execute() {
	// err := c.rcli.Set(key, value, 0).Err()
	_, err := c.rcli.Pipeline().Exec()
	if err != nil {
		panic(err)
	}
}

// TODO: SeMaaS

func (s *stats) Add(nodeType string, id, v int) {
	s.Lock()
	defer s.Unlock()
	if s.batch[nodeType] == nil {
		s.batch[nodeType] = make([]int, 10)
	}
	s.batch[nodeType][id] += v
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

func handler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	nodeType := q.Get("nodeType")
	id, _ := strconv.Atoi(q.Get("id"))

	s.Add(nodeType, id, 1)

	atomic.AddUint64(&count, 1)
}

func PublishHandler(w http.ResponseWriter, r *http.Request, client *RedisCli) {
	if r.Method == "POST" {
		// Decode batch to get signed snapshots and batch version.
		var b BatchSnapshots
		err := json.NewDecoder(r.Body).Decode(&b)
		if err != nil {
			fmt.Println("Error unmarshalling: ", err)
		}

		// Encode each signed snapshot for sending it to DB.
		var buf bytes.Buffer
		encoder := codec.NewEncoder(&buf, &codec.MsgpackHandle{})

		for i, s := range b.Snapshots {
			key := strconv.FormatUint(s.Snapshot.Version, 10)

			_ = encoder.Encode(s)

			client.QueueCommands(key, buf.Bytes())
			if i%len(b.Snapshots) == 0 {
				go client.Execute()
			}
		}

	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

func main() {
	s.batch = make(map[string][]int, 0)
	client := NewRedisClient()

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
	http.HandleFunc("/publish", func(w http.ResponseWriter, r *http.Request) { PublishHandler(w, r, client) })
	log.Fatal(http.ListenAndServe("127.0.0.1:8888", nil))
}
