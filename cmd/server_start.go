/*
   Copyright 2018-2019 Banco Bilbao Vizcaya Argentaria, S.A.

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
	"os"

	"github.com/spf13/cobra"

	"github.com/bbva/qed/log2"
	"github.com/bbva/qed/server"
	"github.com/bbva/qed/util"
)

var serverStart *cobra.Command = &cobra.Command{
	Use:   "start",
	Short: "Stars QED Log service",
	RunE:  runServerStart,
}

func init() {
	serverCmd.AddCommand(serverStart)
}

func runServerStart(cmd *cobra.Command, args []string) error {
	var err error

	conf := serverCtx.Value(k("server.config")).(*server.Config)

	// create main logger
	logOpts := &log2.LoggerOptions{
		Name:            "qed",
		IncludeLocation: true,
		Level:           log2.LevelFromString(conf.Log),
		Output:          log2.DefaultOutput,
		TimeFormat:      log2.DefaultTimeFormat,
	}
	logger := log2.New(logOpts)

	// URL parse
	err = checkServerParams(conf)
	if err != nil {
		logger.Fatalf("Wrong parameters: %v", err)
	}

	if conf.SSLCertificate != "" && conf.SSLCertificateKey != "" {
		if _, err := os.Stat(conf.SSLCertificate); os.IsNotExist(err) {
			logger.Fatalf("Can't find certificate .crt file: %v", err)
		} else if _, err := os.Stat(conf.SSLCertificateKey); os.IsNotExist(err) {
			logger.Fatalf("Can't find certificate .key file: %v", err)
		} else {
			logger.Info("TLS enabled")
			conf.EnableTLS = true
		}
	}

	logger.Infof("Server configuration: \n%+v", conf)

	srv, err := server.NewServerWithLogger(conf, logger.Named("server"))
	if err != nil {
		logger.Fatalf("Can't create QED server: %v", err)
	}

	err = srv.Start()
	if err != nil {
		logger.Fatalf("Can't start QED server: %v", err)
	}

	util.AwaitTermSignal(srv.Stop)

	logger.Info("Stopping server, about to exit...")

	return nil
}

func checkServerParams(conf *server.Config) error {
	var err error

	err = urlParseNoSchemaRequired(conf.GossipAddr, conf.HTTPAddr, conf.MetricsAddr, conf.MgmtAddr, conf.ProfilingAddr, conf.RaftAddr)
	if err != nil {
		return err
	}

	err = urlParseNoSchemaRequired(conf.GossipJoinAddr...)
	if err != nil {
		return err
	}

	err = urlParseNoSchemaRequired(conf.RaftJoinAddr...)
	if err != nil {
		return err
	}

	return nil
}
