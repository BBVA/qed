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
	// "math/rand"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/bbva/qed/testutils/scope"
	"github.com/stretchr/testify/require"
)

func Test_Client_To_Single_Server(t *testing.T) {
	before, after := setupServer(0, "", true, t)
	scenario, let := scope.Scope(t, before, merge(after, delay(2*time.Second)))

	scenario("Add one event through cli and verify it", func() {

		let("Add event", func(t *testing.T) {
			cmd := exec.Command("go",
				"run",
				"./../../main.go",
				fmt.Sprintf("--apikey=%s", APIKey),
				"client",
				fmt.Sprintf("--endpoints=%s", QEDTLS),
				"add",
				"--key='test event'",
				"--value=2",
				"--log=info",
				"--insecure",
			)

			_, err := cmd.CombinedOutput()

			require.NoErrorf(t, err, "Subprocess must not exit with status 1: %v", *cmd)
		})

		let("Verify event with eventDigest", func(t *testing.T) {
			cmd := exec.Command("go",
				"run",
				"./../../main.go",
				fmt.Sprintf("--apikey=%s", APIKey),
				"client",
				fmt.Sprintf("--endpoints=%s", QEDTLS),
				"membership",
				"--hyperDigest=2b0096433f25f59a4310af8a8999abfd9beb6adacbbbb086251f24609c4f5bbf",
				"--historyDigest=0f5129eaf5dbfb1405ff072a04d716aaf4e4ba4247a3322c41582e970dbb7b00",
				"--version=0",
				"--eventDigest=8694718de4363adf07ec3b4aff4c76589f60fe89a7715bee7c8b250e06493922",
				"--log=info",
				"--verify",
				"--insecure",
			)

			stdoutStderr, err := cmd.CombinedOutput()

			require.NoError(t, err, "Subprocess must not exit with status 1")
			require.True(t, strings.Contains(fmt.Sprintf("%s", stdoutStderr), "Verify: OK"), "Must verify with eventDigest")
		})

		let("Verify event with event", func(t *testing.T) {

			cmd := exec.Command("go",
				"run",
				"./../../main.go",
				fmt.Sprintf("--apikey=%s", APIKey),
				"client",
				fmt.Sprintf("--endpoints=%s", QEDTLS),
				"membership",
				"--hyperDigest=2b0096433f25f59a4310af8a8999abfd9beb6adacbbbb086251f24609c4f5bbf",
				"--historyDigest=0f5129eaf5dbfb1405ff072a04d716aaf4e4ba4247a3322c41582e970dbb7b00",
				"--version=0",
				"--key='test event'",
				"--log=info",
				"--verify",
				"--insecure",
			)

			stdoutStderr, err := cmd.CombinedOutput()

			require.NoError(t, err, "Subprocess must not exit with status 1")
			require.True(t, strings.Contains(fmt.Sprintf("%s", stdoutStderr), "Verify: OK"), "Must verify with eventDigest")
		})

	})
}

