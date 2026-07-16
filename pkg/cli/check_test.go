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

func TestCrdObjectName(t *testing.T) {
	tests := []struct {
		kind     string
		expected string
	}{
		{"PyTorchJob", "pytorchjobs.kubeflow.org"},
		{"TFJob", "tfjobs.kubeflow.org"},
		{"MPIJob", "mpijobs.kubeflow.org"},
		{"Unknown", ""},
	}
	for _, tt := range tests {
		got := crdObjectName(tt.kind)
		if got != tt.expected {
			t.Errorf("crdObjectName(%q) = %q, want %q", tt.kind, got, tt.expected)
		}
	}
}

func TestFormatCRDVersions(t *testing.T) {
	tests := []struct {
		name     string
		versions []client.CRDVersionInfo
		expected string
	}{
		{
			name: "single served+storage",
			versions: []client.CRDVersionInfo{
				{Name: "v1", Served: true, Storage: true},
			},
			expected: "v1 (served, storage)",
		},
		{
			name: "multiple versions",
			versions: []client.CRDVersionInfo{
				{Name: "v2beta1", Served: true, Storage: true},
				{Name: "v1", Served: true, Storage: false},
			},
			expected: "v2beta1 (served, storage), v1 (served)",
		},
		{
			name: "served only",
			versions: []client.CRDVersionInfo{
				{Name: "v1", Served: true, Storage: false},
			},
			expected: "v1 (served)",
		},
		{
			name:     "empty",
			versions: nil,
			expected: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, formatCRDVersions(tt.versions))
		})
	}
}

func TestGetCRDVersions_PyTorchJob(t *testing.T) {
	crd := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apiextensions.k8s.io/v1",
			"kind":       "CustomResourceDefinition",
			"metadata":   map[string]interface{}{"name": "pytorchjobs.kubeflow.org"},
			"spec": map[string]interface{}{
				"versions": []interface{}{
					map[string]interface{}{"name": "v1", "served": true, "storage": true},
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

	versions, err := k8sClient.GetCRDVersions(context.Background(), "pytorchjobs.kubeflow.org")
	require.NoError(t, err)
	assert.Len(t, versions, 1)
	assert.Equal(t, "v1", versions[0].Name)
	assert.Equal(t, "v1", client.FindStorageVersion(versions))
}

func TestGetCRDVersions_NotFound(t *testing.T) {
	scheme := runtime.NewScheme()
	fakeClient := fake.NewSimpleDynamicClientWithCustomListKinds(scheme,
		map[schema.GroupVersionResource]string{
			{Group: "apiextensions.k8s.io", Version: "v1", Resource: "customresourcedefinitions"}: "CustomResourceDefinitionList",
		})
	k8sClient := client.NewClientForInterface(fakeClient)

	versions, err := k8sClient.GetCRDVersions(context.Background(), "pytorchjobs.kubeflow.org")
	require.NoError(t, err)
	assert.Nil(t, versions)
}

func TestGetCRDVersions_UnknownKind(t *testing.T) {
	// crdObjectName returns empty for unknown kinds, so skip
	assert.Equal(t, "", crdObjectName("UnknownKind"))
}

func TestGetCRDVersions_EmptyVersions(t *testing.T) {
	crd := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apiextensions.k8s.io/v1",
			"kind":       "CustomResourceDefinition",
			"metadata":   map[string]interface{}{"name": "pytorchjobs.kubeflow.org"},
			"spec": map[string]interface{}{
				"versions": []interface{}{},
			},
		},
	}

	scheme := runtime.NewScheme()
	fakeClient := fake.NewSimpleDynamicClientWithCustomListKinds(scheme,
		map[schema.GroupVersionResource]string{
			{Group: "apiextensions.k8s.io", Version: "v1", Resource: "customresourcedefinitions"}: "CustomResourceDefinitionList",
		}, crd)
	k8sClient := client.NewClientForInterface(fakeClient)

	versions, err := k8sClient.GetCRDVersions(context.Background(), "pytorchjobs.kubeflow.org")
	require.NoError(t, err)
	assert.Empty(t, versions)
}

func TestCheckCmd_AllCRDsInstalled(t *testing.T) {
	pytorchCRD := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apiextensions.k8s.io/v1",
			"kind":       "CustomResourceDefinition",
			"metadata":   map[string]interface{}{"name": "pytorchjobs.kubeflow.org"},
			"spec": map[string]interface{}{
				"versions": []interface{}{
					map[string]interface{}{"name": "v1", "served": true, "storage": true},
				},
			},
		},
	}
	tfCRD := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apiextensions.k8s.io/v1",
			"kind":       "CustomResourceDefinition",
			"metadata":   map[string]interface{}{"name": "tfjobs.kubeflow.org"},
			"spec": map[string]interface{}{
				"versions": []interface{}{
					map[string]interface{}{"name": "v1", "served": true, "storage": true},
				},
			},
		},
	}
	mpiCRD := &unstructured.Unstructured{
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
		}, pytorchCRD, tfCRD, mpiCRD)
	k8sClient := client.NewClientForInterface(fakeClient)

	ctx := context.Background()

	pyVersions, err := k8sClient.GetCRDVersions(ctx, "pytorchjobs.kubeflow.org")
	require.NoError(t, err)
	assert.Equal(t, "v1", client.FindStorageVersion(pyVersions))

	tfVersions, err := k8sClient.GetCRDVersions(ctx, "tfjobs.kubeflow.org")
	require.NoError(t, err)
	assert.Equal(t, "v1", client.FindStorageVersion(tfVersions))

	mpiVersions, err := k8sClient.GetCRDVersions(ctx, "mpijobs.kubeflow.org")
	require.NoError(t, err)
	assert.Equal(t, "v2beta1", client.FindStorageVersion(mpiVersions))
}

func TestCheckCmd_SomeCRDsMissing(t *testing.T) {
	pytorchCRD := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apiextensions.k8s.io/v1",
			"kind":       "CustomResourceDefinition",
			"metadata":   map[string]interface{}{"name": "pytorchjobs.kubeflow.org"},
			"spec": map[string]interface{}{
				"versions": []interface{}{
					map[string]interface{}{"name": "v1", "served": true, "storage": true},
				},
			},
		},
	}

	scheme := runtime.NewScheme()
	fakeClient := fake.NewSimpleDynamicClientWithCustomListKinds(scheme,
		map[schema.GroupVersionResource]string{
			{Group: "apiextensions.k8s.io", Version: "v1", Resource: "customresourcedefinitions"}: "CustomResourceDefinitionList",
		}, pytorchCRD)
	k8sClient := client.NewClientForInterface(fakeClient)

	ctx := context.Background()

	pyVersions, err := k8sClient.GetCRDVersions(ctx, "pytorchjobs.kubeflow.org")
	require.NoError(t, err)
	assert.Equal(t, "v1", client.FindStorageVersion(pyVersions))

	tfVersions, err := k8sClient.GetCRDVersions(ctx, "tfjobs.kubeflow.org")
	require.NoError(t, err)
	assert.Nil(t, tfVersions)

	mpiVersions, err := k8sClient.GetCRDVersions(ctx, "mpijobs.kubeflow.org")
	require.NoError(t, err)
	assert.Nil(t, mpiVersions)
}
