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

func newClientCommand(ctx *cmdContext) *cobra.Command {
	clientCtx := &clientContext{}

	cmd := &cobra.Command{
		Use:   "client",
		Short: "Client mode for qed",
		Long:  `Client process for emitting events to a qed server`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			log.SetLogger("QedClient", ctx.logLevel)

			clientCtx.client = client.NewHTTPClient(client.Config{
				Endpoint: clientCtx.endpoint,
				APIKey:   ctx.apiKey,
				Insecure: clientCtx.insecure,
			})
		},
		TraverseChildren: true,
	}

	cmd.PersistentFlags().StringVarP(&clientCtx.endpoint, "endpoint", "e", "localhost:8080", "Endpoint for REST requests on (host:port)")
	cmd.PersistentFlags().BoolVar(&clientCtx.insecure, "insecure", false, "Disable TLS transport")

	cmd.AddCommand(newAddCommand(clientCtx))
	cmd.AddCommand(newMembershipCommand(clientCtx))
	cmd.AddCommand(newIncrementalCommand(clientCtx))

	return cmd
}
