package main

import (
	"context"
	"flag"
	"github.com/golang/glog"
	"os"
	"verifiabledata/agent"
)

func main() {
	flag.Parse()

	ctx := context.Background()

	glog.V(2).Info("Starting agent")

	agent, err := agent.Run(ctx)

	if err != nil {
		glog.Exitf("Server exited with error: %v", err)
	}

	agent.Echo(os.Stdin)

}
