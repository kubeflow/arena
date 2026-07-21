package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/kubeflow/arena/pkg/client"
	"github.com/kubeflow/arena/pkg/constants"
	"github.com/kubeflow/arena/pkg/log"
	"github.com/kubeflow/arena/pkg/provider"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Verify that required CRDs are installed in the cluster",
	Long:  `Check whether the Kubeflow training operator CRDs (PyTorchJob, TFJob, MPIJob) are installed and accessible in the cluster.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		k8sClient, err := client.NewClient(kubeconfig, kubeContext)
		if err != nil {
			return fmt.Errorf("failed to create K8s client: %w", err)
		}

		mpiAvailable := true
		if err := k8sClient.ResolveMPIVersion(cmdContext(cmd)); err != nil {
			log.Debug("MPIJob CRD not available", "error", err.Error())
			mpiAvailable = false
		}

		kinds := []string{constants.KindPyTorchJob, constants.KindTFJob, constants.KindMPIJob}
		allOk := true

		for _, kind := range kinds {
			if kind == constants.KindMPIJob && !mpiAvailable {
				continue
			}
			crdName := crdObjectName(kind)
			if crdName == "" {
				continue
			}

			versions, err := k8sClient.GetCRDVersions(cmdContext(cmd), crdName)
			if err != nil || versions == nil {
				fmt.Printf("✗ %s: not installed\n", kind)
				allOk = false
				continue
			}

			expected, err := k8sClient.KindToAPIVersion(kind)
			if err != nil {
				fmt.Printf("? %s: version not resolved\n", kind)
				allOk = false
				continue
			}
			versionStr := formatCRDVersions(versions)
			fmt.Printf("✓ %s: installed (expected: %s)\n", kind, expected)
			fmt.Printf("  versions: %s\n", versionStr)

			// MPIJob compatibility check
			if kind == constants.KindMPIJob {
				storageVersion := client.FindStorageVersion(versions)
				supported := provider.MPISupportedVersions()
				if isMPIVersionSupportedByProvider(storageVersion) {
					fmt.Printf("  compatible: ✓ (storage version %s supported by arena)\n", storageVersion)
				} else {
					fmt.Printf("  compatible: ✗ (storage version %s, arena supports: %s)\n",
						storageVersion, strings.Join(supported, ", "))
					allOk = false
				}
			}
		}

		if !allOk {
			return errors.New("one or more CRDs are not installed or incompatible")
		}

		return nil
	},
}

// formatCRDVersions formats version info for display.
// Example: "v2beta1 (served, storage), v1 (served)"
func formatCRDVersions(versions []client.CRDVersionInfo) string {
	if len(versions) == 0 {
		return ""
	}
	parts := make([]string, 0, len(versions))
	for _, v := range versions {
		flags := []string{}
		if v.Served {
			flags = append(flags, "served")
		}
		if v.Storage {
			flags = append(flags, "storage")
		}
		if len(flags) > 0 {
			parts = append(parts, fmt.Sprintf("%s (%s)", v.Name, strings.Join(flags, ", ")))
		} else {
			parts = append(parts, v.Name)
		}
	}
	return strings.Join(parts, ", ")
}

// crdObjectName returns the CRD object name (plural.group) for a given kind.
func crdObjectName(kind string) string {
	group := constants.KubeflowGroup
	switch kind {
	case constants.KindPyTorchJob:
		return "pytorchjobs." + group
	case constants.KindTFJob:
		return "tfjobs." + group
	case constants.KindMPIJob:
		return "mpijobs." + group
	default:
		return ""
	}
}

func init() {
	rootCmd.AddCommand(checkCmd)
}
