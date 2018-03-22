package main

import (
	"bufio"
	"context"
	"flag"
	"os"
	// "os/signal"
	// "sync"

	"github.com/golang/glog"

	"verifiabledata/agent"
)

func main() {
	flag.Parse()

	// var wg sync.WaitGroup

	ctx := context.Background()

	glog.V(2).Info("Starting agent")

	agent, err := agent.Run(ctx)

	if err != nil {
		defer os.Exit(255)
		glog.Exitf("Agent exited with error: %v", err)
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		agent.Add(scanner.Text())
	}

	// c := make(chan os.Signal, 1)
	// wg.Add(1)
	//
	// signal.Notify(c, os.Interrupt)
	// go func() {
	// 	<-c
	// 	defer os.Exit(1)
	// 	close(c)
	// }()
	//
	// wg.Wait()

}
