// Copyright 2018 BBVA. All rights reserved.
// Use of this source code is governed by a Apache 2 License
// that can be found in the LICENSE file

package main

import (
	"log"
	"net/http"
	api "verifiabledata/api/http"
)

func main() {

	http.HandleFunc("/health-check", api.HealthCheckHandler)
	http.Handle("/events", &api.EventInsertHandler{InsertRequestQueue: make(chan *api.InsertRequest)})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
