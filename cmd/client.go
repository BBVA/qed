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

package cmd

import (
	"github.com/bbva/qed/client"
	"github.com/bbva/qed/log"
	"github.com/spf13/cobra"
)

var clientEndpoint string

func newClientCommand() *cobra.Command {

	ctx := newClientContext()

	cmd := &cobra.Command{
		Use:   "client",
		Short: "Client mode for qed",
		Long:  `Client process for emitting events to a qed server`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			log.SetLogger("QedClient", logLevel)
			ctx.client = client.NewHttpClient(clientEndpoint, apiKey)
		},
		TraverseChildren: true,
	}

	cmd.PersistentFlags().StringVarP(
		&clientEndpoint,
		"endpoint",
		"e",
		"http://localhost:8080",
		"Endpoint for REST requests on (host:port)",
	)

	cmd.AddCommand(newAddCommand(ctx))
	cmd.AddCommand(newMembershipCommand(ctx))

	return cmd
}
