package cli

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic/fake"

	"github.com/kubeflow/arena/pkg/client"
	"github.com/kubeflow/arena/pkg/provider"
	"github.com/kubeflow/arena/pkg/task"
)

// rbacListKinds returns the GVR-to-ListKind mapping needed to register
// all resource types that createJobResources may create: ConfigMap,
// ServiceAccount, Role, RoleBinding, and the MPIJob CRD itself.
func rbacListKinds(mpiVersion string) map[schema.GroupVersionResource]string {
	return map[schema.GroupVersionResource]string{
		{Group: "", Version: "v1", Resource: "configmaps"}:                          "ConfigMapList",
		{Group: "", Version: "v1", Resource: "serviceaccounts"}:                     "ServiceAccountList",
		{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "roles"}:       "RoleList",
		{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "rolebindings"}: "RoleBindingList",
		{Group: "kubeflow.org", Version: mpiVersion, Resource: "mpijobs"}:            "MPIJobList",
	}
}

// seedMPIJobCRD creates an MPIJob unstructured object with a UID set,
// simulating a CRD that has already been submitted to the cluster.
func seedMPIJobCRD(name, namespace, mpiVersion, uid string) *unstructured.Unstructured {
	crd := &unstructured.Unstructured{}
	crd.SetAPIVersion("kubeflow.org/" + mpiVersion)
	crd.SetKind("MPIJob")
	crd.SetName(name)
	crd.SetNamespace(namespace)
	crd.SetUID(types.UID(uid))
	return crd
}

// TestCreateJobResources_CreatesRBACResources verifies that createJobResources
// creates the ConfigMap anchor and all three RBAC resources (ServiceAccount,
// Role, RoleBinding) with correct names, namespaces, and ownerReferences
// pointing back to the MPIJob CRD.
func TestCreateJobResources_CreatesRBACResources(t *testing.T) {
	const (
		jobName    = "test-mpi-job"
		namespace  = "default"
		mpiVersion = "v2beta1"
		uid        = "job-uid-123"
	)

	mpiJob := seedMPIJobCRD(jobName, namespace, mpiVersion, uid)

	scheme := runtime.NewScheme()
	fakeClient := fake.NewSimpleDynamicClientWithCustomListKinds(scheme, rbacListKinds(mpiVersion), mpiJob)
	k8sClient := client.NewClientForInterface(fakeClient)
	k8sClient.SetMPIVersion(mpiVersion)

	tk := &task.Task{
		Name:      jobName,
		Namespace: namespace,
		Framework: task.Framework{Name: "mpi"},
		Worker:    &task.Worker{Replicas: 2},
	}
	p := &provider.MPIProvider{APIVersion: provider.MPIAPIVersionV2beta1}

	err := createJobResources(context.Background(), mpiJob, tk, k8sClient, p)
	require.NoError(t, err)

	// Verify ConfigMap was created
	cm, err := k8sClient.Get(context.Background(), "ConfigMap", namespace, jobName)
	require.NoError(t, err)
	assert.Equal(t, "ConfigMap", cm.GetKind())

	// Verify ServiceAccount was created with correct name and namespace
	sa, err := k8sClient.Get(context.Background(), "ServiceAccount", namespace, jobName+"-launcher")
	require.NoError(t, err)
	assert.Equal(t, "ServiceAccount", sa.GetKind())
	assert.Equal(t, namespace, sa.GetNamespace())

	// Role and RoleBinding use rbac.authorization.k8s.io group, which
	// kindToGVR doesn't map — access them via the fake client directly.
	roleGVR := schema.GroupVersionResource{
		Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "roles",
	}
	rbGVR := schema.GroupVersionResource{
		Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "rolebindings",
	}

	role, err := fakeClient.Resource(roleGVR).Namespace(namespace).Get(
		context.Background(), jobName+"-launcher", metav1.GetOptions{})
	require.NoError(t, err)
	assert.Equal(t, "Role", role.GetKind())
	assert.Equal(t, namespace, role.GetNamespace())

	rb, err := fakeClient.Resource(rbGVR).Namespace(namespace).Get(
		context.Background(), jobName+"-launcher", metav1.GetOptions{})
	require.NoError(t, err)
	assert.Equal(t, "RoleBinding", rb.GetKind())
	assert.Equal(t, namespace, rb.GetNamespace())

	// Verify ownerReferences on all RBAC resources point to the MPIJob
	for _, obj := range []*unstructured.Unstructured{sa, role, rb} {
		refs := obj.GetOwnerReferences()
		require.Len(t, refs, 1)
		assert.Equal(t, "MPIJob", refs[0].Kind)
		assert.Equal(t, jobName, refs[0].Name)
		assert.Equal(t, uid, string(refs[0].UID))
		assert.True(t, *refs[0].BlockOwnerDeletion)
		assert.True(t, *refs[0].Controller)
	}

	// Verify RoleBinding subject has correct namespace and SA name
	subjects, found, err := unstructured.NestedSlice(rb.Object, "subjects")
	require.NoError(t, err)
	require.True(t, found)
	require.Len(t, subjects, 1)
	subject := subjects[0].(map[string]interface{})
	assert.Equal(t, "ServiceAccount", subject["kind"])
	assert.Equal(t, jobName+"-launcher", subject["name"])
	assert.Equal(t, namespace, subject["namespace"])
}

