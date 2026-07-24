package cli

import (
	"context"
	"errors"
	"fmt"

	"gopkg.in/yaml.v3"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/kubeflow/arena/pkg/client"
	"github.com/kubeflow/arena/pkg/constants"
	"github.com/kubeflow/arena/pkg/log"
	"github.com/kubeflow/arena/pkg/task"
)

// errV1Job is returned when a job was created by arena v1 and cannot be
// managed by the v2 CLI.
var errV1Job = errors.New("job was created by arena v1")

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
	if err != nil {
		if !apierrors.IsNotFound(err) {
			log.Warning("ConfigMap lookup failed (non-NotFound), falling through to CRD detection", "namespace", namespace, "name", name, "error", err.Error())
		}
	} else {
		kind, found := parseConfigMapJobType(cm)
		if found {
			if kind == constants.KindMPIJob && !mpiAvailable {
				return "", fmt.Errorf("job %q is an MPIJob but MPIJob CRD is not installed", name)
			}
			return kind, nil
		}
	}

	// Phase 2: CRD direct check (v1 detection)
	for _, kind := range supportedJobKinds {
		if kind == constants.KindMPIJob && !mpiAvailable {
			continue
		}
		crd, crdErr := k8sClient.Get(ctx, kind, namespace, name)
		if crdErr != nil {
			if !apierrors.IsNotFound(crdErr) {
				log.Debug("CRD probe failed", "kind", kind, "error", crdErr.Error())
			}
			continue
		}
		// CRD exists — check for v2 label
		labels := crd.GetLabels()
		if _, ok := labels[frameworkLabel]; !ok {
			return "", errV1Job
		}
		// Has label but no ConfigMap (abnormal) — return kind
		return kind, nil
	}

	return "", fmt.Errorf("job %q not found in namespace %q", name, namespace)
}

// parseConfigMapJobType extracts the CRD kind from the arena-v2.yaml data
// stored in a ConfigMap. Returns (kind, true) if found, or ("", false) otherwise.
func parseConfigMapJobType(cm *unstructured.Unstructured) (string, bool) {
	data, found, _ := unstructured.NestedMap(cm.Object, "data")
	if !found {
		return "", false
	}
	yamlContent, ok := data["arena-v2.yaml"].(string)
	if !ok || yamlContent == "" {
		return "", false
	}
	var t task.Task
	if err := yaml.Unmarshal([]byte(yamlContent), &t); err != nil {
		log.Warning("failed to unmarshal ConfigMap data", "namespace", cm.GetNamespace(), "name", cm.GetName(), "error", err.Error())
		return "", false
	}
	if t.Framework.Name == "" {
		return "", false
	}
	kind := frameworkToKind(t.Framework.Name)
	if kind == "" {
		return "", false
	}
	return kind, true
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
		switch ref.Kind {
		case constants.KindPyTorchJob, constants.KindTFJob, constants.KindMPIJob:
			return fmt.Errorf("job %q already exists (type: %s)", name, ref.Kind)
		}
	}

	// ConfigMap exists but no recognized ownerReference (abnormal)
	return fmt.Errorf("job %q already exists", name)
}
