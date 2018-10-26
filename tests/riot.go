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
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/bbva/qed/api/apihttp"
	"github.com/bbva/qed/publish"
)

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
		&apihttp.Event{
			[]byte(fmt.Sprintf("event %d", eventIndex)),
		},
	)
}

func queryMembership(eventIndex int, c *Config) ([]byte, error) {
	return json.Marshal(
		&apihttp.MembershipQuery{
			[]byte(fmt.Sprintf("event %d", eventIndex)),
			c.balloonVersion,
		},
	)
}

func queryIncremental(eventIndex int, c *Config) ([]byte, error) {
	end := uint64(eventIndex)
	start := uint64(math.Max(float64(eventIndex-incrementalDelta), 0.0))
	// start := end >> 1
	return json.Marshal(
		&apihttp.IncrementalRequest{
			Start: start,
			End:   end,
		},
	)
}

func getVersion(eventTemplate string, c *Config) uint64 {
	client := &http.Client{}

	buf := fmt.Sprintf(eventTemplate)

	query, err := json.Marshal(&apihttp.Event{[]byte(buf)})
	if len(query) == 0 {
		log.Fatalf("Empty query: %v", err)
	}

	req, err := http.NewRequest(c.req.method, c.req.endpoint, bytes.NewBuffer(query))
	if err != nil {
		log.Fatalf("Error preparing request: %v", err)
	}

	// Set Api-Key header
	req.Header.Set("Api-Key", c.apiKey)
	res, err := client.Do(req)
	if err != nil {
		log.Fatalf("Unable to perform request: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != 201 {
		log.Fatalf("Server error: %v", err)
	}

	body, _ := ioutil.ReadAll(res.Body)

	var snapshot publish.Snapshot
	json.Unmarshal(body, &snapshot)
	version := snapshot.Version

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

func benchmarkMembership(numFollowers, numReqests, readConcurrency, writeConcurrency int) {
	fmt.Println("\nStarting benchmark run...")
	var queryWg sync.WaitGroup

	client := &http.Client{}

	c := NewDefaultConfig()
	c.req.client = client
	c.numRequests = numReqests
	c.maxGoRoutines = writeConcurrency
	c.req.expectedStatusCode = 201
	c.req.endpoint += "/events"

	fmt.Println("PRELOAD")
	stats(c, addSampleEvents, "Preload")

	config := make([]*Config, 0, numFollowers)
	if numFollowers == 0 {
		c := NewDefaultConfig()
		c.req.client = client
		c.numRequests = numReqests
		c.maxGoRoutines = readConcurrency
		c.req.expectedStatusCode = 200
		c.req.endpoint += "/proofs/membership"

		config = append(config, c)
	}
	for i := 0; i < numFollowers; i++ {
		c := NewDefaultConfig()
		c.req.client = client
		c.numRequests = numReqests
		c.maxGoRoutines = readConcurrency
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
	ca.numRequests = numReqests
	ca.maxGoRoutines = writeConcurrency
	ca.req.expectedStatusCode = 201
	ca.req.endpoint += "/events"
	ca.startVersion = c.numRequests
	ca.continuous = true

	start := time.Now()
	go stats(ca, addSampleEvents, "Leader write")
	queryWg.Wait()
	elapsed := time.Now().Sub(start).Seconds()

	numRequestsf := float64(c.numRequests)
	currentVersion := getVersion("last-event", c)
	fmt.Printf(
		"Leader write throughput: %.0f req/s: (%v reqs in %.3f seconds) | Concurrency: %d\n",
		(float64(currentVersion)-numRequestsf)/elapsed,
		currentVersion-uint64(c.numRequests),
		elapsed,
		c.maxGoRoutines,
	)
}

func benchmarkIncremental(numFollowers, numReqests, readConcurrency, writeConcurrency int) {
	fmt.Println("\nStarting benchmark run...")
	var queryWg sync.WaitGroup

	client := &http.Client{}

	c := NewDefaultConfig()
	c.req.client = client
	c.numRequests = numReqests
	c.maxGoRoutines = writeConcurrency
	c.req.expectedStatusCode = 201
	c.req.endpoint += "/events"

	fmt.Println("PRELOAD")
	stats(c, addSampleEvents, "Preload")

	config := make([]*Config, 0, numFollowers)
	if numFollowers == 0 {
		c := NewDefaultConfig()
		c.req.client = client
		c.numRequests = numReqests
		c.maxGoRoutines = writeConcurrency
		c.req.expectedStatusCode = 200
		c.req.endpoint += "/proofs/incremental"

		config = append(config, c)
	}
	for i := 0; i < numFollowers; i++ {
		c := NewDefaultConfig()
		c.req.client = client
		c.numRequests = numReqests
		c.maxGoRoutines = writeConcurrency
		c.req.expectedStatusCode = 200
		c.req.endpoint = fmt.Sprintf("http://localhost:%d", 8081+i)
		c.req.endpoint += "/proofs/incremental"

		config = append(config, c)
	}

	time.Sleep(1 * time.Second)

	fmt.Println("EXCLUSIVE QUERY INCREMENTAL")
	stats(config[0], queryIncremental, "Follower 1 read")

	fmt.Println("QUERY INCREMENTAL UNDER CONTINUOUS LOAD")
	for i, c := range config {
		queryWg.Add(1)
		go func(i int, c *Config) {
			defer queryWg.Done()
			stats(c, queryIncremental, fmt.Sprintf("Follower %d read", i+1))
		}(i, c)
	}

	fmt.Println("Starting continuous load...")
	ca := NewDefaultConfig()
	ca.req.client = client
	ca.numRequests = numReqests
	ca.maxGoRoutines = writeConcurrency
	ca.req.expectedStatusCode = 201
	ca.req.endpoint += "/events"
	ca.startVersion = c.numRequests
	ca.continuous = true

	start := time.Now()
	go stats(ca, addSampleEvents, "Leader write")
	queryWg.Wait()
	elapsed := time.Now().Sub(start).Seconds()

	numRequestsf := float64(c.numRequests)
	currentVersion := getVersion("last-event", c)
	fmt.Printf(
		"Leader write throughput: %.0f req/s: (%v reqs in %.3f seconds) | Concurrency: %d\n",
		(float64(currentVersion)-numRequestsf)/elapsed,
		currentVersion-uint64(c.numRequests),
		elapsed,
		c.maxGoRoutines,
	)
}

var (
	wantMembership   bool
	incrementalDelta int
	numRequests      int
	readConcurrency  int
	writeConcurrency int
)

func init() {
	const (
		defaultWantMembership   = false
		defaultIncrementalDelta = 1000
		defaultNumRequests      = 100000
		usage                   = "Benchmark MembershipProof"
		usageDelta              = "Specify delta for the IncrementalProof"
		usageNumRequests        = "Number of requests for the attack"
		usageReadConcurrency    = "Set read concurrency value"
		usageWriteConcurrency   = "Set write concurrency value"
	)

	// Create a default config to use as default values in flags
	config := NewDefaultConfig()

	flag.BoolVar(&wantMembership, "membership", defaultWantMembership, usage)
	flag.BoolVar(&wantMembership, "m", defaultWantMembership, usage+" (shorthand)")
	flag.IntVar(&incrementalDelta, "delta", defaultIncrementalDelta, usageDelta)
	flag.IntVar(&incrementalDelta, "d", defaultIncrementalDelta, usageDelta+" (shorthand)")
	flag.IntVar(&numRequests, "n", defaultNumRequests, usageNumRequests)
	flag.IntVar(&readConcurrency, "r", config.maxGoRoutines, usageReadConcurrency)
	flag.IntVar(&writeConcurrency, "w", config.maxGoRoutines, usageWriteConcurrency)
}

func main() {
	var n int
	switch m := os.Getenv("MULTINODE"); m {
	case "":
		n = 0
	case "2":
		n = 2
	case "4":
		n = 4
	default:
		fmt.Println("Error: MULTINODE env var should have values 2 or 4, or not be defined at all.")
	}

	flag.Parse()

	if wantMembership {
		fmt.Println("Benchmark MEMBERSHIP")
		benchmarkMembership(n, numRequests, readConcurrency, writeConcurrency)
	} else {
		fmt.Println("Benchmark INCREMENTAL")
		benchmarkIncremental(n, numRequests, readConcurrency, writeConcurrency)
	}
}
