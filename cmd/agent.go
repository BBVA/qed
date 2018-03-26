package main

import (
	"bufio"
	"context"
	"flag"
	"os"
	// "os/signal"
	"sync"

	"github.com/golang/glog"

	"verifiabledata/agent"
)

func main() {
	flag.Parse()

	var wg sync.WaitGroup

	ctx := context.Background()

	glog.V(2).Info("Starting agent")

	agent, err := agent.Run(ctx)
	if err != nil {
		defer os.Exit(255)
		glog.Exitf("Agent exited with error: %v", err)
	}

	ch := make(chan string, 10000)
	wg.Add(1)
	// cl_ch := make(chan int)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		agent.Add(scanner.Text())
		ch <- scanner.Text()
	}

	go func() {
		glog.Info("init channel")

		for {
			select {
			case msg := <-ch:
				agent.Fetch(msg)

			default:
				close(ch)
				wg.Done()
				return

			}
		}
	}()

	// c := make(chan os.Signal, 1)
	// wg.Add(1)
	//
	// signal.Notify(c, os.Interrupt)
	// go func() {
	// 	<-c
	// 	wg.Done()
	// 	defer os.Exit(1)
	// 	close(c)
	// }()

	wg.Wait()

}
