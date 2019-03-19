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
	"io"
	"net/http"
)

// ReaderFunc is the type of function that can be given natively to NewRetriableRequest.
type ReaderFunc func() (io.Reader, error)

// RetriableRequest wraps the metadata needed to create HTTP requests
// and allow to reused it between retries.
type RetriableRequest struct {
	// body is a seekable reader over the request body payload. This is
	// used to rewind the request data in between retries
	body ReaderFunc

	// Embed an HTTP request directly. This makes a *Request act exactly
	// like an *http.Request so that all meta methods are supported.
	*http.Request
}

// NewRetriableRequest creates a new retriable request.
func NewRetriableRequest(method, url string, rawBody []byte) (*RetriableRequest, error) {

	var body ReaderFunc
	var contentLength int64

	if rawBody != nil {
		body = func() (io.Reader, error) {
			return bytes.NewReader(rawBody), nil
		}
		contentLength = int64(len(rawBody))
	}

	httpReq, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	httpReq.ContentLength = contentLength

	return &RetriableRequest{body, httpReq}, nil
}
