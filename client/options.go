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
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"time"
)

// HTTPClientOptionF is a function that configures an HTTPClient.
type HTTPClientOptionF func(*HTTPClient) error

func configToOptions(conf *Config) ([]HTTPClientOptionF, error) {
	var options []HTTPClientOptionF
	if conf != nil {
		options = []HTTPClientOptionF{
			SetAPIKey(conf.APIKey),
			SetReadPreference(conf.ReadPreference),
			SetMaxRetries(conf.MaxRetries),
			SetTopologyDiscovery(conf.EnableTopologyDiscovery),
			SetHealthchecks(conf.EnableHealthChecks),
			SetHealthCheckTimeout(conf.HealthCheckTimeout),
			SetAttemptToReviveEndpoints(conf.AttemptToReviveEndpoints),
		}
		if len(conf.Endpoints) > 0 {
			options = append(options, SetURLs(conf.Endpoints[0], conf.Endpoints[1:]...))
		}

		defaultTransport := http.DefaultTransport.(*http.Transport)
		options = append(options, SetHttpClient(&http.Client{
			Timeout: conf.Timeout,
			Transport: &http.Transport{
				Dial: (&net.Dialer{
					Timeout: conf.DialTimeout,
				}).Dial,
				Proxy:                 defaultTransport.Proxy,
				DialContext:           defaultTransport.DialContext,
				MaxIdleConns:          defaultTransport.MaxIdleConns,
				IdleConnTimeout:       defaultTransport.IdleConnTimeout,
				ExpectContinueTimeout: defaultTransport.ExpectContinueTimeout,
				TLSClientConfig:       &tls.Config{InsecureSkipVerify: conf.Insecure},
				TLSHandshakeTimeout:   conf.HandshakeTimeout,
			},
		}))
	}
	return options, nil
}

func SetHttpClient(client *http.Client) HTTPClientOptionF {
	return func(c *HTTPClient) error {
		c.httpClient = client
		return nil
	}
}

func SetURLs(primary string, secondaries ...string) HTTPClientOptionF {
	return func(c *HTTPClient) error {
		if len(primary) > 0 {
			c.topology.Update(primary, secondaries...)
			return nil
		}
		return errors.New("Cannot use empty string for the primary url")
	}
}

func SetAttemptToReviveEndpoints(value bool) HTTPClientOptionF {
	return func(c *HTTPClient) error {
		c.topology.attemptToRevive = value
		return nil
	}
}

func SetRequestRetrier(retrier RequestRetrier) HTTPClientOptionF {
	return func(c *HTTPClient) error {
		if retrier != nil {
			c.retrier = retrier
			return nil
		}
		return errors.New("The request retrier cannot be nil")
	}
}

func SetAPIKey(key string) HTTPClientOptionF {
	return func(c *HTTPClient) error {
		c.apiKey = key
		return nil
	}
}

func SetReadPreference(preference ReadPref) HTTPClientOptionF {
	return func(c *HTTPClient) error {
		c.readPreference = preference
		return nil
	}
}

func SetMaxRetries(retries int) HTTPClientOptionF {
	return func(c *HTTPClient) error {
		c.maxRetries = retries
		return nil
	}
}

func SetTopologyDiscovery(enable bool) HTTPClientOptionF {
	return func(c *HTTPClient) error {
		c.discoveryEnabled = enable
		return nil
	}
}

func SetHealthchecks(enable bool) HTTPClientOptionF {
	return func(c *HTTPClient) error {
		c.healthcheckEnabled = enable
		return nil
	}
}

func SetHealthCheckTimeout(seconds time.Duration) HTTPClientOptionF {
	return func(c *HTTPClient) error {
		c.healthcheckTimeout = seconds
		return nil
	}
}