func Test_Client_To_Cluster_With_Leader_Change(t *testing.T) {
	before0, after0 := setupServer(0, "", false, t)
	before1, after1 := setupServer(1, "", false, t)
	before2, after2 := setupServer(2, "", false, t)

	serversHttpAddr := "http://127.0.0.1:8800"

	scenario, let := scope.Scope(t, merge(before0, before1, before2), merge(after1, after2, delay(2*time.Second)))

	scenario("Add one event through cli and verify it", func() {
		let("Add event", func(t *testing.T) {
			cmd := exec.Command("go",
				"run",
				"./../../main.go",
				fmt.Sprintf("--apikey=%s", APIKey),
				"client",
				fmt.Sprintf("--endpoints=%s", serversHttpAddr),
				"add",
				"--key='test event'",
				"--value=2",
				"--log=info",
			)

			_, err := cmd.CombinedOutput()

			require.NoErrorf(t, err, "Subprocess must not exit with status 1: %v", *cmd)
		})

		let("Verify event with eventDigest", func(t *testing.T) {
			cmd := exec.Command("go",
				"run",
				"./../../main.go",
				fmt.Sprintf("--apikey=%s", APIKey),
				"client",
				fmt.Sprintf("--endpoints=%s", serversHttpAddr),
				"membership",
				"--hyperDigest=2b0096433f25f59a4310af8a8999abfd9beb6adacbbbb086251f24609c4f5bbf",
				"--historyDigest=0f5129eaf5dbfb1405ff072a04d716aaf4e4ba4247a3322c41582e970dbb7b00",
				"--version=0",
				"--eventDigest=8694718de4363adf07ec3b4aff4c76589f60fe89a7715bee7c8b250e06493922",
				"--log=info",
				"--verify",
			)

			stdoutStderr, err := cmd.CombinedOutput()

			require.NoError(t, err, "Subprocess must not exit with status 1")
			require.True(t, strings.Contains(fmt.Sprintf("%s", stdoutStderr), "Verify: OK"), "Must verify with eventDigest")
		})

		let("Verify event with event", func(t *testing.T) {

			cmd := exec.Command("go",
				"run",
				"./../../main.go",
				fmt.Sprintf("--apikey=%s", APIKey),
				"client",
				fmt.Sprintf("--endpoints=%s", serversHttpAddr),
				"membership",
				"--hyperDigest=2b0096433f25f59a4310af8a8999abfd9beb6adacbbbb086251f24609c4f5bbf",
				"--historyDigest=0f5129eaf5dbfb1405ff072a04d716aaf4e4ba4247a3322c41582e970dbb7b00",
				"--version=0",
				"--key='test event'",
				"--log=info",
				"--verify",
			)

			stdoutStderr, err := cmd.CombinedOutput()

			require.NoError(t, err, "Subprocess must not exit with status 1")
			require.True(t, strings.Contains(fmt.Sprintf("%s", stdoutStderr), "Verify: OK"), "Must verify with eventDigest")

		})

		let("Kill server 0", func(t *testing.T) {
			after0(t)
			serversHttpAddr = "http://127.0.0.1:8801"

			// Need time to propagate changes via RAFT.
			time.Sleep(10 * time.Second)
		})

		let("Add second event", func(t *testing.T) {
			cmd := exec.Command("go",
				"run",
				"./../../main.go",
				fmt.Sprintf("--apikey=%s", APIKey),
				"client",
				fmt.Sprintf("--endpoints=%s", serversHttpAddr),
				"add",
				"--key='test event 2'",
				"--value=2",
				"--log=info",
			)

			_, err := cmd.CombinedOutput()

			require.NoErrorf(t, err, "Subprocess must not exit with status 1: %v", *cmd)
		})

	})
}

func Test_Client_To_Cluster_With_Bad_Endpoint(t *testing.T) {
	before0, after0 := setupServer(0, "", false, t)
	before1, after1 := setupServer(1, "", false, t)

	scenario, let := scope.Scope(t, merge(before0, before1), merge(after0, after1, delay(2*time.Second)))

	scenario("Success by extracting topology from right endpoint", func() {

		let("Add event with one valid endpoint", func(t *testing.T) {
			cmd := exec.Command("go",
				"run",
				"./../../main.go",
				fmt.Sprintf("--apikey=%s", APIKey),
				"client",
				fmt.Sprintf("--endpoints=badendpoint,http://127.0.0.1:8800"),
				"add",
				"--key='test event'",
				"--value=2",
				"--log=info",
			)

			_, err := cmd.CombinedOutput()

			require.NoErrorf(t, err, "Subprocess must not exit with status 1: %v", *cmd)
		})

		let("Add event with no valid endpoint and fail", func(t *testing.T) {
			cmd := exec.Command("go",
				"run",
				"./../../main.go",
				fmt.Sprintf("--apikey=%s", APIKey),
				"client",
				fmt.Sprintf("--endpoints=badendpoint"),
				"add",
				"--key='test event'",
				"--value=2",
				"--log=info",
			)

			_, err := cmd.CombinedOutput()

			require.Errorf(t, err, "Subprocess must exit with status 1: %v", *cmd)
		})

	})

}

func Test_Client_To_Cluster_Continuous_Load_Node_Fails(t *testing.T) {
	before0, after0 := setupServer(0, "", false, t)
	before1, after1 := setupServer(1, "", false, t)

	serversHttpAddr := "http://127.0.0.1:8800,http://127.0.0.1:8801"

	scenario, let := scope.Scope(t, merge(before0, before1), merge(after1, delay(2*time.Second)))

	scenario("Success by extracting topology from right endpoint", func() {
		let("Add event", func(t *testing.T) {
			cmd := exec.Command("go",
				"run",
				"./../../main.go",
				fmt.Sprintf("--apikey=%s", APIKey),
				"client",
				fmt.Sprintf("--endpoints=%s", serversHttpAddr),
				"add",
				"--key='test event'",
				"--value=2",
				"--log=info",
			)

			_, err := cmd.CombinedOutput()

			require.NoErrorf(t, err, "Subprocess must not exit with status 1: %v", *cmd)
		})

		let("Kill server 0", func(t *testing.T) {
			after0(t)
			serversHttpAddr = "http://127.0.0.1:8081"

			// Need time to propagate changes via RAFT.
			time.Sleep(10 * time.Second)
		})

	})
}
