package main

import (
	"os"
	"verifiabledata/cli"
)

func main() {
	if err := cli.NewQedCommand().Execute(); err != nil {
		os.Exit(-1)
	}
}
