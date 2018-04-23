package cli

import (
	"github.com/spf13/cobra"
)

func NewQedCommand() *cobra.Command {
	//var verbose int
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
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	//cmd.PersistentFlags().CountVarP(&verbose, "", "v", "verbosity (-v or -vv)")
	cmd.AddCommand(newAddCommand())
	return cmd
}
