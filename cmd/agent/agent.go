package main

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"verifiabledata/agent"
)

func main() {
	ctx := context.Background()

	fmt.Println("Starting agent")

	agent, err := agent.Run(ctx)
	if err != nil {
		fmt.Println("Error: ", err)
		panic("Agent exited")
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		fmt.Println(agent.Add(scanner.Text()))
	}
}
