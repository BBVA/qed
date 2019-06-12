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
)

const (
	errMalformedURL     = "Malformed URL"
	errMissingURLScheme = "Missing URL Scheme"
	errMissingURLHost   = "Missing URL Host"
)

func urlParse(endpoint string) error {
	url, err := url.Parse(endpoint)

	if err != nil {
		return errors.New(fmt.Sprintf("%s", errMalformedURL))
	}

	if url.Scheme == "" {
		fmt.Printf("%s in %s\n", errMissingURLScheme, endpoint)
		return errors.New(errMissingURLScheme)
	}

	if url.Hostname() == "" {
		return errors.New(errMissingURLHost)
	}

	return nil
}

func urlParseNoSchemaRequired(endpoint string) error {
	endpoint = "http://" + endpoint
	url, err := url.Parse(endpoint)

	if err != nil {
		return errors.New(errMalformedURL)
	}

	if url.Hostname() == "" {
		return errors.New(errMissingURLHost)
	}

	return nil
}
