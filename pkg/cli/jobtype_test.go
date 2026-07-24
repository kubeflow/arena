package cli

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"
	k8stesting "k8s.io/client-go/testing"

	"github.com/kubeflow/arena/pkg/client"
)

func TestDetectJobType_PyTorchJob(t *testing.T) {
	yamlContent := `
name: test-job
framework:
  name: pytorch
worker:
  replicas: 2
  resources:
    nvidia.com/gpu: "1"
`
	cm := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name":      "test-job",
				"namespace": "default",
			},
			"data": map[string]interface{}{
				"arena-v2.yaml": yamlContent,
			},
		},
	}

	scheme := runtime.NewScheme()
	gvr := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"}
	fakeClient := fake.NewSimpleDynamicClientWithCustomListKinds(
		scheme,
		map[schema.GroupVersionResource]string{gvr: "ConfigMapList"},
		cm,
	)

	k8sClient := client.NewClientForInterface(fakeClient)

	jobType, err := detectJobType(context.Background(), k8sClient, "default", "test-job")

	assert.NoError(t, err)
	assert.Equal(t, "PyTorchJob", jobType)
}

func TestDetectJobType_MPIJob(t *testing.T) {
	yamlContent := `
name: test-job
framework:
  name: mpi
worker:
  replicas: 4
`
	cm := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name":      "test-job",
				"namespace": "default",
			},
			"data": map[string]interface{}{
				"arena-v2.yaml": yamlContent,
			},
		},
	}

	scheme := runtime.NewScheme()
	gvr := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"}
	fakeClient := fake.NewSimpleDynamicClientWithCustomListKinds(
		scheme,
		map[schema.GroupVersionResource]string{gvr: "ConfigMapList"},
		cm,
	)

	k8sClient := client.NewClientForInterface(fakeClient)
	k8sClient.SetMPIVersion("v2beta1")

	jobType, err := detectJobType(context.Background(), k8sClient, "default", "test-job")

	assert.NoError(t, err)
	assert.Equal(t, "MPIJob", jobType)
}

// TestDetectJobType_MPIJob_Phase1ResolvesVersion verifies that when Phase 1
// (ConfigMap anchor) finds an MPIJob, ResolveMPIVersion is called and
// c.mpiVersion is populated. This is the regression test for the bug where
// Phase 1 returned without resolving the version.
func TestDetectJobType_MPIJob_Phase1ResolvesVersion(t *testing.T) {
	yamlContent := `
name: mpi-version-job
framework:
  name: mpi
`
	cm := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name": "mpi-version-job", "namespace": "default",
			},
			"data": map[string]interface{}{
				"arena-v2.yaml": yamlContent,
			},
		},
	}
	// MPIJob CRD object with storage version v2beta1
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

	scheme := runtime.NewScheme()
	listKinds := map[schema.GroupVersionResource]string{
		{Group: "", Version: "v1", Resource: "configmaps"}:                                    "ConfigMapList",
		{Group: "apiextensions.k8s.io", Version: "v1", Resource: "customresourcedefinitions"}: "CustomResourceDefinitionList",
	}
	fakeClient := fake.NewSimpleDynamicClientWithCustomListKinds(scheme, listKinds, cm, mpiCRD)
	k8sClient := client.NewClientForInterface(fakeClient)

	jobType, err := detectJobType(context.Background(), k8sClient, "default", "mpi-version-job")

	assert.NoError(t, err)
	assert.Equal(t, "MPIJob", jobType)
	assert.Equal(t, "v2beta1", k8sClient.GetMPIVersion(),
		"MPIVersion must be resolved after detectJobType returns MPIJob via Phase 1")
}

