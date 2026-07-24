package cli

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"

	"github.com/kubeflow/arena/pkg/client"
	"github.com/kubeflow/arena/pkg/provider"
	"github.com/kubeflow/arena/pkg/task"
)

// TestWorkflow_SubmitAndDetect verifies that after building a CRD and
// creating a ConfigMap anchor, detectJobType correctly identifies the kind.
func TestWorkflow_SubmitAndDetect(t *testing.T) {
	tk := &task.Task{
		Name:      "wf-job",
		Image:     "pytorch:1.13",
		Run:       "python train.py",
		Framework: task.Framework{Name: "pytorch"},
		Worker:    &task.Worker{Replicas: 2},
	}

	p := &provider.PyTorchProvider{}
	crd, err := p.BuildCRD(tk)
	require.NoError(t, err)
	crd.SetNamespace("default")
	setFrameworkLabel(crd, "pytorch")

	scheme := runtime.NewScheme()
	fakeDynamic := fake.NewSimpleDynamicClientWithCustomListKinds(scheme,
		map[schema.GroupVersionResource]string{
			{Group: "", Version: "v1", Resource: "configmaps"}:              "ConfigMapList",
			{Group: "kubeflow.org", Version: "v1", Resource: "pytorchjobs"}: "PyTorchJobList",
		})
	k8sClient := client.NewClientForInterface(fakeDynamic)

	ctx := context.Background()

	// Create the CRD
	err = k8sClient.Create(ctx, crd)
	require.NoError(t, err)

	// Create the ConfigMap anchor
	yamlContent, err := yaml.Marshal(tk)
	require.NoError(t, err)
	cm := buildConfigMap("wf-job", "default", string(yamlContent), metav1.OwnerReference{
		APIVersion: crd.GetAPIVersion(),
		Kind:       crd.GetKind(),
		Name:       crd.GetName(),
	})
	err = k8sClient.Create(ctx, cm)
	require.NoError(t, err)

	// Detect the job type
	kind, err := detectJobType(ctx, k8sClient, "default", "wf-job")
	require.NoError(t, err)
	assert.Equal(t, "PyTorchJob", kind)
}

// TestWorkflow_frameworkLabelPreservation verifies that mpi, horovod, and
// deepspeed all produce MPIJob CRDs but retain distinct framework labels.
func TestWorkflow_frameworkLabelPreservation(t *testing.T) {
	frameworks := []struct {
		name     string
		expected string
	}{
		{"mpi", "mpi"},
		{"horovod", "horovod"},
		{"deepspeed", "deepspeed"},
	}

	for _, fw := range frameworks {
		t.Run(fw.name, func(t *testing.T) {
			tk := &task.Task{
				Name:      "test-" + fw.name,
				Image:     "mpi:latest",
				Run:       "mpirun train",
				Framework: task.Framework{Name: fw.name},
				Worker:    &task.Worker{Replicas: 2},
			}

			p := &provider.MPIProvider{APIVersion: provider.MPIAPIVersionV2beta1}
			crd, err := p.BuildCRD(tk)
			require.NoError(t, err)
			setFrameworkLabel(crd, fw.name)

			assert.Equal(t, "MPIJob", crd.GetKind(), "all three frameworks should produce MPIJob")
			assert.Equal(t, fw.expected, crd.GetLabels()[frameworkLabel],
				"framework label should preserve original name %q", fw.name)
		})
	}
}

// TestWorkflow_DryRunOutputValidation verifies that dry-run output for
// each framework produces valid JSON with the correct kind and framework label.
func TestWorkflow_DryRunOutputValidation(t *testing.T) {
	tests := []struct {
		framework string
		kind      string
	}{
		{"pytorch", "PyTorchJob"},
		{"tensorflow", "TFJob"},
		{"mpi", "MPIJob"},
	}

	for _, tt := range tests {
		t.Run(tt.framework, func(t *testing.T) {
			tk := &task.Task{
				Name:      "dry-" + tt.framework,
				Image:     "image:latest",
				Run:       "echo hello",
				Framework: task.Framework{Name: tt.framework},
				Worker:    &task.Worker{Replicas: 1},
			}

			p, err := getProvider(tt.framework)
			require.NoError(t, err)

			// MPI provider requires an explicit APIVersion before BuildCRD
			if mpiP, ok := p.(*provider.MPIProvider); ok {
				mpiP.APIVersion = provider.MPIAPIVersionV2beta1
			}

			crd, err := p.BuildCRD(tk)
			require.NoError(t, err)
			setFrameworkLabel(crd, tt.framework)

			data, err := json.MarshalIndent(crd.Object, "", "  ")
			require.NoError(t, err)

			var parsed map[string]interface{}
			err = json.Unmarshal(data, &parsed)
			require.NoError(t, err)

			assert.Equal(t, tt.kind, parsed["kind"])
			metadata := parsed["metadata"].(map[string]interface{})
			labels := metadata["labels"].(map[string]interface{})
			assert.Equal(t, tt.framework, labels[frameworkLabel])
		})
	}
}

