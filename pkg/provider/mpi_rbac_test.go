package provider

import (
	"testing"

	"github.com/kubeflow/arena/pkg/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestBuildLauncherServiceAccount(t *testing.T) {
	ownerRef := metav1.OwnerReference{
		APIVersion: "kubeflow.org/v1",
		Kind:       "MPIJob",
		Name:       "test-job",
		UID:        "abc-123",
	}

	sa := buildLauncherServiceAccount("test-job-launcher", "test-job", "default", ownerRef)

	assert.Equal(t, "ServiceAccount", sa.GetKind())
	assert.Equal(t, "v1", sa.GetAPIVersion())
	assert.Equal(t, "test-job-launcher", sa.GetName())
	assert.Equal(t, "default", sa.GetNamespace())

	labels := sa.GetLabels()
	assert.Equal(t, "test-job", labels["app"])

	refs := sa.GetOwnerReferences()
	require.Len(t, refs, 1)
	assert.Equal(t, "test-job", refs[0].Name)
	assert.Equal(t, "MPIJob", refs[0].Kind)
	assert.Equal(t, "abc-123", string(refs[0].UID))
}

func TestBuildLauncherRole(t *testing.T) {
	ownerRef := metav1.OwnerReference{
		APIVersion: "kubeflow.org/v1",
		Kind:       "MPIJob",
		Name:       "test-job",
		UID:        "abc-123",
	}

	role := buildLauncherRole("test-job", "default", 3, ownerRef)

	assert.Equal(t, "Role", role.GetKind())
	assert.Equal(t, "rbac.authorization.k8s.io/v1", role.GetAPIVersion())
	assert.Equal(t, "test-job-launcher", role.GetName())
	assert.Equal(t, "default", role.GetNamespace())

	labels := role.GetLabels()
	assert.Equal(t, "test-job", labels["app"])

	refs := role.GetOwnerReferences()
	require.Len(t, refs, 1)
	assert.Equal(t, "test-job", refs[0].Name)

	rules, found, err := unstructured.NestedSlice(role.Object, "rules")
	require.NoError(t, err)
	assert.True(t, found)
	require.Len(t, rules, 2)

	// Rule 1: get/list/watch on pods
	rule1 := rules[0].(map[string]interface{})
	verbs1, _ := rule1["verbs"].([]interface{})
	assert.Contains(t, verbs1, "get")
	assert.Contains(t, verbs1, "list")
	assert.Contains(t, verbs1, "watch")
	apiGroups1, _ := rule1["apiGroups"].([]interface{})
	assert.Contains(t, apiGroups1, "")
	resources1, _ := rule1["resources"].([]interface{})
	assert.Contains(t, resources1, "pods")

	// Rule 2: create on pods/exec with resourceNames
	rule2 := rules[1].(map[string]interface{})
	verbs2, _ := rule2["verbs"].([]interface{})
	assert.Contains(t, verbs2, "create")
	resources2, _ := rule2["resources"].([]interface{})
	assert.Contains(t, resources2, "pods/exec")
	resourceNames, _ := rule2["resourceNames"].([]interface{})
	require.Len(t, resourceNames, 3)
	assert.Equal(t, "test-job-worker-0", resourceNames[0])
	assert.Equal(t, "test-job-worker-1", resourceNames[1])
	assert.Equal(t, "test-job-worker-2", resourceNames[2])
}

func TestBuildLauncherRoleZeroWorkers(t *testing.T) {
	ownerRef := metav1.OwnerReference{
		APIVersion: "kubeflow.org/v1",
		Kind:       "MPIJob",
		Name:       "test-job",
		UID:        "abc-123",
	}

	role := buildLauncherRole("test-job", "default", 0, ownerRef)

	rules, found, err := unstructured.NestedSlice(role.Object, "rules")
	require.NoError(t, err)
	assert.True(t, found)
	require.Len(t, rules, 1, "zero-worker Role should have only 1 rule (pods get/list/watch)")

	// Rule 1: get/list/watch on pods (the only rule)
	rule1 := rules[0].(map[string]interface{})
	verbs1, _ := rule1["verbs"].([]interface{})
	assert.Contains(t, verbs1, "get")
	assert.Contains(t, verbs1, "list")
	assert.Contains(t, verbs1, "watch")
	resources1, _ := rule1["resources"].([]interface{})
	assert.Contains(t, resources1, "pods")

	// No pods/exec rule should be present
	for _, r := range rules {
		rm := r.(map[string]interface{})
		resources, _ := rm["resources"].([]interface{})
		assert.NotContains(t, resources, "pods/exec", "pods/exec rule must be omitted when workerReplicas == 0")
	}
}

