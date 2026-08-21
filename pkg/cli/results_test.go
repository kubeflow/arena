package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"

	"github.com/kubeflow/arena/pkg/client"
	"github.com/kubeflow/arena/pkg/task"
)

func TestPrintSubmitResult_DefaultText(t *testing.T) {
	orig := outputFormat
	t.Cleanup(func() { outputFormat = orig })
	outputFormat = "table"

	var buf bytes.Buffer
	require.NoError(t, printSubmitResult(&buf, "my-job", "default", "PyTorchJob", "kubeflow.org/v1"))
	assert.Equal(t, "Job my-job submitted successfully\n", buf.String())
}

func TestPrintSubmitResult_JSON(t *testing.T) {
	orig := outputFormat
	t.Cleanup(func() { outputFormat = orig })
	outputFormat = "json"

	var buf bytes.Buffer
	require.NoError(t, printSubmitResult(&buf, "my-job", "default", "PyTorchJob", "kubeflow.org/v1"))

	var parsed SubmitResult
	require.NoError(t, json.Unmarshal(buf.Bytes(), &parsed))
	assert.Equal(t, SubmitResult{
		Name:       "my-job",
		Namespace:  "default",
		Kind:       "PyTorchJob",
		APIVersion: "kubeflow.org/v1",
	}, parsed)
}

func TestPrintSubmitResult_YAML(t *testing.T) {
	orig := outputFormat
	t.Cleanup(func() { outputFormat = orig })
	outputFormat = "yaml"

	var buf bytes.Buffer
	require.NoError(t, printSubmitResult(&buf, "my-job", "default", "PyTorchJob", "kubeflow.org/v1"))
	assert.Contains(t, buf.String(), "name: my-job")
	assert.Contains(t, buf.String(), "kind: PyTorchJob")
	assert.Contains(t, buf.String(), "apiVersion: kubeflow.org/v1")
}

func TestPrintActionResult_DefaultText(t *testing.T) {
	orig := outputFormat
	t.Cleanup(func() { outputFormat = orig })
	outputFormat = "table"

	var buf bytes.Buffer
	require.NoError(t, printActionResult(&buf, "my-job", "default", "PyTorchJob", "deleted"))
	assert.Equal(t, "pytorchjob/my-job deleted\n", buf.String())
}

func TestPrintActionResult_JSON(t *testing.T) {
	orig := outputFormat
	t.Cleanup(func() { outputFormat = orig })
	outputFormat = "json"

	var buf bytes.Buffer
	require.NoError(t, printActionResult(&buf, "my-job", "default", "PyTorchJob", "suspended"))

	var parsed ActionResult
	require.NoError(t, json.Unmarshal(buf.Bytes(), &parsed))
	assert.Equal(t, ActionResult{
		Name:      "my-job",
		Namespace: "default",
		Kind:      "PyTorchJob",
		Action:    "suspended",
	}, parsed)
}

func TestSubmitCRD_JSONOutput(t *testing.T) {
	origFormat := outputFormat
	origNS := namespace
	t.Cleanup(func() {
		outputFormat = origFormat
		namespace = origNS
	})
	outputFormat = "json"
	namespace = "default"

	scheme := runtime.NewScheme()
	listKinds := map[schema.GroupVersionResource]string{
		{Group: "", Version: "v1", Resource: "configmaps"}:                            "ConfigMapList",
		{Group: "", Version: "v1", Resource: "serviceaccounts"}:                       "ServiceAccountList",
		{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "roles"}:        "RoleList",
		{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "rolebindings"}: "RoleBindingList",
		{Group: "kubeflow.org", Version: "v1", Resource: "pytorchjobs"}:               "PyTorchJobList",
	}
	fakeClient := fake.NewSimpleDynamicClientWithCustomListKinds(scheme, listKinds)
	k8sClient := client.NewClientForInterface(fakeClient)

	tk := &task.Task{
		Name:      "json-submit-test",
		Image:     "pytorch:2.1",
		Framework: task.Framework{Name: "pytorch"},
		Run:       "python train.py",
		Worker:    &task.Worker{Replicas: 1},
	}
	tk.SetDefaults()

	var buf bytes.Buffer
	err := submitCRD(context.Background(), &buf, k8sClient, tk, "pytorch", false)
	require.NoError(t, err)

	var parsed SubmitResult
	require.NoError(t, json.Unmarshal(buf.Bytes(), &parsed))
	assert.Equal(t, "json-submit-test", parsed.Name)
	assert.Equal(t, "default", parsed.Namespace)
	assert.Equal(t, "PyTorchJob", parsed.Kind)
	assert.Equal(t, "kubeflow.org/v1", parsed.APIVersion)
}

func TestSubmitCRD_DefaultOutputUnchanged(t *testing.T) {
	origFormat := outputFormat
	origNS := namespace
	t.Cleanup(func() {
		outputFormat = origFormat
		namespace = origNS
	})
	outputFormat = "table"
	namespace = "default"

	scheme := runtime.NewScheme()
	listKinds := map[schema.GroupVersionResource]string{
		{Group: "", Version: "v1", Resource: "configmaps"}:                            "ConfigMapList",
		{Group: "", Version: "v1", Resource: "serviceaccounts"}:                       "ServiceAccountList",
		{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "roles"}:        "RoleList",
		{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "rolebindings"}: "RoleBindingList",
		{Group: "kubeflow.org", Version: "v1", Resource: "pytorchjobs"}:               "PyTorchJobList",
	}
	fakeClient := fake.NewSimpleDynamicClientWithCustomListKinds(scheme, listKinds)
	k8sClient := client.NewClientForInterface(fakeClient)

	tk := &task.Task{
		Name:      "text-submit-test",
		Image:     "pytorch:2.1",
		Framework: task.Framework{Name: "pytorch"},
		Run:       "python train.py",
		Worker:    &task.Worker{Replicas: 1},
	}
	tk.SetDefaults()

	var buf bytes.Buffer
	err := submitCRD(context.Background(), &buf, k8sClient, tk, "pytorch", false)
	require.NoError(t, err)
	assert.Equal(t, "Job text-submit-test submitted successfully\n", buf.String())
}
