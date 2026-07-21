package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/kubeflow/arena/pkg/constants"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
)

// Client wraps the Kubernetes dynamic client; mpiVersion is guarded by RWMutex for concurrent access.
type Client struct {
	mu            sync.RWMutex
	dynamicClient dynamic.Interface
	mpiVersion    string // cached MPIJob API version, set by ResolveMPIVersion
}

// coreResources maps core K8s kinds (group "") to their plural resource names.
var coreResources = map[string]string{
	"ConfigMap":             "configmaps",
	"Pod":                   "pods",
	"Service":               "services",
	"Secret":                "secrets",
	"ServiceAccount":        "serviceaccounts",
	"Namespace":             "namespaces",
	"PersistentVolumeClaim": "persistentvolumeclaims",
}

// nonCoreResources maps K8s kinds that belong to non-empty API groups
// (rbac, apps, etc.) to their GroupVersionResource. This is used by
// kindToGVR for Get/Delete/Patch/List operations that receive only a kind
// string and cannot derive the GVR from an unstructured object.
var nonCoreResources = map[string]schema.GroupVersionResource{
	"Role":        {Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "roles"},
	"RoleBinding": {Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "rolebindings"},
	"Deployment":  {Group: "apps", Version: "v1", Resource: "deployments"},
}

// NewClient creates a new Kubernetes client using the provided kubeconfig path and optional context.
func NewClient(kubeconfig, kubeContext string) (*Client, error) {
	config, err := LoadRestConfig(kubeconfig, kubeContext)
	if err != nil {
		return nil, fmt.Errorf("failed to build config: %w", err)
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	return &Client{dynamicClient: dynamicClient}, nil
}

// NewClientWithNamespace creates a new Kubernetes client and resolves the effective namespace.
// The resolved namespace follows the priority: CLI flag > kubeconfig context > "default".
func NewClientWithNamespace(kubeconfig, kubeContext, cliNamespace string) (*Client, string, error) {
	c, err := NewClient(kubeconfig, kubeContext)
	if err != nil {
		return nil, "", err
	}
	ns := ResolveNamespace(kubeconfig, kubeContext, cliNamespace)
	return c, ns, nil
}

// NewClientForInterface creates a Client wrapping an existing dynamic.Interface (useful for testing).
func NewClientForInterface(client dynamic.Interface) *Client {
	return &Client{dynamicClient: client}
}

// GetMPIVersion returns the cached MPIJob API version.
// Call ResolveMPIVersion first to populate this value.
func (c *Client) GetMPIVersion() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.mpiVersion
}

// SetMPIVersion sets the cached MPIJob API version directly.
// Intended for use in tests; production code should call ResolveMPIVersion.
func (c *Client) SetMPIVersion(version string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.mpiVersion = version
}

// Create submits an unstructured CRD to the cluster.
// The GVR is derived from the object's own GroupVersionKind, which the provider
// sets correctly for every resource type (training CRDs, RBAC, ConfigMaps, etc.).
// This handles arbitrary resource kinds without requiring a static mapping.
func (c *Client) Create(ctx context.Context, crd *unstructured.Unstructured) error {
	gvr := getGVR(crd)
	namespace := crd.GetNamespace()

	_, err := c.dynamicClient.Resource(gvr).Namespace(namespace).Create(
		ctx,
		crd,
		v1.CreateOptions{},
	)
	if err != nil {
		return fmt.Errorf("failed to create %s %q: %w", crd.GetKind(), crd.GetName(), err)
	}

	return nil
}

