// Copyright © 2018 Banco Bilbao Vizcaya Argentaria S.A.  All rights reserved.
// Use of this source code is governed by an Apache 2 License
// that can be found in the LICENSE file

package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"verifiabledata/log"

	"verifiabledata/client"
)

var (
	httpEndpoint string
	apiKey       string
)

func main() {
	flag.StringVar(&httpEndpoint, "http_endpoint", "http://localhost:8080", "Endpoint for send requests on (host:port)")
	flag.StringVar(&apiKey, "api_key", "this-is-my-api-key", "Api auth key")
	flag.Parse()

	logger := log.NewError(os.Stdout, "QedAgent", log.Ldate|log.Ltime|log.Lmicroseconds|log.Llongfile)

	logger.Info("Starting client")

	auditor := client.NewHttpClient(httpEndpoint, apiKey, logger)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		snapshot, err := auditor.Add(scanner.Text())
		if err != nil {
			panic(err)
		}

		proof, err := auditor.Membership(snapshot.Event, snapshot.Version)
		if err != nil {
			panic(err)
		}

		correct := auditor.Verify(proof, snapshot)

		fmt.Println(correct)
	}
}
