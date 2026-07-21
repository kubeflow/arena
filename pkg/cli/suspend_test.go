package cli

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"

	"github.com/kubeflow/arena/pkg/client"
)

func TestSuspendCmd_RequiresArg(t *testing.T) {
	err := suspendCmd.Args(suspendCmd, []string{})
	assert.Error(t, err)
}

func TestSuspendCmd_AcceptsOneArg(t *testing.T) {
	err := suspendCmd.Args(suspendCmd, []string{"my-job"})
	assert.NoError(t, err)
}

func TestSuspendCmd_RejectsExtraArgs(t *testing.T) {
	err := suspendCmd.Args(suspendCmd, []string{"job1", "job2"})
	assert.Error(t, err)
}

func TestSuspendCmd_RegisteredWithJob(t *testing.T) {
	found := false
	for _, cmd := range jobCmd.Commands() {
		if cmd.Name() == "suspend" {
			found = true
			break
		}
	}
	assert.True(t, found, "suspend command should be registered with job command")
}

func TestSuspendCmd_NotFound(t *testing.T) {
	orig := kubeconfig
	defer func() { kubeconfig = orig }()

	kubeconfig = "/nonexistent/kubeconfig"
	err := suspendCmd.RunE(suspendCmd, []string{"nonexistent-job"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create K8s client")
}

func TestSuspendCmd_HasCorrectMetadata(t *testing.T) {
	assert.Equal(t, "suspend <name>", suspendCmd.Use)
	assert.NotEmpty(t, suspendCmd.Short)
}

func TestSuspendCmd_PatchesSuspendField(t *testing.T) {
	yamlContent := `
name: test-job
framework:
  name: pytorch
`
	cm := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name": "test-job", "namespace": "default",
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
				"name": "test-job", "namespace": "default",
			},
			"spec": map[string]interface{}{
				"runPolicy": map[string]interface{}{},
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

	jobType, err := suspendJob(context.Background(), k8sClient, "default", "test-job")
	assert.NoError(t, err)
	assert.Equal(t, "PyTorchJob", jobType)

	updated, _ := k8sClient.Get(context.Background(), "PyTorchJob", "default", "test-job")
	suspend, _, _ := unstructured.NestedBool(updated.Object, "spec", "runPolicy", "suspend")
	assert.True(t, suspend)
}

func TestSuspendCmd_TFJob(t *testing.T) {
	yamlContent := `
name: tf-job
framework:
  name: tensorflow
`
	cm := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name": "tf-job", "namespace": "default",
			},
			"data": map[string]interface{}{
				"arena-v2.yaml": yamlContent,
			},
		},
	}
	job := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "kubeflow.org/v1",
			"kind":       "TFJob",
			"metadata": map[string]interface{}{
				"name": "tf-job", "namespace": "default",
			},
			"spec": map[string]interface{}{
				"runPolicy": map[string]interface{}{},
			},
		},
	}
	scheme := runtime.NewScheme()
	fakeClient := fake.NewSimpleDynamicClientWithCustomListKinds(scheme,
		map[schema.GroupVersionResource]string{
			{Group: "", Version: "v1", Resource: "configmaps"}:         "ConfigMapList",
			{Group: "kubeflow.org", Version: "v1", Resource: "tfjobs"}: "TFJobList",
		}, cm, job)
	k8sClient := client.NewClientForInterface(fakeClient)

	jobType, err := suspendJob(context.Background(), k8sClient, "default", "tf-job")
	assert.NoError(t, err)
	assert.Equal(t, "TFJob", jobType)

	updated, _ := k8sClient.Get(context.Background(), "TFJob", "default", "tf-job")
	suspend, _, _ := unstructured.NestedBool(updated.Object, "spec", "runPolicy", "suspend")
	assert.True(t, suspend)
}

func TestSuspendCmd_MPIJob(t *testing.T) {
	yamlContent := `
name: mpi-job
framework:
  name: mpi
`
	cm := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name": "mpi-job", "namespace": "default",
			},
			"data": map[string]interface{}{
				"arena-v2.yaml": yamlContent,
			},
		},
	}
	job := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "kubeflow.org/v2beta1",
			"kind":       "MPIJob",
			"metadata": map[string]interface{}{
				"name": "mpi-job", "namespace": "default",
			},
			"spec": map[string]interface{}{
				"runPolicy": map[string]interface{}{},
			},
		},
	}
	scheme := runtime.NewScheme()
	fakeClient := fake.NewSimpleDynamicClientWithCustomListKinds(scheme,
		map[schema.GroupVersionResource]string{
			{Group: "", Version: "v1", Resource: "configmaps"}:               "ConfigMapList",
			{Group: "kubeflow.org", Version: "v2beta1", Resource: "mpijobs"}: "MPIJobList",
		}, cm, job)
	k8sClient := client.NewClientForInterface(fakeClient)
	k8sClient.SetMPIVersion("v2beta1")

	jobType, err := suspendJob(context.Background(), k8sClient, "default", "mpi-job")
	assert.NoError(t, err)
	assert.Equal(t, "MPIJob", jobType)

	updated, _ := k8sClient.Get(context.Background(), "MPIJob", "default", "mpi-job")
	suspend, _, _ := unstructured.NestedBool(updated.Object, "spec", "runPolicy", "suspend")
	assert.True(t, suspend)
}

func TestSuspendCmd_JobNotFound(t *testing.T) {
	scheme := runtime.NewScheme()
	fakeClient := fake.NewSimpleDynamicClientWithCustomListKinds(scheme,
		map[schema.GroupVersionResource]string{
			{Group: "", Version: "v1", Resource: "configmaps"}: "ConfigMapList",
		})
	k8sClient := client.NewClientForInterface(fakeClient)

	_, err := suspendJob(context.Background(), k8sClient, "default", "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}
