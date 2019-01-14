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
	"fmt"
	"os/exec"
	"testing"

	"github.com/bbva/qed/testutils/scope"
	assert "github.com/stretchr/testify/require"
)

const (
	QEDProfilingURL = "http://localhost:6060/debug/pprof"
)

// FIXME: This function should also include testing for the other servers, not
// just the profiling one
func TestStart(t *testing.T) {
	bServer, aServer := setupServer(0, "", false, t)

	scenario, let := scope.Scope(t,
		merge(bServer),
		merge(aServer),
	)

	scenario("Test availability of profiling server", func() {
		let("Query to expected context", func(t *testing.T) {
			cmd := exec.Command("curl",
				"--fail",
				"-sS",
				"-XGET",
				"-H", fmt.Sprintf("Api-Key:%s", APIKey),
				"-H", "Content-type: application/json",
				QEDProfilingURL,
			)

			_, err := cmd.CombinedOutput()
			assert.NoError(t, err, "Subprocess must not exit with non-zero status")
		})

		let("Query to unexpected context", func(t *testing.T) {
			cmd := exec.Command("curl",
				"--fail",
				"-sS",
				"-XGET",
				"-H", fmt.Sprintf("Api-Key:%s", APIKey),
				"-H", "Content-type: application/json",
				QEDProfilingURL+"/xD",
			)

			_, err := cmd.CombinedOutput()
			assert.Error(t, err, "Subprocess must exit with non-zero status")
		})
	})
}
