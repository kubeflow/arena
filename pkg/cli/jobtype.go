package cli

import (
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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
		log.Debug("MPIJob CRD not available", "error", err.Error())
		mpiAvailable = false
	}

	// Phase 1: ConfigMap anchor (v2 fast path)
	cm, err := k8sClient.Get(ctx, "ConfigMap", namespace, name)
	if err == nil {
		data, found, _ := unstructured.NestedMap(cm.Object, "data")
		if found {
			if yamlContent, ok := data["arena-v2.yaml"].(string); ok && yamlContent != "" {
				var t task.Task
				if unmarshalErr := yaml.Unmarshal([]byte(yamlContent), &t); unmarshalErr == nil && t.Framework.Name != "" {
					kind := FrameworkToKind(t.Framework.Name)
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
		if apierrors.IsNotFound(err) {
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
