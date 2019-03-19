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
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNoRequestRetrier(t *testing.T) {
	var numFailedReqs int
	fail := func(req *http.Request) (*http.Response, error) {
		numFailedReqs++
		return nil, errors.New("request failed")
	}

	httpClient := NewTestHttpClient(
		NewFailingTransport("/fail", fail, nil),
	)
	retrier := NewNoRequestRetrier(httpClient)

	req, err := NewRetriableRequest("GET", "http://foo.bar/fail", nil)
	require.NoError(t, err)

	resp, err := retrier.DoReq(req)
	require.Error(t, err)
	require.Nil(t, resp)
	require.Equal(t, 1, numFailedReqs, "The expected number of failed requests does not match")

}

func TestBackoffRequestRetrier(t *testing.T) {
	var numFailedReqs int
	fail := func(req *http.Request) (*http.Response, error) {
		numFailedReqs++
		return nil, errors.New("request failed")
	}

	httpClient := NewTestHttpClient(
		NewFailingTransport("/fail", fail, nil),
	)
	maxRetries := 5
	retrier := NewBackoffRequestRetrier(httpClient, maxRetries,
		NewSimpleBackoff(100, 100, 100, 100, 100))

	req, err := NewRetriableRequest("GET", "http://foo.bar/fail", nil)
	require.NoError(t, err)

	resp, err := retrier.DoReq(req)
	require.Error(t, err)
	require.Nil(t, resp)
	require.Equal(t, maxRetries+1, numFailedReqs, "The expected number of failed requests does not match")

}

func TestBackoffRequestRetrierWithStatus(t *testing.T) {
	var numFailedReqs int
	serverFail := func(req *http.Request) (*http.Response, error) {
		numFailedReqs++
		return &http.Response{
			StatusCode: 500,
			Body:       ioutil.NopCloser(bytes.NewBufferString("Internal server error")),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}, nil
	}

	httpClient := NewTestHttpClient(serverFail)
	maxRetries := 5
	retrier := NewBackoffRequestRetrier(httpClient, maxRetries,
		NewSimpleBackoff(100, 100, 100, 100, 100))

	req, err := NewRetriableRequest("GET", "http://foo.bar/fail", nil)
	require.NoError(t, err)

	resp, err := retrier.DoReq(req)
	require.Error(t, err)
	require.Nil(t, resp)
	require.Equal(t, maxRetries+1, numFailedReqs, "The expected number of failed requests does not match")
}

func TestBackoffRequestRetrierTwoRetries(t *testing.T) {
	var numFailedReqs int
	serverTemporaryFail := func(req *http.Request) (*http.Response, error) {
		numFailedReqs++
		if numFailedReqs > 1 {
			return &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewBufferString("OK")),
				// Must be set to non-nil value or it panics
				Header: make(http.Header),
			}, nil
		}
		return &http.Response{
			StatusCode: 500,
			Body:       ioutil.NopCloser(bytes.NewBufferString("Internal server error")),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}, nil
	}

	httpClient := NewTestHttpClient(serverTemporaryFail)
	maxRetries := 5
	retrier := NewBackoffRequestRetrier(httpClient, maxRetries,
		NewSimpleBackoff(100, 100, 100, 100, 100))

	req, err := NewRetriableRequest("GET", "http://foo.bar/fail", nil)
	require.NoError(t, err)

	resp, err := retrier.DoReq(req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 2, numFailedReqs, "The expected number of failed requests does not match")

}
