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
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/bbva/qed/protocol"
)

var (
	endpoint        string
	apiKey          string
	wantAdd         bool
	wantIncremental bool
	wantMembership  bool
	offload         bool

	profiling        bool
	incrementalDelta int
	offset           int
	numRequests      int
	readConcurrency  int
	writeConcurrency int
	delay_ms         int
)

func init() {
	// Create a default config to use as default values in flags
	config := NewDefaultConfig()

	flag.StringVar(&endpoint, "endpoint", "http://localhost:8800", "The endopoint to make the load")
	flag.StringVar(&apiKey, "apikey", "my-key", "The key to use qed servers")
	flag.BoolVar(&wantAdd, "add", false, "Execute add benchmark")
	flag.IntVar(&delay_ms, "delay", 0, "Set request delay in milliseconds")

	usage := "Benchmark MembershipProof"
	flag.BoolVar(&wantMembership, "membership", false, usage)
	flag.BoolVar(&wantMembership, "m", false, usage+" (shorthand)")

	flag.BoolVar(&wantIncremental, "incremental", false, "Execute Incremental benchmark")
	flag.BoolVar(&offload, "offload", false, "Perform reads only on %50 of the cluster size (With cluster size 2 reads will be performed only on follower1)")
	flag.BoolVar(&profiling, "profiling", false, "Enable Go profiling with pprof tool. $ go tool pprof -http : http://localhost:6061 ")

	usageDelta := "Specify delta for the IncrementalProof"
	flag.IntVar(&incrementalDelta, "delta", 1000, usageDelta)
	flag.IntVar(&incrementalDelta, "d", 1000, usageDelta+" (shorthand)")

	flag.IntVar(&numRequests, "n", 10e4, "Number of requests for the attack")
	flag.IntVar(&readConcurrency, "r", config.maxGoRoutines, "Set read concurrency value")
	flag.IntVar(&writeConcurrency, "w", config.maxGoRoutines, "Set write concurrency value")
	flag.IntVar(&offset, "offset", 0, "The starting version from which we start the load")
}

type Config struct {
	maxGoRoutines  int
	numRequests    int
	apiKey         string
	startVersion   int
	continuous     bool
	balloonVersion uint64
	counter        float64
	delay_ms       time.Duration
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
		apiKey:         apiKey,
		startVersion:   0,
		continuous:     false,
		balloonVersion: uint64(numRequests) - 1,
		counter:        0,
		req: HTTPClient{
			client:             nil,
			method:             "POST",
			endpoint:           endpoint,
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

		if res.StatusCode != c.req.expectedStatusCode {
			log.Fatalf("Server error: %v", res)
		}

		c.counter++

		_, _ = io.Copy(ioutil.Discard, res.Body)
		res.Body.Close()

		c.delay_ms = time.Duration(delay_ms)
		time.Sleep(c.delay_ms * time.Millisecond)
	}
	c.counter = 0
}

func addSampleEvents(eventIndex int, c *Config) ([]byte, error) {

	return json.Marshal(
		&protocol.Event{
			[]byte(fmt.Sprintf("event %d", eventIndex)),
		},
	)
}

func queryMembership(eventIndex int, c *Config) ([]byte, error) {
	return json.Marshal(
		&protocol.MembershipQuery{
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
		&protocol.IncrementalRequest{
			Start: start,
			End:   end,
		},
	)
}

func getVersion(eventTemplate string, c *Config) uint64 {
	ssl := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: ssl}

	buf := fmt.Sprintf(eventTemplate)

	query, err := json.Marshal(&protocol.Event{[]byte(buf)})
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

	var signedSnapshot protocol.SignedSnapshot
	json.Unmarshal(body, &signedSnapshot)
	version := signedSnapshot.Snapshot.Version

	return version
}

type axis struct {
	x, y []float64
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
	graph := &axis{}
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

func benchmarkAdd(numFollowers, numReqests, readConcurrency, writeConcurrency, offset int) {
	fmt.Println("\nStarting benchmark run...")

	ssl := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: ssl}

	c := NewDefaultConfig()
	c.req.client = client
	c.numRequests = numReqests
	c.maxGoRoutines = writeConcurrency
	c.startVersion = offset
	c.req.expectedStatusCode = 201
	c.req.endpoint += "/events"
	stats(c, addSampleEvents, "Add")

}

