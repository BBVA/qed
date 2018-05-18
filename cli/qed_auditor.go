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

package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/bbva/qed/api/apihttp"
)

func newAuditorCommand(ctx *Context) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "auditor",
		Short: "Auditor mode for qed",
		Long:  `Auditor process that verifies commitments and events from a qed server`,
		RunE: func(cmd *cobra.Command, args []string) error {

			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				raw := scanner.Text()
				snapshot := &apihttp.Snapshot{}
				json.Unmarshal([]byte(raw), snapshot)

				proof, _ := ctx.client.Membership(snapshot.Event, snapshot.Version)

				if ctx.client.Verify(proof, snapshot) {
					fmt.Println("Verify: OK")
				} else {
					fmt.Printf("Verify: KO, raw: %s\n", raw)
				}
			}

			return nil
		},
	}

	return cmd
}