// TestDetectJobType_MPIJob_CRDNotInstalled verifies that when Phase 1 finds
// an MPIJob via ConfigMap but the MPIJob CRD is not installed on the cluster,
// detectJobType returns an explicit error rather than letting the downstream
// operation fail with "MPIJob version not resolved".
func TestDetectJobType_MPIJob_CRDNotInstalled(t *testing.T) {
	yamlContent := `
name: mpi-no-crd
framework:
  name: mpi
`
	cm := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name": "mpi-no-crd", "namespace": "default",
			},
			"data": map[string]interface{}{
				"arena-v2.yaml": yamlContent,
			},
		},
	}

	scheme := runtime.NewScheme()
	listKinds := map[schema.GroupVersionResource]string{
		{Group: "", Version: "v1", Resource: "configmaps"}: "ConfigMapList",
	}
	fakeClient := fake.NewSimpleDynamicClientWithCustomListKinds(scheme, listKinds, cm)
	k8sClient := client.NewClientForInterface(fakeClient)

	_, err := detectJobType(context.Background(), k8sClient, "default", "mpi-no-crd")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "MPIJob CRD is not installed")
}

func TestDetectJobType_NotFound(t *testing.T) {
	scheme := runtime.NewScheme()
	fakeClient := fake.NewSimpleDynamicClient(scheme)
	k8sClient := client.NewClientForInterface(fakeClient)

	_, err := detectJobType(context.Background(), k8sClient, "default", "nonexistent-job")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestDetectJobType_InvalidYAML(t *testing.T) {
	cm := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name":      "test-job",
				"namespace": "default",
			},
			"data": map[string]interface{}{
				"arena-v2.yaml": "invalid: yaml: content:",
			},
		},
	}

	scheme := runtime.NewScheme()
	gvr := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"}
	fakeClient := fake.NewSimpleDynamicClientWithCustomListKinds(
		scheme,
		map[schema.GroupVersionResource]string{gvr: "ConfigMapList"},
		cm,
	)

	k8sClient := client.NewClientForInterface(fakeClient)

	_, err := detectJobType(context.Background(), k8sClient, "default", "test-job")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestCheckJobExists_ReturnsErrorWhenExists(t *testing.T) {
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
				"name":      "test-job",
				"namespace": "default",
				"ownerReferences": []interface{}{
					map[string]interface{}{
						"apiVersion": "kubeflow.org/v1",
						"kind":       "PyTorchJob",
						"name":       "test-job",
					},
				},
			},
			"data": map[string]interface{}{
				"arena-v2.yaml": yamlContent,
			},
		},
	}

	scheme := runtime.NewScheme()
	gvr := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"}
	fakeClient := fake.NewSimpleDynamicClientWithCustomListKinds(
		scheme,
		map[schema.GroupVersionResource]string{gvr: "ConfigMapList"},
		cm,
	)

	k8sClient := client.NewClientForInterface(fakeClient)

	err := checkJobExists(context.Background(), k8sClient, "default", "test-job")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
	assert.Contains(t, err.Error(), "PyTorchJob")
}

func TestCheckJobExists_ReturnsNilWhenNotExists(t *testing.T) {
	scheme := runtime.NewScheme()
	fakeClient := fake.NewSimpleDynamicClient(scheme)
	k8sClient := client.NewClientForInterface(fakeClient)

	err := checkJobExists(context.Background(), k8sClient, "default", "nonexistent-job")

	assert.NoError(t, err)
}

// TestCheckJobExists_DistinguishesNotFoundFromAPIError verifies that
// non-NotFound API errors are propagated rather than silently treated as "not found".
func TestCheckJobExists_DistinguishesNotFoundFromAPIError(t *testing.T) {
	scheme := runtime.NewScheme()
	fakeClient := fake.NewSimpleDynamicClient(scheme)

	// Inject an InternalError reactor so Get("configmaps", ...) returns a 500-style API error.
	apiErr := apierrors.NewInternalError(fmt.Errorf("etcd unavailable"))
	fakeClient.PrependReactor("get", "configmaps", func(_ k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, apiErr
	})

	k8sClient := client.NewClientForInterface(fakeClient)

	err := checkJobExists(context.Background(), k8sClient, "default", "some-job")

	assert.Error(t, err, "API errors must be propagated, not silently treated as not-found")
	assert.Contains(t, err.Error(), "failed to check if job exists",
		"error should be wrapped with context")
	assert.Contains(t, err.Error(), "etcd unavailable",
		"underlying API error should be visible in the wrapped message")
}

