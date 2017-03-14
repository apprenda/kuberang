package main

import (
	"os"

	"github.com/apprenda/kuberang/pkg/util"
)

// Set via linker flag
var version string
var buildDate string

func main() {
	cmd := NewKuberangCommand(version, os.Stdin, os.Stdout)

	if err := cmd.Execute(); err != nil {
		util.PrintColor(os.Stderr, util.Red, "Error running command: %v\n", err)
		os.Exit(1)
	}
}
