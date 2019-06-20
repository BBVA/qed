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
	"fmt"
	"net/url"
	"strings"
)

const (
	errMalformedURL     = "malformed URL"
	errMissingURLScheme = "missing URL Scheme"
	errMissingURLHost   = "missing URL Host"
	errMissingURLPort   = "missing URL Port"
	errUnexpectedScheme = "unexpected URL Scheme"
)

// urlParse function checks that given string parameters are valid URLs for
// REST requests: schema + hostname [ + port ]
func urlParse(endpoints ...string) error {
	for _, endpoint := range endpoints {
		url, err := url.Parse(endpoint)

		if err != nil {
			return fmt.Errorf("%s in %s", errMalformedURL, endpoint)
		}

		if url.Scheme == "" {
			return fmt.Errorf("%s in %s", errMissingURLScheme, endpoint)
		}

		if url.Hostname() == "" {
			return fmt.Errorf("%s in %s", errMissingURLHost, endpoint)
		}
	}
	return nil
}

// urlParse function checks that given string parameters are valid IPs for
// binding services: hostname + port. No schema is required.
func urlParseNoSchemaRequired(endpoints ...string) error {
	for _, endpoint := range endpoints {

		if strings.Contains(endpoint, "://") {
			return fmt.Errorf("%s in %s", errUnexpectedScheme, endpoint)
		}

		// Add fake scheme to get an expected result from url.Parse
		url, err := url.Parse("http://" + endpoint)

		if err != nil {
			return fmt.Errorf("%s in %s", errMalformedURL, endpoint)
		}

		if url.Hostname() == "" {
			return fmt.Errorf("%s in %s", errMissingURLHost, endpoint)
		}

		if url.Port() == "" {
			return fmt.Errorf("%s in %s", errMissingURLPort, endpoint)
		}
	}
	return nil
}
