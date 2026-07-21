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

// buildLauncherRole creates a Role granting the launcher permission to get/list/watch
// pods and exec into specific worker pods. The pods/exec permission is restricted
// to worker pod names via resourceNames (least-privilege). When workerReplicas is 0,
// the pods/exec rule is omitted entirely because an empty resourceNames list would
// grant exec on ALL pods (the opposite of least-privilege).
func buildLauncherRole(jobName, namespace string, workerReplicas int, ownerRef metav1.OwnerReference) *unstructured.Unstructured {
	rules := []interface{}{
		map[string]interface{}{
			"verbs":     []interface{}{"get", "list", "watch"},
			"apiGroups": []interface{}{""},
			"resources": []interface{}{"pods"},
		},
	}

	if workerReplicas > 0 {
		// Pod names follow the training-operator convention: <jobName>-<replicaType>-<index>.
		// The MPI operator creates worker pods named <jobName>-worker-0, <jobName>-worker-1, etc.
		// This must stay in sync with the operator's naming scheme; if the operator changes
		// its naming convention, this list will need to be updated accordingly.
		podNames := make([]interface{}, workerReplicas)
		for i := 0; i < workerReplicas; i++ {
			podNames[i] = fmt.Sprintf("%s-worker-%d", jobName, i)
		}
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
