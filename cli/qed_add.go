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
