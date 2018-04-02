// Copyright 2018 BBVA. All rights reserved.
// Use of this source code is governed by a Apache 2 License
// that can be found in the LICENSE file.

package http

//
// import (
// 	"bytes"
// 	"encoding/json"
// 	"net/http"
// 	// "github.com/golang/glog"
// )
//
// type FetchData struct {
// 	Message string `json: "message"`
// }
//
// type FetchResponse struct {
// 	Index uint64 `json:"index"`
// }
//
// type FetchRequest struct {
// 	FetchData       FetchData
// 	ResponseChannel chan *FetchResponse
// }
//
// func FetchEvent(eventIndex chan *FetchRequest) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		// Make our endpoint cand only be called with HTTP GET method
// 		if r.Method != "GET" {
// 			w.Header().Set("Allow", "GET")
// 			w.WriteHeader(http.StatusMethodNotAllowed)
// 			return
// 		}
//
// 		// Check if the request body is empty
// 		if r.Body == nil {
// 			http.Error(w, "Please send a request body", http.StatusBadRequest)
// 		}
//
// 		var fetch FetchData
// 		err := json.NewDecoder(r.Body).Decode(&fetch)
// 		if err != nil {
// 			http.Error(w, err.Error(), http.StatusBadRequest)
// 			return
//
// 		}
//
// 		responseChannel := make(chan *FetchResponse)
// 		defer close(responseChannel)
//
// 		eventRequest := &FetchRequest{
// 			FetchData:       fetch,
// 			ResponseChannel: responseChannel,
// 		}
//
// 		eventIndex <- eventRequest
//
// 		// Wait for the response
// 		response := <-responseChannel
//
// 		out, err := json.Marshal(response)
// 		if err != nil {
// 			http.Error(w, err.Error(), http.StatusInternalServerError)
// 			return
// 		}
//
// 		w.WriteHeader(http.StatusOK)
// 		w.Write(out)
// 		return
// 	}
// }
//
// func MembershipProof(eventIndex chan *FetchRequest) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
//
// 		// Make our endpoint cand only be called with HTTP GET method
// 		if r.Method != "GET" {
// 			w.Header().Set("Allow", "GET")
// 			w.WriteHeader(http.StatusMethodNotAllowed)
// 			return
// 		}
//
// 		// Check if the request body is empty
// 		if r.Body == nil {
// 			http.Error(w, "Please send a request body", http.StatusBadRequest)
// 		}
//
// 		var fetch FetchData
// 		err := json.NewDecoder(r.Body).Decode(&fetch)
// 		if err != nil {
// 			http.Error(w, err.Error(), http.StatusBadRequest)
// 			return
//
// 		}
//
// 		responseChannel := make(chan *FetchResponse)
// 		defer close(responseChannel)
//
// 		eventRequest := &FetchRequest{
// 			FetchData:       fetch,
// 			ResponseChannel: responseChannel,
// 		}
//
// 		eventIndex <- eventRequest
//
// 		// Wait for the response
// 		response := <-responseChannel
//
// 		out, err := json.Marshal(response)
// 		if err != nil {
// 			http.Error(w, err.Error(), http.StatusInternalServerError)
// 			return
// 		}
//
// 		w.WriteHeader(http.StatusOK)
// 		w.Write(out)
// 		return
// 		return
// 	}
// }
