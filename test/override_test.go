package integration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kubeflow/arena/pkg/task"
)

func TestSubmitOverrideFlow(t *testing.T) {
	taskObj := &task.Task{
		Framework: task.Framework{Name: "pytorch"},
		Worker:    &task.Worker{Replicas: 1},
	}

	flags := map[string]interface{}{
		"name":           "my-pytorch-job",
		"image":          "pytorch/pytorch:2.1",
		"run":            "python train.py --lr 0.001",
		"workers":        4,
		"gpus":           2,
		"cpus":           "4",
		"mem":            "16Gi",
		"framework":      "pytorch",
		"nproc-per-node": "auto",
	}
	task.ApplyOverrides(taskObj, flags)

	require.NoError(t, task.Validate(taskObj))

	assert.Equal(t, "my-pytorch-job", taskObj.Name)
	assert.Equal(t, "pytorch/pytorch:2.1", taskObj.Image)
	assert.Equal(t, "python train.py --lr 0.001", taskObj.Run)
	assert.Equal(t, 4, taskObj.Worker.Replicas)
	assert.Equal(t, "2", taskObj.Worker.Resources["nvidia.com/gpu"])
	assert.Equal(t, "auto", taskObj.Framework.Options.NprocPerNode)

	p := providerFor("pytorch")
	crd, err := p.BuildCRD(taskObj)
	require.NoError(t, err)
	assert.Equal(t, "PyTorchJob", crd.GetKind())
	assert.Equal(t, "my-pytorch-job", crd.GetName())

	ctx := context.Background()
	k8sClient := newFakeK8sClient(t)
	crd.SetNamespace("default")
	require.NoError(t, k8sClient.Create(ctx, crd))

	obj, err := k8sClient.Get(ctx, "PyTorchJob", "default", "my-pytorch-job")
	require.NoError(t, err)
	assert.Equal(t, "my-pytorch-job", obj.GetName())
}

func TestApplyOverridesNamespace(t *testing.T) {
	taskObj := &task.Task{
		Name:      "ns-override",
		Image:     "pytorch:1.13",
		Framework: task.Framework{Name: "pytorch"},
		Worker:    &task.Worker{Replicas: 1},
	}

	flags := map[string]interface{}{
		"namespace": "custom-ns",
	}
	task.ApplyOverrides(taskObj, flags)

	assert.Equal(t, "custom-ns", taskObj.Namespace)
}
