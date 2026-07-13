package cli

import (
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
	RunE: func(cmd *cobra.Command, args []string) error {
		if runFile == "" {
			return fmt.Errorf("--file is required")
		}

		log.Debug("loading task from file", "file", runFile)

		// Read raw YAML
		yamlData, err := os.ReadFile(runFile)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", runFile, err)
		}

		// Apply --set overrides on raw YAML before parsing
		mergedData, err := task.ApplySetOverrides(yamlData, runSetExprs)
		if err != nil {
			return fmt.Errorf("failed to apply --set overrides: %w", err)
		}

		// Parse merged YAML into Task
		t, err := task.LoadFromBytes(mergedData)
		if err != nil {
			log.Error(err, "failed to load task", "file", runFile)
			return fmt.Errorf("failed to load task: %w", err)
		}

		log.Debug("task loaded", "name", t.Name, "framework", t.Framework.Name)

		// Get provider
		p, err := getProvider(t.Framework.Name)
		if err != nil {
			log.Error(err, "failed to get provider", "framework", t.Framework.Name)
			return err
		}

		// Create K8s client early for version detection
		k8sClient, err := client.NewClient(kubeconfig, kubeContext)
		if err != nil {
			log.Error(err, "failed to create Kubernetes client")
			return fmt.Errorf("failed to create Kubernetes client: %w", err)
		}

		// For MPI-family frameworks, detect cluster version
		if isMPIFamily(t.Framework.Name) {
			if mpiP, ok := p.(*provider.MPIProvider); ok {
				version, err := resolveMPIAPIVersion(cmdContext(cmd), k8sClient)
				if err != nil {
					return err
				}
				mpiP.APIVersion = version
			}
		}

		log.Debug("building CRD", "framework", t.Framework.Name)

		// Build CRD
		crd, err := p.BuildCRD(t)
		if err != nil {
			log.Error(err, "failed to build CRD", "framework", t.Framework.Name)
			return fmt.Errorf("failed to build CRD: %w", err)
		}

		// Set namespace from resolved value
		ns := resolveNS(t.Namespace)
		crd.SetNamespace(ns)

		setFrameworkLabel(crd, t.Framework.Name)

		log.Debug("CRD built", "kind", crd.GetKind(), "name", crd.GetName(), "namespace", ns)

		// Dry-run: print and exit
		if runDryRun {
			log.Info("dry-run mode, not submitting")
			return printDryRun(crd, t)
		}

		log.Debug("checking if job exists", "name", t.Name, "namespace", ns)

		// Check if job already exists
		if err := checkJobExists(cmdContext(cmd), k8sClient, ns, t.Name); err != nil {
			return err
		}

		log.Debug("submitting job", "kind", crd.GetKind(), "name", crd.GetName(), "namespace", ns)

		// Submit to cluster
		if err := k8sClient.Create(cmdContext(cmd), crd); err != nil {
			log.Error(err, "job creation failed", "kind", crd.GetKind(), "name", crd.GetName())
			return fmt.Errorf("failed to submit job: %w", err)
		}

		log.Debug("creating auxiliary resources", "name", crd.GetName())

		// Create auxiliary resources (ConfigMap anchor, TensorBoard)
		if err := createJobResources(cmdContext(cmd), crd, t, k8sClient); err != nil {
			log.Error(err, "failed to create auxiliary resources", "name", crd.GetName())
			return err
		}

		log.Info("job submitted successfully", "name", t.Name, "namespace", ns)
		fmt.Printf("Job %s submitted successfully\n", t.Name)
		return nil
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
		return nil, fmt.Errorf("ray provider is not yet implemented")
	default:
		return nil, fmt.Errorf("unsupported framework: %s", frameworkName)
	}
}

func init() {
	runCmd.Flags().StringVarP(&runFile, "file", "f", "", "path to YAML file")
	runCmd.Flags().BoolVar(&runDryRun, "dry-run", false, "print CRD as JSON without submitting")
	runCmd.Flags().StringArrayVar(&runSetExprs, "set", nil,
		"override YAML field (Helm-style: key=value, repeatable)")

	jobCmd.AddCommand(runCmd)
}
