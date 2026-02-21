package main

import (
	"os"

	"github.com/selesy/git-bug-ax/pkg/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}

	os.Exit(0)
}
