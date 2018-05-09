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
	"github.com/spf13/cobra"

	"qed/client"
	"qed/log"
)

func NewQedCommand(ctx *Context) *cobra.Command {
	var (
		endpoint, apiKey, logLevel string
	)

	cmd := &cobra.Command{
		Use:       "qed",
		Short:     "QED is a client for the verifiable log server",
		Long:      `blah blah`,
		ValidArgs: []string{"add", "verify"},
		Args: func(cmd *cobra.Command, args []string) error {
			err1 := cobra.MinimumNArgs(1)(cmd, args)
			if err1 != nil {
				return err1
			}
			err2 := cobra.OnlyValidArgs(cmd, args)
			return err2
		},
		PersistentPreRun: func(cmd *cobra.Command, args []string) {

			log.SetLogger("QedClient", logLevel)
			ctx.client = client.NewHttpClient(endpoint, apiKey)

		},
		TraverseChildren: true,
	}

	cmd.PersistentFlags().StringVarP(&logLevel, "log", "l", "error", "Choose between log levels: silent, error, info and debug")
	cmd.PersistentFlags().StringVarP(&endpoint, "endpoint", "e", "", "Server endpoint")
	cmd.PersistentFlags().StringVarP(&apiKey, "apikey", "k", "", "Server api key")
	cmd.MarkPersistentFlagRequired("endpoint")
	cmd.MarkPersistentFlagRequired("apikey")

	cmd.AddCommand(newClientCommand(ctx))
	cmd.AddCommand(newAuditorCommand(ctx))

	cmd.AddCommand(newAddCommand(ctx))
	cmd.AddCommand(newMembershipCommand(ctx))

	return cmd
}
