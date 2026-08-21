package main

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kubeflow/arena/pkg/cli"
)

func TestFormatError_PlainError(t *testing.T) {
	assert.Equal(t, "Error: boom\n", formatError(errors.New("boom"), false, false))
}

func TestFormatError_CLIErrorText(t *testing.T) {
	err := &cli.CLIError{Message: "invalid value", ValidValues: []string{"a", "b"}}
	assert.Equal(t, "Error: invalid value; must be one of: a, b\n", formatError(err, false, false))
}

func TestFormatError_CLIErrorJSONEnvelope(t *testing.T) {
	err := &cli.CLIError{Message: "invalid value", ValidValues: []string{"json", "yaml"}}
	out := formatError(err, false, true)
	assert.Contains(t, out, `"message": "invalid value"`)
	assert.Contains(t, out, `"validValues": [`)
	assert.NotContains(t, out, "Error:")
	assert.True(t, strings.HasSuffix(out, "\n"), "envelope ends with newline")
}

func TestFormatError_NonCLIErrorInJSONModeRendersEnvelope(t *testing.T) {
	out := formatError(errors.New("api failure"), false, true)
	assert.Contains(t, out, `"message": "api failure"`)
	assert.NotContains(t, out, "Error:", "JSON mode must not render prose errors")
	assert.True(t, strings.HasSuffix(out, "\n"), "envelope ends with newline")
}

func TestFormatError_WrappedErrorInJSONModeRendersEnvelope(t *testing.T) {
	err := fmt.Errorf("failed to create K8s client: %w", errors.New("no such file"))
	out := formatError(err, false, true)
	assert.Contains(t, out, `"message": "failed to create K8s client: no such file"`)
	assert.NotContains(t, out, "Error:")
}

func TestFormatError_DebugChain(t *testing.T) {
	err := fmt.Errorf("outer: %w", errors.New("root cause"))
	out := formatError(err, true, false)
	assert.Contains(t, out, "Error: outer: root cause\n")
	assert.Contains(t, out, "Full error chain:")
	assert.Contains(t, out, "  - outer: root cause\n")
	assert.Contains(t, out, "  - root cause\n")
}

func TestFormatError_CLIErrorJSONWithDebugChain(t *testing.T) {
	err := &cli.CLIError{Message: "invalid value"}
	out := formatError(err, true, true)
	assert.Contains(t, out, `"message": "invalid value"`)
	assert.Contains(t, out, "Full error chain:")
}
