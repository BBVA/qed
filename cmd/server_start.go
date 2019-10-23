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

	"github.com/bbva/qed/log"
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
	logOpts := &log.LoggerOptions{
		Name:            "qed",
		IncludeLocation: true,
		Level:           log.LevelFromString(conf.Log),
		Output:          log.DefaultOutput,
		TimeFormat:      log.DefaultTimeFormat,
	}
	log.SetDefault(log.New(logOpts))

	// URL parse
	err = checkServerParams(conf)
	if err != nil {
		log.L().Fatalf("Wrong parameters: %v", err)
	}

	if conf.EnableTLS && conf.TLSCertPath != "" && conf.TLSKeyPath != "" {
		if _, err := os.Stat(conf.TLSCertPath); os.IsNotExist(err) {
			log.L().Fatalf("Can't find certificate .crt file: %v", err)
		} else if _, err := os.Stat(conf.TLSKeyPath); os.IsNotExist(err) {
			log.L().Fatalf("Can't find certificate .key file: %v", err)
		} else {
			log.L().Info("TLS enabled")
		}
	}

	log.L().Infof("Server configuration: \n%+v", conf)

	srv, err := server.NewServerWithLogger(conf, log.L().Named("server"))
	if err != nil {
		log.L().Fatalf("Can't create QED server: %v", err)
	}

	err = srv.Start()
	if err != nil {
		log.L().Fatalf("Can't start QED server: %v", err)
	}

	util.AwaitTermSignal(srv.Stop)

	log.L().Info("Stopping server, about to exit...")

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
