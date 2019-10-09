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

package e2e

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"

	"github.com/bbva/qed/testutils/spec"
)

func TestClientToSingleServer(t *testing.T) {
	b0, a0 := newServerSetup(0, true)
	let, report := spec.New()
	defer func() {
		a0()
		t.Logf(report())
	}()

	_, err := b0()
	spec.NoError(t, err, "Error starting server")

	let(t, "Add one event through cli and verify it", func(t *testing.T) {

		let(t, "Add event", func(t *testing.T) {

			cmd := exec.Command("go",
				"run",
				"./../../main.go",
				"client",
				"--endpoints=https://127.0.0.1:8800",
				"add",
				"--event='test event'",
				"--attempt-to-revive-endpoints",
				"--max-retries=3",
				"--log=error",
				"--insecure",
			)

			output, err := cmd.CombinedOutput()

			fmt.Println(string(output))

			spec.NoError(t, err, "Client returned error")
		})

		let(t, "Verify membership proof with an event digest", func(t *testing.T) {
			cmd := exec.Command("go",
				"run",
				"./../../main.go",
				"client",
				"--endpoints=https://127.0.0.1:8800",
				"membership",
				"--event-digest=8694718de4363adf07ec3b4aff4c76589f60fe89a7715bee7c8b250e06493922",
				"--hyper-digest=2b0096433f25f59a4310af8a8999abfd9beb6adacbbbb086251f24609c4f5bbf",
				"--history-digest=0f5129eaf5dbfb1405ff072a04d716aaf4e4ba4247a3322c41582e970dbb7b00",
				"--verify",
				"--version=0",
				"--attempt-to-revive-endpoints",
				"--max-retries=3",
				"--log=error",
				"--insecure",
			)

			output, err := cmd.CombinedOutput()

			fmt.Println(string(output))

			spec.NoError(t, err, "Client returned an error")
			// This check depends on the client print format
			// which makes it fragile.
			// IF the client returns a 0, the the command is succesfull and no
			// furhter check should be needed.
			spec.True(t, strings.Contains(string(output), "Verify: OK"), "Must verify with eventDigest")
		})

		let(t, "Verify membership proof with a plain event", func(t *testing.T) {

			cmd := exec.Command("go",
				"run",
				"./../../main.go",
				"client",
				"--endpoints=https://127.0.0.1:8800",
				"membership",
				"--hyper-digest=2b0096433f25f59a4310af8a8999abfd9beb6adacbbbb086251f24609c4f5bbf",
				"--history-digest=0f5129eaf5dbfb1405ff072a04d716aaf4e4ba4247a3322c41582e970dbb7b00",
				"--version=0",
				"--event='test event'",
				"--log=info",
				"--verify",
				"--insecure",
			)

			stdoutStderr, err := cmd.CombinedOutput()

			spec.NoError(t, err, "Subprocess must not exit with status 1")
			spec.True(t, strings.Contains(fmt.Sprintf("%s", stdoutStderr), "Verify: OK"), "Must verify with eventDigest")
		})

	})
}

