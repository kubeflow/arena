package task

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadFromFile(t *testing.T) {
	// Create temp file
	tmpFile, err := os.CreateTemp("", "task-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	yamlContent := `
name: test-job
image: pytorch:2.1
framework:
  name: pytorch
  options:
    nproc_per_node: auto
run: python train.py
worker:
  replicas: 4
  resources:
    nvidia.com/gpu: "2"
`
	_, err = tmpFile.WriteString(yamlContent)
	require.NoError(t, err)
	tmpFile.Close()

	task, err := LoadFromFile(tmpFile.Name())
	require.NoError(t, err)

	assert.Equal(t, "test-job", task.Name)
	assert.Equal(t, "pytorch", task.Framework.Name)
	assert.Equal(t, 4, task.Worker.Replicas)
	assert.Equal(t, "2", task.Worker.Resources["nvidia.com/gpu"])
}

func TestLoadFromFile_NotFound(t *testing.T) {
	_, err := LoadFromFile("/nonexistent/path/task.yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read file")
}

func TestLoadFromFile_InvalidYAML(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "task-invalid-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// Truly malformed YAML that will fail at parse time, not validation
	invalidYAML := `
name: test-job
image: [invalid
  broken: {yaml
`
	_, err = tmpFile.WriteString(invalidYAML)
	require.NoError(t, err)
	tmpFile.Close()

	_, err = LoadFromFile(tmpFile.Name())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse YAML")
}

func TestLoadFromBytes_ValidationError(t *testing.T) {
	// Missing required name field
	yamlData := []byte(`
image: pytorch:2.1
framework:
  name: pytorch
run: train
worker:
  replicas: 1
`)

	_, err := LoadFromBytes(yamlData)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")
	assert.Contains(t, err.Error(), "name is required")
}

func TestLoadFromBytes_MissingRun(t *testing.T) {
	yamlData := []byte(`
name: test
image: pytorch:2.1
framework:
  name: pytorch
worker:
  replicas: 1
`)
	_, err := LoadFromBytes(yamlData)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "run is required")
}

func TestLoadFromBytes_MissingFramework(t *testing.T) {
	yamlData := []byte(`
name: test
image: pytorch:2.1
run: train
worker:
  replicas: 1
`)
	_, err := LoadFromBytes(yamlData)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported framework")
}
