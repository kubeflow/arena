package cli

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic/fake"
	ktesting "k8s.io/client-go/testing"

	"github.com/kubeflow/arena/pkg/client"
	"github.com/kubeflow/arena/pkg/provider"
	"github.com/kubeflow/arena/pkg/task"
)

// rbacListKinds returns the GVR-to-ListKind mapping needed to register
// all resource types that preCreateRBAC and finalizeJobResources may create:
// ConfigMap, ServiceAccount, Role, RoleBinding, and the MPIJob CRD itself.
func rbacListKinds(mpiVersion string) map[schema.GroupVersionResource]string {
	return map[schema.GroupVersionResource]string{
		{Group: "", Version: "v1", Resource: "configmaps"}:                            "ConfigMapList",
		{Group: "", Version: "v1", Resource: "serviceaccounts"}:                       "ServiceAccountList",
		{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "roles"}:        "RoleList",
		{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "rolebindings"}: "RoleBindingList",
		{Group: "kubeflow.org", Version: mpiVersion, Resource: "mpijobs"}:             "MPIJobList",
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

// TestPreCreateAndFinalize_CreatesRBACResources verifies the full new flow:
// preCreateRBAC creates RBAC without ownerRef, then finalizeJobResources
// patches ownerRef and creates ConfigMap.
func TestPreCreateAndFinalize_CreatesRBACResources(t *testing.T) {
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

	// Step 1: Pre-create RBAC (no ownerRef)
	rbacResources, err := preCreateRBAC(context.Background(), tk, k8sClient, p)
	require.NoError(t, err)
	assert.Len(t, rbacResources, 3)

	// Step 2: Finalize (creates ConfigMap + patches ownerRef)
	err = finalizeJobResources(context.Background(), mpiJob, tk, k8sClient, p, rbacResources)
	require.NoError(t, err)

	// Verify ConfigMap was created
	cm, err := k8sClient.Get(context.Background(), "ConfigMap", namespace, jobName)
	require.NoError(t, err)
	assert.Equal(t, "ConfigMap", cm.GetKind())

	// Verify ServiceAccount with ownerRef
	sa, err := k8sClient.Get(context.Background(), "ServiceAccount", namespace, jobName+"-launcher")
	require.NoError(t, err)
	assert.Equal(t, "ServiceAccount", sa.GetKind())
	assert.Equal(t, namespace, sa.GetNamespace())

	// Role and RoleBinding via fake client directly
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

	rb, err := fakeClient.Resource(rbGVR).Namespace(namespace).Get(
		context.Background(), jobName+"-launcher", metav1.GetOptions{})
	require.NoError(t, err)
	assert.Equal(t, "RoleBinding", rb.GetKind())

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

	// Verify RoleBinding subject
	subjects, found, err := unstructured.NestedSlice(rb.Object, "subjects")
	require.NoError(t, err)
	require.True(t, found)
	require.Len(t, subjects, 1)
	subject := subjects[0].(map[string]interface{})
	assert.Equal(t, "ServiceAccount", subject["kind"])
	assert.Equal(t, jobName+"-launcher", subject["name"])
	assert.Equal(t, namespace, subject["namespace"])
}

// TestPreCreateRBAC_NamespacePropagation verifies that namespace propagation
// works correctly in the new flow.
func TestPreCreateRBAC_NamespacePropagation(t *testing.T) {
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

			scheme := runtime.NewScheme()
			fakeClient := fake.NewSimpleDynamicClientWithCustomListKinds(
				scheme, rbacListKinds(mpiVersion))
			k8sClient := client.NewClientForInterface(fakeClient)
			k8sClient.SetMPIVersion(mpiVersion)

			tk := &task.Task{
				Name:      jobName,
				Namespace: tt.namespace,
				Framework: task.Framework{Name: "mpi"},
				Worker:    &task.Worker{Replicas: 1},
			}
			p := &provider.MPIProvider{APIVersion: provider.MPIAPIVersionV2beta1}

			_, err := preCreateRBAC(context.Background(), tk, k8sClient, p)
			require.NoError(t, err)

			// Verify RoleBinding subject namespace matches
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

			// Verify SA in the correct namespace
			sa, err := k8sClient.Get(context.Background(), "ServiceAccount", tt.namespace, jobName+"-launcher")
			require.NoError(t, err)
			assert.Equal(t, tt.namespace, sa.GetNamespace())
		})
	}
}

