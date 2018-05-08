// Copyright © 2018 Banco Bilbao Vizcaya Argentaria S.A.  All rights reserved.
// Use of this source code is governed by an Apache 2 License
// that can be found in the LICENSE file

package cli

import (
	"github.com/spf13/cobra"

	"verifiabledata/balloon/storage"
	"verifiabledata/log"
	"verifiabledata/server"
)

func NewServerCommand() *cobra.Command {
	var (
		logLevel, endpoint, dbPath, apiKey, storageName string
		cacheSize                                       uint64
	)

	cmd := &cobra.Command{
		Use:   "server",
		Short: "The server for the verifiable log QED",
		Long:  ``,
		// Args:  cobra.NoArgs(),

		Run: func(cmd *cobra.Command, args []string) {

			log.SetLogger("QedServer", logLevel)

			s := server.NewServer(
				endpoint,
				dbPath,
				apiKey,
				cacheSize,
				storageName,
			)

			err := s.ListenAndServe()
			if err != nil {
				log.Errorf("Can't start HTTP Server: ", err)
			}

		},
	}

	cmd.Flags().StringVarP(&apiKey, "apikey", "k", "", "Server api key")
	cmd.Flags().StringVarP(&endpoint, "endpoint", "e", "0.0.0.0:8080", "Endpoint for REST requests on (host:port)")
	cmd.Flags().StringVarP(&dbPath, "path", "p", "/tmp/balloon.db", "Set default storage path.")
	cmd.Flags().Uint64VarP(&cacheSize, "cache", "c", storage.SIZE25, "Initialize and reserve custom cache size.")
	cmd.Flags().StringVarP(&storageName, "storage", "s", "badger", "Choose between different storage backends. Eg badge|bolt")
	cmd.Flags().StringVarP(&logLevel, "log", "l", "error", "Choose between log levels: silent, error, info and debug")

	cmd.MarkFlagRequired("apikey")

	return cmd
}