func benchmarkMembership(numFollowers, numReqests, readConcurrency, writeConcurrency int) {
	fmt.Println("\nStarting benchmark run...")
	var queryWg sync.WaitGroup
	ssl := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: ssl}

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
		c.req.endpoint = fmt.Sprintf("http://localhost:%d", 8801+i)
		c.req.endpoint += "/proofs/membership"

		config = append(config, c)
	}

	time.Sleep(1 * time.Second)

	fmt.Println("EXCLUSIVE QUERY MEMBERSHIP")
	stats(config[0], queryMembership, "Follower-1-read")

	go hotParams(config)
	fmt.Println("QUERY MEMBERSHIP UNDER CONTINUOUS LOAD")
	for i, c := range config {
		queryWg.Add(1)
		go func(i int, c *Config) {
			defer queryWg.Done()
			stats(c, queryMembership, fmt.Sprintf("Follower-%d-read-mixed", i+1))
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
	go stats(ca, addSampleEvents, "Leader-write-mixed")
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
	ssl := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: ssl}

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
		c.req.endpoint += "/proofs/incremental"

		config = append(config, c)
	}
	for i := 0; i < numFollowers; i++ {
		c := NewDefaultConfig()
		c.req.client = client
		c.numRequests = numReqests
		c.maxGoRoutines = readConcurrency
		c.req.expectedStatusCode = 200
		c.req.endpoint = fmt.Sprintf("http://localhost:%d", 8801+i)
		c.req.endpoint += "/proofs/incremental"

		config = append(config, c)
	}

	time.Sleep(1 * time.Second)
	fmt.Println("EXCLUSIVE QUERY INCREMENTAL")
	stats(config[0], queryIncremental, "Follower-1-read")

	go hotParams(config)
	fmt.Println("QUERY INCREMENTAL UNDER CONTINUOUS LOAD")
	for i, c := range config {
		queryWg.Add(1)
		go func(i int, c *Config) {
			defer queryWg.Done()
			stats(c, queryIncremental, fmt.Sprintf("Follower-%d-read-mixed", i+1))
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
	go stats(ca, addSampleEvents, "Leader-write-mixed")
	queryWg.Wait()
	elapsed := time.Now().Sub(start).Seconds()

	numRequestsf := float64(c.numRequests)
	currentVersion := getVersion("last-event", c)
	fmt.Printf(
		"Leader-write-mixed throughput: %.0f req/s: (%v reqs in %.3f seconds) | Concurrency: %d\n",
		(float64(currentVersion)-numRequestsf)/elapsed,
		currentVersion-uint64(c.numRequests),
		elapsed,
		c.maxGoRoutines,
	)
}
func hotParams(config []*Config) {
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		value := scanner.Text()

		switch t := value[0:2]; t {
		case "mr":
			i, _ := strconv.ParseInt(value[2:], 10, 64)
			d := time.Duration(i)
			for _, c := range config {
				c.delay_ms = d
			}
			fmt.Printf("Read throughtput set to: %d\n", i)
		case "ir":
			i, _ := strconv.ParseInt(value[2:], 10, 64)
			d := time.Duration(i)
			for _, c := range config {
				c.delay_ms = d
			}
			fmt.Printf("Read throughtput set to: %d\n", i)
		default:
			fmt.Println("Invalid command - Valid commands: mr100|ir200")
		}

	}
}

func main() {
	var n int
	switch m := os.Getenv("CLUSTER_SIZE"); m {
	case "":
		n = 0
	case "2":
		n = 2
	case "4":
		n = 4
	default:
		fmt.Println("Error: CLUSTER_SIZE env var should have values 2 or 4, or not be defined at all.")
	}

	flag.Parse()

	if profiling {
		go func() {
			fmt.Print("Go profiling enabled\n")
			log.Print(http.ListenAndServe(":6061", nil))
		}()
	}

	if offload {
		n = n / 2
		fmt.Printf("Offload: %v | %d\n", offload, n)
	}

	if wantAdd {
		fmt.Print("Benchmark ADD")
		benchmarkAdd(n, numRequests, readConcurrency, writeConcurrency, offset)
	}

	if wantMembership {
		fmt.Print("Benchmark MEMBERSHIP")
		benchmarkMembership(n, numRequests, readConcurrency, writeConcurrency)
	}

	if wantIncremental {
		fmt.Print("Benchmark INCREMENTAL")
		benchmarkIncremental(n, numRequests, readConcurrency, writeConcurrency)
	}
}
