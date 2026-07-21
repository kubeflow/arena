package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/kubeflow/arena/pkg/client"
	"github.com/kubeflow/arena/pkg/constants"
	"github.com/kubeflow/arena/pkg/log"
	"github.com/kubeflow/arena/pkg/provider"
	"github.com/kubeflow/arena/pkg/task"
)

var (
	runFile     string
	runDryRun   bool
	runSetExprs []string
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Submit a training job from YAML file",
	Long: `Submit a training job to Kubernetes by loading a YAML specification file.
Use --set to override YAML fields with Helm-style dot-notation paths.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if runFile == "" {
			return errors.New("--file is required")
		}

		log.Debug("loading task from file", "file", runFile)

		// Read raw YAML
		yamlData, err := os.ReadFile(runFile)
		if err != nil {
			return fmt.Errorf("failed to read file %q: %w", runFile, err)
		}

		// Apply --set overrides on raw YAML before parsing
		mergedData, err := task.ApplySetOverrides(yamlData, runSetExprs)
		if err != nil {
			return fmt.Errorf("failed to apply --set overrides: %w", err)
		}

		// Parse merged YAML into Task
		t, err := task.LoadFromBytes(mergedData)
		if err != nil {
			return fmt.Errorf("failed to load task: %w", err)
		}

		log.Debug("task loaded", "name", t.Name, "framework", t.Framework.Name)

		var k8sClient *client.Client
		if !runDryRun {
			k8sClient, err = client.NewClient(kubeconfig, kubeContext)
			if err != nil {
				return fmt.Errorf("failed to create K8s client: %w", err)
			}
		}
		return submitCRD(cmdContext(cmd), k8sClient, t, t.Framework.Name, runDryRun)
	},
}

// getProvider returns the appropriate provider based on the framework name.
func getProvider(frameworkName string) (provider.Provider, error) {
	switch frameworkName {
	case constants.FrameworkPyTorch:
		return &provider.PyTorchProvider{}, nil
	case constants.FrameworkTensorFlow:
		return &provider.TensorFlowProvider{}, nil
	case constants.FrameworkMPI, constants.FrameworkHorovod, constants.FrameworkDeepSpeed:
		return &provider.MPIProvider{}, nil
	case constants.FrameworkRay:
		return nil, errors.New("ray provider is not yet implemented")
	default:
		return nil, fmt.Errorf("unsupported framework: %q", frameworkName)
	}
}

func init() {
	runCmd.Flags().StringVarP(&runFile, "file", "f", "", "path to YAML file")
	runCmd.Flags().BoolVar(&runDryRun, "dry-run", false, "print CRD as JSON without submitting")
	runCmd.Flags().StringArrayVar(&runSetExprs, "set", nil,
		"override YAML field (Helm-style: key=value, repeatable)")

	jobCmd.AddCommand(runCmd)
}
