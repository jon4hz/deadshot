package main

import (
	"os"

	"github.com/jon4hz/deadshot/cmd/deadshot/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
