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
	"github.com/bbva/qed/crypto"
	"github.com/spf13/cobra"

	"github.com/bbva/qed/server"
)

var generateKeypair *cobra.Command = &cobra.Command{
	Use:   "keypair",
	Short: "Generate keypair",
	Run:   runGenerateKeypair,
}

func init() {
	generateCmd.AddCommand(generateKeypair)
}

func runGenerateKeypair(cmd *cobra.Command, args []string) {

	conf := generateCtx.Value(k("server.config")).(*server.GenerateConfig)

	crypto.NewEd25519KeyPairFile(conf.Path)

}
