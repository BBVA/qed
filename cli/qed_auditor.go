// Copyright © 2018 Banco Bilbao Vizcaya Argentaria S.A.  All rights reserved.
// Use of this source code is governed by an Apache 2 License
// that can be found in the LICENSE file

package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"verifiabledata/api/apihttp"
)

func newAuditorCommand(ctx *Context) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "auditor",
		Short: "Auditor mode for qed",
		Long:  `Auditor process that verifies commitments and events from a qed server`,
		RunE: func(cmd *cobra.Command, args []string) error {

			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				snapshot := &apihttp.Snapshot{}
				json.Unmarshal([]byte(scanner.Text()), snapshot)

				// fmt.Println(snapshot.Version)
				proof, _ := ctx.client.Membership(snapshot.Event, snapshot.Version)
				correct := ctx.client.Verify(proof, snapshot)
				fmt.Println(correct)
			}

			return nil
		},
	}

	return cmd
}
