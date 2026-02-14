package main

import (
	"os"

	"github.com/igorlopes-orca/generate-cost-savings-from-json/cmd"
)

func main() {
	if err := cmd.NewRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