// TestPreCreateRBAC_RollbackOnFailure verifies that when an RBAC resource
// creation fails, previously created RBAC resources are rolled back.
func TestPreCreateRBAC_RollbackOnFailure(t *testing.T) {
	const (
		jobName    = "rollback-pre-rbac"
		namespace  = "default"
		mpiVersion = "v2beta1"
	)

	scheme := runtime.NewScheme()
	fakeClient := fake.NewSimpleDynamicClientWithCustomListKinds(
		scheme, rbacListKinds(mpiVersion))

	// Make RoleBinding CreateOrUpdate fail by failing on the "create" action
	// for rolebindings (the 3rd resource: SA succeeds, Role succeeds, RB fails).
	fakeClient.PrependReactor(
		"create", "rolebindings",
		func(ktesting.Action) (bool, runtime.Object, error) {
			return true, nil, fmt.Errorf("simulated RoleBinding creation failure")
		},
	)

	k8sClient := client.NewClientForInterface(fakeClient)
	k8sClient.SetMPIVersion(mpiVersion)

	tk := &task.Task{
		Name:      jobName,
		Namespace: namespace,
		Framework: task.Framework{Name: "mpi"},
		Worker:    &task.Worker{Replicas: 2},
	}
	p := &provider.MPIProvider{APIVersion: provider.MPIAPIVersionV2beta1}

	_, err := preCreateRBAC(context.Background(), tk, k8sClient, p)
	require.Error(t, err, "preCreateRBAC should fail when RoleBinding creation fails")

	// ServiceAccount was created before the failure, so it must have been
	// rolled back (deleted).
	_, err = k8sClient.Get(context.Background(), "ServiceAccount", namespace, jobName+"-launcher")
	assert.Error(t, err,
		"ServiceAccount should have been rolled back after RoleBinding creation failure")

	// Verify Role was also rolled back
	roleGVR := schema.GroupVersionResource{
		Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "roles",
	}
	_, err = fakeClient.Resource(roleGVR).Namespace(namespace).Get(
		context.Background(), jobName+"-launcher", metav1.GetOptions{})
	assert.Error(t, err, "Role should have been rolled back after RoleBinding creation failure")
}

// TestPreCreateRBAC_CreatesResources verifies that preCreateRBAC creates
// ServiceAccount, Role, and RoleBinding without ownerReferences (ownerRef
// is patched later by finalizeJobResources after CRD creation).
func TestPreCreateRBAC_CreatesResources(t *testing.T) {
	const (
		jobName    = "pre-rbac-job"
		namespace  = "default"
		mpiVersion = "v2beta1"
	)

	scheme := runtime.NewScheme()
	fakeClient := fake.NewSimpleDynamicClientWithCustomListKinds(scheme, rbacListKinds(mpiVersion))
	k8sClient := client.NewClientForInterface(fakeClient)
	k8sClient.SetMPIVersion(mpiVersion)

	tk := &task.Task{
		Name:      jobName,
		Namespace: namespace,
		Framework: task.Framework{Name: "mpi"},
		Worker:    &task.Worker{Replicas: 2},
	}
	p := &provider.MPIProvider{APIVersion: provider.MPIAPIVersionV2beta1}

	resources, err := preCreateRBAC(context.Background(), tk, k8sClient, p)
	require.NoError(t, err)
	assert.Len(t, resources, 3)

	// Verify ServiceAccount exists and has no ownerReferences
	sa, err := k8sClient.Get(context.Background(), "ServiceAccount", namespace, jobName+"-launcher")
	require.NoError(t, err)
	assert.Empty(t, sa.GetOwnerReferences(), "preCreateRBAC should not set ownerReferences")

	// Verify Role exists and has no ownerReferences
	roleGVR := schema.GroupVersionResource{
		Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "roles",
	}
	role, err := fakeClient.Resource(roleGVR).Namespace(namespace).Get(
		context.Background(), jobName+"-launcher", metav1.GetOptions{})
	require.NoError(t, err)
	assert.Empty(t, role.GetOwnerReferences(), "Role should have no ownerReferences")

	// Verify RoleBinding exists and has no ownerReferences
	rbGVR := schema.GroupVersionResource{
		Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "rolebindings",
	}
	rb, err := fakeClient.Resource(rbGVR).Namespace(namespace).Get(
		context.Background(), jobName+"-launcher", metav1.GetOptions{})
	require.NoError(t, err)
	assert.Empty(t, rb.GetOwnerReferences(), "RoleBinding should have no ownerReferences")
}

