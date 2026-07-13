package integration

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kubeflow/arena/pkg/cli"
	"github.com/kubeflow/arena/pkg/task"
)

func TestErrorDisplay_InvalidFile(t *testing.T) {
	_, err := task.LoadFromFile("/nonexistent/invalid-file.yaml")
	require.Error(t, err, "loading a nonexistent file must return an error")

	errMsg := err.Error()
	assert.Contains(t, errMsg, "failed to read file", "error should mention file read failure")
	assert.Contains(t, errMsg, "invalid-file.yaml", "error should include the filename for context")
}

func TestErrorDisplay_InvalidYAML(t *testing.T) {
	_, err := task.LoadFromBytes([]byte("{{not valid yaml"))
	require.Error(t, err, "loading invalid YAML must return an error")

	errMsg := err.Error()
	assert.Contains(t, errMsg, "failed to parse YAML", "error should mention YAML parse failure")
}

func TestErrorDisplay_ValidationError(t *testing.T) {
	tests := []struct {
		name     string
		yaml     string
		expected string
	}{
		{
			name: "missing name",
			yaml: `image: pytorch:1.13
framework:
  name: pytorch
worker:
  replicas: 1
`,
			expected: "name is required",
		},
		{
			name: "missing image",
			yaml: `name: test
framework:
  name: pytorch
worker:
  replicas: 1
`,
			expected: "image is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := task.LoadFromBytes([]byte(tt.yaml))
			require.Error(t, err, "validation failure must return an error")
			assert.Contains(t, err.Error(), tt.expected, "error should describe the validation failure")
		})
	}
}

func TestErrorDisplay_CLIExecuteReturnsError(t *testing.T) {
	t.Run("nonexistent file", func(t *testing.T) {
		err := cli.ExecuteWithArgs([]string{"job", "run", "-f", "/nonexistent/file.yaml"})
		require.Error(t, err, "CLI must return an error for invalid input")
		assert.Contains(t, err.Error(), "failed to read file", "error should describe the failure")
	})
}

func TestErrorDisplay_CLIExecuteSuccess(t *testing.T) {
	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "valid.yaml")
	content := `name: success-test
image: pytorch:1.13
framework:
  name: pytorch
worker:
  replicas: 1
run: echo hello
`
	err := os.WriteFile(yamlPath, []byte(content), 0644)
	require.NoError(t, err)

	err = cli.ExecuteWithArgs([]string{"job", "run", "-f", yamlPath, "--dry-run"})
	assert.NoError(t, err, "dry-run with valid YAML should succeed (exit code 0)")
}

func TestErrorDisplay_CLIErrorChain(t *testing.T) {
	err := cli.ExecuteWithArgs([]string{"job", "run", "-f", "/nonexistent/cli-error-test.yaml"})
	require.Error(t, err, "CLI must return an error for nonexistent file")

	errMsg := err.Error()
	assert.Contains(t, errMsg, "failed to read file", "CLI error should include file read failure context")
	assert.Contains(t, errMsg, "cli-error-test.yaml", "CLI error should include the filename")

	var chain []string
	current := err
	for current != nil {
		chain = append(chain, current.Error())
		current = errors.Unwrap(current)
	}
	assert.GreaterOrEqual(t, len(chain), 2, "CLI error chain should have at least 2 levels")
}

func TestErrorDisplay_ErrorChainUnwrap(t *testing.T) {
	_, err := task.LoadFromFile("/nonexistent/file.yaml")
	require.Error(t, err)

	var chain []string
	current := err
	for current != nil {
		chain = append(chain, current.Error())
		current = errors.Unwrap(current)
	}

	require.GreaterOrEqual(t, len(chain), 2, "error chain should have at least 2 levels")
	assert.Contains(t, chain[0], "failed to read file", "outermost error should describe the file read failure")

	foundOSError := false
	for _, msg := range chain[1:] {
		if containsAny(msg, "no such file", "does not exist") {
			foundOSError = true
			break
		}
	}
	assert.True(t, foundOSError, "error chain should include the underlying OS error")
}

func TestErrorDisplay_FormatErrorOutput(t *testing.T) {
	innerErr := errors.New("connection refused")
	wrappedErr := fmt.Errorf("failed to connect: %w", innerErr)
	outerErr := fmt.Errorf("failed to create client: %w", wrappedErr)

	basicMsg := fmt.Sprintf("Error: %s", outerErr.Error())
	assert.Contains(t, basicMsg, "Error: failed to create client")
	assert.Contains(t, basicMsg, "failed to connect")
	assert.Contains(t, basicMsg, "connection refused")

	var chain []string
	current := outerErr
	for current != nil {
		chain = append(chain, current.Error())
		current = errors.Unwrap(current)
	}
	assert.Equal(t, 3, len(chain), "error chain should have 3 levels")
	assert.Contains(t, chain[0], "failed to create client")
	assert.Contains(t, chain[1], "failed to connect")
	assert.Contains(t, chain[2], "connection refused")
}
