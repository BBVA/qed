// Copyright 2018 BBVA. All rights reserved.
// Use of this source code is governed by a Apache 2 License
// that can be found in the LICENSE file.

/*
	Package API implements the interface to use the Verification Data Service
*/
package api

import (
	"io"
	"net/http"
)

// This handler checks the system status and returns it accordinly.
// The http call it answer is:
//	GET /health-check
//
// The following statuses are expected:
//
// If everything is allright, the HTTP status is 200 and the body contains:
//	 {"version": "0", "status":"ok"}
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	// A very simple health check.
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	// In the future we could report back on the status of our DB, or our cache
	// (e.g. Redis) by performing a simple PING, and include them in the response.
	io.WriteString(w, `{"version": "0", "status":"ok"}`)
}
