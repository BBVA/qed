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
	"context"

	"github.com/bbva/qed/log"
	"github.com/octago/sflags/gen/gpflag"
	"github.com/spf13/cobra"
)

type GenerateConfig struct {
	Path     string // Path to the private key file used to sign snapshots.
	Hostname string // Hostname of IPAddr for which the certificates will be generated.
}

func GenerateDefaultConfig() *GenerateConfig {
	return &GenerateConfig{
		Path:     "/var/tmp",
		Hostname: "127.0.0.1",
	}
}

var generateCmd *cobra.Command = &cobra.Command{
	Use:   "generate",
	Short: "Generates configuration files,keys and cerificates for QED",
	Long: `This command generates config files, keys and certificates
required to run QED server.`,
	TraverseChildren: true,
}

var generateCtx context.Context = generateConfig()

func init() {
	Root.AddCommand(generateCmd)
}

func generateConfig() context.Context {

	conf := GenerateDefaultConfig()

	err := gpflag.ParseTo(conf, generateCmd.PersistentFlags())
	if err != nil {
		log.Fatalf("err: %v", err)
	}
	return context.WithValue(Ctx, k("generate.config"), conf)
}