// TestWorkflow_VersionValidation verifies YAML version field handling.
func TestWorkflow_VersionValidation(t *testing.T) {
	t.Run("valid version", func(t *testing.T) {
		content := "version: 0.1.0\nname: ver-test\nimage: pytorch:1.13\nrun: echo hello\nframework:\n  name: pytorch\nworker:\n  replicas: 1\n"
		tk, err := task.LoadFromBytes([]byte(content))
		require.NoError(t, err)
		assert.Equal(t, "0.1.0", tk.Version)
	})

	t.Run("no version defaults", func(t *testing.T) {
		content := "name: no-ver\nimage: pytorch:1.13\nrun: echo hello\nframework:\n  name: pytorch\nworker:\n  replicas: 1\n"
		tk, err := task.LoadFromBytes([]byte(content))
		require.NoError(t, err)
		// Version field is optional
		assert.True(t, tk.Version == "" || tk.Version == "0.1.0")
	})
}

// TestWorkflow_ErrorPaths verifies error handling for invalid inputs.
func TestWorkflow_ErrorPaths(t *testing.T) {
	t.Run("unsupported framework", func(t *testing.T) {
		_, err := getProvider("jax")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported framework")
	})

	t.Run("detect nonexistent job", func(t *testing.T) {
		scheme := runtime.NewScheme()
		fakeClient := fake.NewSimpleDynamicClientWithCustomListKinds(scheme,
			map[schema.GroupVersionResource]string{
				{Group: "", Version: "v1", Resource: "configmaps"}: "ConfigMapList",
			})
		k8sClient := client.NewClientForInterface(fakeClient)

		_, err := detectJobType(context.Background(), k8sClient, "default", "ghost")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("checkJobExists when no job exists", func(t *testing.T) {
		scheme := runtime.NewScheme()
		fakeClient := fake.NewSimpleDynamicClientWithCustomListKinds(scheme,
			map[schema.GroupVersionResource]string{
				{Group: "", Version: "v1", Resource: "configmaps"}: "ConfigMapList",
			})
		k8sClient := client.NewClientForInterface(fakeClient)

		err := checkJobExists(context.Background(), k8sClient, "default", "no-such-job")
		assert.NoError(t, err, "no error when no job exists")
	})
}

// TestWorkflow_SuspendAndResumeRoundtrip verifies suspend then resume round-trip.
func TestWorkflow_SuspendAndResumeRoundtrip(t *testing.T) {
	yamlContent := "name: roundtrip\nframework:\n  name: pytorch\n"
	cm := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata":   map[string]interface{}{"name": "roundtrip", "namespace": "default"},
			"data":       map[string]interface{}{"arena-v2.yaml": yamlContent},
		},
	}
	job := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "kubeflow.org/v1",
			"kind":       "PyTorchJob",
			"metadata":   map[string]interface{}{"name": "roundtrip", "namespace": "default"},
			"spec":       map[string]interface{}{"runPolicy": map[string]interface{}{}},
		},
	}
	scheme := runtime.NewScheme()
	fakeClient := fake.NewSimpleDynamicClientWithCustomListKinds(scheme,
		map[schema.GroupVersionResource]string{
			{Group: "", Version: "v1", Resource: "configmaps"}:              "ConfigMapList",
			{Group: "kubeflow.org", Version: "v1", Resource: "pytorchjobs"}: "PyTorchJobList",
		}, cm, job)
	k8sClient := client.NewClientForInterface(fakeClient)
	ctx := context.Background()

	// Suspend the job
	jobType, err := suspendJob(ctx, k8sClient, "default", "roundtrip")
	require.NoError(t, err)
	assert.Equal(t, "PyTorchJob", jobType)

	updated, _ := k8sClient.Get(ctx, "PyTorchJob", "default", "roundtrip")
	suspend, _, _ := unstructured.NestedBool(updated.Object, "spec", "runPolicy", "suspend")
	assert.True(t, suspend, "job should be suspended after suspend")

	// Resume the job
	jobType, err = resumeJob(ctx, k8sClient, "default", "roundtrip")
	require.NoError(t, err)
	assert.Equal(t, "PyTorchJob", jobType)

	updated, _ = k8sClient.Get(ctx, "PyTorchJob", "default", "roundtrip")
	suspend, _, _ = unstructured.NestedBool(updated.Object, "spec", "runPolicy", "suspend")
	assert.False(t, suspend, "job should be running after resume")
}
