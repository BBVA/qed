// Copyright 2018 BBVA. All rights reserved.
// Use of this source code is governed by a Apache 2 License
// that can be found in the LICENSE file.

package http

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"net/http"

	"github.com/golang/glog"

	"verifiabledata/merkle/history"
)

// FIXME: temporal mock insead of the SMT
var smt_mock = make(map[string]*InsertResponse)

var tree = history.NewInmemoryTree()

// Create a JSON struct for our HealthCheck
type HealthCheckResponse struct {
	Version int    `json:"version"`
	Status  string `json:"status"`
}

// This handler checks the system status and returns it accordinly.
// The http call it answer is:
//	GET /health-check
//
// The following statuses are expected:
//
// If everything is allright, the HTTP status is 200 and the body contains:
//	 {"version": "0", "status":"ok"}
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	result := HealthCheckResponse{
		Version: 0,
		Status:  "ok",
	}

	resultJson, err := json.Marshal(result)
	if err != nil {
		panic(err)
	}

	// A very simple health check.
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	// In the future we could report back on the status of our DB, or our cache
	// (e.g. Redis) by performing a simple PING, and include them in the response.
	out := new(bytes.Buffer)
	json.Compact(out, resultJson)

	w.Write(out.Bytes())
}

type InsertData struct {
	Message string `json:"message"`
}

// Create a JSON struct for our event
type InsertResponse struct {
	Hash       string `json:"hash"`
	Commitment string `json:"commitment"`
	Index      uint64 `json:"index"`
}

type InsertRequest struct {
	InsertData      InsertData
	ResponseChannel chan *InsertResponse
	ProcessResponse func(data InsertData, responseChannel chan *InsertResponse)
}

// This handler posts an event into the system:
// The http post url is:
//  POST /events
//
// The follwing statuses are expected:
// If everything is allright, the HTTP status is 201 and the body contains:
//	{
//	"hash": "B8E1F80BD70AE0784C7855A451731B745FDDB67749D23F637BE9082B75E9575B",
//	"commitment": "6A19F0FB4BE54511524BCD5B0C98B38DA1EE049A39735C39311E10336024436F",
//	"index": 1
//	}

func InsertEvent(insertRequestQueue chan *InsertRequest) http.HandlerFunc {

	ProcessInsertResponse := func(data InsertData, responseChannel chan *InsertResponse) {
		commitment, node, err := tree.Add([]byte(data.Message))
		if err != nil {
			panic(err)
		}

		response := InsertResponse{
			Hash:       string(node.Digest),
			Commitment: string(commitment.Digest),
			Index:      node.Pos.Index,
		}

		// FIXME: temporal mock insead of the SMT
		smt_mock[data.Message] = &response

		glog.Infof("New event inserted with index [%d]: %s", response.Index,
			hex.EncodeToString([]byte(response.Commitment)))

		responseChannel <- &response
	}

	return func(w http.ResponseWriter, r *http.Request) {

		// Make sure we can only be called with an HTTP POST request.
		if r.Method != "POST" {
			w.Header().Set("Allow", "POST")
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if r.Body == nil {
			http.Error(w, "Please send a request body", http.StatusBadRequest)
			return
		}

		var event InsertData
		err := json.NewDecoder(r.Body).Decode(&event)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		responseChannel := make(
			chan *InsertResponse,
		)
		defer close(responseChannel)

		eventRequest := &InsertRequest{
			InsertData:      event,
			ResponseChannel: responseChannel,
			ProcessResponse: ProcessInsertResponse,
		}

		insertRequestQueue <- eventRequest

		// Wait for the response
		response := <-responseChannel

		out, err := json.Marshal(response)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Write(out)
		return
	}
}

type FetchData struct {
	Message string `json: "message"`
}

type FetchResponse struct {
	Index uint64 `json:"index"`
}

type FetchRequest struct {
	FetchData       FetchData
	ResponseChannel chan *FetchResponse
	ProcessResponse func(data FetchData, responseChannel chan *FetchResponse)
}

func GetEvent(eventIndex chan *FetchRequest, process func(fetch FetchData, responseChannel chan *FetchResponse)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Make our endpoint cand only be called with HTTP GET method
		if r.Method != "GET" {
			w.Header().Set("Allow", "GET")
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// Check if the request body is empty
		if r.Body == nil {
			http.Error(w, "Please send a request body", http.StatusBadRequest)
		}

		var fetch FetchData
		err := json.NewDecoder(r.Body).Decode(&fetch)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return

		}

		responseChannel := make(chan *FetchResponse)
		defer close(responseChannel)

		eventRequest := &FetchRequest{
			FetchData:       fetch,
			ResponseChannel: responseChannel,
			ProcessResponse: process,
		}

		eventIndex <- eventRequest

		// Wait for the response
		response := <-responseChannel

		out, err := json.Marshal(response)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(out)
		return
	}
}

func FetchEvent(eventIndex chan *FetchRequest) http.HandlerFunc {
	process := func(fetch FetchData, responseChannel chan *FetchResponse) {
		// FIXME: temporal mock insead of the SMT
		responseChannel <- &FetchResponse{
			Index: smt_mock[fetch.Message].Index,
		}
	}

	return GetEvent(eventIndex, process)
}

func GetMembershipProof(eventIndex chan *FetchRequest) http.HandlerFunc {
	process := func(fetch FetchData, responseChannel chan *FetchResponse) {

	}

	return GetEvent(eventIndex, process)
}

// AuthHandlerMiddleware function is an HTTP handler wrapper that validates our requests
func AuthHandlerMiddleware(handler http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Check if Api-Key header is empty
		if r.Header.Get("Api-Key") == "" {
			http.Error(w, "Missing Api-Key header", http.StatusUnauthorized)
			return
		}

		handler.ServeHTTP(w, r)
	})
}