// TestCreateJobResources_NamespacePropagation is a regression test for the
// critical bug where t.Namespace was not set after resolving the namespace
// in submit.go/run.go. When t.Namespace is correctly propagated (as the fix
// ensures), the RoleBinding subject namespace must match the CRD namespace.
func TestCreateJobResources_NamespacePropagation(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
	}{
		{"default namespace", "default"},
		{"custom namespace", "ml-pipeline"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jobName := "ns-test-job"
			mpiVersion := "v2beta1"

			mpiJob := seedMPIJobCRD(jobName, tt.namespace, mpiVersion, "uid-ns-1")

			scheme := runtime.NewScheme()
			fakeClient := fake.NewSimpleDynamicClientWithCustomListKinds(
				scheme, rbacListKinds(mpiVersion), mpiJob)
			k8sClient := client.NewClientForInterface(fakeClient)
			k8sClient.SetMPIVersion(mpiVersion)

			// Simulate the fix: t.Namespace is set to the resolved namespace
			// before calling createJobResources (as submit.go/run.go do).
			tk := &task.Task{
				Name:      jobName,
				Namespace: tt.namespace,
				Framework: task.Framework{Name: "mpi"},
				Worker:    &task.Worker{Replicas: 1},
			}
			p := &provider.MPIProvider{APIVersion: provider.MPIAPIVersionV2beta1}

			err := createJobResources(context.Background(), mpiJob, tk, k8sClient, p)
			require.NoError(t, err)

			// Verify RoleBinding subject namespace matches CRD namespace
			rbGVR := schema.GroupVersionResource{
				Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "rolebindings",
			}
			rb, err := fakeClient.Resource(rbGVR).Namespace(tt.namespace).Get(
				context.Background(), jobName+"-launcher", metav1.GetOptions{})
			require.NoError(t, err)

			subjects, found, err := unstructured.NestedSlice(rb.Object, "subjects")
			require.NoError(t, err)
			require.True(t, found)
			subject := subjects[0].(map[string]interface{})
			assert.Equal(t, tt.namespace, subject["namespace"],
				"RoleBinding subject namespace must match CRD namespace")

			// Verify SA also in the correct namespace
			sa, err := k8sClient.Get(context.Background(), "ServiceAccount", tt.namespace, jobName+"-launcher")
			require.NoError(t, err)
			assert.Equal(t, tt.namespace, sa.GetNamespace())
		})
	}
}

// TestCreateJobResources_NilRBACProvider verifies that a provider returning
// nil RBAC resources (e.g. PyTorchProvider) does not cause errors and the
// ConfigMap is still created.
func TestCreateJobResources_NilRBACProvider(t *testing.T) {
	const (
		jobName    = "pytorch-job"
		namespace  = "default"
		uid        = "pt-uid-456"
	)

	crd := &unstructured.Unstructured{}
	crd.SetAPIVersion("kubeflow.org/v1")
	crd.SetKind("PyTorchJob")
	crd.SetName(jobName)
	crd.SetNamespace(namespace)
	crd.SetUID(types.UID(uid))

	listKinds := map[schema.GroupVersionResource]string{
		{Group: "", Version: "v1", Resource: "configmaps"}:              "ConfigMapList",
		{Group: "kubeflow.org", Version: "v1", Resource: "pytorchjobs"}: "PyTorchJobList",
	}
	scheme := runtime.NewScheme()
	fakeClient := fake.NewSimpleDynamicClientWithCustomListKinds(scheme, listKinds, crd)
	k8sClient := client.NewClientForInterface(fakeClient)

	tk := &task.Task{
		Name:      jobName,
		Namespace: namespace,
		Framework: task.Framework{Name: "pytorch"},
		Worker:    &task.Worker{Replicas: 2},
	}
	p := &provider.PyTorchProvider{}

	err := createJobResources(context.Background(), crd, tk, k8sClient, p)
	require.NoError(t, err)

	// ConfigMap should still be created
	cm, err := k8sClient.Get(context.Background(), "ConfigMap", namespace, jobName)
	require.NoError(t, err)
	assert.Equal(t, "ConfigMap", cm.GetKind())

	// No ServiceAccount should exist (PyTorch provider returns nil RBAC)
	_, err = k8sClient.Get(context.Background(), "ServiceAccount", namespace, jobName+"-launcher")
	assert.Error(t, err, "no ServiceAccount should be created for non-MPI providers")
}
