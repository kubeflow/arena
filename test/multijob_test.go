package integration

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/kubeflow/arena/pkg/task"
)

func TestMultipleJobsAcrossFrameworks(t *testing.T) {
	ctx := context.Background()
	k8sClient := newFakeK8sClient(t)

	examples := []struct {
		file      string
		framework string
		kind      string
		name      string
	}{
		{"pytorch-simple.yaml", "pytorch", "PyTorchJob", "pytorch-example"},
		{"tensorflow-simple.yaml", "tensorflow", "TFJob", "tensorflow-example"},
		{"mpi-simple.yaml", "mpi", "MPIJob", "mpi-example"},
	}

	for _, ex := range examples {
		yamlPath := filepath.Join(examplesDir(t), ex.file)
		taskObj, err := task.LoadFromFile(yamlPath)
		require.NoError(t, err)

		p := providerFor(ex.framework)
		crd, err := p.BuildCRD(taskObj)
		require.NoError(t, err)
		crd.SetNamespace("default")

		err = k8sClient.Create(ctx, crd)
		require.NoError(t, err)
	}

	for _, ex := range examples {
		jobs, err := k8sClient.List(ctx, ex.kind, "default", "")
		require.NoError(t, err)
		require.Len(t, jobs, 1, "expected 1 %s", ex.kind)
		assert.Equal(t, ex.name, jobs[0].GetName())
	}

	require.NoError(t, k8sClient.Delete(ctx, "TFJob", "default", "tensorflow-example"))

	tfJobs, err := k8sClient.List(ctx, "TFJob", "default", "")
	require.NoError(t, err)
	assert.Empty(t, tfJobs)

	ptJobs, err := k8sClient.List(ctx, "PyTorchJob", "default", "")
	require.NoError(t, err)
	assert.Len(t, ptJobs, 1)

	mpiJobs, err := k8sClient.List(ctx, "MPIJob", "default", "")
	require.NoError(t, err)
	assert.Len(t, mpiJobs, 1)
}

func TestDuplicateJobCreationFails(t *testing.T) {
	ctx := context.Background()
	k8sClient := newFakeK8sClient(t)

	crd := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "kubeflow.org/v1",
			"kind":       "PyTorchJob",
			"metadata": map[string]interface{}{
				"name":      "dup-job",
				"namespace": "default",
			},
			"spec": map[string]interface{}{},
		},
	}

	require.NoError(t, k8sClient.Create(ctx, crd))

	err := k8sClient.Create(ctx, crd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}
