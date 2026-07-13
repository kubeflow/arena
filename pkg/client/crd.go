package client

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

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
// Returns nil, nil if the CRD is not found (k8s API returns not-found error).
func (c *Client) GetCRDVersions(ctx context.Context, crdName string) ([]CRDVersionInfo, error) {
	obj, err := c.dynamicClient.Resource(crdGVR).Get(ctx, crdName, v1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get CRD %s: %w", crdName, err)
	}

	return parseCRDVersions(obj)
}

// parseCRDVersions extracts version info from a CRD unstructured object.
func parseCRDVersions(obj *unstructured.Unstructured) ([]CRDVersionInfo, error) {
	versions, found, err := unstructured.NestedSlice(obj.Object, "spec", "versions")
	if err != nil {
		return nil, fmt.Errorf("failed to parse spec.versions: %w", err)
	}
	if !found {
		return nil, nil
	}

	var result []CRDVersionInfo
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

// ResolveMPIVersion queries the MPIJob CRD and caches the storage version in c.MPIVersion.
// Returns immediately if MPIVersion is already set (cached).
func (c *Client) ResolveMPIVersion(ctx context.Context) error {
	if c.MPIVersion != "" {
		return nil // cached
	}

	versions, err := c.GetCRDVersions(ctx, "mpijobs.kubeflow.org")
	if err != nil {
		return fmt.Errorf("failed to get MPIJob CRD versions: %w", err)
	}
	if versions == nil {
		return fmt.Errorf("MPIJob CRD not found in cluster")
	}

	storageVersion := FindStorageVersion(versions)
	if storageVersion == "" {
		return fmt.Errorf("MPIJob CRD has no storage version configured")
	}

	c.MPIVersion = storageVersion
	return nil
}
