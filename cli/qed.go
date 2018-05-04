// Copyright © 2018 Banco Bilbao Vizcaya Argentaria S.A.  All rights reserved.
// Use of this source code is governed by an Apache 2 License
// that can be found in the LICENSE file

package cli

import (
	"os"
	"verifiabledata/client"
	"verifiabledata/log"

	"github.com/spf13/cobra"
)

func NewQedCommand(ctx *Context) *cobra.Command {
	var endpoint string
	var apikey string
	var verbose int
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
			//ctx.viper.Set("verbose", verbose)
			//ctx.viper.Set("apikey", apikey)
			//ctx.viper.Set("endpoint", endpoint)
			var logger log.Logger
			if verbose == 1 {
				logger = log.NewInfo(os.Stdout, "QedClient", log.Ldate|log.Ltime|log.Lmicroseconds|log.Llongfile)
			} else if verbose > 1 {
				logger = log.NewDebug(os.Stdout, "QedClient", log.Ldate|log.Ltime|log.Lmicroseconds|log.Llongfile)
			}
			ctx.client = client.NewHttpClient(endpoint, apikey, logger)
		},
		TraverseChildren: true,
	}

	cmd.PersistentFlags().CountVarP(&verbose, "verbose", "v", "verbosity (-v or -vv)")
	cmd.PersistentFlags().StringVarP(&endpoint, "endpoint", "e", "", "Server endpoint")
	cmd.PersistentFlags().StringVarP(&apikey, "apikey", "k", "", "Server api key")
	cmd.MarkPersistentFlagRequired("endpoint")
	cmd.MarkPersistentFlagRequired("apikey")

	cmd.AddCommand(newClientCommand(ctx))
	cmd.AddCommand(newAuditorCommand(ctx))

	cmd.AddCommand(newAddCommand(ctx))
	cmd.AddCommand(newMembershipCommand(ctx))
	return cmd
}
