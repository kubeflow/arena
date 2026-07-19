package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"

	"github.com/kubeflow/arena/pkg/client"
	"github.com/kubeflow/arena/pkg/provider"
)

func newFakeK8sClient(t *testing.T) *client.Client {
	t.Helper()
	scheme := runtime.NewScheme()
	listKinds := map[schema.GroupVersionResource]string{
		{Group: "kubeflow.org", Version: "v1", Resource: "pytorchjobs"}:  "PyTorchJobList",
		{Group: "kubeflow.org", Version: "v1", Resource: "tfjobs"}:       "TFJobList",
		{Group: "kubeflow.org", Version: "v1", Resource: "mpijobs"}:      "MPIJobList",
		{Group: "kubeflow.org", Version: "v2beta1", Resource: "mpijobs"}: "MPIJobList",
	}
	fakeDynamic := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(scheme, listKinds)
	c := client.NewClientForInterface(fakeDynamic)
	c.SetMPIVersion("v1")
	return c
}

func providerFor(framework string) provider.Provider {
	switch framework {
	case "pytorch":
		return &provider.PyTorchProvider{}
	case "tensorflow":
		return &provider.TensorFlowProvider{}
	case "mpi":
		return &provider.MPIProvider{APIVersion: provider.MPIAPIVersionV1}
	default:
		return nil
	}
}

func examplesDir(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	require.NoError(t, err)
	return filepath.Join(wd, "..", "examples", "v2")
}

func testdataDir(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	require.NoError(t, err)
	return filepath.Join(wd, "testdata")
}

func containsAny(s string, substrings ...string) bool {
	for _, sub := range substrings {
		if len(s) >= len(sub) {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
		}
	}
	return false
}
