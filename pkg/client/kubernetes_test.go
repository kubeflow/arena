package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
)

func newFakeClient() *Client {
	scheme := runtime.NewScheme()
	listKinds := map[schema.GroupVersionResource]string{
		{Group: "kubeflow.org", Version: "v1", Resource: "pytorchjobs"}:      "PyTorchJobList",
		{Group: "kubeflow.org", Version: "v1", Resource: "tfjobs"}:           "TFJobList",
		{Group: "kubeflow.org", Version: "v2beta1", Resource: "mpijobs"}:     "MPIJobList",
	}
	fakeDynamic := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(scheme, listKinds)
	return &Client{dynamicClient: fakeDynamic, mpiVersion: "v2beta1"}
}

func newPyTorchJobCRD(name, namespace string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "kubeflow.org/v1",
			"kind":       "PyTorchJob",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
			},
			"spec": map[string]interface{}{
				"pytorchReplicaSpecs": map[string]interface{}{
					"Worker": map[string]interface{}{
						"replicas": int64(2),
					},
				},
			},
		},
	}
}

func TestClientCreate(t *testing.T) {
	client := newFakeClient()
	crd := newPyTorchJobCRD("test-job", "default")

	err := client.Create(context.Background(), crd)
	require.NoError(t, err)

	// Verify the object was created
	obj, err := client.Get(context.Background(), "PyTorchJob", "default", "test-job")
	require.NoError(t, err)
	require.Equal(t, "test-job", obj.GetName())
	require.Equal(t, "default", obj.GetNamespace())
}

func TestClientGet(t *testing.T) {
	client := newFakeClient()
	crd := newPyTorchJobCRD("get-test-job", "default")

	err := client.Create(context.Background(), crd)
	require.NoError(t, err)

	obj, err := client.Get(context.Background(), "PyTorchJob", "default", "get-test-job")
	require.NoError(t, err)
	require.Equal(t, "get-test-job", obj.GetName())
	require.Equal(t, "PyTorchJob", obj.GetKind())
}

func TestClientGetNotFound(t *testing.T) {
	client := newFakeClient()

	_, err := client.Get(context.Background(), "PyTorchJob", "default", "nonexistent")
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

func TestClientList(t *testing.T) {
	client := newFakeClient()

	// Create multiple jobs
	for _, name := range []string{"job-1", "job-2", "job-3"} {
		crd := newPyTorchJobCRD(name, "default")
		err := client.Create(context.Background(), crd)
		require.NoError(t, err)
	}

	list, err := client.List(context.Background(), "PyTorchJob", "default", "")
	require.NoError(t, err)
	require.Len(t, list, 3)
}

func TestClientListEmpty(t *testing.T) {
	client := newFakeClient()

	list, err := client.List(context.Background(), "PyTorchJob", "default", "")
	require.NoError(t, err)
	require.Empty(t, list)
}

func TestClientListWithLabelSelector(t *testing.T) {
	client := newFakeClient()

	crd := newPyTorchJobCRD("labeled-job", "default")
	crd.SetLabels(map[string]string{"arena.io/framework": "pytorch"})
	err := client.Create(context.Background(), crd)
	require.NoError(t, err)

	// Empty selector returns all jobs
	list, err := client.List(context.Background(), "PyTorchJob", "default", "")
	require.NoError(t, err)
	require.Len(t, list, 1)

	// Non-empty selector is accepted (fake client does not filter server-side)
	list, err = client.List(context.Background(), "PyTorchJob", "default", "arena.io/framework")
	require.NoError(t, err)
	// Fake client returns all objects regardless of selector; this verifies
	// the parameter is accepted without error
	require.Len(t, list, 1)
}

func TestClientDelete(t *testing.T) {
	client := newFakeClient()
	crd := newPyTorchJobCRD("delete-test-job", "default")

	err := client.Create(context.Background(), crd)
	require.NoError(t, err)

	err = client.Delete(context.Background(), "PyTorchJob", "default", "delete-test-job")
	require.NoError(t, err)

	// Verify the object is gone
	_, err = client.Get(context.Background(), "PyTorchJob", "default", "delete-test-job")
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

func TestClientDeleteNotFound(t *testing.T) {
	client := newFakeClient()

	err := client.Delete(context.Background(), "PyTorchJob", "default", "nonexistent")
	require.Error(t, err)
}

func TestClientPatch(t *testing.T) {
	client := newFakeClient()
	crd := newPyTorchJobCRD("patch-test-job", "default")

	err := client.Create(context.Background(), crd)
	require.NoError(t, err)

	// Apply a merge patch to add annotations
	patch := []byte(`{"metadata":{"annotations":{"arena.io/stop":"true"}}}`)
	err = client.Patch(context.Background(), "PyTorchJob", "default", "patch-test-job", patch)
	require.NoError(t, err)

	// Verify the annotation was applied
	obj, err := client.Get(context.Background(), "PyTorchJob", "default", "patch-test-job")
	require.NoError(t, err)
	annotations := obj.GetAnnotations()
	require.Equal(t, "true", annotations["arena.io/stop"])
}

func TestClientPatchNotFound(t *testing.T) {
	client := newFakeClient()

	patch := []byte(`{"metadata":{"annotations":{"arena.io/stop":"true"}}}`)
	err := client.Patch(context.Background(), "PyTorchJob", "default", "nonexistent", patch)
	require.Error(t, err)
}

func TestClient_kindToGVR(t *testing.T) {
	tests := []struct {
		name       string
		mpiVersion string
		kind       string
		want       schema.GroupVersionResource
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "MPIJob with v2beta1",
			mpiVersion: "v2beta1",
			kind:       "MPIJob",
			want: schema.GroupVersionResource{
				Group:    "kubeflow.org",
				Version:  "v2beta1",
				Resource: "mpijobs",
			},
		},
		{
			name:       "MPIJob with v1",
			mpiVersion: "v1",
			kind:       "MPIJob",
			want: schema.GroupVersionResource{
				Group:    "kubeflow.org",
				Version:  "v1",
				Resource: "mpijobs",
			},
		},
		{
			name:       "MPIJob without version returns error",
			mpiVersion: "",
			kind:       "MPIJob",
			wantErr:    true,
			errMsg:     "MPIJob version not resolved",
		},
		{
			name:       "PyTorchJob uses hardcoded v1",
			mpiVersion: "",
			kind:       "PyTorchJob",
			want: schema.GroupVersionResource{
				Group:    "kubeflow.org",
				Version:  "v1",
				Resource: "pytorchjobs",
			},
		},
		{
			name:       "TFJob uses hardcoded v1",
			mpiVersion: "",
			kind:       "TFJob",
			want: schema.GroupVersionResource{
				Group:    "kubeflow.org",
				Version:  "v1",
				Resource: "tfjobs",
			},
		},
		{
			name:       "ConfigMap is core resource",
			mpiVersion: "",
			kind:       "ConfigMap",
			want: schema.GroupVersionResource{
				Group:    "",
				Version:  "v1",
				Resource: "configmaps",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{mpiVersion: tt.mpiVersion}
			gvr, err := c.kindToGVR(tt.kind)
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, gvr)
		})
	}
}

