// Copyright © 2018 Banco Bilbao Vizcaya Argentaria S.A.  All rights reserved.
// Use of this source code is governed by an Apache 2 License
// that can be found in the LICENSE file

package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"

	"verifiabledata/agent"
)

func main() {
	ctx := context.Background()

	log.Println("Starting agent")

	agent, err := agent.Run(ctx)
	if err != nil {
		log.Panicln("Agent exited with error:", err)
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		fmt.Println(agent.Add(scanner.Text()))
	}
}
