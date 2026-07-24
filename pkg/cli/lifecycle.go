package cli

import (
	"context"
	"fmt"
	"regexp"

	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/kubeflow/arena/pkg/client"
	"github.com/kubeflow/arena/pkg/constants"
	"github.com/kubeflow/arena/pkg/log"
	"github.com/kubeflow/arena/pkg/provider"
	"github.com/kubeflow/arena/pkg/task"
)

// secretPattern matches environment variable names that commonly hold sensitive
// values (tokens, keys, secrets, passwords, credentials).
var secretPattern = regexp.MustCompile(`(?i)(token|key|secret|password|credential)`)

// createdResource tracks a single resource that was successfully created,
// so it can be rolled back on partial failure.
type createdResource struct {
	kind      string
	namespace string
	name      string
}

// rollback deletes previously created resources in reverse order (best-effort).
// Errors during cleanup are logged but not returned.
func rollback(ctx context.Context, k8sClient *client.Client, created []createdResource) {
	for i := len(created) - 1; i >= 0; i-- {
		r := created[i]
		if err := k8sClient.Delete(ctx, r.kind, r.namespace, r.name); err != nil {
			log.Warning("failed to clean up resource after partial failure",
				"kind", r.kind, "name", r.name, "namespace", r.namespace, "error", err.Error())
		}
	}
}

// isLikelySecretKey returns true if the given env var name matches common
// secret-naming patterns, suggesting the value should use a secretKeyRef
// instead of being stored in plaintext.
func isLikelySecretKey(name string) bool {
	return secretPattern.MatchString(name)
}

// preCreateRBAC creates or updates RBAC resources before the CRD is submitted.
// Resources are created without ownerReferences — these are patched onto
// the resources by finalizeJobResources after the CRD exists and its UID is known.
// Returns the list of created/updated resources for rollback tracking and
// ownerReference patching.
//
// Caveat: a hard crash before finalizeJobResources leaves RBAC resources owner-less;
// Kubernetes GC will not reclaim them. A re-submit will correctly patch ownerReferences
// onto the existing resources via CreateOrUpdate, so no manual cleanup is needed.
func preCreateRBAC(
	ctx context.Context,
	t *task.Task,
	k8sClient *client.Client,
	p provider.Provider,
) ([]createdResource, error) {
	rbacResources, err := p.BuildRBAC(t, metav1.OwnerReference{})
	if err != nil {
		return nil, fmt.Errorf("failed to build RBAC resources: %w", err)
	}

	var created []createdResource
	for _, r := range rbacResources {
		r.SetNamespace(t.Namespace)
		r.SetOwnerReferences(nil)
		if err := k8sClient.CreateOrUpdate(ctx, r); err != nil {
			rollback(ctx, k8sClient, created)
			return nil, err
		}
		created = append(created, createdResource{
			kind:      r.GetKind(),
			namespace: r.GetNamespace(),
			name:      r.GetName(),
		})
	}
	return created, nil
}

