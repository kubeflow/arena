package client

import (
	"context"
	"errors"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ErrCRDNotFound is returned when the requested CRD does not exist in the cluster.
var ErrCRDNotFound = errors.New("crd not found")

// MPISupportedVersions lists the MPIJob API versions that Arena supports.
var MPISupportedVersions = []string{"v1", "v2beta1"}

// crdGVR is the GroupVersionResource for CustomResourceDefinition objects.
var crdGVR = schema.GroupVersionResource{
	Group:    "apiextensions.k8s.io",
	Version:  "v1",
	Resource: "customresourcedefinitions",
}

// CRDVersionInfo represents one entry from a CRD's spec.versions.
type CRDVersionInfo struct {
	Name    string // e.g., "v2beta1"
	Served  bool
	Storage bool
}

// FindStorageVersion returns the storage version name from a version list.
// Returns "" if no storage version exists or the list is empty.
func FindStorageVersion(versions []CRDVersionInfo) string {
	for _, v := range versions {
		if v.Storage {
			return v.Name
		}
	}
	return ""
}

// GetCRDVersions queries the CRD object by name and returns all declared versions.
// Returns ErrCRDNotFound if the CRD is not installed in the cluster.
// Returns an empty slice if the CRD exists but has no versions defined.
func (c *Client) GetCRDVersions(ctx context.Context, crdName string) ([]CRDVersionInfo, error) {
	obj, err := c.dynamicClient.Resource(crdGVR).Get(ctx, crdName, v1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, ErrCRDNotFound
		}
		return nil, fmt.Errorf("failed to get CRD %s: %w", crdName, err)
	}

	return parseCRDVersions(obj)
}

// ResolveMPIVersion queries the MPIJob CRD and caches the storage version for the Client's lifetime.
func (c *Client) ResolveMPIVersion(ctx context.Context) error {
	c.mu.RLock()
	cached := c.mpiVersion
	c.mu.RUnlock()
	if cached != "" {
		return nil // cached
	}

	versions, err := c.GetCRDVersions(ctx, "mpijobs.kubeflow.org")
	if err != nil {
		if errors.Is(err, ErrCRDNotFound) {
			return fmt.Errorf("mpijob crd not found in cluster: %w", ErrCRDNotFound)
		}
		return fmt.Errorf("failed to get MPIJob CRD versions: %w", err)
	}

	storageVersion := FindStorageVersion(versions)
	if storageVersion == "" {
		return errors.New("mpijob crd has no storage version configured")
	}

	// Validate the resolved version is one we support.
	if !IsMPIVersionSupported(storageVersion) {
		return fmt.Errorf("mpijob storage version %q is not supported, supported versions: %v",
			storageVersion, MPISupportedVersions)
	}

	c.mu.Lock()
	c.mpiVersion = storageVersion
	c.mu.Unlock()
	return nil
}

// IsMPIVersionSupported reports whether version is in the supported set.
func IsMPIVersionSupported(version string) bool {
	for _, v := range MPISupportedVersions {
		if v == version {
			return true
		}
	}
	return false
}

// parseCRDVersions extracts version info from a CRD unstructured object.
// Returns an empty slice if the CRD has no spec.versions field.
func parseCRDVersions(obj *unstructured.Unstructured) ([]CRDVersionInfo, error) {
	versions, found, err := unstructured.NestedSlice(obj.Object, "spec", "versions")
	if err != nil {
		return nil, fmt.Errorf("failed to parse spec.versions: %w", err)
	}
	if !found {
		return []CRDVersionInfo{}, nil
	}

	result := make([]CRDVersionInfo, 0, len(versions))
	for _, v := range versions {
		vMap, ok := v.(map[string]interface{})
		if !ok {
			continue
		}
		name, _, _ := unstructured.NestedString(vMap, "name")
		served, _, _ := unstructured.NestedBool(vMap, "served")
		storage, _, _ := unstructured.NestedBool(vMap, "storage")
		result = append(result, CRDVersionInfo{
			Name:    name,
			Served:  served,
			Storage: storage,
		})
	}
	return result, nil
}
