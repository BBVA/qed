package cli

import (
	"verifiabledata/client"

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
			if verbose == 1 {
				ctx.Logger() // should set level to info
			} else if verbose > 1 {
				ctx.Logger() // should set level to debug
			}
			ctx.client = client.NewHttpClient(endpoint, apikey)
		},
		TraverseChildren: true,
	}

	cmd.PersistentFlags().CountVarP(&verbose, "verbose", "v", "verbosity (-v or -vv)")
	cmd.PersistentFlags().StringVarP(&endpoint, "endpoint", "e", "", "Server endpoint")
	cmd.PersistentFlags().StringVarP(&apikey, "apikey", "k", "", "Server api key")
	cmd.MarkPersistentFlagRequired("endpoint")
	cmd.MarkPersistentFlagRequired("apikey")

	cmd.AddCommand(newAddCommand(ctx))
	return cmd
}
