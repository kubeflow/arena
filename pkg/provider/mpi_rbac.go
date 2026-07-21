package provider

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// buildLauncherServiceAccount creates a ServiceAccount for the MPIJob launcher pod.
// The SA is named <jobName>-launcher and carries an ownerReference to the MPIJob CRD
// so it is garbage-collected when the job is deleted.
func buildLauncherServiceAccount(saName, jobName, namespace string, ownerRef metav1.OwnerReference) *unstructured.Unstructured {
	sa := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ServiceAccount",
			"metadata": map[string]interface{}{
				"name":      saName,
				"namespace": namespace,
				"labels": map[string]interface{}{
					"app": jobName,
				},
			},
		},
	}
	sa.SetOwnerReferences([]metav1.OwnerReference{ownerRef})
	return sa
}

// buildLauncherRole creates a least-privilege Role for the launcher; list/watch are namespace-wide, get/exec are restricted to worker pod names.
func buildLauncherRole(jobName, namespace string, workerReplicas int, ownerRef metav1.OwnerReference) *unstructured.Unstructured {
	// list/watch must be namespace-wide (K8s RBAC cannot restrict them with
	// resourceNames). When there are no workers, get is also namespace-wide
	// because there are no specific pod names to restrict it to.
	podVerbs := []interface{}{"list", "watch"}
	if workerReplicas == 0 {
		podVerbs = []interface{}{"get", "list", "watch"}
	}

	rules := []interface{}{
		map[string]interface{}{
			"verbs":     podVerbs,
			"apiGroups": []interface{}{""},
			"resources": []interface{}{"pods"},
		},
	}

	if workerReplicas > 0 {
		// Pod names follow the training-operator convention: <jobName>-worker-<index>.
		podNames := make([]interface{}, workerReplicas)
		for i := 0; i < workerReplicas; i++ {
			podNames[i] = fmt.Sprintf("%s-worker-%d", jobName, i)
		}
		// Restrict the get verb to specific worker pod names (least-privilege).
		rules = append(rules, map[string]interface{}{
			"verbs":         []interface{}{"get"},
			"apiGroups":     []interface{}{""},
			"resources":     []interface{}{"pods"},
			"resourceNames": podNames,
		})
		rules = append(rules, map[string]interface{}{
			"verbs":         []interface{}{"create"},
			"apiGroups":     []interface{}{""},
			"resources":     []interface{}{"pods/exec"},
			"resourceNames": podNames,
		})
	}

	role := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "rbac.authorization.k8s.io/v1",
			"kind":       "Role",
			"metadata": map[string]interface{}{
				"name":      jobName + "-launcher",
				"namespace": namespace,
				"labels": map[string]interface{}{
					"app": jobName,
				},
			},
			"rules": rules,
		},
	}
	role.SetOwnerReferences([]metav1.OwnerReference{ownerRef})
	return role
}

// buildLauncherRoleBinding creates a RoleBinding that binds the launcher Role
// to the launcher ServiceAccount.
func buildLauncherRoleBinding(jobName, namespace, saName string, ownerRef metav1.OwnerReference) *unstructured.Unstructured {
	rb := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "rbac.authorization.k8s.io/v1",
			"kind":       "RoleBinding",
			"metadata": map[string]interface{}{
				"name":      jobName + "-launcher",
				"namespace": namespace,
				"labels": map[string]interface{}{
					"app": jobName,
				},
			},
			"subjects": []interface{}{
				map[string]interface{}{
					"kind":      "ServiceAccount",
					"name":      saName,
					"namespace": namespace,
				},
			},
			"roleRef": map[string]interface{}{
				"apiGroup": "rbac.authorization.k8s.io",
				"kind":     "Role",
				"name":     jobName + "-launcher",
			},
		},
	}
	rb.SetOwnerReferences([]metav1.OwnerReference{ownerRef})
	return rb
}
