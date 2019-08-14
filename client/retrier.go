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
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/bbva/qed/log2"
)

var (
	// We need to consume response bodies to maintain http connections, but
	// limit the size we consume to respReadLimit.
	respReadLimit = int64(4096)
)

// RequestRetrier decides whether to retry a failed HTTP request.
type RequestRetrier interface {
	// DoReq executes the given request and if fails, it decides whether to retry
	// the call, how long to wait for the next call, or whether to return an
	// error (which will be returned to the service that started the HTTP
	// request in the first place).
	DoReq(req *RetriableRequest) (*http.Response, error)
}

// NoRequestRetrier is an implementation that does no retries.
type NoRequestRetrier struct {
	*http.Client
}

// NewNoRequestRetrier returns a retrier that does no retries.
func NewNoRequestRetrier(httpClient *http.Client) *NoRequestRetrier {
	return &NoRequestRetrier{Client: httpClient}
}

func (r *NoRequestRetrier) DoReq(req *RetriableRequest) (*http.Response, error) {
	// always rewind
	if req.body != nil {
		body, err := req.body()
		if err != nil {
			return nil, err
		}
		if c, ok := body.(io.ReadCloser); ok {
			req.Request.Body = c
		} else {
			req.Request.Body = ioutil.NopCloser(body)
		}
	}
	resp, err := r.Do(req.Request)
	// Check the response code. We retry on 500-range responses to allow
	// the server time to recover, as 500's are typically not permanent
	// errors and may relate to outages on the server side. This will catch
	// invalid reponse codes as well, like 0.
	if err == nil && resp.StatusCode > 0 && resp.StatusCode < 500 {
		return resp, nil
	}
	return nil, fmt.Errorf("%s %s: giving up after %d attempts",
		req.Method, req.URL, 1)
}

// BackoffRequestRetrier is an implementation that uses the given backoff strategy.
type BackoffRequestRetrier struct {
	*http.Client
	maxRetries int
	backoff    Backoff
	log        log2.Logger
}

// NewBackoffRequestRetrier returns a retrier that uses the given backoff strategy.
func NewBackoffRequestRetrier(httpClient *http.Client, maxRetries int, backoff Backoff) *BackoffRequestRetrier {
	return &BackoffRequestRetrier{
		Client:     httpClient,
		maxRetries: maxRetries,
		backoff:    backoff,
		log:        log2.L(),
	}
}

// NewBackoffRequestRetrierWithLogger returns a retrier that uses the given backoff strategy.
func NewBackoffRequestRetrierWithLogger(httpClient *http.Client, maxRetries int, backoff Backoff, logger log2.Logger) *BackoffRequestRetrier {
	return &BackoffRequestRetrier{
		Client:     httpClient,
		maxRetries: maxRetries,
		backoff:    backoff,
		log:        logger,
	}
}

func (r *BackoffRequestRetrier) DoReq(req *RetriableRequest) (*http.Response, error) {

	var resp *http.Response
	var err error

	for i := 0; ; i++ {

		var code int // HTTP response status code

		// always rewind
		if req.body != nil {
			body, err := req.body()
			if err != nil {
				return resp, err
			}
			if c, ok := body.(io.ReadCloser); ok {
				req.Request.Body = c
			} else {
				req.Request.Body = ioutil.NopCloser(body)
			}
		}

		// attempt the request
		resp, err = r.Do(req.Request)
		if resp != nil {
			code = resp.StatusCode
		}
		if err != nil {
			r.log.Infof("%s %s request failed: %v", req.Method, req.URL, err)
		}

		// Check the response code. We retry on 500-range responses to allow
		// the server time to recover, as 500's are typically not permanent
		// errors and may relate to outages on the server side. This will catch
		// invalid reponse codes as well, like 0.
		if err == nil && resp.StatusCode > 0 && resp.StatusCode < 500 {
			return resp, nil
		}

		// we decide to continue with retrying

		// We do this before drainBody beause there's no need for the I/O if
		// we're breaking out
		remain := r.maxRetries - i
		if remain <= 0 {
			break
		}

		// We're going to retry, consume any response to reuse the connection.
		if err == nil && resp != nil {
			// drain body
			_, err := io.Copy(ioutil.Discard, io.LimitReader(resp.Body, respReadLimit))
			if err != nil {
				r.log.Infof("Error reading response body: %v", err)
			}
		}

		wait, goahead := r.backoff.Next(i)
		if !goahead {
			break
		}

		desc := fmt.Sprintf("%s %s", req.Method, req.URL)
		if code > 0 {
			desc = fmt.Sprintf("%s (status: %d)", desc, code)
		}
		r.log.Infof("%s: retrying in %s (%d left)", desc, wait, remain)

		time.Sleep(wait)

	}

	// By default, we close the response body and return an error without
	// returning the response
	if resp != nil {
		resp.Body.Close()
	}

	return nil, fmt.Errorf("%s %s giving up after %d attempts",
		req.Method, req.URL, r.maxRetries+1)

}