// TestCheckJobExists_DistinguishesNotFoundFromForbiddenError verifies that
// RBAC/Forbidden errors are propagated rather than silently treated as "not found".
func TestCheckJobExists_DistinguishesNotFoundFromForbiddenError(t *testing.T) {
	scheme := runtime.NewScheme()
	fakeClient := fake.NewSimpleDynamicClient(scheme)

	// Inject a Forbidden error.
	cmGVR := schema.GroupResource{Group: "", Resource: "configmaps"}
	forbiddenErr := apierrors.NewForbidden(cmGVR, "some-job", fmt.Errorf("User cannot get resource"))
	fakeClient.PrependReactor("get", "configmaps", func(_ k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, forbiddenErr
	})

	k8sClient := client.NewClientForInterface(fakeClient)

	err := checkJobExists(context.Background(), k8sClient, "default", "some-job")

	assert.Error(t, err, "Forbidden errors must be propagated, not silently treated as not-found")
	assert.Contains(t, err.Error(), "failed to check if job exists")
}

// TestCheckJobExists_HandlesCorruptConfigMap verifies that when a ConfigMap
// exists but cannot be parsed, checkJobExists still reports the job as existing
// rather than silently discarding the error.
func TestCheckJobExists_HandlesCorruptConfigMap(t *testing.T) {
	// ConfigMap exists but has no arena-v2.yaml key and no ownerReferences.
	cm := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name":      "corrupt-job",
				"namespace": "default",
			},
			"data": map[string]interface{}{
				"other-key.yaml": "some: content",
			},
		},
	}

	scheme := runtime.NewScheme()
	gvr := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"}
	fakeClient := fake.NewSimpleDynamicClientWithCustomListKinds(
		scheme,
		map[schema.GroupVersionResource]string{gvr: "ConfigMapList"},
		cm,
	)

	k8sClient := client.NewClientForInterface(fakeClient)

	err := checkJobExists(context.Background(), k8sClient, "default", "corrupt-job")

	assert.Error(t, err, "corrupt ConfigMap should still report job as existing")
	assert.Contains(t, err.Error(), "already exists",
		"error should indicate the job already exists")
}

// TestCheckJobExists_OwnerReferences verifies that checkJobExists extracts
// the CRD kind from ConfigMap ownerReferences.
func TestCheckJobExists_OwnerReferences(t *testing.T) {
	tests := []struct {
		name string
		kind string
	}{
		{"PyTorchJob ownerRef", "PyTorchJob"},
		{"TFJob ownerRef", "TFJob"},
		{"MPIJob ownerRef", "MPIJob"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "ConfigMap",
					"metadata": map[string]interface{}{
						"name":      "owned-job",
						"namespace": "default",
						"ownerReferences": []interface{}{
							map[string]interface{}{
								"apiVersion": "kubeflow.org/v1",
								"kind":       tt.kind,
								"name":       "owned-job",
							},
						},
					},
					"data": map[string]interface{}{
						"arena-v2.yaml": "name: owned-job\nframework:\n  name: pytorch\n",
					},
				},
			}

			scheme := runtime.NewScheme()
			gvr := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"}
			fakeClient := fake.NewSimpleDynamicClientWithCustomListKinds(
				scheme,
				map[schema.GroupVersionResource]string{gvr: "ConfigMapList"},
				cm,
			)
			k8sClient := client.NewClientForInterface(fakeClient)

			err := checkJobExists(context.Background(), k8sClient, "default", "owned-job")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "already exists")
			assert.Contains(t, err.Error(), tt.kind)
		})
	}
}

