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
