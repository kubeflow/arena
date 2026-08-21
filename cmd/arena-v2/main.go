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
		fmt.Fprint(os.Stderr, formatError(err, cli.DebugMode(), cli.JSONErrorRequested()))
		os.Exit(1)
	}
}

// formatError formats an error for display to the user. In machine-readable
// mode (-o json/yaml) every error renders as a JSON envelope so stderr stays
// parseable; otherwise the error renders as prose. Debug mode appends the
// full error chain.
func formatError(err error, debug, jsonMode bool) string {
	if jsonMode {
		s := cli.ErrorEnvelope(err)
		if debug {
			s += errorChain(err)
		}
		return s
	}

	var sb strings.Builder
	sb.WriteString("Error: ")
	sb.WriteString(err.Error())
	sb.WriteString("\n")
	if debug {
		sb.WriteString(errorChain(err))
	}
	return sb.String()
}

func errorChain(err error) string {
	var sb strings.Builder
	sb.WriteString("\nFull error chain:\n")
	for current := err; current != nil; current = errors.Unwrap(current) {
		fmt.Fprintf(&sb, "  - %v\n", current)
	}
	return sb.String()
}