func TestBuildLauncherRoleBinding(t *testing.T) {
	ownerRef := metav1.OwnerReference{
		APIVersion: "kubeflow.org/v1",
		Kind:       "MPIJob",
		Name:       "test-job",
		UID:        "abc-123",
	}

	rb := buildLauncherRoleBinding("test-job", "default", "test-job-launcher", ownerRef)

	assert.Equal(t, "RoleBinding", rb.GetKind())
	assert.Equal(t, "rbac.authorization.k8s.io/v1", rb.GetAPIVersion())
	assert.Equal(t, "test-job-launcher", rb.GetName())
	assert.Equal(t, "default", rb.GetNamespace())

	labels := rb.GetLabels()
	assert.Equal(t, "test-job", labels["app"])

	refs := rb.GetOwnerReferences()
	require.Len(t, refs, 1)
	assert.Equal(t, "test-job", refs[0].Name)

	// Verify subjects
	subjects, found, err := unstructured.NestedSlice(rb.Object, "subjects")
	require.NoError(t, err)
	assert.True(t, found)
	require.Len(t, subjects, 1)
	subject := subjects[0].(map[string]interface{})
	assert.Equal(t, "ServiceAccount", subject["kind"])
	assert.Equal(t, "test-job-launcher", subject["name"])
	assert.Equal(t, "default", subject["namespace"])

	// Verify roleRef
	roleRef, found, err := unstructured.NestedMap(rb.Object, "roleRef")
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "rbac.authorization.k8s.io", roleRef["apiGroup"])
	assert.Equal(t, "Role", roleRef["kind"])
	assert.Equal(t, "test-job-launcher", roleRef["name"])
}

func TestMPIProviderBuildRBAC_DefaultSA(t *testing.T) {
	ownerRef := metav1.OwnerReference{
		APIVersion: "kubeflow.org/v1",
		Kind:       "MPIJob",
		Name:       "test-job",
		UID:        "abc-123",
	}

	tk := &task.Task{
		Name:      "test-job",
		Namespace: "default",
		Framework: task.Framework{Name: "mpi"},
		Worker:    &task.Worker{Replicas: 2},
	}

	p := &MPIProvider{APIVersion: MPIAPIVersionV2beta1}
	resources, err := p.BuildRBAC(tk, ownerRef)
	require.NoError(t, err)
	require.Len(t, resources, 3)

	// Resource 0: ServiceAccount
	sa := resources[0]
	assert.Equal(t, "ServiceAccount", sa.GetKind())
	assert.Equal(t, "test-job-launcher", sa.GetName())

	// Resource 1: Role
	role := resources[1]
	assert.Equal(t, "Role", role.GetKind())
	assert.Equal(t, "test-job-launcher", role.GetName())

	// Resource 2: RoleBinding
	rb := resources[2]
	assert.Equal(t, "RoleBinding", rb.GetKind())
	assert.Equal(t, "test-job-launcher", rb.GetName())

	// All resources should have ownerReferences
	for _, r := range resources {
		refs := r.GetOwnerReferences()
		assert.Len(t, refs, 1)
		assert.Equal(t, "test-job", refs[0].Name)
	}
}

