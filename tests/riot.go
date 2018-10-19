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
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type Event struct {
	Event []byte
}
type MembershipQuery struct {
	Key     []byte
	Version uint64
}
type SignedSnapshot struct {
	Snapshot  *Snapshot
	Signature []byte
}
type Snapshot struct {
	HistoryDigest []byte
	HyperDigest   []byte
	Version       uint64
	Event         []byte
}

type Config struct {
	maxGoRoutines  int
	numRequests    int
	apiKey         string
	startVersion   int
	continuous     bool
	balloonVersion uint64
	counter        float64
	req            HTTPClient
}
type HTTPClient struct {
	client             *http.Client
	method             string
	endpoint           string
	expectedStatusCode int
}

// type Config map[string]interface{}
func NewDefaultConfig() *Config {
	numRequests := 100000
	return &Config{
		maxGoRoutines:  10,
		numRequests:    numRequests,
		apiKey:         "pepe",
		startVersion:   0,
		continuous:     false,
		balloonVersion: uint64(numRequests) - 1,
		counter:        0,
		req: HTTPClient{
			client:             nil,
			method:             "POST",
			endpoint:           "http://localhost:8080",
			expectedStatusCode: 200,
		},
	}
}

type Task func(goRoutineId int, c *Config) ([]byte, error)

// func (t *Task) Timeout()event
func SpawnerOfEvil(c *Config, t Task) {
	// TODO: only one client per run MAYBE
	var wg sync.WaitGroup

	for goRoutineId := 0; goRoutineId < c.maxGoRoutines; goRoutineId++ {
		wg.Add(1)
		go func(goRoutineId int) {
			defer wg.Done()
			Attacker(goRoutineId, c, t)
		}(goRoutineId)
	}
	wg.Wait()
}

func Attacker(goRoutineId int, c *Config, f func(j int, c *Config) ([]byte, error)) {

	for eventIndex := c.startVersion + goRoutineId; eventIndex < c.startVersion+c.numRequests || c.continuous; eventIndex += c.maxGoRoutines {
		query, err := f(eventIndex, c)
		if len(query) == 0 {
			log.Fatalf("Empty query: %v", err)
		}

		req, err := http.NewRequest(c.req.method, c.req.endpoint, bytes.NewBuffer(query))
		if err != nil {
			log.Fatalf("Error preparing request: %v", err)
		}

		// Set Api-Key header
		req.Header.Set("Api-Key", c.apiKey)
		res, err := c.req.client.Do(req)
		if err != nil {
			log.Fatalf("Unable to perform request: %v", err)
		}
		defer res.Body.Close()

		if res.StatusCode != c.req.expectedStatusCode {
			log.Fatalf("Server error: %v", err)
		}
		c.counter++
		io.Copy(ioutil.Discard, res.Body)
	}
	c.counter = 0
}

func addSampleEvents(eventIndex int, c *Config) ([]byte, error) {

	return json.Marshal(
		&Event{
			[]byte(fmt.Sprintf("event %d", eventIndex)),
		},
	)
}

func queryMembership(eventIndex int, c *Config) ([]byte, error) {
	return json.Marshal(
		&MembershipQuery{
			[]byte(fmt.Sprintf("event %d", eventIndex)),
			c.balloonVersion,
		},
	)
}

