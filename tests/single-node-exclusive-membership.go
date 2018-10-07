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
	"os/exec"
	"sync"
	"time"

	"github.com/bbva/qed/api/apihttp"
)

const (
	maxGoRutines = 10
	apiKey       = "pepe"
	numRequests  = 100000
)

func startServer() {
	cmd := exec.Command("./start_server")
	go cmd.Run()
	time.Sleep(5 * time.Second)
}

func stopServer() {
	fmt.Println("Shutting down server...")
	cmd := exec.Command("./stop_server")
	cmd.Run()
}

// func BenchmarkMembership(b *testing.B) {
func AddSampleEvents(baseVersion int, continuous bool) {
	client := &http.Client{}
	var wg sync.WaitGroup
	maxRequests := numRequests

	if continuous == true {
		maxRequests *= 100
	}

	for i := 0; i < maxGoRutines; i++ {
		wg.Add(1)
		go func(goRutineId int) {
			defer wg.Done()
			for j := baseVersion + goRutineId; j < baseVersion+maxRequests; j += maxGoRutines {

				buf := []byte(fmt.Sprintf("event %d", j))
				query, err := json.Marshal(&apihttp.Event{buf})
				if len(query) == 0 {
					log.Fatalf("Empty query: %v", err)
				}

				req, err := http.NewRequest("POST", "http://localhost:8080/events", bytes.NewBuffer(query))
				if err != nil {
					log.Fatalf("Error preparing request: %v", err)
				}

				// Set Api-Key header
				req.Header.Set("Api-Key", apiKey)
				res, err := client.Do(req)
				defer res.Body.Close()
				if err != nil {
					log.Fatalf("Unable to perform request: %v", err)
				}
				if res.StatusCode != 201 {
					log.Fatalf("Server error: %v", err)
				}

				io.Copy(ioutil.Discard, res.Body)
			}
		}(i)
	}

	wg.Wait()
}

func getVersion(eventTemplate string) uint64 {
	client := &http.Client{}

	buf := fmt.Sprintf(eventTemplate)

	query, err := json.Marshal(&apihttp.Event{[]byte(buf)})
	if len(query) == 0 {
		log.Fatalf("Empty query: %v", err)
	}

	req, err := http.NewRequest("POST", "http://localhost:8080/events", bytes.NewBuffer(query))
	if err != nil {
		log.Fatalf("Error preparing request: %v", err)
	}

	// Set Api-Key header
	req.Header.Set("Api-Key", apiKey)
	res, err := client.Do(req)
	defer res.Body.Close()
	if err != nil {
		log.Fatalf("Unable to perform request: %v", err)
	}
	if res.StatusCode != 201 {
		log.Fatalf("Server error: %v", err)
	}

	body, _ := ioutil.ReadAll(res.Body)

	var signedSnapshot apihttp.SignedSnapshot
	json.Unmarshal(body, &signedSnapshot)
	version := signedSnapshot.Snapshot.Version

	io.Copy(ioutil.Discard, res.Body)

	return version
}

func QueryMembership(baseVersion int, continuous bool) {
	client := &http.Client{}

	var wg sync.WaitGroup
	var version uint64 = uint64(numRequests - 1)
	maxRequests := numRequests

	if continuous == true {
		maxRequests *= 100
	}

	for i := 0; i < maxGoRutines; i++ {
		wg.Add(1)
		go func(goRutineId int) {
			defer wg.Done()
			for j := baseVersion + goRutineId; j < baseVersion+maxRequests; j += maxGoRutines {

				buf := []byte(fmt.Sprintf("event %d", j))
				query, err := json.Marshal(apihttp.MembershipQuery{
					buf,
					version,
				})
				if len(query) == 0 {
					log.Fatalf("Empty query: %v", err)
				}

				req, err := http.NewRequest("POST", "http://localhost:8080/proofs/membership", bytes.NewBuffer(query))
				if err != nil {
					log.Fatalf("Error preparing request: %v", err)
				}

				// Set Api-Key header
				req.Header.Set("Api-Key", apiKey)
				res, err := client.Do(req)
				defer res.Body.Close()
				if err != nil {
					log.Fatalf("Unable to perform request: %v", err)
				}
				if res.StatusCode != 200 {
					log.Fatalf("Server error: %v", err)
				}

				io.Copy(ioutil.Discard, res.Body)
			}
		}(i)
	}

	wg.Wait()
}

func main() {
	fmt.Println("Starting contest...")
	numRequestsf := float64(numRequests)
	startServer()
	defer stopServer()

	start := time.Now()
	fmt.Println("Preloading events...")
	AddSampleEvents(0, false)
	elapsed := time.Now().Sub(start).Seconds()
	fmt.Printf("Preload done. Throughput: %.0f req/s: (%v reqs in %.3f seconds) | Concurrency: %d\n", numRequestsf/elapsed, numRequests, elapsed, maxGoRutines)

	start = time.Now()
	fmt.Println("Starting exclusive Query Membership...")
	QueryMembership(0, false)
	elapsed = time.Now().Sub(start).Seconds()
	fmt.Printf("Query done. Throughput: %.0f req/s: (%v reqs in %.3f seconds) | Concurrency: %d\n", numRequestsf/elapsed, numRequests, elapsed, maxGoRutines)

	fmt.Println("Starting continuous load...")
	go AddSampleEvents(numRequests, true)

	//go currentVersion()
	start = time.Now()
	fmt.Println("Starting Query Membership with continuous load...")
	QueryMembership(0, false)
	elapsed = time.Now().Sub(start).Seconds()
	currentVersion := getVersion("last-event")
	fmt.Printf("Query done. Reading Throughput: %.0f req/s: (%v reqs in %.3f seconds) | Concurrency: %d\n", numRequestsf/elapsed, numRequests, elapsed, maxGoRutines)
	fmt.Printf("Query done. Writing Throughput: %.0f req/s: (%v reqs in %.3f seconds) | Concurrency: %d\n", (float64(currentVersion)-numRequestsf)/elapsed, currentVersion-numRequests, elapsed, maxGoRutines)
}
