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

var generateCsrRequest *cobra.Command = &cobra.Command{
	Use:   "csr",
	Short: "Generate Certificate Signing Request",
	RunE:  runCsr,
}

func init() {
	generateCmd.AddCommand(generateCsrRequest)
}

func runCsr(cmd *cobra.Command, args []string) error {
	var err error
	conf := generateCtx.Value(k("generate.config")).(*GenerateConfig)

	err = isValidFQDN(conf.Host)
	if err != nil {
		return fmt.Errorf("Invalid FQDN: %v", err)
	}

	cert, key, err := crypto.NewCsrRequest(conf.Path, conf.Host)
	if err != nil {
		return fmt.Errorf("Error: %v", err)
	}
	fmt.Printf("New Certificate Signing Request and Private Key generated at:\n%v\n%v\n", cert, key)

	return nil
}
