// Copyright 2018 BBVA. All rights reserved.
// Use of this source code is governed by a Apache 2 License
// that can be found in the LICENSE file
package apihttp

import (
	"net/http"

	"verifiabledata/balloon"
)

// NewServer returns a new *http.ServeMux containing all the API handlers
// already configured
func New(balloon balloon.Balloon) *http.ServeMux {

	api := http.NewServeMux()
	api.HandleFunc("/health-check", AuthHandlerMiddleware(HealthCheckHandler))
	api.HandleFunc("/events", AuthHandlerMiddleware(InsertEvent(balloon)))

	return api
}
