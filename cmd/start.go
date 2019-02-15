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
	"fmt"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	v "github.com/spf13/viper"

	"github.com/bbva/qed/log"
	"github.com/bbva/qed/server"
)

func newStartCommand(ctx *cmdContext) *cobra.Command {
	conf := server.DefaultConfig()

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the server for the verifiable log QED",
		Run: func(cmd *cobra.Command, args []string) {
			var err error

			log.SetLogger("QEDServer", ctx.logLevel)

			// Bindings
			conf.APIKey = ctx.apiKey
			conf.NodeID = v.GetString("server.node-id")
			conf.EnableProfiling = v.GetBool("server.profiling")
			conf.PrivateKeyPath, _ = homedir.Expand(v.GetString("server.key"))
			conf.SSLCertificate, _ = homedir.Expand(v.GetString("server.tls.certificate"))
			conf.SSLCertificateKey, _ = homedir.Expand(v.GetString("server.tls.certificate_key"))
			conf.HTTPAddr = v.GetString("server.addr.http")
			conf.RaftAddr = v.GetString("server.addr.raft")
			conf.MgmtAddr = v.GetString("server.addr.mgmt")
			conf.MetricsAddr = v.GetString("server.addr.metrics")
			conf.RaftJoinAddr = v.GetStringSlice("server.addr.raft_join")
			conf.GossipAddr = v.GetString("server.addr.gossip")
			conf.GossipJoinAddr = v.GetStringSlice("server.addr.gossip_join")
			conf.DBPath = fmt.Sprintf("%s/%s", ctx.path, "db")
			conf.RaftPath = fmt.Sprintf("%s/%s", ctx.path, "wal")

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

			// cmd.DisableSuggestions = true
			srv, err := server.NewServer(conf)
			if err != nil {
				log.Fatalf("Can't start QED server: %v", err)
			}

			err = srv.Start()
			if err != nil {
				log.Fatalf("Can't start QED server: %v", err)
			}

		},
	}

	f := cmd.Flags()
	hostname, _ := os.Hostname()
	f.StringVar(&conf.NodeID, "node-id", hostname, "Unique name for node. If not set, fallback to hostname")
	f.BoolVarP(&conf.EnableProfiling, "profiling", "f", false, "Allow a pprof url (localhost:6060) for profiling purposes")
	f.StringVar(&conf.PrivateKeyPath, "keypath", fmt.Sprintf("%s/%s", ctx.path, "id_ed25519"), "Server Singning private key file path")
	f.StringVar(&conf.SSLCertificate, "certificate", fmt.Sprintf("%s/%s", ctx.path, "server.crt"), "Server crt file")
	f.StringVar(&conf.SSLCertificateKey, "certificate-key", fmt.Sprintf("%s/%s", ctx.path, "server.key"), "Server key file")

	f.StringVar(&conf.HTTPAddr, "http-addr", ":8800", "Endpoint for REST requests on (host:port)")
	f.StringVar(&conf.RaftAddr, "raft-addr", ":8500", "Raft bind address (host:port)")
	f.StringVar(&conf.MgmtAddr, "mgmt-addr", ":8700", "Management endpoint bind address (host:port)")
	f.StringVar(&conf.MetricsAddr, "metrics-addr", ":8600", "Metrics export bind address (host:port)")
	f.StringSliceVar(&conf.RaftJoinAddr, "join-addr", []string{}, "Raft: Comma-delimited list of nodes ([host]:port), through which a cluster can be joined")
	f.StringVar(&conf.GossipAddr, "gossip-addr", ":8400", "Gossip: management endpoint bind address (host:port)")
	f.StringSliceVar(&conf.GossipJoinAddr, "gossip-join-addr", []string{}, "Gossip: Comma-delimited list of nodes ([host]:port), through which a cluster can be joined")

	// INFO: testing purposes
	f.BoolVar(&conf.EnableTampering, "tampering", false, "Allow tampering api for proof demostrations")
	_ = f.MarkHidden("tampering")

	// Lookups
	v.BindPFlag("server.node-id", f.Lookup("node-id"))
	v.BindPFlag("server.profiling", f.Lookup("profiling"))
	v.BindPFlag("server.key", f.Lookup("keypath"))
	v.BindPFlag("server.tls.certificate", f.Lookup("certificate"))
	v.BindPFlag("server.tls.certificate_key", f.Lookup("certificate-key"))

	v.BindPFlag("server.addr.http", f.Lookup("http-addr"))
	v.BindPFlag("server.addr.mgmt", f.Lookup("mgmt-addr"))
	v.BindPFlag("server.addr.metrics", f.Lookup("metrics-addr"))
	v.BindPFlag("server.addr.raft", f.Lookup("raft-addr"))
	v.BindPFlag("server.addr.raft_join", f.Lookup("join-addr"))
	v.BindPFlag("server.addr.gossip", f.Lookup("gossip-addr"))
	v.BindPFlag("server.addr.gossip_join", f.Lookup("gossip-join-addr"))

	v.BindPFlag("server.path.db", f.Lookup("dbpath"))
	v.BindPFlag("server.path.wal", f.Lookup("raftpath"))

	return cmd
}
