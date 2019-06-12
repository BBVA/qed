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

package cmd

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUrlParse(t *testing.T) {

	testCases := []struct {
		endpoints     []string
		expectedError error
	}{
		{
			endpoints:     []string{"http://localhost", "http://localhost:8080", "https://127.0.0.1", "https://127.0.0.1:8080"},
			expectedError: nil,
		},
		{
			endpoints:     []string{"localhost", "127.0.0.1", "http//localhost"},
			expectedError: errors.New(errMissingURLScheme),
		},
		{
			endpoints:     []string{"http://", "https:/localhost", "http://:8080", "localhost:8080"},
			expectedError: errors.New(errMissingURLHost),
		},
		{
			endpoints:     []string{"127.0.0.1:8080"},
			expectedError: errors.New(errMalformedURL),
		},
	}

	for _, c := range testCases {
		for _, e := range c.endpoints {
			fmt.Printf("Input: %v ", e)
			err := urlParse(e)
			require.Equal(t, err, c.expectedError, "Errors do not match")
		}
	}
}

func TestUrlParseNoSchemaRequired(t *testing.T) {

	testCases := []struct {
		endpoints     []string
		expectedError error
	}{
		{
			endpoints:     []string{"localhost", "localhost:8080", "127.0.0.1", "127.0.0.1:8080"},
			expectedError: nil,
		},
		// {
		// 	endpoints:     []string{"http//localhost"},
		// 	expectedError: errors.New(errMissingURLScheme),
		// },
		{
			endpoints:     []string{""}, // , "http://:8080"},
			expectedError: errors.New(errMissingURLHost),
		},
	}

	for _, c := range testCases {
		for _, e := range c.endpoints {
			err := urlParseNoSchemaRequired(e)
			require.Equal(t, err, c.expectedError, "Errors do not match")
		}
	}
}
