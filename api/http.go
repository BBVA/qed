// Copyright 2018 BBVA. All rights reserved.
// Use of this source code is governed by a Apache 2 License
// that can be found in the LICENSE file.

package api

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)

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
	Index      int64  `json:"index"`
}

type InsertRequest struct {
	InsertData      InsertData
	ResponseChannel chan *InsertResponse
}

type EventInsertHandler struct {
	InsertRequestQueue chan *InsertRequest
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
func (handler *EventInsertHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

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

	responseChannel := make(chan *InsertResponse)
	eventRequest := &InsertRequest{
		InsertData:      event,
		ResponseChannel: responseChannel,
	}

	log.Print(eventRequest)

	// Wait for the response
	response := InsertResponse{
		Hash:       "B8E1F80BD70AE0784C7855A451731B745FDDB67749D23F637BE9082B75E9575B",
		Commitment: "6A19F0FB4BE54511524BCD5B0C98B38DA1EE049A39735C39311E10336024436F",
		Index:      1,
	}

	// Close shannel afte response (MOVE after response code!!!)
	close(responseChannel)

	out, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(out)
	return
}
