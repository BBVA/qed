// Copyright © 2018 Banco Bilbao Vizcaya Argentaria S.A.  All rights reserved.
// Use of this source code is governed by an Apache 2 License
// that can be found in the LICENSE file

package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"

	"verifiabledata/agent"
)

var (
	httpEndpoint string
)

func main() {
	flag.StringVar(&httpEndpoint, "http_endpoint", "http://localhost:8080", "Endpoint for send requests on (host:port)")
	flag.Parse()

	log.Println("Starting agent")

	agent, err := agent.Run(httpEndpoint)
	if err != nil {
		log.Panicln("Agent exited with error:", err)
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		fmt.Println(agent.Add(scanner.Text()))
	}
}
