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

import "errors"

var (
	// ErrNoEndpoint is raised when no QED node is available.
	ErrNoEndpoint = errors.New("no QED node available")

	// ErrNoPrimary is raised when no QED primary node is available.
	ErrNoPrimary = errors.New("no QED primary node available")

	// ErrNoPrimary is raised when the current QED primary node is dead.
	ErrPrimaryDead = errors.New("current QED primary node is dead")

	// ErrRetry is raised when a request cannot be executed after
	// the configured number of retries.
	ErrRetry = errors.New("cannot connect after serveral retries")

	// ErrTimeout is raised when a request timed out.
	ErrTimeout = errors.New("timeout")
)
