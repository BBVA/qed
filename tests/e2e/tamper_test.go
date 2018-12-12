/*
   Copyright 2018 Banco Bilbao Vizcaya Argentaria, S.A.

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
package e2e

import (
	"testing"

	"github.com/bbva/qed/testutils/rand"
	"github.com/bbva/qed/testutils/scope"
	assert "github.com/stretchr/testify/require"
)

func TestTamper(t *testing.T) {
	bStore, aStore := setupStore(t)
	bServer, aServer := setupServer(0, "", t)
	bAuditor, aAuditor := setupAuditor(0, t)
	bMonitor, aMonitor := setupMonitor(0, t)
	bPublisher, aPublisher := setupPublisher(0, t)

	scenario, let := scope.Scope(t,
		merge(bStore, bServer, bAuditor, bMonitor, bPublisher),
		merge(aStore, aServer, aAuditor, aMonitor, aPublisher),
	)

	client := getClient(0)

	event := rand.RandomString(10)

	scenario("S", func() {
		var err error

		let("Add event", func(t *testing.T) {
			_, err = client.Add(event)
			assert.NoError(t, err)
		})
	})
}