// finalizeJobResources creates the ConfigMap anchor, patches ownerReferences
// onto pre-created RBAC resources, and creates TensorBoard resources if enabled.
// Called after the CRD has been submitted so its UID is available for ownerReferences.
func finalizeJobResources(
	ctx context.Context,
	crd *unstructured.Unstructured,
	t *task.Task,
	k8sClient *client.Client,
	p provider.Provider,
	rbacResources []createdResource,
) error {
	// Re-fetch CRD to get UID assigned by the API server
	created, err := k8sClient.Get(ctx, crd.GetKind(), crd.GetNamespace(), crd.GetName())
	if err != nil {
		return fmt.Errorf("failed to re-fetch CRD for UID: %w", err)
	}

	ownerRef := metav1.OwnerReference{
		APIVersion:         crd.GetAPIVersion(),
		Kind:               crd.GetKind(),
		Name:               crd.GetName(),
		UID:                created.GetUID(),
		BlockOwnerDeletion: ptrBool(true),
		Controller:         ptrBool(true),
	}

	createdResources := make([]createdResource, 0)

	// SEC-7: Warn about env vars that look like secrets but are stored in
	// plaintext in the ConfigMap.
	for k, v := range t.Envs {
		if isLikelySecretKey(k) && v.Value != "" && v.Secret == nil {
			log.Warning("env var looks like a secret but is stored in plaintext in ConfigMap",
				"key", k, "hint", "use secretKeyRef instead")
		}
	}

	// Create ConfigMap anchor with ownerRef at creation time
	yamlContent, err := yaml.Marshal(t)
	if err != nil {
		return fmt.Errorf("failed to marshal task configuration: %w", err)
	}
	cm := buildConfigMap(
		crd.GetName(),
		crd.GetNamespace(),
		string(yamlContent),
		ownerRef,
	)
	if err := k8sClient.CreateOrUpdate(ctx, cm); err != nil {
		return fmt.Errorf("failed to create ConfigMap: %w", err)
	}
	createdResources = append(createdResources, createdResource{
		kind: "ConfigMap", namespace: crd.GetNamespace(), name: crd.GetName(),
	})

	// Patch ownerReferences onto pre-created RBAC resources
	for _, r := range rbacResources {
		if err := k8sClient.PatchOwnerReferences(
			ctx, r.kind, r.namespace, r.name, []metav1.OwnerReference{ownerRef},
		); err != nil {
			rollback(ctx, k8sClient, createdResources)
			return fmt.Errorf("failed to patch ownerReference onto %s %q: %w", r.kind, r.name, err)
		}
		createdResources = append(createdResources, r)
	}

	// Create TensorBoard resources if enabled
	if t.Logging.TensorBoard != nil && t.Logging.TensorBoard.Enabled {
		log.Warning("TensorBoard has no built-in authentication; the UI will be accessible to any pod with network access to this namespace",
			"port", constants.DefaultTensorBoardPort)
		tbResources, err := createTensorBoardResources(ctx, crd, t, k8sClient, ownerRef)
		if err != nil {
			rollback(ctx, k8sClient, tbResources)
			rollback(ctx, k8sClient, createdResources)
			return err
		}
	}

	return nil
}

// createTensorBoardResources creates a TensorBoard Deployment and Service
// with ownerReferences pointing to the main training job CRD.
// Returns the list of successfully created resources for rollback purposes.
func createTensorBoardResources(
	ctx context.Context,
	crd *unstructured.Unstructured,
	t *task.Task,
	k8sClient *client.Client,
	ownerRef metav1.OwnerReference,
) ([]createdResource, error) {
	tbName := crd.GetName() + "-tensorboard"
	tbImage := constants.DefaultTensorBoardImage
	if t.Logging.TensorBoard.Image != "" {
		tbImage = t.Logging.TensorBoard.Image
	}
	logDir := t.Logging.TensorBoard.LogDir

	created := make([]createdResource, 0, 2)

	deploy := buildTensorBoardDeployment(
		tbName,
		crd.GetName(),
		crd.GetNamespace(),
		tbImage,
		logDir,
		t,
		ownerRef,
	)
	if err := k8sClient.CreateOrUpdate(ctx, deploy); err != nil {
		return created, fmt.Errorf("failed to create TensorBoard Deployment: %w", err)
	}
	created = append(created, createdResource{
		kind: "Deployment", namespace: crd.GetNamespace(), name: tbName,
	})

	svc := buildTensorBoardService(
		tbName,
		crd.GetName(),
		crd.GetNamespace(),
		ownerRef,
	)
	if err := k8sClient.CreateOrUpdate(ctx, svc); err != nil {
		return created, fmt.Errorf("failed to create TensorBoard Service: %w", err)
	}
	created = append(created, createdResource{
		kind: "Service", namespace: crd.GetNamespace(), name: tbName,
	})

	return created, nil
}

// buildConfigMap creates a ConfigMap that stores the arena-v2.yaml task
// configuration. The ConfigMap has an ownerReference to the CRD so that
// deleting the CRD cascades to remove it via K8s garbage collection.
func buildConfigMap(name, namespace, yamlContent string, ownerRef metav1.OwnerReference) *unstructured.Unstructured {
	cm := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
			},
			"data": map[string]interface{}{
				"arena-v2.yaml": yamlContent,
			},
		},
	}
	cm.SetOwnerReferences([]metav1.OwnerReference{ownerRef})
	return cm
}

