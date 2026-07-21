package cli

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"

	"github.com/kubeflow/arena/pkg/client"
)

func TestDeleteCmd_AcceptsOneArg(t *testing.T) {
	err := deleteCmd.Args(deleteCmd, []string{"my-job"})
	assert.NoError(t, err)
}

func TestDeleteCmd_AcceptsZeroArgs(t *testing.T) {
	// With -f flag support, zero positional args is valid (name comes from file).
	err := deleteCmd.Args(deleteCmd, []string{})
	assert.NoError(t, err)
}

func TestDeleteCmd_RejectsExtraArgs(t *testing.T) {
	err := deleteCmd.Args(deleteCmd, []string{"a", "b"})
	assert.Error(t, err)
}

func TestDeleteCmd_RegisteredWithJob(t *testing.T) {
	found := false
	for _, cmd := range jobCmd.Commands() {
		if cmd.Name() == "delete" {
			found = true
			break
		}
	}
	assert.True(t, found, "delete command should be registered with job command")
}

func TestDeleteCmd_HasCorrectMetadata(t *testing.T) {
	assert.Equal(t, "delete [name]", deleteCmd.Use)
	assert.NotEmpty(t, deleteCmd.Short)
}

func TestDeleteCmd_HasFileFlag(t *testing.T) {
	f := deleteCmd.Flags().Lookup("file")
	require.NotNil(t, f, "delete command should have a --file flag")
	assert.Equal(t, "f", f.Shorthand)
}

func TestDeleteCmd_NotFound(t *testing.T) {
	orig := kubeconfig
	defer func() { kubeconfig = orig }()

	t.Setenv("KUBECONFIG", "/nonexistent/env-kubeconfig")
	kubeconfig = "/nonexistent/kubeconfig"
	err := deleteCmd.RunE(deleteCmd, []string{"nonexistent-job"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create K8s client")
}

func TestDeleteCmd_DeletesJobViaFakeClient(t *testing.T) {
	yamlContent := "name: del-job\nframework:\n  name: pytorch\n"
	cm := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name": "del-job", "namespace": "default",
			},
			"data": map[string]interface{}{
				"arena-v2.yaml": yamlContent,
			},
		},
	}
	job := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "kubeflow.org/v1",
			"kind":       "PyTorchJob",
			"metadata": map[string]interface{}{
				"name": "del-job", "namespace": "default",
			},
		},
	}
	scheme := runtime.NewScheme()
	fakeClient := fake.NewSimpleDynamicClientWithCustomListKinds(scheme,
		map[schema.GroupVersionResource]string{
			{Group: "", Version: "v1", Resource: "configmaps"}:              "ConfigMapList",
			{Group: "kubeflow.org", Version: "v1", Resource: "pytorchjobs"}: "PyTorchJobList",
		}, cm, job)
	k8sClient := client.NewClientForInterface(fakeClient)

	// Verify detectJobType works
	kind, err := detectJobType(context.Background(), k8sClient, "default", "del-job")
	assert.NoError(t, err)
	assert.Equal(t, "PyTorchJob", kind)

	// Delete the job
	err = k8sClient.Delete(context.Background(), "PyTorchJob", "default", "del-job")
	assert.NoError(t, err)

	// Verify it's gone
	_, err = k8sClient.Get(context.Background(), "PyTorchJob", "default", "del-job")
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// New tests for -f flag support
// ---------------------------------------------------------------------------

