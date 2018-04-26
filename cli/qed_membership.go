package cli

import "github.com/spf13/cobra"

func newMembershipCommand() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "membership",
		Short: "Query for membership",
		Long: `Query for membership of an event to the authenticated data structure.
			It also verifies the proofs provided by the server`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	return cmd
}
