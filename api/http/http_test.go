// Copyright 2018 BBVA. All rights reserved.
// Use of this source code is governed by a Apache 2 License
// that can be found in the LICENSE file

package http

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthCheckHandler(t *testing.T) {
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req, err := http.NewRequest("GET", "/health-check", nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HealthCheckHandler)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response body is what we expect.
	expected := `{"version":0,"status":"ok"}`
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

func TestEventInsertHandler(t *testing.T) {
	// Create a request to pass to our handler. We pass a message as a data.
	// If it's nil it will fail.
	data := []byte(`{"message": "this is a sample event"}`)

	req, err := http.NewRequest("POST", "/events", bytes.NewBuffer(data))
	if len(data) == 0 {
		t.Fatal(err)
	}

	fakeRequestQueue := make(chan *InsertRequest)
	go func() {
		select {
		case request := <-fakeRequestQueue:
			response := InsertResponse{
				Hash:       "B8E1F80BD70AE0784C7855A451731B745FDDB67749D23F637BE9082B75E9575B",
				Commitment: "6A19F0FB4BE54511524BCD5B0C98B38DA1EE049A39735C39311E10336024436F",
				Index:      1,
			}
			request.ResponseChannel <- &response
		}
	}()

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := EventInsertHandler{InsertRequestQueue: fakeRequestQueue}

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusCreated)
	}
}
func TestAuthHandler(t *testing.T){

	req, err := http.NewRequest("GET", "/auth", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Set Api-Key header
	req.Header.Set("Api-Key", "this-is-my-api-key")

	// Test our handler wrapper
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// if status := rr.Code; status != http.StatusOK {
		// 	t.Errorf("handler returned wrong status code: got %v want %v",
		// 		status, http.StatusOK)
		// }
	})

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder() 
	handler := AuthHandler(testHandler) 
	
	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}
