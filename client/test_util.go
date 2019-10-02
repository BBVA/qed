/*
   Copyright 2018-2019 Banco Bilbao Vizcaya Argentaria, S.A.

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

package client

import (
	"net/http"
	"strings"
)

type RoundTripFunc func(req *http.Request) (*http.Response, error)

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// NewTestHttpClient returns *http.Client with Transport replaced to avoid making real calls
func NewTestHttpClient(fn RoundTripFunc) *http.Client {

	checkRedirect := func(req *http.Request, via []*http.Request) error {
		req.Host = req.URL.Host
		return nil
	}

	return &http.Client{
		Transport:     fn,
		CheckRedirect: checkRedirect,
	}
}

// NewFailingTransport will run a fail callback if it sees a given URL path prefix.
func NewFailingTransport(path string, fail RoundTripFunc, next http.RoundTripper) RoundTripFunc {
	return func(req *http.Request) (*http.Response, error) {
		if strings.HasPrefix(req.URL.Path, path) && fail != nil {
			return fail(req)
		}
		if next != nil {
			return next.RoundTrip(req)
		}
		return http.DefaultTransport.RoundTrip(req)
	}
}