func TestMPIProviderBuildRBAC_UserSA(t *testing.T) {
	ownerRef := metav1.OwnerReference{
		APIVersion: "kubeflow.org/v1",
		Kind:       "MPIJob",
		Name:       "test-job",
		UID:        "abc-123",
	}

	tk := &task.Task{
		Name:           "test-job",
		Namespace:      "default",
		ServiceAccount: "my-existing-sa",
		Framework:      task.Framework{Name: "mpi"},
		Worker:         &task.Worker{Replicas: 2},
	}

	p := &MPIProvider{APIVersion: MPIAPIVersionV2beta1}
	resources, err := p.BuildRBAC(tk, ownerRef)
	require.NoError(t, err)
	require.Len(t, resources, 2)

	// No ServiceAccount resource (user specified existing one)
	assert.Equal(t, "Role", resources[0].GetKind())
	assert.Equal(t, "RoleBinding", resources[1].GetKind())

	// RoleBinding should reference the user's SA
	rb := resources[1]
	subjects, found, err := unstructured.NestedSlice(rb.Object, "subjects")
	require.NoError(t, err)
	assert.True(t, found)
	subject := subjects[0].(map[string]interface{})
	assert.Equal(t, "my-existing-sa", subject["name"])
}

func TestMPIProviderBuildRBAC_ZeroWorkers(t *testing.T) {
	ownerRef := metav1.OwnerReference{
		APIVersion: "kubeflow.org/v1",
		Kind:       "MPIJob",
		Name:       "test-job",
		UID:        "abc-123",
	}

	tk := &task.Task{
		Name:      "test-job",
		Namespace: "default",
		Framework: task.Framework{Name: "mpi"},
		Worker:    &task.Worker{Replicas: 0},
	}

	p := &MPIProvider{APIVersion: MPIAPIVersionV2beta1}
	resources, err := p.BuildRBAC(tk, ownerRef)
	require.NoError(t, err)
	require.Len(t, resources, 3)

	role := resources[1]
	rules, found, err := unstructured.NestedSlice(role.Object, "rules")
	require.NoError(t, err)
	assert.True(t, found)
	require.Len(t, rules, 1, "zero-worker Role should have only 1 rule (pods get/list/watch)")

	// The single rule must be the pods get/list/watch rule
	rule1 := rules[0].(map[string]interface{})
	resources1, _ := rule1["resources"].([]interface{})
	assert.Contains(t, resources1, "pods")
	assert.NotContains(t, resources1, "pods/exec", "pods/exec rule must be omitted when workerReplicas == 0")
}

func TestMPIProviderBuildRBAC_NilWorker(t *testing.T) {
	ownerRef := metav1.OwnerReference{
		APIVersion: "kubeflow.org/v1",
		Kind:       "MPIJob",
		Name:       "test-job",
		UID:        "abc-123",
	}

	tk := &task.Task{
		Name:      "test-job",
		Namespace: "default",
		Framework: task.Framework{Name: "mpi"},
		Worker:    nil,
	}

	p := &MPIProvider{APIVersion: MPIAPIVersionV2beta1}

	require.NotPanics(t, func() {
		resources, err := p.BuildRBAC(tk, ownerRef)
		require.NoError(t, err)
		require.Len(t, resources, 3, "nil Worker should still produce SA, Role, RoleBinding")

		// Role should have only 1 rule (pods get/list/watch), no pods/exec
		role := resources[1]
		assert.Equal(t, "Role", role.GetKind())

		rules, found, err := unstructured.NestedSlice(role.Object, "rules")
		require.NoError(t, err)
		assert.True(t, found)
		require.Len(t, rules, 1, "nil Worker means 0 replicas, so no pods/exec rule")

		rule1 := rules[0].(map[string]interface{})
		resources1, _ := rule1["resources"].([]interface{})
		assert.Contains(t, resources1, "pods")
		assert.NotContains(t, resources1, "pods/exec")
	})
}

func TestBuildLauncherRole_SingleWorker(t *testing.T) {
	ownerRef := metav1.OwnerReference{
		APIVersion: "kubeflow.org/v1",
		Kind:       "MPIJob",
		Name:       "single-worker-job",
		UID:        "uid-789",
	}

	role := buildLauncherRole("single-worker-job", "default", 1, ownerRef)

	rules, found, err := unstructured.NestedSlice(role.Object, "rules")
	require.NoError(t, err)
	assert.True(t, found)
	require.Len(t, rules, 2, "1 worker should produce 2 rules (pods + pods/exec)")

	// pods/exec rule should have exactly 1 resourceName
	rule2 := rules[1].(map[string]interface{})
	resourceNames, _ := rule2["resourceNames"].([]interface{})
	require.Len(t, resourceNames, 1)
	assert.Equal(t, "single-worker-job-worker-0", resourceNames[0])
}
