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
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/bbva/qed/log"
	"github.com/bbva/qed/server"
	"github.com/bbva/qed/util"
)

var serverStart *cobra.Command = &cobra.Command{
	Use:   "start",
	Short: "Stars QED Log service",
	Run:   runServerStart,
}

func init() {
	serverCmd.AddCommand(serverStart)
}

func runServerStart(cmd *cobra.Command, args []string) {
	var err error

	conf := serverCtx.Value(k("server.config")).(*server.Config)

	if conf.SSLCertificate != "" && conf.SSLCertificateKey != "" {
		if _, err := os.Stat(conf.SSLCertificate); os.IsNotExist(err) {
			log.Infof("Can't find certificate .crt file: %v", err)
		} else if _, err := os.Stat(conf.SSLCertificateKey); os.IsNotExist(err) {
			log.Infof("Can't find certificate .key file: %v", err)
		} else {
			log.Info("EnabledTLS")
			conf.EnableTLS = true
		}
	}

	log.SetLogger("server", conf.Log)
	fmt.Printf("CONF: %+v\n", conf)
	srv, err := server.NewServer(conf)
	if err != nil {
		log.Fatalf("Can't start QED server: %v", err)
	}

	err = srv.Start()
	if err != nil {
		log.Fatalf("Can't start QED server: %v", err)
	}

	util.AwaitTermSignal(srv.Stop)

	log.Debug("Stopping server, about to exit...")

}