// TestDetectJobType_V1Job_PyTorch verifies that a CRD without the
// arena.io/framework label is identified as a v1 job.
func TestDetectJobType_V1Job_PyTorch(t *testing.T) {
	// v1 CRD: no arena.io/framework label, no ConfigMap
	v1Job := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "kubeflow.org/v1",
			"kind":       "PyTorchJob",
			"metadata": map[string]interface{}{
				"name":      "v1-job",
				"namespace": "default",
			},
		},
	}

	scheme := runtime.NewScheme()
	listKinds := map[schema.GroupVersionResource]string{
		{Group: "", Version: "v1", Resource: "configmaps"}:               "ConfigMapList",
		{Group: "kubeflow.org", Version: "v1", Resource: "pytorchjobs"}:  "PyTorchJobList",
		{Group: "kubeflow.org", Version: "v1", Resource: "tfjobs"}:       "TFJobList",
		{Group: "kubeflow.org", Version: "v2beta1", Resource: "mpijobs"}: "MPIJobList",
	}
	fakeClient := fake.NewSimpleDynamicClientWithCustomListKinds(scheme, listKinds, v1Job)
	k8sClient := client.NewClientForInterface(fakeClient)

	_, err := detectJobType(context.Background(), k8sClient, "default", "v1-job")
	assert.ErrorIs(t, err, errV1Job)
}

// TestDetectJobType_V1Job_TFJob verifies v1 detection for TFJob kind.
func TestDetectJobType_V1Job_TFJob(t *testing.T) {
	v1Job := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "kubeflow.org/v1",
			"kind":       "TFJob",
			"metadata": map[string]interface{}{
				"name":      "v1-tf-job",
				"namespace": "default",
			},
		},
	}

	scheme := runtime.NewScheme()
	listKinds := map[schema.GroupVersionResource]string{
		{Group: "", Version: "v1", Resource: "configmaps"}:               "ConfigMapList",
		{Group: "kubeflow.org", Version: "v1", Resource: "pytorchjobs"}:  "PyTorchJobList",
		{Group: "kubeflow.org", Version: "v1", Resource: "tfjobs"}:       "TFJobList",
		{Group: "kubeflow.org", Version: "v2beta1", Resource: "mpijobs"}: "MPIJobList",
	}
	fakeClient := fake.NewSimpleDynamicClientWithCustomListKinds(scheme, listKinds, v1Job)
	k8sClient := client.NewClientForInterface(fakeClient)

	_, err := detectJobType(context.Background(), k8sClient, "default", "v1-tf-job")
	assert.ErrorIs(t, err, errV1Job)
}

// TestDetectJobType_V1Job_MPIJob verifies v1 detection for MPIJob kind.
func TestDetectJobType_V1Job_MPIJob(t *testing.T) {
	v1Job := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "kubeflow.org/v2beta1",
			"kind":       "MPIJob",
			"metadata": map[string]interface{}{
				"name":      "v1-mpi-job",
				"namespace": "default",
			},
		},
	}

	scheme := runtime.NewScheme()
	listKinds := map[schema.GroupVersionResource]string{
		{Group: "", Version: "v1", Resource: "configmaps"}:               "ConfigMapList",
		{Group: "kubeflow.org", Version: "v1", Resource: "pytorchjobs"}:  "PyTorchJobList",
		{Group: "kubeflow.org", Version: "v1", Resource: "tfjobs"}:       "TFJobList",
		{Group: "kubeflow.org", Version: "v2beta1", Resource: "mpijobs"}: "MPIJobList",
	}
	fakeClient := fake.NewSimpleDynamicClientWithCustomListKinds(scheme, listKinds, v1Job)
	k8sClient := client.NewClientForInterface(fakeClient)
	k8sClient.SetMPIVersion("v2beta1")

	_, err := detectJobType(context.Background(), k8sClient, "default", "v1-mpi-job")
	assert.ErrorIs(t, err, errV1Job)
}

