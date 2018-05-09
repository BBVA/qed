package cli

import (
	"github.com/spf13/cobra"

	"qed/log"
)

func newAddCommand(ctx *Context) *cobra.Command {

	var key, value string

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add an event",
		Long:  `Add an event to the authenticated data structure`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Infof("Adding key [ %s ] with value [ %s ]\n", key, value)

			snapshot, err := ctx.client.Add(key)
			if err != nil {
				return err
			}

			log.Infof("Received snapshot with values: \n\tEvent: %s\n\tHyperDigest: %x\n\tHistoryDigest: %x\n\tVersion: %d\n",
				snapshot.Event, snapshot.HyperDigest, snapshot.HistoryDigest, snapshot.Version)

			return nil
		},
	}

	cmd.Flags().StringVar(&key, "key", "", "Key to add")
	cmd.Flags().StringVar(&value, "value", "", "Value to add")
	cmd.MarkFlagRequired("key")
	cmd.MarkFlagRequired("value")

	return cmd
}
