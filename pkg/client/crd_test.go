package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"
)

func TestFindStorageVersion_Single(t *testing.T) {
	versions := []CRDVersionInfo{
		{Name: "v2beta1", Served: true, Storage: true},
	}
	assert.Equal(t, "v2beta1", FindStorageVersion(versions))
}

func TestFindStorageVersion_MultipleVersions(t *testing.T) {
	versions := []CRDVersionInfo{
		{Name: "v1", Served: true, Storage: false},
		{Name: "v2beta1", Served: true, Storage: true},
	}
	assert.Equal(t, "v2beta1", FindStorageVersion(versions))
}

func TestFindStorageVersion_NoStorage(t *testing.T) {
	versions := []CRDVersionInfo{
		{Name: "v1", Served: true, Storage: false},
		{Name: "v2beta1", Served: true, Storage: false},
	}
	assert.Equal(t, "", FindStorageVersion(versions))
}

func TestFindStorageVersion_Empty(t *testing.T) {
	assert.Equal(t, "", FindStorageVersion(nil))
	assert.Equal(t, "", FindStorageVersion([]CRDVersionInfo{}))
}

func TestGetCRDVersions_MPIJob(t *testing.T) {
	crd := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apiextensions.k8s.io/v1",
			"kind":       "CustomResourceDefinition",
			"metadata":   map[string]interface{}{"name": "mpijobs.kubeflow.org"},
			"spec": map[string]interface{}{
				"versions": []interface{}{
					map[string]interface{}{"name": "v1", "served": true, "storage": false},
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
	k8sClient := NewClientForInterface(fakeClient)

	versions, err := k8sClient.GetCRDVersions(context.Background(), "mpijobs.kubeflow.org")
	assert.NoError(t, err)
	assert.Len(t, versions, 2)
	assert.Equal(t, "v1", versions[0].Name)
	assert.True(t, versions[0].Served)
	assert.False(t, versions[0].Storage)
	assert.Equal(t, "v2beta1", versions[1].Name)
	assert.True(t, versions[1].Served)
	assert.True(t, versions[1].Storage)
}

func TestGetCRDVersions_NotFound(t *testing.T) {
	scheme := runtime.NewScheme()
	fakeClient := fake.NewSimpleDynamicClientWithCustomListKinds(scheme,
		map[schema.GroupVersionResource]string{
			{Group: "apiextensions.k8s.io", Version: "v1", Resource: "customresourcedefinitions"}: "CustomResourceDefinitionList",
		})
	k8sClient := NewClientForInterface(fakeClient)

	versions, err := k8sClient.GetCRDVersions(context.Background(), "mpijobs.kubeflow.org")
	assert.NoError(t, err)
	assert.Nil(t, versions)
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
	k8sClient := NewClientForInterface(fakeClient)

	versions, err := k8sClient.GetCRDVersions(context.Background(), "pytorchjobs.kubeflow.org")
	assert.NoError(t, err)
	assert.Empty(t, versions)
}

// buildFakeMPICRD creates a fake MPIJob CRD unstructured object with the given versions.
func buildFakeMPICRD(versions []CRDVersionInfo) *unstructured.Unstructured {
	var versionList []interface{}
	for _, v := range versions {
		versionList = append(versionList, map[string]interface{}{
			"name":    v.Name,
			"served":  v.Served,
			"storage": v.Storage,
		})
	}

	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apiextensions.k8s.io/v1",
			"kind":       "CustomResourceDefinition",
			"metadata": map[string]interface{}{
				"name": "mpijobs.kubeflow.org",
			},
			"spec": map[string]interface{}{
				"versions": versionList,
			},
		},
	}
}

func newFakeCRDClient(objects ...*unstructured.Unstructured) *Client {
	scheme := runtime.NewScheme()
	runtimeObjects := make([]runtime.Object, len(objects))
	for i, obj := range objects {
		runtimeObjects[i] = obj
	}
	fakeClient := fake.NewSimpleDynamicClientWithCustomListKinds(scheme,
		map[schema.GroupVersionResource]string{
			crdGVR: "CustomResourceDefinitionList",
		}, runtimeObjects...)
	return NewClientForInterface(fakeClient)
}

func TestClient_ResolveMPIVersion(t *testing.T) {
	tests := []struct {
		name          string
		crd           *unstructured.Unstructured
		wantVersion   string
		wantErr       string
	}{
		{
			name: "storage version v2beta1",
			crd: buildFakeMPICRD([]CRDVersionInfo{
				{Name: "v1", Served: true, Storage: false},
				{Name: "v2beta1", Served: true, Storage: true},
			}),
			wantVersion: "v2beta1",
		},
		{
			name: "storage version v1",
			crd: buildFakeMPICRD([]CRDVersionInfo{
				{Name: "v1", Served: true, Storage: true},
				{Name: "v2beta1", Served: true, Storage: false},
			}),
			wantVersion: "v1",
		},
		{
			name:    "CRD not found",
			crd:     nil, // no CRD object seeded
			wantErr: "MPIJob CRD not found in cluster",
		},
		{
			name: "no storage version configured",
			crd: buildFakeMPICRD([]CRDVersionInfo{
				{Name: "v1", Served: true, Storage: false},
				{Name: "v2beta1", Served: true, Storage: false},
			}),
			wantErr: "MPIJob CRD has no storage version configured",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var c *Client
			if tt.crd != nil {
				c = newFakeCRDClient(tt.crd)
			} else {
				c = newFakeCRDClient()
			}

			err := c.ResolveMPIVersion(context.Background())
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
				assert.Empty(t, c.MPIVersion)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantVersion, c.MPIVersion)
			}
		})
	}
}

func TestClient_ResolveMPIVersion_Caching(t *testing.T) {
	crd := buildFakeMPICRD([]CRDVersionInfo{
		{Name: "v2beta1", Served: true, Storage: true},
	})
	c := newFakeCRDClient(crd)

	// First call resolves from cluster
	err := c.ResolveMPIVersion(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "v2beta1", c.MPIVersion)

	// Mutate the cached value to prove the second call skips resolution
	c.MPIVersion = "cached-value"

	err = c.ResolveMPIVersion(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "cached-value", c.MPIVersion, "second call should use cached value")
}