// TestDetectJobType_V2CRDWithoutConfigMap verifies that a CRD with the
// arena.io/framework label but no ConfigMap (abnormal state) is still
// handled by returning the kind.
func TestDetectJobType_V2CRDWithoutConfigMap(t *testing.T) {
	v2CRD := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "kubeflow.org/v1",
			"kind":       "PyTorchJob",
			"metadata": map[string]interface{}{
				"name":      "orphan-crd",
				"namespace": "default",
				"labels": map[string]interface{}{
					"arena.io/framework": "pytorch",
				},
			},
		},
	}

	scheme := runtime.NewScheme()
	listKinds := map[schema.GroupVersionResource]string{
		{Group: "", Version: "v1", Resource: "configmaps"}:               "ConfigMapList",
		{Group: "kubeflow.org", Version: "v1", Resource: "pytorchjobs"}:  "PyTorchJobList",
		{Group: "kubeflow.org", Version: "v1", Resource: "tfjobs"}:       "TFJobList",
		{Group: "kubeflow.org", Version: "v2beta1", Resource: "mpijobs"}: "MPIJobList",
	}
	fakeClient := fake.NewSimpleDynamicClientWithCustomListKinds(scheme, listKinds, v2CRD)
	k8sClient := client.NewClientForInterface(fakeClient)

	kind, err := detectJobType(context.Background(), k8sClient, "default", "orphan-crd")
	assert.NoError(t, err)
	assert.Equal(t, "PyTorchJob", kind)
}

// TestDetectJobType_V2MPIJobWithoutConfigMap verifies that a v2 MPIJob CRD
// with the arena.io/framework label but no ConfigMap (abnormal state) is
// detected correctly with MPIVersion resolved from the cluster CRD.
func TestDetectJobType_V2MPIJobWithoutConfigMap(t *testing.T) {
	v2CRD := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "kubeflow.org/v2beta1",
			"kind":       "MPIJob",
			"metadata": map[string]interface{}{
				"name":      "orphan-mpi-job",
				"namespace": "default",
				"labels": map[string]interface{}{
					"arena.io/framework": "mpi",
				},
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

	scheme := runtime.NewScheme()
	listKinds := map[schema.GroupVersionResource]string{
		{Group: "", Version: "v1", Resource: "configmaps"}:                                    "ConfigMapList",
		{Group: "apiextensions.k8s.io", Version: "v1", Resource: "customresourcedefinitions"}: "CustomResourceDefinitionList",
		{Group: "kubeflow.org", Version: "v2beta1", Resource: "mpijobs"}:                      "MPIJobList",
	}
	fakeClient := fake.NewSimpleDynamicClientWithCustomListKinds(scheme, listKinds, v2CRD, mpiCRD)
	k8sClient := client.NewClientForInterface(fakeClient)

	kind, err := detectJobType(context.Background(), k8sClient, "default", "orphan-mpi-job")
	assert.NoError(t, err)
	assert.Equal(t, "MPIJob", kind)
	assert.Equal(t, "v2beta1", k8sClient.GetMPIVersion())
}

// TestDetectJobType_MPIFamilyFrameworks verifies that DeepSpeed and Horovod
// frameworks map to MPIJob and trigger the same edge-case check when the
// MPIJob CRD is not installed.
func TestDetectJobType_MPIFamilyFrameworks(t *testing.T) {
	tests := []struct {
		name      string
		framework string
	}{
		{"deepspeed", "deepspeed"},
		{"horovod", "horovod"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			yamlContent := "name: " + tt.name + "-job\nframework:\n  name: " + tt.framework + "\n"
			cm := &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "ConfigMap",
					"metadata": map[string]interface{}{
						"name": tt.name + "-job", "namespace": "default",
					},
					"data": map[string]interface{}{
						"arena-v2.yaml": yamlContent,
					},
				},
			}

			scheme := runtime.NewScheme()
			listKinds := map[schema.GroupVersionResource]string{
				{Group: "", Version: "v1", Resource: "configmaps"}: "ConfigMapList",
			}
			fakeClient := fake.NewSimpleDynamicClientWithCustomListKinds(scheme, listKinds, cm)
			k8sClient := client.NewClientForInterface(fakeClient)

			_, err := detectJobType(context.Background(), k8sClient, "default", tt.name+"-job")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "MPIJob CRD is not installed")
		})
	}
}