func TestDelete_ByName(t *testing.T) {
	// Providing a positional arg should still work (no -f flag).
	// We verify RunE fails at client creation (no real cluster), proving
	// the name was accepted and execution proceeded past arg parsing.
	orig := kubeconfig
	defer func() { kubeconfig = orig }()

	t.Setenv("KUBECONFIG", "/nonexistent/env-kubeconfig")
	kubeconfig = "/nonexistent/kubeconfig"

	origFile := deleteFile
	defer func() { deleteFile = origFile }()
	deleteFile = ""

	err := deleteCmd.RunE(deleteCmd, []string{"my-job"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create K8s client")
}

func TestDelete_ByFile(t *testing.T) {
	// Create a valid YAML file; RunE should extract the name and proceed
	// to client creation (which fails because there's no real cluster).
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "job.yaml")
	content := `name: file-delete-job
image: pytorch:1.13
run: python train.py
framework:
  name: pytorch
worker:
  replicas: 1
`
	err := os.WriteFile(tmpFile, []byte(content), 0600)
	require.NoError(t, err)

	orig := kubeconfig
	defer func() { kubeconfig = orig }()

	t.Setenv("KUBECONFIG", "/nonexistent/env-kubeconfig")
	kubeconfig = "/nonexistent/kubeconfig"

	origFile := deleteFile
	defer func() { deleteFile = origFile }()
	deleteFile = tmpFile

	err = deleteCmd.RunE(deleteCmd, nil)
	assert.Error(t, err)
	// Should get past file loading and fail at client creation.
	assert.Contains(t, err.Error(), "failed to create K8s client")
}

func TestDelete_FileNotFound(t *testing.T) {
	origFile := deleteFile
	defer func() { deleteFile = origFile }()
	deleteFile = "/nonexistent/path/to/missing.yaml"

	orig := kubeconfig
	defer func() { kubeconfig = orig }()
	t.Setenv("KUBECONFIG", "/nonexistent/env-kubeconfig")
	kubeconfig = "/nonexistent/kubeconfig"

	err := deleteCmd.RunE(deleteCmd, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load file")
}

func TestDelete_NoNameNoFile(t *testing.T) {
	origFile := deleteFile
	defer func() { deleteFile = origFile }()
	deleteFile = ""

	err := deleteCmd.RunE(deleteCmd, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "either job name or -f flag is required")
}

func TestDelete_FileWithEmptyName(t *testing.T) {
	// A YAML file with an empty name should fail validation during LoadFromFile.
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "empty-name.yaml")
	content := `name: ""
image: pytorch:1.13
run: python train.py
framework:
  name: pytorch
worker:
  replicas: 1
`
	err := os.WriteFile(tmpFile, []byte(content), 0600)
	require.NoError(t, err)

	origFile := deleteFile
	defer func() { deleteFile = origFile }()
	deleteFile = tmpFile

	err = deleteCmd.RunE(deleteCmd, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load file")
}

// TestDeleteCmd_DeletesMPIJobViaFakeClient verifies the end-to-end flow that
// was broken by the original bug: detectJobType resolves MPIVersion via the
// ConfigMap anchor (Phase 1), then Delete succeeds. Before the fix, Phase 1
// returned without calling ResolveMPIVersion, causing Delete to fail with
// "MPIJob version not resolved".
func TestDeleteCmd_DeletesMPIJobViaFakeClient(t *testing.T) {
	yamlContent := "name: del-mpi-job\nframework:\n  name: mpi\n"
	cm := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name": "del-mpi-job", "namespace": "default",
			},
			"data": map[string]interface{}{
				"arena-v2.yaml": yamlContent,
			},
		},
	}
	mpiCRD := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apiextensions.k8s.io/v1",
			"kind":       "CustomResourceDefinition",
			"metadata": map[string]interface{}{
				"name": "mpijobs.kubeflow.org",
			},
			"spec": map[string]interface{}{
				"versions": []interface{}{
					map[string]interface{}{
						"name":    "v2beta1",
						"served":  true,
						"storage": true,
					},
				},
			},
		},
	}
	job := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "kubeflow.org/v2beta1",
			"kind":       "MPIJob",
			"metadata": map[string]interface{}{
				"name": "del-mpi-job", "namespace": "default",
			},
		},
	}
	scheme := runtime.NewScheme()
	listKinds := map[schema.GroupVersionResource]string{
		{Group: "", Version: "v1", Resource: "configmaps"}:                                    "ConfigMapList",
		{Group: "apiextensions.k8s.io", Version: "v1", Resource: "customresourcedefinitions"}: "CustomResourceDefinitionList",
		{Group: "kubeflow.org", Version: "v2beta1", Resource: "mpijobs"}:                      "MPIJobList",
	}
	fakeClient := fake.NewSimpleDynamicClientWithCustomListKinds(scheme, listKinds, cm, mpiCRD, job)
	k8sClient := client.NewClientForInterface(fakeClient)

	// detectJobType should resolve MPIVersion automatically via the CRD definition
	kind, err := detectJobType(context.Background(), k8sClient, "default", "del-mpi-job")
	assert.NoError(t, err)
	assert.Equal(t, "MPIJob", kind)
	assert.Equal(t, "v2beta1", k8sClient.GetMPIVersion(), "MPIVersion should be resolved by detectJobType")

	// Delete should succeed — this was the original bug
	err = k8sClient.Delete(context.Background(), "MPIJob", "default", "del-mpi-job")
	assert.NoError(t, err)

	// Verify it's gone
	_, err = k8sClient.Get(context.Background(), "MPIJob", "default", "del-mpi-job")
	assert.Error(t, err)
}
