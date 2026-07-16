package integration

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestStopResumeFlow(t *testing.T) {
	ctx := context.Background()
	k8sClient := newFakeK8sClient(t)

	crd := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "kubeflow.org/v1",
			"kind":       "PyTorchJob",
			"metadata": map[string]interface{}{
				"name":      "stop-resume-job",
				"namespace": "default",
			},
			"spec": map[string]interface{}{
				"runPolicy": map[string]interface{}{},
				"pytorchReplicaSpecs": map[string]interface{}{
					"Master": map[string]interface{}{
						"replicas": int64(1),
					},
				},
			},
		},
	}
	require.NoError(t, k8sClient.Create(ctx, crd))

	stopPatch := map[string]interface{}{
		"spec": map[string]interface{}{
			"runPolicy": map[string]interface{}{
				"suspend": true,
			},
		},
	}
	stopBytes, err := json.Marshal(stopPatch)
	require.NoError(t, err)

	err = k8sClient.Patch(ctx, "PyTorchJob", "default", "stop-resume-job", stopBytes)
	require.NoError(t, err)

	obj, err := k8sClient.Get(ctx, "PyTorchJob", "default", "stop-resume-job")
	require.NoError(t, err)
	suspend, found, err := unstructured.NestedBool(obj.Object, "spec", "runPolicy", "suspend")
	require.NoError(t, err)
	assert.True(t, found)
	assert.True(t, suspend)

	resumePatch := map[string]interface{}{
		"spec": map[string]interface{}{
			"runPolicy": map[string]interface{}{
				"suspend": false,
			},
		},
	}
	resumeBytes, err := json.Marshal(resumePatch)
	require.NoError(t, err)

	err = k8sClient.Patch(ctx, "PyTorchJob", "default", "stop-resume-job", resumeBytes)
	require.NoError(t, err)

	obj, err = k8sClient.Get(ctx, "PyTorchJob", "default", "stop-resume-job")
	require.NoError(t, err)
	suspend, _, err = unstructured.NestedBool(obj.Object, "spec", "runPolicy", "suspend")
	require.NoError(t, err)
	assert.False(t, suspend)
}