// TestDetectJobType_ConfigMapWithoutArenaYAML verifies that a ConfigMap
// without the arena-v2.yaml key falls through to Phase 2.
func TestDetectJobType_ConfigMapWithoutArenaYAML(t *testing.T) {
	cm := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name": "plain-cm", "namespace": "default",
			},
			"data": map[string]interface{}{
				"other-key": "some-value",
			},
		},
	}

	scheme := runtime.NewScheme()
	listKinds := map[schema.GroupVersionResource]string{
		{Group: "", Version: "v1", Resource: "configmaps"}: "ConfigMapList",
	}
	fakeClient := fake.NewSimpleDynamicClientWithCustomListKinds(scheme, listKinds, cm)
	k8sClient := client.NewClientForInterface(fakeClient)

	_, err := detectJobType(context.Background(), k8sClient, "default", "plain-cm")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// TestDetectJobType_UnknownFrameworkFallsThrough verifies that a ConfigMap
// with an unknown framework name falls through to Phase 2 (frameworkToKind
// returns "" for unrecognized frameworks).
func TestDetectJobType_UnknownFrameworkFallsThrough(t *testing.T) {
	yamlContent := "name: unknown-fw-job\nframework:\n  name: unknown-framework\n"
	cm := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name": "unknown-fw-job", "namespace": "default",
			},
			"data": map[string]interface{}{
				"arena-v2.yaml": yamlContent,
			},
		},
	}

	scheme := runtime.NewScheme()
	listKinds := map[schema.GroupVersionResource]string{
		{Group: "", Version: "v1", Resource: "configmaps"}: "ConfigMapList",
	}
	fakeClient := fake.NewSimpleDynamicClientWithCustomListKinds(scheme, listKinds, cm)
	k8sClient := client.NewClientForInterface(fakeClient)

	_, err := detectJobType(context.Background(), k8sClient, "default", "unknown-fw-job")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// TestDetectJobType_EmptyFrameworkNameFallsThrough verifies that a ConfigMap
// with an empty framework name falls through to Phase 2 (the condition
// t.Framework.Name != "" prevents entering the frameworkToKind branch).
func TestDetectJobType_EmptyFrameworkNameFallsThrough(t *testing.T) {
	yamlContent := "name: empty-fw-job\nframework:\n  name: \"\"\n"
	cm := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name": "empty-fw-job", "namespace": "default",
			},
			"data": map[string]interface{}{
				"arena-v2.yaml": yamlContent,
			},
		},
	}

	scheme := runtime.NewScheme()
	listKinds := map[schema.GroupVersionResource]string{
		{Group: "", Version: "v1", Resource: "configmaps"}: "ConfigMapList",
	}
	fakeClient := fake.NewSimpleDynamicClientWithCustomListKinds(scheme, listKinds, cm)
	k8sClient := client.NewClientForInterface(fakeClient)

	_, err := detectJobType(context.Background(), k8sClient, "default", "empty-fw-job")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestDuplicateJobCreationFails(t *testing.T) {
	ctx := context.Background()
	k8sClient := newFakeK8sClient(t)

	crd := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "kubeflow.org/v1",
			"kind":       "PyTorchJob",
			"metadata": map[string]interface{}{
				"name":      "dup-job",
				"namespace": "default",
			},
			"spec": map[string]interface{}{},
		},
	}

	require.NoError(t, k8sClient.Create(ctx, crd))

	err := k8sClient.Create(ctx, crd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}