// TestPreCreateRBAC_NilRBACProvider verifies that a provider returning nil
// RBAC resources (e.g. PyTorchProvider) returns an empty resource list.
func TestPreCreateRBAC_NilRBACProvider(t *testing.T) {
	const (
		jobName   = "pytorch-pre-rbac"
		namespace = "default"
	)

	scheme := runtime.NewScheme()
	listKinds := map[schema.GroupVersionResource]string{
		{Group: "", Version: "v1", Resource: "configmaps"}:              "ConfigMapList",
		{Group: "kubeflow.org", Version: "v1", Resource: "pytorchjobs"}: "PyTorchJobList",
	}
	fakeClient := fake.NewSimpleDynamicClientWithCustomListKinds(scheme, listKinds)
	k8sClient := client.NewClientForInterface(fakeClient)

	tk := &task.Task{
		Name:      jobName,
		Namespace: namespace,
		Framework: task.Framework{Name: "pytorch"},
		Worker:    &task.Worker{Replicas: 2},
	}
	p := &provider.PyTorchProvider{}

	resources, err := preCreateRBAC(context.Background(), tk, k8sClient, p)
	require.NoError(t, err)
	assert.Empty(t, resources)
}

// TestPreCreateRBAC_UpdatesStaleResources verifies that preCreateRBAC
// updates existing RBAC resources with current content (create-or-update).
func TestPreCreateRBAC_UpdatesStaleResources(t *testing.T) {
	const (
		jobName    = "stale-rbac-job"
		namespace  = "default"
		mpiVersion = "v2beta1"
	)

	scheme := runtime.NewScheme()
	fakeClient := fake.NewSimpleDynamicClientWithCustomListKinds(scheme, rbacListKinds(mpiVersion))
	k8sClient := client.NewClientForInterface(fakeClient)
	k8sClient.SetMPIVersion(mpiVersion)

	// First call: create with 2 workers
	tk2 := &task.Task{
		Name:      jobName,
		Namespace: namespace,
		Framework: task.Framework{Name: "mpi"},
		Worker:    &task.Worker{Replicas: 2},
	}
	p := &provider.MPIProvider{APIVersion: provider.MPIAPIVersionV2beta1}

	_, err := preCreateRBAC(context.Background(), tk2, k8sClient, p)
	require.NoError(t, err)

	// Verify Role has 2 worker pod names in the get rule
	roleGVR := schema.GroupVersionResource{
		Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "roles",
	}
	role, err := fakeClient.Resource(roleGVR).Namespace(namespace).Get(
		context.Background(), jobName+"-launcher", metav1.GetOptions{})
	require.NoError(t, err)
	getRule, _, err := unstructured.NestedSlice(role.Object, "rules")
	require.NoError(t, err)
	// With 2 workers: 3 rules (list/watch pods, get pods by name, create pods/exec by name)
	assert.Len(t, getRule, 3)

	// Second call: update with 4 workers
	tk4 := &task.Task{
		Name:      jobName,
		Namespace: namespace,
		Framework: task.Framework{Name: "mpi"},
		Worker:    &task.Worker{Replicas: 4},
	}

	_, err = preCreateRBAC(context.Background(), tk4, k8sClient, p)
	require.NoError(t, err)

	// Verify Role now has 4 worker pod names
	role, err = fakeClient.Resource(roleGVR).Namespace(namespace).Get(
		context.Background(), jobName+"-launcher", metav1.GetOptions{})
	require.NoError(t, err)
	rules, _, err := unstructured.NestedSlice(role.Object, "rules")
	require.NoError(t, err)
	// The get-by-name rule should now have 4 resourceNames
	var getByNameRule map[string]interface{}
	for _, r := range rules {
		rule := r.(map[string]interface{})
		verbs, _ := rule["verbs"].([]interface{})
		if len(verbs) == 1 && verbs[0] == "get" {
			getByNameRule = rule
			break
		}
	}
	require.NotNil(t, getByNameRule, "should find get-by-name rule")
	resourceNames, _ := getByNameRule["resourceNames"].([]interface{})
	assert.Len(t, resourceNames, 4, "Role should have 4 worker pod names after update")
}

