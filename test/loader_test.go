package integration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kubeflow/arena/pkg/provider"
	"github.com/kubeflow/arena/pkg/task"
)

func TestLoadFromBytesRoundTrip(t *testing.T) {
	yamlData := `
name: full-task
image: pytorch/pytorch:2.1
framework:
  name: pytorch
  options:
    nproc_per_node: "auto"
worker:
  replicas: 8
  resources:
    nvidia.com/gpu: "4"
    cpu: "8"
    memory: "32Gi"
envs:
  NCCL_DEBUG: INFO
  SECRET_KEY:
    secret: my-secret
    key: api-key
labels:
  team: ml-platform
annotations:
  description: large-scale training
run: python train.py --epochs 100
`
	taskObj, err := task.LoadFromBytes([]byte(yamlData))
	require.NoError(t, err)

	assert.Equal(t, "full-task", taskObj.Name)
	assert.Equal(t, "pytorch/pytorch:2.1", taskObj.Image)
	assert.Equal(t, "pytorch", taskObj.Framework.Name)
	assert.Equal(t, "auto", taskObj.Framework.Options.NprocPerNode)
	assert.Equal(t, 8, taskObj.Worker.Replicas)
	assert.Equal(t, "4", taskObj.Worker.Resources["nvidia.com/gpu"])
	assert.Equal(t, "8", taskObj.Worker.Resources["cpu"])
	assert.Equal(t, "32Gi", taskObj.Worker.Resources["memory"])

	assert.Equal(t, "INFO", taskObj.Envs["NCCL_DEBUG"].Value)
	require.NotNil(t, taskObj.Envs["SECRET_KEY"].Secret)
	assert.Equal(t, "my-secret", taskObj.Envs["SECRET_KEY"].Secret.Name)
	assert.Equal(t, "api-key", taskObj.Envs["SECRET_KEY"].Secret.Key)

	assert.Equal(t, "ml-platform", taskObj.Labels["team"])
	assert.Equal(t, "large-scale training", taskObj.Annotations["description"])

	p := &provider.PyTorchProvider{}
	crd, err := p.BuildCRD(taskObj)
	require.NoError(t, err)
	assert.Equal(t, "PyTorchJob", crd.GetKind())
	assert.Equal(t, "full-task", crd.GetName())
}

func TestLoadFromFileNotFound(t *testing.T) {
	_, err := task.LoadFromFile("/nonexistent/path/to/file.yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read file")
}

func TestLoadFromBytesInvalidYAML(t *testing.T) {
	_, err := task.LoadFromBytes([]byte("{{invalid yaml content"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse YAML")
}

func TestNamespaceIsSetOnCRD(t *testing.T) {
	taskObj := &task.Task{
		Name:      "ns-test",
		Image:     "pytorch:1.13",
		Framework: task.Framework{Name: "pytorch"},
		Worker:    &task.Worker{Replicas: 1},
	}

	p := &provider.PyTorchProvider{}
	crd, err := p.BuildCRD(taskObj)
	require.NoError(t, err)

	assert.Empty(t, crd.GetNamespace())

	crd.SetNamespace("training")
	assert.Equal(t, "training", crd.GetNamespace())

	ctx := context.Background()
	k8sClient := newFakeK8sClient(t)
	require.NoError(t, k8sClient.Create(ctx, crd))

	obj, err := k8sClient.Get(ctx, "PyTorchJob", "training", "ns-test")
	require.NoError(t, err)
	assert.Equal(t, "training", obj.GetNamespace())
}
