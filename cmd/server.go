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
	"fmt"

	"github.com/bbva/qed/server"
	"github.com/octago/sflags/gen/gpflag"
	"github.com/spf13/cobra"
)

var serverCmd *cobra.Command = &cobra.Command{
	Use:   "server",
	Short: "Provides access to the QED log server commands",
	Long: `QED server provides a REST API to the QED Log. The API is documented
elsewhere.`,
	TraverseChildren: true,
}

var serverCtx context.Context = configServer()

func init() {
	Root.AddCommand(serverCmd)
}

func configServer() context.Context {

	conf := server.DefaultConfig()

	err := gpflag.ParseTo(conf, serverCmd.PersistentFlags())
	if err != nil {
		panic(fmt.Sprintf("Unable to parse server config: %v", err))
	}
	return context.WithValue(Ctx, k("server.config"), conf)
}
