package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/kubeflow/arena/pkg/client"
	"github.com/kubeflow/arena/pkg/constants"
	"github.com/kubeflow/arena/pkg/log"
	outputpkg "github.com/kubeflow/arena/pkg/output"
	"github.com/kubeflow/arena/pkg/provider"
	"github.com/kubeflow/arena/pkg/task"
)

// dryRunFormat returns the output format for dry-run.
// Only json and yaml are meaningful for CRD output.
func dryRunFormat() (outputpkg.Format, error) {
	f := outputpkg.Format(outputFormat)
	switch f {
	case outputpkg.FormatYAML, outputpkg.FormatJSON:
		return f, nil
	default:
		return "", fmt.Errorf("dry-run only supports -o json or yaml, got %q", outputFormat)
	}
}

// applyDryRunOutputDefault switches the shared output format from the table
// default to json when --dry-run is set and the user did not pass -o/--output.
func applyDryRunOutputDefault(cmd *cobra.Command, dryRun bool) {
	if dryRun && !cmd.Flags().Changed("output") {
		outputFormat = string(outputpkg.FormatJSON)
	}
}

// printCRD marshals a CRD as indented JSON or YAML and writes it to out.
func printCRD(out io.Writer, crd *unstructured.Unstructured, format outputpkg.Format) error {
	switch format {
	case outputpkg.FormatYAML:
		var buf bytes.Buffer
		enc := yaml.NewEncoder(&buf)
		enc.SetIndent(2)
		if err := enc.Encode(crd.Object); err != nil {
			return fmt.Errorf("failed to marshal CRD as YAML: %w", err)
		}
		if err := enc.Close(); err != nil {
			return fmt.Errorf("failed to close YAML encoder: %w", err)
		}
		fmt.Fprint(out, buf.String())
	case outputpkg.FormatJSON:
		data, err := json.MarshalIndent(crd.Object, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal CRD as JSON: %w", err)
		}
		fmt.Fprintln(out, string(data))
	default:
		return fmt.Errorf("unsupported CRD output format %q", format)
	}
	return nil
}

// printDryRun writes the CRD and all auxiliary resources that would be created
// during a real submission (TensorBoard Deployment and Service if enabled).
// Output format respects -o/--output (json or yaml; defaults to json).
// Multiple resources are separated by "---" for readability.
func printDryRun(out io.Writer, crd *unstructured.Unstructured, t *task.Task) error {
	format, err := dryRunFormat()
	if err != nil {
		return err
	}

	if err := printCRD(out, crd, format); err != nil {
		return fmt.Errorf("failed to print CRD: %w", err)
	}

	// Print TensorBoard resources if enabled
	if t.Logging.TensorBoard != nil && t.Logging.TensorBoard.Enabled {
		tbName := crd.GetName() + "-tensorboard"
		tbImage := constants.DefaultTensorBoardImage
		if t.Logging.TensorBoard.Image != "" {
			tbImage = t.Logging.TensorBoard.Image
		}
		logDir := t.Logging.TensorBoard.LogDir

		// Use a placeholder ownerRef for dry-run (UID is not available)
		ownerRef := metav1.OwnerReference{
			APIVersion:         crd.GetAPIVersion(),
			Kind:               crd.GetKind(),
			Name:               crd.GetName(),
			UID:                "dry-run",
			BlockOwnerDeletion: ptrBool(true),
			Controller:         ptrBool(true),
		}

		deploy := buildTensorBoardDeployment(
			tbName,
			crd.GetName(),
			crd.GetNamespace(),
			tbImage,
			logDir,
			t,
			ownerRef,
		)
		fmt.Fprintln(out, "---")
		if err := printCRD(out, deploy, format); err != nil {
			return fmt.Errorf("failed to print TensorBoard Deployment: %w", err)
		}

		svc := buildTensorBoardService(
			tbName,
			crd.GetName(),
			crd.GetNamespace(),
			ownerRef,
		)
		fmt.Fprintln(out, "---")
		if err := printCRD(out, svc, format); err != nil {
			return fmt.Errorf("failed to print TensorBoard Service: %w", err)
		}
	}

	return nil
}

// resolveNS resolves the effective namespace using the 4-level priority chain:
// CLI -n flag > YAML namespace > kubeconfig context namespace > "default".
func resolveNS(yamlNamespace string) string {
	ns := namespace
	if ns == "" {
		ns = yamlNamespace
	}
	if ns == "" {
		ns = client.ResolveNamespace(kubeconfig, kubeContext, "")
	}
	if isSystemNamespace(ns) {
		log.Warning("creating resources in system namespace — ensure this is intentional", "namespace", ns)
	}
	return ns
}

