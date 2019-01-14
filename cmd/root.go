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

// Package cli implements the command line commands qed and server.
package cmd

import (
	"fmt"

	"github.com/bbva/qed/log"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	v "github.com/spf13/viper"
)

// NewRootCommand is the main Parser for the qed cli.
func NewRootCommand() *cobra.Command {
	ctx := &cmdContext{}

	cmd := &cobra.Command{
		Use:              "qed",
		Short:            "QED is a client for the verifiable log server",
		TraverseChildren: true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			fmt.Println(">>>>>>>>>>>>>>>>>>>>>>>>>", "root.persistentprerun", ctx.path)

			if ctx.configFile != "" {
				v.SetConfigFile(ctx.configFile)
			} else {
				v.SetConfigName("config")
				v.AddConfigPath(ctx.path)
				v.AddConfigPath(".")
			}

			if !ctx.disableConfig {
				// read in environment variables that match.
				// ex: `QED_API_KEY=environ-key`
				v.SetEnvPrefix("QED")
				v.AutomaticEnv()

				err := v.ReadInConfig()
				if _, ok := err.(v.ConfigFileNotFoundError); err != nil && !ok {
					log.Error("Can't read config file.", err)
				}

				// Runtime Binding
				ctx.logLevel = v.GetString("log")
				ctx.apiKey = v.GetString("api_key")
				ctx.path, err = homedir.Expand(v.GetString("path"))
				if err != nil {
					log.Fatalf("Can't expand global path: %v", err)
				}

			}

			markStringRequired(ctx.apiKey, "apikey")

		},
	}

	f := cmd.PersistentFlags()
	f.StringVarP(&ctx.configFile, "config-file", "c", "", "Qed config file")
	f.BoolVarP(&ctx.disableConfig, "no-conf", "n", false, "Disable config file loading")
	f.StringVarP(&ctx.logLevel, "log", "l", "error", "Choose between log levels: silent, error, info and debug")
	f.StringVarP(&ctx.apiKey, "apikey", "k", "", "Server api key")
	f.StringVarP(&ctx.path, "path", "p", "/var/tmp/qed", "Qed root path for storage configuration and credentials")

	// Lookups
	v.BindPFlag("log", f.Lookup("log"))
	v.BindPFlag("api_key", f.Lookup("apikey"))
	v.BindPFlag("path", f.Lookup("path"))

	cmd.AddCommand(
		newStartCommand(ctx),
		newClientCommand(ctx),
		newAgentCommand(ctx),
	)

	return cmd
}
