// Copyright © 2018 Banco Bilbao Vizcaya Argentaria S.A.  All rights reserved.
// Use of this source code is governed by an Apache 2 License
// that can be found in the LICENSE file

package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newAddCommand(ctx *Context) *cobra.Command {

	var key, value string

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add an event",
		Long:  `Add an event to the authenticated data structure`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Adding key [ %s ] with value [ %s ]\n", key, value)

			snapshot, err := ctx.client.Add(key)
			if err != nil {
				return err
			}

			fmt.Printf("Received snapshot with values: \n\tEvent: %s\n\tHyperDigest: %x\n\tHistoryDigest: %x\n\tVersion: %d\n",
				snapshot.Event, snapshot.HyperDigest, snapshot.HistoryDigest, snapshot.Version)
			// ctx.Logger().Printf("Reponse Status: %s\n", resp.Status)
			// ctx.Logger().Printf("Reponse Headers: %v\n", resp.Header)
			// body, _ := ioutil.ReadAll(resp.Body)
			// ctx.Logger().Printf("Reponse Body: %v\n", string(body))
			return nil
		},
	}

	cmd.Flags().StringVar(&key, "key", "", "Key to add")
	cmd.Flags().StringVar(&value, "value", "", "Value to add")
	cmd.MarkFlagRequired("key")
	cmd.MarkFlagRequired("value")

	return cmd
}
