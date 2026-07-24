package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/kubeflow/arena/pkg/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprint(os.Stderr, formatError(err, cli.DebugMode()))
		os.Exit(1)
	}
}

// formatError formats an error for display to the user.
// If debug is true, includes full error chain.
func formatError(err error, debug bool) string {
	var sb strings.Builder

	sb.WriteString("Error: ")
	sb.WriteString(err.Error())
	sb.WriteString("\n")

	// In debug mode, show the full error chain
	if debug {
		sb.WriteString("\nFull error chain:\n")
		current := err
		for current != nil {
			fmt.Fprintf(&sb, "  - %v\n", current)
			current = errors.Unwrap(current)
		}
	}

	return sb.String()
}
