// Copyright 2018 BBVA. All rights reserved.
// Use of this source code is governed by a Apache 2 License
// that can be found in the LICENSE file

package main

import (
	"context"
	"flag"
	"verifiabledata/server"

	"github.com/golang/glog"
)

var (
	httpEndpoint = flag.String("http_endpoint", "localhost:8080", "Endpoint for REST requests on (host:port)")
	configFile   = flag.String("config", "", "Config file containing flags. Its contents can be overriden by command line flags.")
)

func main() {

	flag.Parse()

	// TODO merge config file flags

	ctx := context.Background()

	srv := server.Server{
		HTTPEndpoint: *httpEndpoint,
	}

	if err := srv.Run(ctx); err != nil {
		glog.Exitf("Server exited with error: %v", err)
	}

}
