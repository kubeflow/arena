package client

import (
	"context"
	"fmt"
	"strings"

	"github.com/kubeflow/arena/pkg/constants"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
)

// Client wraps the Kubernetes dynamic client for working with training job CRDs.
type Client struct {
	dynamicClient dynamic.Interface
	mpiVersion    string // cached MPIJob API version, set by ResolveMPIVersion
}

// GetMPIVersion returns the cached MPIJob API version.
// Call ResolveMPIVersion first to populate this value.
func (c *Client) GetMPIVersion() string {
	return c.mpiVersion
}

// SetMPIVersion sets the cached MPIJob API version directly.
// Intended for use in tests; production code should call ResolveMPIVersion.
func (c *Client) SetMPIVersion(version string) {
	c.mpiVersion = version
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

// Create submits an unstructured CRD to the cluster.
func (c *Client) Create(ctx context.Context, crd *unstructured.Unstructured) error {
	gvr := getGVR(crd)
	namespace := crd.GetNamespace()

	_, err := c.dynamicClient.Resource(gvr).Namespace(namespace).Create(
		ctx,
		crd,
		v1.CreateOptions{},
	)
	if err != nil {
		return fmt.Errorf("failed to create %s %s: %w", crd.GetKind(), crd.GetName(), err)
	}

	return nil
}

// Get retrieves a CRD by kind, namespace, and name.
func (c *Client) Get(ctx context.Context, kind, namespace, name string) (*unstructured.Unstructured, error) {
	gvr, err := c.kindToGVR(kind)
	if err != nil {
		return nil, fmt.Errorf("failed to get %s %s: %w", kind, name, err)
	}
	obj, err := c.dynamicClient.Resource(gvr).Namespace(namespace).Get(
		ctx,
		name,
		v1.GetOptions{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get %s %s: %w", kind, name, err)
	}
	return obj, nil
}

// List returns all CRDs of a given kind in a namespace.
// An optional labelSelector restricts results (empty string means no filtering).
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

	var results []*unstructured.Unstructured
	for i := range list.Items {
		results = append(results, &list.Items[i])
	}
	return results, nil
}

// Delete removes a CRD by kind, namespace, and name.
func (c *Client) Delete(ctx context.Context, kind, namespace, name string) error {
	gvr, err := c.kindToGVR(kind)
	if err != nil {
		return fmt.Errorf("failed to delete %s %s: %w", kind, name, err)
	}
	err = c.dynamicClient.Resource(gvr).Namespace(namespace).Delete(
		ctx,
		name,
		v1.DeleteOptions{},
	)
	if err != nil {
		return fmt.Errorf("failed to delete %s %s: %w", kind, name, err)
	}
	return nil
}

// GetCRDObject retrieves a cluster-scoped CRD object by name.
// Only works for apiextensions.k8s.io/v1 resources (e.g., CustomResourceDefinition).
// Note: pluralization uses a naive "kind + s" approach which is correct for
// CustomResourceDefinition → customresourcedefinitions but would not work for
// irregular plurals. If this function is extended for other kinds, use a lookup table.
func (c *Client) GetCRDObject(ctx context.Context, kind, name string) (*unstructured.Unstructured, error) {
	gvr := schema.GroupVersionResource{
		Group:    "apiextensions.k8s.io",
		Version:  "v1",
		Resource: strings.ToLower(kind) + "s",
	}
	return c.dynamicClient.Resource(gvr).Get(ctx, name, v1.GetOptions{})
}

// Patch applies a JSON merge patch to a CRD.
func (c *Client) Patch(ctx context.Context, kind, namespace, name string, patch []byte) error {
	gvr, err := c.kindToGVR(kind)
	if err != nil {
		return fmt.Errorf("failed to patch %s %s: %w", kind, name, err)
	}
	_, err = c.dynamicClient.Resource(gvr).Namespace(namespace).Patch(
		ctx,
		name,
		types.MergePatchType,
		patch,
		v1.PatchOptions{},
	)
	if err != nil {
		return fmt.Errorf("failed to patch %s %s: %w", kind, name, err)
	}
	return nil
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

// kindToGVR maps a known training job kind to its GVR.
// Core K8s resources (ConfigMap, Pod, etc.) use the empty group.
// MPIJob uses the resolved mpiVersion from the Client; all other kubeflow kinds use v1.
// Note: PyTorchJob and TFJob currently always use v1 in the Training Operator,
// so hardcoding constants.KubeflowVersion ("v1") is correct for these frameworks.
func (c *Client) kindToGVR(kind string) (schema.GroupVersionResource, error) {
	version := constants.KubeflowVersion
	if kind == constants.KindMPIJob {
		if c.mpiVersion == "" {
			return schema.GroupVersionResource{},
				fmt.Errorf("MPIJob version not resolved; call ResolveMPIVersion first")
		}
		version = c.mpiVersion
	}

	// Check if it's a core K8s resource (empty group).
	if resource, ok := coreResources[kind]; ok {
		return schema.GroupVersionResource{
			Group:    "",
			Version:  "v1",
			Resource: resource,
		}, nil
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
	return strings.ToLower(kind) + "s"
}
