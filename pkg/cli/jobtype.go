package cli

import (
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"gopkg.in/yaml.v3"

	"github.com/kubeflow/arena/pkg/client"
	"github.com/kubeflow/arena/pkg/constants"
	"github.com/kubeflow/arena/pkg/log"
	"github.com/kubeflow/arena/pkg/task"
)

// ErrV1Job is returned when a job was created by arena v1 and cannot be
// managed by the v2 CLI.
var ErrV1Job = fmt.Errorf("this job was created by arena v1, please use the arena v1 CLI to manage it")

// detectJobType identifies the CRD kind for a job using a two-phase lookup:
// Phase 1: ConfigMap anchor (fast path for v2 jobs)
// Phase 2: CRD direct check (detects v1 jobs or abnormal v2 state)
func detectJobType(ctx context.Context, k8sClient *client.Client, namespace, name string) (string, error) {
	// Resolve MPI version upfront (cached, needed if job is MPIJob).
	// This ensures c.mpiVersion is set before any return path that might
	// return MPIJob, including the Phase 1 fast path.
	mpiAvailable := true
	if err := k8sClient.ResolveMPIVersion(ctx); err != nil {
		log.Warning("MPIJob CRD not available", "error", err.Error())
		mpiAvailable = false
	}

	// Phase 1: ConfigMap anchor (v2 fast path)
	cm, err := k8sClient.Get(ctx, "ConfigMap", namespace, name)
	if err == nil {
		data, found, _ := getNestedMap(cm.Object, "data")
		if found {
			if yamlContent, ok := data["arena-v2.yaml"].(string); ok && yamlContent != "" {
				var t task.Task
				if unmarshalErr := yaml.Unmarshal([]byte(yamlContent), &t); unmarshalErr == nil && t.Framework.Name != "" {
					kind := frameworkToKind(t.Framework.Name)
					if kind != "" {
						if kind == constants.KindMPIJob && !mpiAvailable {
							return "", fmt.Errorf("job %q is an MPIJob but MPIJob CRD is not installed", name)
						}
						return kind, nil
					}
				}
			}
		}
	}

	// Phase 2: CRD direct check (v1 detection)
	for _, kind := range supportedJobKinds {
		if kind == constants.KindMPIJob && !mpiAvailable {
			continue
		}
		crd, crdErr := k8sClient.Get(ctx, kind, namespace, name)
		if crdErr != nil {
			continue
		}
		// CRD exists — check for v2 label
		labels := crd.GetLabels()
		if _, ok := labels[FrameworkLabel]; !ok {
			return "", ErrV1Job
		}
		// Has label but no ConfigMap (abnormal) — return kind
		return kind, nil
	}

	return "", fmt.Errorf("job %q not found in namespace %q", name, namespace)
}

// checkJobExists returns an error if a v2 job with the given name already exists.
// It checks the ConfigMap anchor and extracts the CRD kind from ownerReferences.
// v1 CRD name collisions are not detected here; they are caught by the K8s API
// server when the duplicate CRD creation is attempted.
func checkJobExists(ctx context.Context, k8sClient *client.Client, namespace, name string) error {
	cm, err := k8sClient.Get(ctx, "ConfigMap", namespace, name)
	if err != nil {
		if isNotFoundError(err) {
			return nil
		}
		return fmt.Errorf("failed to check if job exists: %w", err)
	}

	// ConfigMap exists — extract CRD kind from ownerReferences
	refs := cm.GetOwnerReferences()
	for _, ref := range refs {
		if ref.Kind == constants.KindPyTorchJob || ref.Kind == constants.KindTFJob || ref.Kind == constants.KindMPIJob {
			return fmt.Errorf("job %q already exists (type: %s)", name, ref.Kind)
		}
	}

	// ConfigMap exists but no recognized ownerReference (abnormal)
	return fmt.Errorf("job %q already exists", name)
}

// isNotFoundError checks if an error indicates a resource was not found
// using the standard Kubernetes apierrors.IsNotFound() function.
func isNotFoundError(err error) bool {
	return apierrors.IsNotFound(err)
}

// frameworkToKind maps a framework name to its CRD kind.
func frameworkToKind(framework string) string {
	switch framework {
	case constants.FrameworkPyTorch:
		return constants.KindPyTorchJob
	case constants.FrameworkTensorFlow, "tf":
		return constants.KindTFJob
	case constants.FrameworkMPI, constants.FrameworkHorovod, constants.FrameworkDeepSpeed:
		return constants.KindMPIJob
	default:
		return ""
	}
}

// getNestedMap safely extracts a nested map from an unstructured object.
func getNestedMap(obj map[string]interface{}, fields ...string) (map[string]interface{}, bool, error) {
	val, found, err := nestedFieldNoCopy(obj, fields...)
	if !found || err != nil {
		return nil, found, err
	}

	m, ok := val.(map[string]interface{})
	if !ok {
		return nil, false, fmt.Errorf("%v is not a map", fields)
	}
	return m, true, nil
}

// nestedFieldNoCopy returns the value at the specified path without copying.
func nestedFieldNoCopy(obj map[string]interface{}, fields ...string) (interface{}, bool, error) {
	var val interface{} = obj
	for _, field := range fields {
		m, ok := val.(map[string]interface{})
		if !ok {
			return nil, false, fmt.Errorf("%v is not a map", field)
		}
		val, ok = m[field]
		if !ok {
			return nil, false, nil
		}
	}
	return val, true, nil
}
