// Package main is the entry-point for the kubectl-redroid plugin.
// Install: go build -o kubectl-redroid ./cmd/kubectl-redroid  then add it to your $PATH.
// Usage:   kubectl redroid <command>
package main

import (
	"os"

	"github.com/isning/redroid-operator/cmd/kubectl-redroid/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
