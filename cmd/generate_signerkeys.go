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

	"github.com/bbva/qed/crypto"
	"github.com/spf13/cobra"
)

var generateSignerKeys *cobra.Command = &cobra.Command{
	Use:   "signerkeys",
	Short: "Generate Signer Keys",
	RunE:  runGenerateSignerKeys,
}

func init() {
	generateCmd.AddCommand(generateSignerKeys)
}

func runGenerateSignerKeys(cmd *cobra.Command, args []string) error {

	conf := generateCtx.Value(k("generate.config")).(*GenerateConfig)

	pubKey, priKey, err := crypto.NewEd25519SignerKeysFile(conf.Path)
	if err != nil {
		return err
	}
	fmt.Printf("New signer keys generated at:\n%v\n%v\n", pubKey, priKey)

	return nil
}
