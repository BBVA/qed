package main

import (
	"os"
	"verifiabledata/cli"
)

func main() {
	ctx := cli.NewContext()
	if err := cli.NewQedCommand(ctx).Execute(); err != nil {
		os.Exit(-1)
	}
}
