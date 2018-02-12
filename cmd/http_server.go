// Copyright 2018 BBVA. All rights reserved.
// Use of this source code is governed by a Apache 2 License
//  that can be found in the LICENSE file

package main

import (
	"log"
	"net/http"
	"verifiabledata/api"
)

func main() {

	http.HandleFunc("/health-check", api.HealthCheckHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
