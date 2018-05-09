package main

import (
	"os"

	"qed/cli"
)

func main() {
	ctx := cli.NewContext()
	if err := cli.NewQedCommand(ctx).Execute(); err != nil {
		os.Exit(-1)
	}
}