// isSystemNamespace returns true for Kubernetes system namespaces where
// creating user resources is typically unintended.
func isSystemNamespace(ns string) bool {
	switch ns {
	case "kube-system", "kube-public", "kube-node-lease":
		return true
	default:
		return false
	}
}

// resolveMPIAPIVersion detects the cluster's MPIJob storage version.
// Returns the storage version if supported, or an error.
func resolveMPIAPIVersion(ctx context.Context, k8sClient *client.Client) (string, error) {
	if err := k8sClient.ResolveMPIVersion(ctx); err != nil {
		return "", err
	}

	if !isMPIVersionSupportedByProvider(k8sClient.GetMPIVersion()) {
		return "", fmt.Errorf("cluster MPIJob storage version is %s, arena supports: %s",
			k8sClient.GetMPIVersion(), strings.Join(provider.MPISupportedVersions(), ", "))
	}

	return k8sClient.GetMPIVersion(), nil
}

// isMPIVersionSupportedByProvider checks if a version is in the provider's supported set.
func isMPIVersionSupportedByProvider(version string) bool {
	for _, v := range provider.MPISupportedVersions() {
		if v == version {
			return true
		}
	}
	return false
}

// submitCRD handles the common CRD submission flow shared by `submit` and `run`:
// provider lookup, MPI version detection, CRD build, namespace resolution,
// dry-run, existence check, RBAC pre-creation, CRD submit, and auxiliary
// resource finalisation (ConfigMap + ownerRef patching) with rollback on failure.
// All human- and machine-readable output goes to out.
func submitCRD(ctx context.Context, out io.Writer, k8sClient *client.Client, t *task.Task, frameworkLabel string, dryRun bool) error {
	p, err := getProvider(t.Framework.Name)
	if err != nil {
		return err
	}

	if a := t.Scheduling.Affinity; a != nil && (a.Policy == "" || a.Policy == "none") && len(a.Rules) > 0 {
		log.Warning("affinity.policy is unset or 'none'; affinity rules are ignored and no affinity is generated",
			"rules", len(a.Rules))
	}

	if isMPIFamily(t.Framework.Name) {
		if mpiP, ok := p.(*provider.MPIProvider); ok {
			if k8sClient != nil {
				version, err := resolveMPIAPIVersion(ctx, k8sClient)
				if err != nil {
					return err
				}
				mpiP.APIVersion = version
			} else {
				mpiP.APIVersion = provider.MPIAPIVersionV1
			}
		}
	}

	log.Debug("building CRD", "framework", t.Framework.Name)
	crd, err := p.BuildCRD(t)
	if err != nil {
		return fmt.Errorf("failed to build CRD: %w", err)
	}

	ns := resolveNS(t.Namespace)
	log.Debug("resolved target namespace", "namespace", ns)
	crd.SetNamespace(ns)
	t.Namespace = ns
	setFrameworkLabel(crd, frameworkLabel)

	log.Debug("CRD built", "kind", crd.GetKind(), "name", crd.GetName(), "namespace", ns)

	if dryRun {
		return printDryRun(out, crd, t)
	}

	log.Debug("checking if job exists", "name", t.Name, "namespace", ns)
	if err := checkJobExists(ctx, k8sClient, ns, t.Name); err != nil {
		return err
	}

	// Pre-create RBAC resources before CRD so they exist when the
	// operator reconciles the CRD (eliminates pod startup race).
	log.Debug("pre-creating RBAC resources", "name", t.Name)
	rbacResources, err := preCreateRBAC(ctx, t, k8sClient, p)
	if err != nil {
		return fmt.Errorf("failed to pre-create RBAC resources: %w", err)
	}

	log.Debug("submitting job", "kind", crd.GetKind(), "name", crd.GetName(), "namespace", ns)
	if err := k8sClient.Create(ctx, crd); err != nil {
		rollback(ctx, k8sClient, rbacResources)
		return fmt.Errorf("failed to submit job: %w", err)
	}

	log.Debug("finalizing auxiliary resources", "name", crd.GetName())
	if err := finalizeJobResources(ctx, crd, t, k8sClient, p, rbacResources); err != nil {
		log.Debug("rolling back after auxiliary resource failure", "kind", crd.GetKind(), "name", crd.GetName())
		rollback(ctx, k8sClient, rbacResources)
		if delErr := k8sClient.Delete(ctx, crd.GetKind(), ns, t.Name); delErr != nil {
			log.Warning("failed to clean up CRD after partial failure",
				"kind", crd.GetKind(), "name", crd.GetName(), "error", delErr.Error())
		}
		return fmt.Errorf("failed to create auxiliary resources: %w", err)
	}

	fmt.Fprintf(out, "Job %s submitted successfully\n", t.Name)
	return nil
}