// Update replaces an existing resource with the provided object.
// The resourceVersion is fetched from the current server state to satisfy
// optimistic concurrency requirements.
func (c *Client) Update(ctx context.Context, obj *unstructured.Unstructured) error {
	gvr := getGVR(obj)
	namespace := obj.GetNamespace()
	name := obj.GetName()

	existing, err := c.dynamicClient.Resource(gvr).Namespace(namespace).Get(ctx, name, v1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get %s %q for update: %w", obj.GetKind(), name, err)
	}

	obj.SetResourceVersion(existing.GetResourceVersion())
	_, err = c.dynamicClient.Resource(gvr).Namespace(namespace).Update(ctx, obj, v1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update %s %q: %w", obj.GetKind(), name, err)
	}
	return nil
}

// CreateOrUpdate creates the resource if it doesn't exist, or updates it
// with the provided object if it does. On conflict, retries once with a
// fresh resourceVersion. The update path delegates to Update, which
// handles Get + SetResourceVersion + Update internally.
func (c *Client) CreateOrUpdate(ctx context.Context, obj *unstructured.Unstructured) error {
	gvr := getGVR(obj)
	namespace := obj.GetNamespace()
	name := obj.GetName()
	kind := obj.GetKind()

	existing, err := c.dynamicClient.Resource(gvr).Namespace(namespace).Get(ctx, name, v1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("failed to get %s %q: %w", kind, name, err)
		}
		// Not found → Create
		_, err = c.dynamicClient.Resource(gvr).Namespace(namespace).Create(ctx, obj, v1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create %s %q: %w", kind, name, err)
		}
		return nil
	}

	// Found → Update with the current resourceVersion
	obj.SetResourceVersion(existing.GetResourceVersion())
	if err := c.Update(ctx, obj); err != nil {
		if !apierrors.IsConflict(err) {
			return err
		}
		// Conflict → retry once (Update re-fetches and re-applies)
		return c.Update(ctx, obj)
	}
	return nil
}

