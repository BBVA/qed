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
	"strings"
	"testing"

	"github.com/bbva/qed/testutils/scope"

	assert "github.com/stretchr/testify/require"
)

func TestCli(t *testing.T) {
	before, after := setup(0, "", t)
	scenario, let := scope.Scope(t, before, after)

	scenario("Add one event through cli and verify it", func() {

		let("Add event", func(t *testing.T) {
			cmd := exec.Command("go",
				"run",
				"./../../main.go",
				"--apikey=my-key",
				"client",
				"--endpoint=http://localhost:8500",
				"add",
				"--key='test event'",
				"--value=2",
				"--log=info",
			)

			_, err := cmd.CombinedOutput()

			assert.NoError(t, err, "Subprocess must not exit with status 1")
		})

		let("verify event with eventDigest", func(t *testing.T) {
			cmd := exec.Command("go",
				"run",
				"./../../main.go",
				"--apikey=my-key",
				"client",
				"--endpoint=http://localhost:8500",
				"membership",
				"--hyperDigest=81ae2d8f6ecec9c5837d12a09e3b42a1c880b6c77f81ff1f85aef36dac4fdf6a",
				"--historyDigest=0f5129eaf5dbfb1405ff072a04d716aaf4e4ba4247a3322c41582e970dbb7b00",
				"--version=0",
				"--eventDigest=8694718de4363adf07ec3b4aff4c76589f60fe89a7715bee7c8b250e06493922",
				"--log=info",
				"--verify",
			)

			stdoutStderr, err := cmd.CombinedOutput()

			assert.NoError(t, err, "Subprocess must not exit with status 1")
			assert.True(t, strings.Contains(fmt.Sprintf("%s", stdoutStderr), "Verify: OK"), "Must verify with eventDigest")

		})

		let("verify event with eventDigest", func(t *testing.T) {

			cmd := exec.Command("go",
				"run",
				"./../../main.go",
				"--apikey=my-key",
				"client",
				"--endpoint=http://localhost:8500",
				"membership",
				"--hyperDigest=81ae2d8f6ecec9c5837d12a09e3b42a1c880b6c77f81ff1f85aef36dac4fdf6a",
				"--historyDigest=0f5129eaf5dbfb1405ff072a04d716aaf4e4ba4247a3322c41582e970dbb7b00",
				"--version=0",
				"--key='test event'",
				"--log=info",
				"--verify",
			)

			stdoutStderr, err := cmd.CombinedOutput()

			assert.NoError(t, err, "Subprocess must not exit with status 1")
			assert.True(t, strings.Contains(fmt.Sprintf("%s", stdoutStderr), "Verify: OK"), "Must verify with eventDigest")

		})

	})
}
