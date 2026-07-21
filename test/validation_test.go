package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kubeflow/arena/pkg/provider"
	"github.com/kubeflow/arena/pkg/task"
)

func TestValidationRejectsInvalidTasks(t *testing.T) {
	tests := []struct {
		name string
		yaml string
		err  string
	}{
		{
			name: "missing name",
			yaml: `
image: pytorch:1.13
framework:
  name: pytorch
worker:
  replicas: 1
`,
			err: "name is required",
		},
		{
			name: "missing image",
			yaml: `
name: test
framework:
  name: pytorch
worker:
  replicas: 1
`,
			err: "image is required",
		},
		{
			name: "unsupported framework",
			yaml: `
name: test
image: mxnet:1.0
run: echo test
framework:
  name: mxnet
worker:
  replicas: 1
`,
			err: "unsupported framework",
		},
		{
			name: "zero replicas",
			yaml: `
name: test
image: pytorch:1.13
run: echo test
framework:
  name: pytorch
worker:
  replicas: 0
`,
			err: "worker.replicas must be > 0",
		},
		{
			name: "invalid cleanPodPolicy",
			yaml: `
name: test
image: pytorch:1.13
run: echo test
framework:
  name: pytorch
worker:
  replicas: 1
lifecycle:
  clean_pod_policy: InvalidPolicy
`,
			err: "invalid clean_pod_policy",
		},
		{
			name: "invalid restart",
			yaml: `
name: test
image: pytorch:1.13
run: echo test
framework:
  name: pytorch
worker:
  replicas: 1
restart: BadPolicy
`,
			err: "invalid restart",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := task.LoadFromBytes([]byte(tt.yaml))
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.err)
		})
	}
}

func TestValidationRejectsInvalidNprocPerNode(t *testing.T) {
	yamlData := `
name: test
image: pytorch:1.13
run: echo test
framework:
  name: pytorch
  options:
    nproc_per_node: "not-a-number"
worker:
  replicas: 1
`
	_, err := task.LoadFromBytes([]byte(yamlData))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nproc_per_node")
}

func TestProviderRejectsWrongFramework(t *testing.T) {
	tk := &task.Task{
		Name:      "wrong",
		Image:     "pytorch:1.13",
		Framework: task.Framework{Name: "pytorch"},
		Worker:    &task.Worker{Replicas: 1},
	}

	tests := []struct {
		name     string
		provider provider.Provider
		err      string
	}{
		{"TensorFlow rejects PyTorch", &provider.TensorFlowProvider{}, "tensorflow"},
		{"MPI rejects PyTorch", &provider.MPIProvider{}, "mpi"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.provider.BuildCRD(tk)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.err)
		})
	}
}

func TestProviderJobTypes(t *testing.T) {
	tests := []struct {
		provider provider.Provider
		expected string
	}{
		{&provider.PyTorchProvider{}, "PyTorchJob"},
		{&provider.TensorFlowProvider{}, "TFJob"},
		{&provider.MPIProvider{}, "MPIJob"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.provider.GetJobType())
		})
	}
}
