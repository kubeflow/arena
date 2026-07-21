package integration

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kubeflow/arena/pkg/task"
)

func TestEndToEndPyTorchFlow(t *testing.T) {
	yamlPath := filepath.Join(examplesDir(t), "pytorch-simple.yaml")
	taskObj, err := task.LoadFromFile(yamlPath)
	require.NoError(t, err, "LoadFromFile should succeed for pytorch-simple.yaml")

	assert.Equal(t, "pytorch-example", taskObj.Name)
	assert.Equal(t, "docker.io/pytorch/pytorch:1.13-cuda11.6-cudnn8-runtime", taskObj.Image)
	assert.Equal(t, "pytorch", taskObj.Framework.Name)
	assert.Equal(t, 4, taskObj.Worker.Replicas)

	p := providerFor(taskObj.Framework.Name)
	require.NotNil(t, p)

	crd, err := p.BuildCRD(taskObj)
	require.NoError(t, err)
	assert.Equal(t, "PyTorchJob", crd.GetKind())
	assert.Equal(t, "pytorch-example", crd.GetName())
	assert.Equal(t, "kubeflow.org/v1", crd.GetAPIVersion())

	spec := crd.Object["spec"].(map[string]interface{})
	replicaSpecs := spec["pytorchReplicaSpecs"].(map[string]interface{})

	master := replicaSpecs["Master"].(map[string]interface{})
	assert.Equal(t, int64(1), master["replicas"])

	worker := replicaSpecs["Worker"].(map[string]interface{})
	assert.Equal(t, int64(4), worker["replicas"])

	ctx := context.Background()
	k8sClient := newFakeK8sClient(t)
	crd.SetNamespace("default")
	err = k8sClient.Create(ctx, crd)
	require.NoError(t, err, "Create should succeed with fake client")

	obj, err := k8sClient.Get(ctx, "PyTorchJob", "default", "pytorch-example")
	require.NoError(t, err)
	assert.Equal(t, "pytorch-example", obj.GetName())
	assert.Equal(t, "default", obj.GetNamespace())

	jobs, err := k8sClient.List(ctx, "PyTorchJob", "default", "")
	require.NoError(t, err)
	require.Len(t, jobs, 1)
	assert.Equal(t, "pytorch-example", jobs[0].GetName())

	err = k8sClient.Delete(ctx, "PyTorchJob", "default", "pytorch-example")
	require.NoError(t, err)

	_, err = k8sClient.Get(ctx, "PyTorchJob", "default", "pytorch-example")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestEndToEndTensorFlowFlow(t *testing.T) {
	yamlPath := filepath.Join(examplesDir(t), "tensorflow-simple.yaml")
	taskObj, err := task.LoadFromFile(yamlPath)
	require.NoError(t, err)

	assert.Equal(t, "tensorflow-example", taskObj.Name)
	assert.Equal(t, "tensorflow", taskObj.Framework.Name)
	assert.Equal(t, 2, taskObj.Worker.Replicas)

	p := providerFor(taskObj.Framework.Name)
	require.NotNil(t, p)

	crd, err := p.BuildCRD(taskObj)
	require.NoError(t, err)
	assert.Equal(t, "TFJob", crd.GetKind())
	assert.Equal(t, "tensorflow-example", crd.GetName())

	spec := crd.Object["spec"].(map[string]interface{})
	replicaSpecs := spec["tfReplicaSpecs"].(map[string]interface{})

	worker := replicaSpecs["Worker"].(map[string]interface{})
	assert.Equal(t, int64(2), worker["replicas"])

	ctx := context.Background()
	k8sClient := newFakeK8sClient(t)
	crd.SetNamespace("default")
	require.NoError(t, k8sClient.Create(ctx, crd))

	jobs, err := k8sClient.List(ctx, "TFJob", "default", "")
	require.NoError(t, err)
	require.Len(t, jobs, 1)

	require.NoError(t, k8sClient.Delete(ctx, "TFJob", "default", "tensorflow-example"))

	_, err = k8sClient.Get(ctx, "TFJob", "default", "tensorflow-example")
	require.Error(t, err)
}

func TestEndToEndMPIFlow(t *testing.T) {
	yamlPath := filepath.Join(examplesDir(t), "mpi-simple.yaml")
	taskObj, err := task.LoadFromFile(yamlPath)
	require.NoError(t, err)

	assert.Equal(t, "mpi-example", taskObj.Name)
	assert.Equal(t, "mpi", taskObj.Framework.Name)
	assert.Equal(t, 4, taskObj.Worker.Replicas)

	p := providerFor(taskObj.Framework.Name)
	require.NotNil(t, p)

	crd, err := p.BuildCRD(taskObj)
	require.NoError(t, err)
	assert.Equal(t, "MPIJob", crd.GetKind())
	assert.Equal(t, "mpi-example", crd.GetName())

	spec := crd.Object["spec"].(map[string]interface{})
	assert.Equal(t, int64(4), spec["slotsPerWorker"])

	replicaSpecs := spec["mpiReplicaSpecs"].(map[string]interface{})

	launcher := replicaSpecs["Launcher"].(map[string]interface{})
	assert.Equal(t, int64(1), launcher["replicas"])

	worker := replicaSpecs["Worker"].(map[string]interface{})
	assert.Equal(t, int64(4), worker["replicas"])

	ctx := context.Background()
	k8sClient := newFakeK8sClient(t)
	crd.SetNamespace("default")
	require.NoError(t, k8sClient.Create(ctx, crd))

	jobs, err := k8sClient.List(ctx, "MPIJob", "default", "")
	require.NoError(t, err)
	require.Len(t, jobs, 1)

	require.NoError(t, k8sClient.Delete(ctx, "MPIJob", "default", "mpi-example"))

	_, err = k8sClient.Get(ctx, "MPIJob", "default", "mpi-example")
	require.Error(t, err)
}
