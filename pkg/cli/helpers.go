package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/kubeflow/arena/pkg/client"
	"github.com/kubeflow/arena/pkg/constants"
	"github.com/kubeflow/arena/pkg/provider"
	"github.com/kubeflow/arena/pkg/task"
)

// printCRD marshals a CRD as indented JSON and prints it to stdout.
func printCRD(crd *unstructured.Unstructured) error {
	data, err := json.MarshalIndent(crd.Object, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal CRD: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

// printDryRun prints the CRD and all auxiliary resources that would be created
// during a real submission (TensorBoard Deployment and Service if enabled).
// Each resource is printed as indented JSON separated by "---" for readability.
func printDryRun(crd *unstructured.Unstructured, t *task.Task) error {
	if err := printCRD(crd); err != nil {
		return err
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

		deploy := buildTensorBoardDeployment(tbName, crd.GetName(), crd.GetNamespace(), tbImage, logDir, t, ownerRef)
		fmt.Println("---")
		if err := printCRD(deploy); err != nil {
			return err
		}

		svc := buildTensorBoardService(tbName, crd.GetName(), crd.GetNamespace(), ownerRef)
		fmt.Println("---")
		if err := printCRD(svc); err != nil {
			return err
		}
	}

	return nil
}

// resolveNS resolves the effective namespace using the 4-level priority chain:
// CLI -n flag > YAML namespace > kubeconfig context namespace > "default".
func resolveNS(yamlNamespace string) string {
	if namespace != "" {
		return namespace
	}
	if yamlNamespace != "" {
		return yamlNamespace
	}
	return client.ResolveNamespace(kubeconfig, kubeContext, "")
}

// isMPIFamily returns true if the framework uses the MPIJob CRD.
func isMPIFamily(framework string) bool { return IsMPIFamily(framework) }

// resolveMPIAPIVersion detects the cluster's MPIJob storage version.
// Returns the storage version if supported, or an error.
func resolveMPIAPIVersion(ctx context.Context, k8sClient *client.Client) (string, error) {
	if err := k8sClient.ResolveMPIVersion(ctx); err != nil {
		return "", err
	}

	if !isMPIVersionSupportedByProvider(k8sClient.MPIVersion) {
		return "", fmt.Errorf("cluster MPIJob storage version is %s, arena supports: %s",
			k8sClient.MPIVersion, strings.Join(provider.MPISupportedVersions(), ", "))
	}

	return k8sClient.MPIVersion, nil
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
