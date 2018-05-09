package main

import (
	"os"

	"qed/cli"
)

func main() {
	if err := cli.NewServerCommand().Execute(); err != nil {
		os.Exit(-1)
	}
}