func TestGetGVR(t *testing.T) {
	crd := newPyTorchJobCRD("test", "default")
	gvr := getGVR(crd)

	require.Equal(t, "kubeflow.org", gvr.Group)
	require.Equal(t, "v1", gvr.Version)
	require.Equal(t, "pytorchjobs", gvr.Resource)
}

func TestPluralize(t *testing.T) {
	require.Equal(t, "pytorchjobs", pluralize("PyTorchJob"))
	require.Equal(t, "tfjobs", pluralize("TFJob"))
	require.Equal(t, "mpijobs", pluralize("MPIJob"))
	require.Equal(t, "customkinds", pluralize("CustomKind"))
}

func TestTFJobCRD(t *testing.T) {
	client := newFakeClient()

	crd := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "kubeflow.org/v1",
			"kind":       "TFJob",
			"metadata": map[string]interface{}{
				"name":      "tf-test-job",
				"namespace": "default",
			},
		},
	}

	err := client.Create(context.Background(), crd)
	require.NoError(t, err)

	obj, err := client.Get(context.Background(), "TFJob", "default", "tf-test-job")
	require.NoError(t, err)
	require.Equal(t, "tf-test-job", obj.GetName())
}

func TestMPIJobCRD(t *testing.T) {
	client := newFakeClient()

	crd := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "kubeflow.org/v2beta1",
			"kind":       "MPIJob",
			"metadata": map[string]interface{}{
				"name":      "mpi-test-job",
				"namespace": "default",
			},
		},
	}

	err := client.Create(context.Background(), crd)
	require.NoError(t, err)

	obj, err := client.Get(context.Background(), "MPIJob", "default", "mpi-test-job")
	require.NoError(t, err)
	require.Equal(t, "mpi-test-job", obj.GetName())
}

func TestClient_KindToAPIVersion(t *testing.T) {
	tests := []struct {
		name       string
		mpiVersion string
		kind       string
		want       string
		wantErr    bool
	}{
		{name: "PyTorchJob", mpiVersion: "", kind: "PyTorchJob", want: "kubeflow.org/v1"},
		{name: "TFJob", mpiVersion: "", kind: "TFJob", want: "kubeflow.org/v1"},
		{name: "MPIJob v2beta1", mpiVersion: "v2beta1", kind: "MPIJob", want: "kubeflow.org/v2beta1"},
		{name: "MPIJob v1", mpiVersion: "v1", kind: "MPIJob", want: "kubeflow.org/v1"},
		{name: "MPIJob no version", mpiVersion: "", kind: "MPIJob", want: "", wantErr: true},
		{name: "ConfigMap", mpiVersion: "", kind: "ConfigMap", want: "v1"},
		{name: "Pod", mpiVersion: "", kind: "Pod", want: "v1"},
		{name: "UnknownKind", mpiVersion: "", kind: "UnknownKind", want: "kubeflow.org/v1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{mpiVersion: tt.mpiVersion}
			got, err := c.KindToAPIVersion(tt.kind)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}