// buildTensorBoardDeployment creates a TensorBoard Deployment with a single
// replica running the TensorBoard server. All task storages are injected as
// volumes; when TensorBoardConfig.Mounts is set, only the listed storages
// receive volumeMounts (with mount fields overriding storage defaults).
func buildTensorBoardDeployment(
	name, jobName, namespace, image, logDir string,
	t *task.Task,
	ownerRef metav1.OwnerReference,
) *unstructured.Unstructured {
	labels := tensorBoardPodLabels(jobName)

	args := []interface{}{"--host", "0.0.0.0"}
	if logDir != "" {
		args = append(args, "--logdir", logDir)
	}

	container := map[string]interface{}{
		"name":  "tensorboard",
		"image": image,
		"command": []interface{}{
			"tensorboard",
		},
		"args": args,
		"ports": []interface{}{
			map[string]interface{}{
				"containerPort": int64(constants.DefaultTensorBoardPort),
			},
		},
	}

	// Build volumes and volumeMounts from task storages, respecting the
	// optional TensorBoard mounts override (identical to sync/init behavior):
	// - No storages: empty volumes and volumeMounts
	// - Storages, no mounts field: inject all storages (volumes + volumeMounts)
	// - Storages + mounts field: all storages become volumes, but only the
	//   mounted storages get volumeMounts (mount fields override storage defaults)
	tbVolumes := []interface{}{}
	tbVolumeMounts := []interface{}{}
	if t != nil && len(t.Storages) > 0 {
		allVolumes, _ := provider.BuildVolumes(t)
		tbVolumes = allVolumes

		var tbMounts []task.Mount
		if t.Logging.TensorBoard != nil {
			tbMounts = t.Logging.TensorBoard.Mounts
		}
		tbVolumeMounts = provider.ResolveContainerMounts(tbMounts, t)
	}
	if len(tbVolumeMounts) > 0 {
		container["volumeMounts"] = tbVolumeMounts
	}

	podSpec := map[string]interface{}{
		"containers": []interface{}{container},
	}
	if len(tbVolumes) > 0 {
		podSpec["volumes"] = tbVolumes
	}

	deploy := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
			},
			"spec": map[string]interface{}{
				"replicas": int64(1),
				"selector": map[string]interface{}{
					"matchLabels": labels,
				},
				"template": map[string]interface{}{
					"metadata": map[string]interface{}{
						"labels": labels,
					},
					"spec": podSpec,
				},
			},
		},
	}
	deploy.SetOwnerReferences([]metav1.OwnerReference{ownerRef})
	return deploy
}

// buildTensorBoardService creates a Service exposing the TensorBoard
// Deployment on port 6006.
func buildTensorBoardService(
	name, jobName, namespace string,
	ownerRef metav1.OwnerReference,
) *unstructured.Unstructured {
	svc := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Service",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
			},
			"spec": map[string]interface{}{
				"selector": tensorBoardPodLabels(jobName),
				"ports": []interface{}{
					map[string]interface{}{
						"port":       int64(constants.DefaultTensorBoardPort),
						"targetPort": int64(constants.DefaultTensorBoardPort),
					},
				},
			},
		},
	}
	svc.SetOwnerReferences([]metav1.OwnerReference{ownerRef})
	return svc
}

// tensorBoardPodLabels returns the pod labels used by a TensorBoard Deployment
// and matched by its Service selector. The jobName value enables direct lookup
// of a TensorBoard by training job name.
func tensorBoardPodLabels(jobName string) map[string]interface{} {
	return map[string]interface{}{
		constants.LabelComponent:    constants.ComponentTensorBoard,
		constants.LabelArenaJobName: jobName,
	}
}

// ptrBool returns a pointer to a bool value.
func ptrBool(b bool) *bool {
	return &b
}
