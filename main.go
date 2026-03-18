package main

import (
	"os"

	"github.com/selesy/git-bug-agent/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}

	os.Exit(0)
}