// Get retrieves a CRD by kind, namespace, and name.
func (c *Client) Get(ctx context.Context, kind, namespace, name string) (*unstructured.Unstructured, error) {
	gvr, err := c.kindToGVR(kind)
	if err != nil {
		return nil, fmt.Errorf("failed to get %s %q: %w", kind, name, err)
	}
	obj, err := c.dynamicClient.Resource(gvr).Namespace(namespace).Get(
		ctx,
		name,
		v1.GetOptions{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get %s %q: %w", kind, name, err)
	}
	return obj, nil
}

// List returns all CRDs of a given kind in a namespace.
// An optional labelSelector restricts results (empty string means no filtering).
//
// NOTE: The returned pointers alias into the internal list.Items slice.
// Callers must not modify them in place; deep-copy if mutation is needed.
func (c *Client) List(ctx context.Context, kind, namespace, labelSelector string) ([]*unstructured.Unstructured, error) {
	gvr, err := c.kindToGVR(kind)
	if err != nil {
		return nil, fmt.Errorf("failed to list %s: %w", kind, err)
	}
	listOpts := v1.ListOptions{}
	if labelSelector != "" {
		listOpts.LabelSelector = labelSelector
	}
	list, err := c.dynamicClient.Resource(gvr).Namespace(namespace).List(
		ctx,
		listOpts,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list %s: %w", kind, err)
	}

	results := make([]*unstructured.Unstructured, 0, len(list.Items))
	for i := range list.Items {
		results = append(results, &list.Items[i])
	}
	return results, nil
}

// Delete removes a CRD by kind, namespace, and name.
func (c *Client) Delete(ctx context.Context, kind, namespace, name string) error {
	gvr, err := c.kindToGVR(kind)
	if err != nil {
		return fmt.Errorf("failed to delete %s %q: %w", kind, name, err)
	}
	err = c.dynamicClient.Resource(gvr).Namespace(namespace).Delete(
		ctx,
		name,
		v1.DeleteOptions{},
	)
	if err != nil {
		return fmt.Errorf("failed to delete %s %q: %w", kind, name, err)
	}
	return nil
}

// Patch applies a JSON merge patch to a CRD.
func (c *Client) Patch(ctx context.Context, kind, namespace, name string, patch []byte) error {
	gvr, err := c.kindToGVR(kind)
	if err != nil {
		return fmt.Errorf("failed to patch %s %q: %w", kind, name, err)
	}
	_, err = c.dynamicClient.Resource(gvr).Namespace(namespace).Patch(
		ctx,
		name,
		types.MergePatchType,
		patch,
		v1.PatchOptions{},
	)
	if err != nil {
		return fmt.Errorf("failed to patch %s %q: %w", kind, name, err)
	}
	return nil
}

// PatchOwnerReferences patches the metadata.ownerReferences field of a resource,
// replacing the entire array. Uses JSON merge patch internally.
func (c *Client) PatchOwnerReferences(ctx context.Context, kind, namespace, name string, ownerRefs []v1.OwnerReference) error {
	ownerRefJSON, err := json.Marshal(ownerRefs)
	if err != nil {
		return fmt.Errorf("failed to marshal ownerReferences for %s %q: %w", kind, name, err)
	}
	patch := []byte(fmt.Sprintf(`{"metadata":{"ownerReferences":%s}}`, ownerRefJSON))
	return c.Patch(ctx, kind, namespace, name, patch)
}

// getGVR derives the GroupVersionResource from an unstructured object's GVK.
func getGVR(crd *unstructured.Unstructured) schema.GroupVersionResource {
	gvk := crd.GroupVersionKind()
	return schema.GroupVersionResource{
		Group:    gvk.Group,
		Version:  gvk.Version,
		Resource: pluralize(gvk.Kind),
	}
}

// kindToGVR maps a CRD kind to its GVR; MPIJob uses the resolved version, others use v1.
func (c *Client) kindToGVR(kind string) (schema.GroupVersionResource, error) {
	version := constants.KubeflowVersion
	if kind == constants.KindMPIJob {
		c.mu.RLock()
		mpiVer := c.mpiVersion
		c.mu.RUnlock()
		if mpiVer == "" {
			return schema.GroupVersionResource{},
				errors.New("mpijob version not resolved")
		}
		version = mpiVer
	}

	// Check if it's a core K8s resource (empty group).
	if resource, ok := coreResources[kind]; ok {
		return schema.GroupVersionResource{
			Group:    "",
			Version:  "v1",
			Resource: resource,
		}, nil
	}

	// Check if it's a non-core K8s resource (rbac, apps, etc.).
	if gvr, ok := nonCoreResources[kind]; ok {
		return gvr, nil
	}

	return schema.GroupVersionResource{
		Group:    constants.KubeflowGroup,
		Version:  version,
		Resource: pluralize(kind),
	}, nil
}

// KindToAPIVersion returns the full group/version string for a given CRD kind.
// Core K8s resources (group "") return just the version (e.g., "v1").
// Kubeflow resources return "group/version" (e.g., "kubeflow.org/v1").
// Returns an error if the kind cannot be resolved (e.g., MPIJob without version).
func (c *Client) KindToAPIVersion(kind string) (string, error) {
	gvr, err := c.kindToGVR(kind)
	if err != nil {
		return "", err
	}
	if gvr.Group == "" {
		return gvr.Version, nil
	}
	return gvr.Group + "/" + gvr.Version, nil
}

// pluralize handles the specific CRD kinds Arena uses.
func pluralize(kind string) string {
	resourceMap := map[string]string{
		constants.KindPyTorchJob: "pytorchjobs",
		constants.KindTFJob:      "tfjobs",
		constants.KindMPIJob:     "mpijobs",
	}
	if r, ok := resourceMap[kind]; ok {
		return r
	}
	// Check core resources too
	if r, ok := coreResources[kind]; ok {
		return r
	}
	// Best-effort fallback for unknown kinds; not reliable for arbitrary K8s resources
	// (e.g., "Policy" -> "policys"). Arena only uses known kinds above.
	return strings.ToLower(kind) + "s"
}