// TestFinalizeJobResources_PatchesOwnerRef verifies that finalizeJobResources
// creates the ConfigMap with ownerRef and patches ownerRef onto RBAC resources
// that were pre-created by preCreateRBAC.
func TestFinalizeJobResources_PatchesOwnerRef(t *testing.T) {
	const (
		jobName    = "finalize-job"
		namespace  = "default"
		mpiVersion = "v2beta1"
		uid        = "finalize-uid-123"
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

	// Pre-create RBAC (simulates preCreateRBAC step)
	rbacResources, err := preCreateRBAC(context.Background(), tk, k8sClient, p)
	require.NoError(t, err)

	// Verify RBAC has no ownerRef before finalize
	sa, err := k8sClient.Get(context.Background(), "ServiceAccount", namespace, jobName+"-launcher")
	require.NoError(t, err)
	assert.Empty(t, sa.GetOwnerReferences(), "SA should have no ownerRef before finalize")

	// Finalize
	err = finalizeJobResources(context.Background(), mpiJob, tk, k8sClient, p, rbacResources)
	require.NoError(t, err)

	// Verify ConfigMap has ownerRef at creation
	cm, err := k8sClient.Get(context.Background(), "ConfigMap", namespace, jobName)
	require.NoError(t, err)
	cmRefs := cm.GetOwnerReferences()
	require.Len(t, cmRefs, 1)
	assert.Equal(t, "MPIJob", cmRefs[0].Kind)
	assert.Equal(t, uid, string(cmRefs[0].UID))

	// Verify RBAC resources now have ownerRef patched
	sa, err = k8sClient.Get(context.Background(), "ServiceAccount", namespace, jobName+"-launcher")
	require.NoError(t, err)
	saRefs := sa.GetOwnerReferences()
	require.Len(t, saRefs, 1)
	assert.Equal(t, "MPIJob", saRefs[0].Kind)
	assert.Equal(t, uid, string(saRefs[0].UID))

	// Verify Role has ownerRef
	roleGVR := schema.GroupVersionResource{
		Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "roles",
	}
	role, err := fakeClient.Resource(roleGVR).Namespace(namespace).Get(
		context.Background(), jobName+"-launcher", metav1.GetOptions{})
	require.NoError(t, err)
	roleRefs := role.GetOwnerReferences()
	require.Len(t, roleRefs, 1)
	assert.Equal(t, uid, string(roleRefs[0].UID))
}

// TestFinalizeJobResources_NilRBACProvider verifies that finalizeJobResources
// works correctly when the provider returns no RBAC resources (e.g. PyTorch).
// It should still create the ConfigMap with ownerRef.
func TestFinalizeJobResources_NilRBACProvider(t *testing.T) {
	const (
		jobName   = "pytorch-finalize"
		namespace = "default"
		uid       = "pt-finalize-uid"
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

	// Empty RBAC resources (PyTorch provider returns nil)
	err := finalizeJobResources(context.Background(), crd, tk, k8sClient, p, nil)
	require.NoError(t, err)

	// ConfigMap should be created with ownerRef
	cm, err := k8sClient.Get(context.Background(), "ConfigMap", namespace, jobName)
	require.NoError(t, err)
	refs := cm.GetOwnerReferences()
	require.Len(t, refs, 1)
	assert.Equal(t, "PyTorchJob", refs[0].Kind)
	assert.Equal(t, uid, string(refs[0].UID))
}

// TestFinalizeJobResources_TensorBoard verifies that finalizeJobResources
// creates TensorBoard Deployment and Service with ownerReferences pointing
// to the CRD when TensorBoard is enabled.
func TestFinalizeJobResources_TensorBoard(t *testing.T) {
	const (
		jobName    = "tb-finalize-job"
		namespace  = "default"
		mpiVersion = "v2beta1"
		uid        = "tb-finalize-uid-123"
	)

	mpiJob := seedMPIJobCRD(jobName, namespace, mpiVersion, uid)

	// listKinds includes deployments and services for TensorBoard resources
	listKinds := map[schema.GroupVersionResource]string{
		{Group: "", Version: "v1", Resource: "configmaps"}:                            "ConfigMapList",
		{Group: "", Version: "v1", Resource: "services"}:                              "ServiceList",
		{Group: "", Version: "v1", Resource: "serviceaccounts"}:                       "ServiceAccountList",
		{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "roles"}:        "RoleList",
		{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "rolebindings"}: "RoleBindingList",
		{Group: "apps", Version: "v1", Resource: "deployments"}:                       "DeploymentList",
		{Group: "kubeflow.org", Version: mpiVersion, Resource: "mpijobs"}:             "MPIJobList",
	}

	scheme := runtime.NewScheme()
	fakeClient := fake.NewSimpleDynamicClientWithCustomListKinds(scheme, listKinds, mpiJob)
	k8sClient := client.NewClientForInterface(fakeClient)
	k8sClient.SetMPIVersion(mpiVersion)

	tk := &task.Task{
		Name:      jobName,
		Namespace: namespace,
		Framework: task.Framework{Name: "mpi"},
		Worker:    &task.Worker{Replicas: 2},
		Logging: task.Logging{
			TensorBoard: &task.TensorBoardConfig{
				Enabled: true,
				Image:   "tensorflow/tensorflow:2.21.0",
				LogDir:  "/logs/tb",
			},
		},
	}
	p := &provider.MPIProvider{APIVersion: provider.MPIAPIVersionV2beta1}

	// Step 1: Pre-create RBAC (no ownerRef)
	rbacResources, err := preCreateRBAC(context.Background(), tk, k8sClient, p)
	require.NoError(t, err)

	// Step 2: Finalize (creates ConfigMap, patches ownerRef, creates TensorBoard)
	err = finalizeJobResources(context.Background(), mpiJob, tk, k8sClient, p, rbacResources)
	require.NoError(t, err)

	// Verify TensorBoard Deployment was created with ownerRef
	deployGVR := schema.GroupVersionResource{
		Group: "apps", Version: "v1", Resource: "deployments",
	}
	tbName := jobName + "-tensorboard"
	deploy, err := fakeClient.Resource(deployGVR).Namespace(namespace).Get(
		context.Background(), tbName, metav1.GetOptions{})
	require.NoError(t, err)
	assert.Equal(t, "Deployment", deploy.GetKind())
	assert.Equal(t, tbName, deploy.GetName())

	deployRefs := deploy.GetOwnerReferences()
	require.Len(t, deployRefs, 1)
	assert.Equal(t, "MPIJob", deployRefs[0].Kind)
	assert.Equal(t, jobName, deployRefs[0].Name)
	assert.Equal(t, uid, string(deployRefs[0].UID))
	assert.True(t, *deployRefs[0].BlockOwnerDeletion)
	assert.True(t, *deployRefs[0].Controller)

	// Verify TensorBoard Service was created with ownerRef
	svcGVR := schema.GroupVersionResource{
		Group: "", Version: "v1", Resource: "services",
	}
	svc, err := fakeClient.Resource(svcGVR).Namespace(namespace).Get(
		context.Background(), tbName, metav1.GetOptions{})
	require.NoError(t, err)
	assert.Equal(t, "Service", svc.GetKind())
	assert.Equal(t, tbName, svc.GetName())

	svcRefs := svc.GetOwnerReferences()
	require.Len(t, svcRefs, 1)
	assert.Equal(t, "MPIJob", svcRefs[0].Kind)
	assert.Equal(t, jobName, svcRefs[0].Name)
	assert.Equal(t, uid, string(svcRefs[0].UID))
	assert.True(t, *svcRefs[0].BlockOwnerDeletion)
	assert.True(t, *svcRefs[0].Controller)
}

// TestFinalizeJobResources_RollbackOnPatchFailure verifies that when
// patching ownerReferences onto an RBAC resource fails, finalizeJobResources
// rolls back the ConfigMap it created and returns an error.
func TestFinalizeJobResources_RollbackOnPatchFailure(t *testing.T) {
	const (
		jobName    = "rollback-patch-job"
		namespace  = "default"
		mpiVersion = "v2beta1"
		uid        = "rollback-patch-uid-123"
	)

	mpiJob := seedMPIJobCRD(jobName, namespace, mpiVersion, uid)

	scheme := runtime.NewScheme()
	fakeClient := fake.NewSimpleDynamicClientWithCustomListKinds(scheme, rbacListKinds(mpiVersion), mpiJob)
	k8sClient := client.NewClientForInterface(fakeClient)
	k8sClient.SetMPIVersion(mpiVersion)

	// Make the patch action on rolebindings fail
	fakeClient.PrependReactor(
		"patch", "rolebindings",
		func(ktesting.Action) (bool, runtime.Object, error) {
			return true, nil, fmt.Errorf("simulated patch failure on rolebindings")
		},
	)

	tk := &task.Task{
		Name:      jobName,
		Namespace: namespace,
		Framework: task.Framework{Name: "mpi"},
		Worker:    &task.Worker{Replicas: 2},
	}
	p := &provider.MPIProvider{APIVersion: provider.MPIAPIVersionV2beta1}

	// Step 1: Pre-create RBAC (no ownerRef)
	rbacResources, err := preCreateRBAC(context.Background(), tk, k8sClient, p)
	require.NoError(t, err)

	// Step 2: Finalize should fail because patching rolebindings fails
	err = finalizeJobResources(context.Background(), mpiJob, tk, k8sClient, p, rbacResources)
	require.Error(t, err, "finalizeJobResources should return an error when patch fails")

	// Verify the ConfigMap was rolled back (deleted)
	_, err = k8sClient.Get(context.Background(), "ConfigMap", namespace, jobName)
	assert.Error(t, err, "ConfigMap should have been rolled back after patch failure")
}
