package cli

import (
	"errors"
	"fmt"
	"strconv"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// parseResourceValue extracts a resource quantity from nested maps.
// K8s stores resource values as strings (e.g., "8"), but may also be int.
func parseResourceValue(obj map[string]interface{}, fields ...string) (int64, error) {
	val, found, err := unstructured.NestedFieldNoCopy(obj, fields...)
	if err != nil || !found {
		return 0, errors.New("not found")
	}

	switch v := val.(type) {
	case string:
		parsed, err := strconv.Atoi(v)
		if err != nil {
			return 0, fmt.Errorf("parse error: %w", err)
		}
		return int64(parsed), nil
	case int64:
		return v, nil
	case int:
		return int64(v), nil
	case float64:
		return int64(v), nil
	default:
		return 0, fmt.Errorf("unsupported type %T", v)
	}
}

// extractGPURequested calculates the total GPU requests across all roles in a CRD.
// Returns 0 if no GPU resources are requested.
func extractGPURequested(obj *unstructured.Unstructured) int {
	spec, found, err := unstructured.NestedMap(obj.Object, "spec")
	if err != nil || !found {
		return 0
	}

	total := 0
	// Iterate over spec to find <framework>ReplicaSpecs
	for _, replicaSpecsVal := range spec {
		replicaSpecs, ok := replicaSpecsVal.(map[string]interface{})
		if !ok {
			continue
		}

		// Iterate over roles (Worker, Master, Launcher, etc.)
		for _, roleSpecVal := range replicaSpecs {
			roleSpec, ok := roleSpecVal.(map[string]interface{})
			if !ok {
				continue
			}

			// Get replicas
			replicas, found, err := unstructured.NestedInt64(roleSpec, "replicas")
			if err != nil || !found {
				replicas = 1
			}

			// Get containers[0].resources.requests["nvidia.com/gpu"]
			containers, found, err := unstructured.NestedSlice(roleSpec, "template", "spec", "containers")
			if err != nil || !found || len(containers) == 0 {
				continue
			}

			container, ok := containers[0].(map[string]interface{})
			if !ok {
				continue
			}

			gpu, err := parseResourceValue(container, "resources", "requests", "nvidia.com/gpu")
			if err != nil {
				continue
			}

			total += int(replicas * gpu)
		}
	}

	return total
}