func TestClientToClusterWithLeaderChange(t *testing.T) {
	b0, a0 := newServerSetup(0, true)
	b1, a1 := newServerSetup(1, true)
	b2, a2 := newServerSetup(2, true)
	let, report := spec.New()
	defer func() {
		// a0()
		a1()
		a2()
		t.Logf(report())
	}()

	_, err := b0()
	spec.NoError(t, err, "Error starting node 0")
	_, err = b1()
	spec.NoError(t, err, "Error starting node 1")
	_, err = b2()
	spec.NoError(t, err, "Error starting node 2")

	let(t, "Add one event through cli and verify it", func(t *testing.T) {
		let(t, "Add event", func(t *testing.T) {

			cmd := exec.Command("go",
				"run",
				"./../../main.go",
				"client",
				"--endpoints=https://127.0.0.1:8800,https://127.0.0.1:8801,https://127.0.0.1:8802",
				"add",
				"--event='test event'",
				"--attempt-to-revive-endpoints",
				"--max-retries=3",
				"--log=error",
				"--insecure",
			)

			output, err := cmd.CombinedOutput()

			fmt.Println(string(output))

			spec.NoError(t, err, "Client returned error")
		})

		let(t, "Verify membership proof with an event digest", func(t *testing.T) {
			cmd := exec.Command("go",
				"run",
				"./../../main.go",
				"client",
				"--endpoints=https://127.0.0.1:8800,https://127.0.0.1:8801,https://127.0.0.1:8802",
				"membership",
				"--event-digest=8694718de4363adf07ec3b4aff4c76589f60fe89a7715bee7c8b250e06493922",
				"--hyper-digest=2b0096433f25f59a4310af8a8999abfd9beb6adacbbbb086251f24609c4f5bbf",
				"--history-digest=0f5129eaf5dbfb1405ff072a04d716aaf4e4ba4247a3322c41582e970dbb7b00",
				"--verify",
				"--version=0",
				"--attempt-to-revive-endpoints",
				"--max-retries=3",
				"--log=error",
				"--insecure",
			)

			output, err := cmd.CombinedOutput()

			fmt.Println(string(output))

			spec.NoError(t, err, "Client returned an error")
			// This check depends on the client print format
			// which makes it fragile.
			// IF the client returns a 0, the the command is succesfull and no
			// furhter check should be needed.
			spec.True(t, strings.Contains(string(output), "Verify: OK"), "Must verify with eventDigest")
		})

		let(t, "Verify membership proof with a plain event", func(t *testing.T) {

			cmd := exec.Command("go",
				"run",
				"./../../main.go",
				"client",
				"--endpoints=https://127.0.0.1:8800,https://127.0.0.1:8801,https://127.0.0.1:8802",
				"membership",
				"--hyper-digest=2b0096433f25f59a4310af8a8999abfd9beb6adacbbbb086251f24609c4f5bbf",
				"--history-digest=0f5129eaf5dbfb1405ff072a04d716aaf4e4ba4247a3322c41582e970dbb7b00",
				"--version=0",
				"--event='test event'",
				"--log=info",
				"--verify",
				"--insecure",
			)

			stdoutStderr, err := cmd.CombinedOutput()

			spec.NoError(t, err, "Subprocess must not exit with status 1")
			spec.True(t, strings.Contains(fmt.Sprintf("%s", stdoutStderr), "Verify: OK"), "Must verify with eventDigest")
		})

		let(t, "Shutdown server 0", func(t *testing.T) {
			err = a0()
			spec.NoError(t, err, "error stoping server")

		})

		let(t, "Add second event", func(t *testing.T) {

			cmd := exec.Command("go",
				"run",
				"./../../main.go",
				"client",
				"--endpoints=https://127.0.0.1:8800,https://127.0.0.1:8801,https://127.0.0.1:8802",
				"add",
				"--event='test event 2'",
				"--max-retries=5",
				"--log=error",
				"--insecure",
			)

			output, err := cmd.CombinedOutput()

			fmt.Println(string(output))

			spec.NoError(t, err, "Client returned error")
		})

	})
}

func TestClientToClusterWithBadEndpoint(t *testing.T) {
	b0, a0 := newServerSetup(0, true)
	b1, a1 := newServerSetup(1, true)

	let, report := spec.New()

	_, err := b0()
	spec.NoError(t, err, "Error starting node 0")
	_, err = b1()
	spec.NoError(t, err, "Error starting node 1")

	defer func() {
		a0()
		a1()
		t.Logf(report())
	}()

	let(t, "Success by extracting topology from right endpoint", func(t *testing.T) {

		let(t, "Add event with one valid endpoint", func(t *testing.T) {

			cmd := exec.Command("go",
				"run",
				"./../../main.go",
				"client",
				"--endpoints=https://127.0.0.1:8800",
				"add",
				"--event='test event'",
				"--attempt-to-revive-endpoints",
				"--max-retries=3",
				"--log=error",
				"--insecure",
			)

			output, err := cmd.CombinedOutput()

			fmt.Println(string(output))

			spec.NoError(t, err, "Client returned error")
		})

		let(t, "Add event with no valid endpoint and fail", func(t *testing.T) {

			cmd := exec.Command("go",
				"run",
				"./../../main.go",
				"client",
				"--endpoints=badendpoint",
				"add",
				"--event='test event'",
				"--attempt-to-revive-endpoints",
				"--max-retries=3",
				"--log=error",
				"--insecure",
			)

			output, err := cmd.CombinedOutput()

			fmt.Println(string(output))

			spec.Error(t, err, "Client must return error")
		})
	})
}
