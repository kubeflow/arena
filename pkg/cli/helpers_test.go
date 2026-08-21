package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"

	"github.com/kubeflow/arena/pkg/client"
	outputpkg "github.com/kubeflow/arena/pkg/output"
)

func TestDryRunFormat(t *testing.T) {
	orig := outputFormat
	t.Cleanup(func() { outputFormat = orig })

	tests := []struct {
		format  string
		want    outputpkg.Format
		wantErr bool
	}{
		{"json", outputpkg.FormatJSON, false},
		{"yaml", outputpkg.FormatYAML, false},
		{"table", "", true},
		{"wide", "", true},
		{"", "", true},
	}
	for _, tt := range tests {
		t.Run("format_"+tt.format, func(t *testing.T) {
			outputFormat = tt.format
			got, err := dryRunFormat()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "dry-run only supports -o json or yaml")
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestPrintCRD_UnsupportedFormat(t *testing.T) {
	crd := &unstructured.Unstructured{Object: map[string]interface{}{"kind": "PyTorchJob"}}
	var buf bytes.Buffer
	err := printCRD(&buf, crd, outputpkg.FormatTable)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported CRD output format")
}

func TestPrintCRD_YAMLAndJSON(t *testing.T) {
	crd := &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "kubeflow.org/v1",
		"kind":       "PyTorchJob",
	}}
	var yamlBuf bytes.Buffer
	require.NoError(t, printCRD(&yamlBuf, crd, outputpkg.FormatYAML))
	assert.Contains(t, yamlBuf.String(), "kind: PyTorchJob")

	var jsonBuf bytes.Buffer
	require.NoError(t, printCRD(&jsonBuf, crd, outputpkg.FormatJSON))
	assert.Contains(t, jsonBuf.String(), `"kind": "PyTorchJob"`)
}

func TestApplyDryRunOutputDefault(t *testing.T) {
	orig := outputFormat
	t.Cleanup(func() { outputFormat = orig })

	newCmd := func() *cobra.Command {
		cmd := &cobra.Command{Use: "x"}
		cmd.Flags().StringVarP(&outputFormat, "output", "o", string(outputpkg.DefaultFormat), "")
		return cmd
	}

	// dry-run without explicit -o: defaults to json
	outputFormat = string(outputpkg.DefaultFormat)
	applyDryRunOutputDefault(newCmd(), true)
	assert.Equal(t, string(outputpkg.FormatJSON), outputFormat)

	// dry-run with explicit -o yaml: kept
	cmd := newCmd()
	require.NoError(t, cmd.Flags().Set("output", "yaml"))
	applyDryRunOutputDefault(cmd, true)
	assert.Equal(t, "yaml", outputFormat)

	// no dry-run: untouched
	outputFormat = string(outputpkg.DefaultFormat)
	applyDryRunOutputDefault(newCmd(), false)
	assert.Equal(t, string(outputpkg.DefaultFormat), outputFormat)
}

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
	assert.Contains(t, err.Error(), "not supported")
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
	assert.Contains(t, err.Error(), "mpijob crd not found")
}

// newFakeK8sClient creates a fake K8s client with PyTorchJob, TFJob, and MPIJob
// GVRs pre-registered for Create/Get/List/Delete operations.
func newFakeK8sClient(t *testing.T) *client.Client {
	t.Helper()
	scheme := runtime.NewScheme()
	listKinds := map[schema.GroupVersionResource]string{
		{Group: "kubeflow.org", Version: "v1", Resource: "pytorchjobs"}:  "PyTorchJobList",
		{Group: "kubeflow.org", Version: "v1", Resource: "tfjobs"}:       "TFJobList",
		{Group: "kubeflow.org", Version: "v1", Resource: "mpijobs"}:      "MPIJobList",
		{Group: "kubeflow.org", Version: "v2beta1", Resource: "mpijobs"}: "MPIJobList",
	}
	fakeDynamic := fake.NewSimpleDynamicClientWithCustomListKinds(scheme, listKinds)
	c := client.NewClientForInterface(fakeDynamic)
	c.SetMPIVersion("v1")
	return c
}

// examplesDir resolves the path to examples/v2/quickstart relative to pkg/cli/.
func examplesDir(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	require.NoError(t, err)
	return filepath.Join(wd, "..", "..", "examples", "v2", "quickstart")
}
