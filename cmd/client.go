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

package cmd

import (
	"github.com/spf13/cobra"
	v "github.com/spf13/viper"

	"github.com/bbva/qed/client"
	"github.com/bbva/qed/log"
)

func newClientCommand(ctx *cmdContext) *cobra.Command {
	clientCtx := &clientContext{}
	clientCtx.config = client.DefaultConfig()

	cmd := &cobra.Command{
		Use:   "client",
		Short: "Client mode for qed",
		Long:  `Client process for emitting events to a qed server`,
	}

	f := cmd.PersistentFlags()
	f.StringVarP(&clientCtx.config.Endpoint, "endpoint", "e", "127.0.0.1:8080", "Endpoint for REST requests on (host:port)")
	f.BoolVar(&clientCtx.config.Insecure, "insecure", false, "Allow self signed certificates")
	f.IntVar(&clientCtx.config.TimeoutSeconds, "timeout-seconds", 10, "Seconds to cut the connection")
	f.IntVar(&clientCtx.config.DialTimeoutSeconds, "dial-timeout-seconds", 5, "Seconds to cut the dialing")
	f.IntVar(&clientCtx.config.HandshakeTimeoutSeconds, "handshake-timeout-seconds", 5, "Seconds to cut the handshaking")

	// Lookups
	v.BindPFlag("client.endpoint", f.Lookup("endpoint"))
	v.BindPFlag("client.insecure", f.Lookup("insecure"))
	v.BindPFlag("client.timeout.connection", f.Lookup("timeout-seconds"))
	v.BindPFlag("client.timeout.dial", f.Lookup("dial-timeout-seconds"))
	v.BindPFlag("client.timeout.handshake", f.Lookup("handshake-timeout-seconds"))

	clientPreRun := func(cmd *cobra.Command, args []string) {
		log.SetLogger("QEDClient", ctx.logLevel)

		clientCtx.config.APIKey = ctx.apiKey
		clientCtx.config.Endpoint = v.GetString("client.endpoint")
		clientCtx.config.Insecure = v.GetBool("client.insecure")
		clientCtx.config.TimeoutSeconds = v.GetInt("client.timeout.connection")
		clientCtx.config.DialTimeoutSeconds = v.GetInt("client.timeout.dial")
		clientCtx.config.HandshakeTimeoutSeconds = v.GetInt("client.timeout.handshake")

		clientCtx.client = client.NewHTTPClient(*clientCtx.config)

	}

	cmd.AddCommand(
		newAddCommand(clientCtx, clientPreRun),
		newMembershipCommand(clientCtx, clientPreRun),
		newIncrementalCommand(clientCtx, clientPreRun),
	)

	return cmd
}
