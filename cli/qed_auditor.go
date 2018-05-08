// Copyright © 2018 Banco Bilbao Vizcaya Argentaria S.A.  All rights reserved.
// Use of this source code is governed by an Apache 2 License
// that can be found in the LICENSE file

package cli

import (
	"bufio"
	"encoding/json"
	"os"

	"github.com/spf13/cobra"

	"verifiabledata/api/apihttp"

	"verifiabledata/log"
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
				correct := ctx.client.Verify(proof, snapshot)

				if ctx.client.Verify(proof, snapshot) {
					log.Info("Verify: OK")
				} else {
					log.Errorf("Verify: KO, raw: %s", raw)
				}
			}

			return nil
		},
	}

	return cmd
}