func getVersion(eventTemplate string) uint64 {
	client := &http.Client{}

	buf := fmt.Sprintf(eventTemplate)

	query, err := json.Marshal(&Event{[]byte(buf)})
	if len(query) == 0 {
		log.Fatalf("Empty query: %v", err)
	}

	req, err := http.NewRequest("POST", "http://localhost:8080/events", bytes.NewBuffer(query))
	if err != nil {
		log.Fatalf("Error preparing request: %v", err)
	}

	// Set Api-Key header
	// TODO: remove pepe and pass a config var
	req.Header.Set("Api-Key", "pepe")
	res, err := client.Do(req)
	if err != nil {
		log.Fatalf("Unable to perform request: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != 201 {
		log.Fatalf("Server error: %v", err)
	}

	body, _ := ioutil.ReadAll(res.Body)

	var signedSnapshot SignedSnapshot
	json.Unmarshal(body, &signedSnapshot)
	version := signedSnapshot.Snapshot.Version

	io.Copy(ioutil.Discard, res.Body)

	return version
}

func summary(message string, numRequestsf, elapsed float64, c *Config) {
	fmt.Printf(
		"%s throughput: %.0f req/s: (%v reqs in %.3f seconds) | Concurrency: %d\n",
		message,
		numRequestsf/elapsed,
		c.numRequests,
		elapsed,
		c.maxGoRoutines,
	)
}

func summaryPerDuration(message string, numRequestsf, elapsed float64, c *Config) {
	fmt.Printf(
		"%s throughput: %.0f req/s | Concurrency: %d | Elapsed time: %.3f seconds\n",
		message,
		c.counter/elapsed,
		c.maxGoRoutines,
		elapsed,
	)
}

func stats(c *Config, t Task, message string) {
	ticker := time.NewTicker(1 * time.Second)
	numRequestsf := float64(c.numRequests)
	start := time.Now()
	defer ticker.Stop()
	done := make(chan bool)
	go func() {
		SpawnerOfEvil(c, t)
		elapsed := time.Now().Sub(start).Seconds()
		fmt.Println("Task done.")
		summary(message, numRequestsf, elapsed, c)
		done <- true
	}()
	for {
		select {
		case <-done:
			return
		case t := <-ticker.C:
			_ = t
			elapsed := time.Now().Sub(start).Seconds()
			summaryPerDuration(message, numRequestsf, elapsed, c)
		}
	}
}

func singleNode() {
	fmt.Println("\nStarting single-node contest...")

	client := &http.Client{}

	c := NewDefaultConfig()
	c.req.client = client
	c.req.expectedStatusCode = 201
	c.req.endpoint += "/events"

	numRequestsf := float64(c.numRequests)

	fmt.Println("Preloading events...")
	stats(c, addSampleEvents, "Preload")

	fmt.Println("Starting exclusive Query Membership...")
	cq := NewDefaultConfig()
	cq.req.client = client
	cq.req.expectedStatusCode = 200
	cq.req.endpoint += "/proofs/membership"
	stats(cq, queryMembership, "Query")

	fmt.Println("Starting continuous load...")
	ca := NewDefaultConfig()
	ca.req.client = client
	ca.req.expectedStatusCode = 201
	ca.req.endpoint += "/events"
	ca.startVersion = c.numRequests
	ca.continuous = true
	go stats(ca, addSampleEvents, "Write")
	fmt.Println("Starting Query Membership with continuous load...")
	//	stats(c, QueryMembership, "Read query")
	start := time.Now()
	stats(cq, queryMembership, "Query")
	elapsed := time.Now().Sub(start).Seconds()
	fmt.Printf(
		"Query done. Reading Throughput: %.0f req/s: (%v reqs in %.3f seconds) | Concurrency: %d\n",
		numRequestsf/elapsed,
		cq.numRequests,
		elapsed,
		cq.maxGoRoutines,
	)

	currentVersion := getVersion("last-event")
	fmt.Printf(
		"Query done. Writing Throughput: %.0f req/s: (%v reqs in %.3f seconds) | Concurrency: %d\n",
		(float64(currentVersion)-numRequestsf)/elapsed,
		currentVersion-uint64(c.numRequests),
		elapsed,
		c.maxGoRoutines,
	)
}

func multiNode(numFollowers int) {
	fmt.Println("\nStarting multi-node contest...")
	var queryWg sync.WaitGroup

	client := &http.Client{}

	c := NewDefaultConfig()
	c.req.client = client
	c.req.expectedStatusCode = 201
	c.req.endpoint += "/events"

	fmt.Println("PRELOAD")
	stats(c, addSampleEvents, "Preload")

	config := make([]*Config, 0, numFollowers)
	for i := 0; i < numFollowers; i++ {
		c := NewDefaultConfig()
		c.req.client = client
		c.req.expectedStatusCode = 200
		c.req.endpoint = fmt.Sprintf("http://localhost:%d", 8081+i)
		c.req.endpoint += "/proofs/membership"

		config = append(config, c)
	}

	time.Sleep(1 * time.Second)

	fmt.Println("EXCLUSIVE QUERY MEMBERSHIP")
	stats(config[0], queryMembership, "Follower 1 read")

	fmt.Println("QUERY MEMBERSHIP UNDER CONTINUOUS LOAD")
	for i, c := range config {
		queryWg.Add(1)
		go func(i int, c *Config) {
			defer queryWg.Done()
			stats(c, queryMembership, fmt.Sprintf("Follower %d read", i+1))
		}(i, c)

	}

	fmt.Println("Starting continuous load...")
	ca := NewDefaultConfig()
	ca.req.client = client
	ca.req.expectedStatusCode = 201
	ca.req.endpoint += "/events"
	ca.startVersion = c.numRequests
	ca.continuous = true

	start := time.Now()
	go stats(ca, addSampleEvents, "Leader write")
	queryWg.Wait()
	elapsed := time.Now().Sub(start).Seconds()

	numRequestsf := float64(c.numRequests)
	currentVersion := getVersion("last-event")
	fmt.Printf(
		"Leader write throughput: %.0f req/s: (%v reqs in %.3f seconds) | Concurrency: %d\n",
		(float64(currentVersion)-numRequestsf)/elapsed,
		currentVersion-uint64(c.numRequests),
		elapsed,
		c.maxGoRoutines,
	)
}

func main() {
	switch m := os.Getenv("MULTINODE"); m {
	case "":
		singleNode()
	case "2":
		multiNode(2)
	case "4":
		multiNode(4)
	default:
		fmt.Println("Error: MULTINODE env var should have values 2 or 4, or not be defined at all.")
	}
}
