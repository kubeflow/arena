package cli

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"

	"github.com/kubeflow/arena/pkg/client"
)

func TestIsMPIFamily(t *testing.T) {
	tests := []struct {
		framework string
		expected  bool
	}{
		{"mpi", true},
		{"horovod", true},
		{"deepspeed", true},
		{"pytorch", false},
		{"tensorflow", false},
		{"ray", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.framework, func(t *testing.T) {
			assert.Equal(t, tt.expected, isMPIFamily(tt.framework))
		})
	}
}

func TestResolveMPIAPIVersion_StorageV2beta1(t *testing.T) {
	crd := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apiextensions.k8s.io/v1",
			"kind":       "CustomResourceDefinition",
			"metadata":   map[string]interface{}{"name": "mpijobs.kubeflow.org"},
			"spec": map[string]interface{}{
				"versions": []interface{}{
					map[string]interface{}{"name": "v2beta1", "served": true, "storage": true},
				},
			},
		},
	}

	scheme := runtime.NewScheme()
	fakeClient := fake.NewSimpleDynamicClientWithCustomListKinds(scheme,
		map[schema.GroupVersionResource]string{
			{Group: "apiextensions.k8s.io", Version: "v1", Resource: "customresourcedefinitions"}: "CustomResourceDefinitionList",
		}, crd)
	k8sClient := client.NewClientForInterface(fakeClient)

	version, err := resolveMPIAPIVersion(context.Background(), k8sClient)
	require.NoError(t, err)
	assert.Equal(t, "v2beta1", version)
}

func TestResolveMPIAPIVersion_StorageV1(t *testing.T) {
	crd := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apiextensions.k8s.io/v1",
			"kind":       "CustomResourceDefinition",
			"metadata":   map[string]interface{}{"name": "mpijobs.kubeflow.org"},
			"spec": map[string]interface{}{
				"versions": []interface{}{
					map[string]interface{}{"name": "v1", "served": true, "storage": true},
					map[string]interface{}{"name": "v2beta1", "served": true, "storage": false},
				},
			},
		},
	}

	scheme := runtime.NewScheme()
	fakeClient := fake.NewSimpleDynamicClientWithCustomListKinds(scheme,
		map[schema.GroupVersionResource]string{
			{Group: "apiextensions.k8s.io", Version: "v1", Resource: "customresourcedefinitions"}: "CustomResourceDefinitionList",
		}, crd)
	k8sClient := client.NewClientForInterface(fakeClient)

	version, err := resolveMPIAPIVersion(context.Background(), k8sClient)
	require.NoError(t, err)
	assert.Equal(t, "v1", version)
}

func TestResolveMPIAPIVersion_UnsupportedStorageVersion(t *testing.T) {
	crd := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apiextensions.k8s.io/v1",
			"kind":       "CustomResourceDefinition",
			"metadata":   map[string]interface{}{"name": "mpijobs.kubeflow.org"},
			"spec": map[string]interface{}{
				"versions": []interface{}{
					map[string]interface{}{"name": "v1alpha1", "served": true, "storage": true},
				},
			},
		},
	}

	scheme := runtime.NewScheme()
	fakeClient := fake.NewSimpleDynamicClientWithCustomListKinds(scheme,
		map[schema.GroupVersionResource]string{
			{Group: "apiextensions.k8s.io", Version: "v1", Resource: "customresourcedefinitions"}: "CustomResourceDefinitionList",
		}, crd)
	k8sClient := client.NewClientForInterface(fakeClient)

	_, err := resolveMPIAPIVersion(context.Background(), k8sClient)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "v1alpha1")
	assert.Contains(t, err.Error(), "arena supports")
}

func TestResolveMPIAPIVersion_CRDNotFound(t *testing.T) {
	scheme := runtime.NewScheme()
	fakeClient := fake.NewSimpleDynamicClientWithCustomListKinds(scheme,
		map[schema.GroupVersionResource]string{
			{Group: "apiextensions.k8s.io", Version: "v1", Resource: "customresourcedefinitions"}: "CustomResourceDefinitionList",
		})
	k8sClient := client.NewClientForInterface(fakeClient)

	_, err := resolveMPIAPIVersion(context.Background(), k8sClient)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "MPIJob CRD not found")
}
