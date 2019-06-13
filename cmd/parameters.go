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
	"net/url"
	"strings"
)

const (
	errMalformedURL     = "Malformed URL"
	errMissingURLScheme = "Missing URL Scheme"
	errMissingURLHost   = "Missing URL Host"
	errUnexpectedScheme = "Unexpected URL Scheme"
)

func urlParse(endpoints ...string) error {
	for _, endpoint := range endpoints {
		url, err := url.Parse(endpoint)

		if err != nil {
			return errors.New(fmt.Sprintf("%s in %s.", errMalformedURL, endpoint))
		}

		if url.Scheme == "" {
			return errors.New(fmt.Sprintf("%s in %s.", errMissingURLScheme, endpoint))
		}

		if url.Hostname() == "" {
			return errors.New(fmt.Sprintf("%s in %s.", errMissingURLHost, endpoint))
		}
	}
	return nil
}

func urlParseNoSchemaRequired(endpoints ...string) error {
	for _, endpoint := range endpoints {

		if strings.Index(endpoint, "://") != -1 {
			return errors.New(fmt.Sprintf("%s in %s.", errUnexpectedScheme, endpoint))
		}

		// Add fake scheme to get an expected result from url.Parse
		newEndpoint := "http://" + endpoint
		url, err := url.Parse(newEndpoint)

		if err != nil {
			return errors.New(fmt.Sprintf("%s in %s.", errMalformedURL, endpoint))
		}

		if url.Hostname() == "" {
			return errors.New(fmt.Sprintf("%s in %s.", errMissingURLHost, endpoint))
		}
	}
	return nil
}
