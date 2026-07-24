package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestKindToFramework(t *testing.T) {
	tests := []struct {
		kind     string
		expected string
	}{
		{"PyTorchJob", "pytorch"},
		{"TFJob", "tensorflow"},
		{"MPIJob", "mpi"},
		{"UnknownJob", "unknownjob"},
	}

	for _, tt := range tests {
		result := kindToFramework(tt.kind)
		if result != tt.expected {
			t.Errorf("kindToFramework(%q) = %q, want %q", tt.kind, result, tt.expected)
		}
	}
}

func TestSetFrameworkLabel_NilLabels(t *testing.T) {
	crd := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "kubeflow.org/v1",
			"kind":       "PyTorchJob",
			"metadata":   map[string]interface{}{"name": "test", "namespace": "default"},
		},
	}
	assert.Nil(t, crd.GetLabels())

	setFrameworkLabel(crd, "pytorch")

	labels := crd.GetLabels()
	assert.Equal(t, "pytorch", labels[frameworkLabel])
}

func TestSetFrameworkLabel_ExistingLabels(t *testing.T) {
	crd := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "kubeflow.org/v1",
			"kind":       "MPIJob",
			"metadata": map[string]interface{}{
				"name":      "test",
				"namespace": "default",
				"labels": map[string]interface{}{
					"app": "my-training",
				},
			},
		},
	}

	setFrameworkLabel(crd, "horovod")

	labels := crd.GetLabels()
	assert.Equal(t, "horovod", labels[frameworkLabel])
	assert.Equal(t, "my-training", labels["app"], "existing labels should be preserved")
}

func TestSetFrameworkLabel_OverwriteExisting(t *testing.T) {
	crd := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "kubeflow.org/v1",
			"kind":       "MPIJob",
			"metadata": map[string]interface{}{
				"name":      "test",
				"namespace": "default",
				"labels": map[string]interface{}{
					frameworkLabel: "mpi",
				},
			},
		},
	}

	setFrameworkLabel(crd, "deepspeed")

	assert.Equal(t, "deepspeed", crd.GetLabels()[frameworkLabel])
}

func TestSetFrameworkLabel_AllFrameworks(t *testing.T) {
	frameworks := []string{"pytorch", "tensorflow", "mpi", "horovod", "deepspeed"}
	for _, fw := range frameworks {
		crd := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "kubeflow.org/v1",
				"kind":       "PyTorchJob",
				"metadata":   map[string]interface{}{"name": "test", "namespace": "default"},
			},
		}
		setFrameworkLabel(crd, fw)
		assert.Equal(t, fw, crd.GetLabels()[frameworkLabel], "framework %q should be set", fw)
	}
}
