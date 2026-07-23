package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/kubeflow/arena/pkg/client"
	"github.com/kubeflow/arena/pkg/constants"
	"github.com/kubeflow/arena/pkg/task"
)

var (
	submitName               string
	submitImage              string
	submitWorkers            int
	submitGPUs               int
	submitCPUs               string
	submitMem                string
	submitEnvs               []string
	submitData               []string
	submitLabels             []string
	submitAnnotations        []string
	submitSelectors          []string
	submitTolerations        []string
	submitPriority           int
	submitPriorityClass      string
	submitGang               bool
	submitSchedulerName      string
	submitCleanPodPolicy     string
	submitActiveDeadline     string
	submitTTLAfterFinished   string
	submitBackoffLimit       int
	submitImagePullPolicy    string
	submitImagePullSecret    []string
	submitServiceAccount     string
	submitRestart            string
	submitHostNetwork        bool
	submitHostIPC            bool
	submitHostPID            bool
	submitWorkingDir         string
	submitShell              string
	submitSHM                string
	submitDevice             []string
	submitGPUType            string
	submitTensorBoard        bool
	submitTBLogDir           string
	submitTBImage            string
	submitNprocPerNode       string
	submitPSCount            int
	submitChief              bool
	submitEvaluator          bool
	submitSlotsPerWorker     int
	submitGPUTopology        bool
	submitMountsOnLauncher   bool
	submitAffinityPolicy     string
	submitAffinityConstraint string
	submitAffinityTarget     string
	submitSuccessPolicy      string
	submitDryRun             bool
	submitQueue              string
	submitDataDir            []string
	submitConfigFile         []string
)

var submitCmd = &cobra.Command{
	Use:   "submit <type>",
	Short: "Submit a training job",
	Long: `Submit a training job using CLI flags.
Supported types: pytorch, tensorflow, mpi, horovod, deepspeed (case-insensitive).
Trailing arguments after -- are used as the run command.`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		framework := normalizeFramework(args[0])
		if framework == "" {
			return fmt.Errorf("unsupported framework type: %q (must be pytorch, tensorflow, mpi, horovod, deepspeed, or ray)",
				args[0])
		}

		// Trailing args after -- become the run command
		trailingArgs := []string{}
		if len(args) > 1 {
			trailingArgs = args[1:]
		}

		// Build the task with PyTorch N-1 conversion
		t := buildSubmitTask(framework, trailingArgs)

		// Build overrides map
		flags := buildSubmitFlags()

		// Apply overrides
		if err := task.ApplyOverrides(t, flags); err != nil {
			return fmt.Errorf("failed to apply overrides: %w", err)
		}

		// Validate
		t.SetDefaults()
		if err := task.Validate(t); err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}

		// Wire TensorBoard configuration from CLI flags
		if submitTensorBoard {
			if t.Logging.TensorBoard == nil {
				t.Logging.TensorBoard = &task.TensorBoardConfig{}
			}
			t.Logging.TensorBoard.Enabled = true
			if submitTBLogDir != "" {
				t.Logging.TensorBoard.LogDir = submitTBLogDir
			}
			if submitTBImage != "" {
				t.Logging.TensorBoard.Image = submitTBImage
			}
		}

		var (
			k8sClient *client.Client
			err       error
		)
		if !submitDryRun {
			k8sClient, err = client.NewClient(kubeconfig, kubeContext)
			if err != nil {
				return fmt.Errorf("failed to create K8s client: %w", err)
			}
		}
		return submitCRD(cmdContext(cmd), k8sClient, t, originalFramework(args[0]), submitDryRun)
	},
}

// buildSubmitTask constructs a Task from submit CLI flags with framework-specific conversions.
// PyTorch: --workers N means N total processes (1 master + N-1 workers), so worker.replicas = N-1.
// Other frameworks: --workers N means N workers, so worker.replicas = N.
func buildSubmitTask(framework string, trailingArgs []string) *task.Task {
	t := &task.Task{
		Name:      submitName,
		Image:     submitImage,
		Framework: task.Framework{Name: framework},
		Worker:    &task.Worker{Replicas: submitWorkers},
	}
	if t.Worker.Replicas < 1 {
		t.Worker.Replicas = 1
	}

	// v1 PyTorch compat: --workers N means N total (1 master + N-1 workers)
	// Conversion happens here in submit path (not in launch path)
	if framework == constants.FrameworkPyTorch {
		if submitWorkers <= 1 {
			// --workers=1 means master-only (no worker block)
			t.Worker = nil
			t.Master = &task.RoleConfig{}
		} else {
			t.Worker.Replicas = submitWorkers - 1
		}
	}

	// Set run from trailing args
	if len(trailingArgs) > 0 {
		t.Run = strings.Join(trailingArgs, " ")
	}

	// Map --chief/--evaluator/--ps-count to role sections
	if submitChief {
		t.Chief = &task.RoleConfig{} // nil resources → inherit from worker
	}
	if submitEvaluator {
		t.Evaluator = &task.RoleConfig{}
	}
	if submitPSCount > 0 {
		t.PS = &task.RoleConfig{Replicas: submitPSCount}
	}

	return t
}

// buildSubmitFlags builds the overrides map from all submit flag values.
func buildSubmitFlags() map[string]interface{} {
	flags := make(map[string]interface{})

	// Identity
	if submitName != "" {
		flags["name"] = submitName
	}

	// Resources
	if submitGPUs > 0 {
		flags["gpus"] = submitGPUs
	}
	if submitCPUs != "" {
		flags["cpus"] = submitCPUs
	}
	if submitMem != "" {
		flags["mem"] = submitMem
	}

	// Environment
	if len(submitEnvs) > 0 {
		flags["env"] = submitEnvs
	}

	// Data
	if len(submitData) > 0 {
		flags["data"] = submitData
	}
	if len(submitDataDir) > 0 {
		flags["data-dir"] = submitDataDir
	}
	if len(submitConfigFile) > 0 {
		flags["config-file"] = submitConfigFile
	}

	// Labels
	if len(submitLabels) > 0 {
		flags["label"] = submitLabels
	}

	// Annotations
	if len(submitAnnotations) > 0 {
		flags["annotation"] = submitAnnotations
	}

	// Scheduling
	if len(submitSelectors) > 0 {
		flags["selector"] = submitSelectors
	}
	if len(submitTolerations) > 0 {
		flags["toleration"] = submitTolerations
	}
	if submitPriority > 0 {
		flags["priority"] = submitPriority
	}
	if submitPriorityClass != "" {
		flags["priority-class-name"] = submitPriorityClass
	}
	if submitGang {
		flags["gang"] = submitGang
	}
	if submitSchedulerName != "" {
		flags["scheduler-name"] = submitSchedulerName
	}
	if submitAffinityPolicy != "" {
		flags["affinity-policy"] = submitAffinityPolicy
	}
	if submitAffinityConstraint != "" {
		flags["affinity-constraint"] = submitAffinityConstraint
	}
	if submitAffinityTarget != "" {
		flags["affinity-target"] = submitAffinityTarget
	}
	if submitQueue != "" {
		flags["queue"] = submitQueue
	}

	// Lifecycle
	if submitCleanPodPolicy != "" {
		flags["clean-pod-policy"] = submitCleanPodPolicy
	}
	if submitActiveDeadline != "" {
		flags["active-deadline"] = submitActiveDeadline
	}
	if submitTTLAfterFinished != "" {
		flags["ttl-after-finished"] = submitTTLAfterFinished
	}
	if submitBackoffLimit > 0 {
		flags["backoff-limit"] = submitBackoffLimit
	}
	if submitSuccessPolicy != "" {
		flags["success-policy"] = submitSuccessPolicy
	}

	// Runtime
	if submitImagePullPolicy != "" {
		flags["image-pull-policy"] = submitImagePullPolicy
	}
	if len(submitImagePullSecret) > 0 {
		flags["image-pull-secret"] = submitImagePullSecret
	}
	if submitServiceAccount != "" {
		flags["service-account"] = submitServiceAccount
	}
	if submitRestart != "" {
		flags["restart"] = submitRestart
	}
	if submitHostNetwork {
		flags["host-network"] = submitHostNetwork
	}
	if submitHostIPC {
		flags["host-ipc"] = submitHostIPC
	}
	if submitHostPID {
		flags["host-pid"] = submitHostPID
	}

	// Task
	if submitWorkingDir != "" {
		flags["working-dir"] = submitWorkingDir
	}
	if submitShell != "" {
		flags["shell"] = submitShell
	}
	if submitSHM != "" {
		flags["shm"] = submitSHM
	}
	if len(submitDevice) > 0 {
		flags["device"] = submitDevice
	}
	if submitGPUType != "" {
		flags["gpu-type"] = submitGPUType
	}

	// Logging
	if submitTensorBoard {
		flags["tensorboard"] = submitTensorBoard
	}
	if submitTBLogDir != "" {
		flags["tensorboard-logdir"] = submitTBLogDir
	}
	if submitTBImage != "" {
		flags["tensorboard-image"] = submitTBImage
	}

	// Framework-specific
	if submitNprocPerNode != "" {
		flags["nproc-per-node"] = submitNprocPerNode
	}
	if submitPSCount > 0 {
		flags["ps-count"] = submitPSCount
	}
	if submitChief {
		flags["chief"] = submitChief
	}
	if submitEvaluator {
		flags["evaluator"] = submitEvaluator
	}
	if submitSlotsPerWorker > 0 {
		flags["slots-per-worker"] = submitSlotsPerWorker
	}
	if submitGPUTopology {
		flags["gpu-topology"] = submitGPUTopology
		flags["host-network"] = true
		var labels []string
		if existing, ok := flags["label"].([]string); ok {
			labels = existing
		} else {
			labels = []string{}
		}
		flags["label"] = append(labels, "gpu-topology=true", "gpu-topology-replica=true")
	}
	if submitMountsOnLauncher {
		flags["mounts-on-launcher"] = submitMountsOnLauncher
	}

	return flags
}

func init() {
	// Required flags
	submitCmd.Flags().StringVar(&submitName, "name", "", "job name (required)")
	submitCmd.Flags().StringVar(&submitImage, "image", "", "container image (required)")

	// Worker configuration
	submitCmd.Flags().IntVar(&submitWorkers, "workers", 1, "number of worker replicas")

	// Resource flags
	submitCmd.Flags().IntVar(&submitGPUs, "gpus", 0, "number of GPUs per worker")
	submitCmd.Flags().StringVar(&submitCPUs, "cpus", "", "CPU request (e.g. 500m, 2)")
	submitCmd.Flags().StringVar(&submitMem, "mem", "", "memory request (e.g. 1Gi, 512Mi)")

	// Environment and data
	submitCmd.Flags().StringSliceVarP(&submitEnvs, "env", "e", nil, "environment variable (key=value, repeatable)")
	submitCmd.Flags().StringSliceVarP(&submitData, "data", "d", nil, "data volume (name:path:pvc, repeatable)")
	submitCmd.Flags().StringSliceVar(&submitDataDir, "data-dir", nil, "host path volume (name:path:hostpath, repeatable)")
	submitCmd.Flags().StringSliceVar(&submitConfigFile, "config-file", nil, "configmap volume (name:path:configmap, repeatable)")
	submitCmd.Flags().StringSliceVarP(&submitLabels, "label", "l", nil, "label (key=value, repeatable)")
	submitCmd.Flags().StringSliceVarP(&submitAnnotations, "annotation", "a", nil, "annotation (key=value, repeatable)")

	// Scheduling
	submitCmd.Flags().StringSliceVar(&submitSelectors, "selector", nil, "node selector (key=value, repeatable)")
	submitCmd.Flags().StringSliceVar(&submitTolerations, "toleration", nil, "toleration (key=value:effect, repeatable)")
	submitCmd.Flags().IntVar(&submitPriority, "priority", 0, "pod priority value")
	submitCmd.Flags().StringVar(&submitPriorityClass, "priority-class-name", "", "priority class name")
	submitCmd.Flags().BoolVar(&submitGang, "gang", false, "enable gang scheduling")
	submitCmd.Flags().StringVar(&submitSchedulerName, "scheduler-name", "", "custom scheduler name")
	submitCmd.Flags().StringVar(&submitAffinityPolicy, "affinity-policy", "", "affinity policy")
	submitCmd.Flags().StringVar(&submitAffinityConstraint, "affinity-constraint", "", "affinity constraint")
	submitCmd.Flags().StringVar(&submitAffinityTarget, "affinity-target", "", "affinity target (pod or node)")
	submitCmd.Flags().StringVar(&submitQueue, "queue", "", "scheduling queue name")

	// Lifecycle
	submitCmd.Flags().StringVar(&submitCleanPodPolicy, "clean-pod-policy", "", "clean pod policy (None, Running, All)")
	submitCmd.Flags().StringVar(&submitActiveDeadline, "active-deadline", "", "active deadline (e.g. 2h, 7d)")
	submitCmd.Flags().StringVar(&submitTTLAfterFinished, "ttl-after-finished", "", "TTL after finished (e.g. 7d)")
	submitCmd.Flags().IntVar(&submitBackoffLimit, "backoff-limit", 0, "backoff limit for retries")
	submitCmd.Flags().StringVar(&submitSuccessPolicy, "success-policy", "", "success policy (ChiefWorker, AllWorkers, TF only). ChiefWorker is an alias for the default \"\"")

	// Runtime
	submitCmd.Flags().StringVar(&submitImagePullPolicy, "image-pull-policy", "", "image pull policy (Always, IfNotPresent, Never)")
	submitCmd.Flags().StringSliceVar(&submitImagePullSecret, "image-pull-secret", nil, "image pull secret name (repeatable)")
	submitCmd.Flags().StringVar(&submitServiceAccount, "service-account", "", "service account name")
	submitCmd.Flags().StringVar(&submitRestart, "restart", "", "restart policy (Always, OnFailure, Never)")
	submitCmd.Flags().BoolVar(&submitHostNetwork, "host-network", false, "use host network")
	submitCmd.Flags().BoolVar(&submitHostIPC, "host-ipc", false, "use host IPC namespace")
	submitCmd.Flags().BoolVar(&submitHostPID, "host-pid", false, "use host PID namespace")

	// Task
	submitCmd.Flags().StringVar(&submitWorkingDir, "working-dir", "", "working directory in container")
	submitCmd.Flags().StringVar(&submitShell, "shell", "", "shell to use (default /bin/sh)")
	submitCmd.Flags().StringVar(&submitSHM, "shm", "", "shared memory size (e.g. 8Gi)")
	submitCmd.Flags().StringSliceVar(&submitDevice, "device", nil, "extended resource (name=count, repeatable)")
	submitCmd.Flags().StringVar(&submitGPUType, "gpu-type", "", "GPU type (sets node selector nvidia.com/gpu.product)")

	// Logging / TensorBoard
	submitCmd.Flags().BoolVar(&submitTensorBoard, "tensorboard", false, "enable TensorBoard sidecar (TensorBoard has no built-in authentication)")
	submitCmd.Flags().StringVar(&submitTBLogDir, "tensorboard-logdir", "", "TensorBoard log directory")
	submitCmd.Flags().StringVar(&submitTBImage, "tensorboard-image", "", "TensorBoard container image")

	// Framework-specific: PyTorch
	submitCmd.Flags().StringVar(&submitNprocPerNode, "nproc-per-node", "", "PyTorch: processes per node (auto, gpu, cpu, or int)")

	// Framework-specific: TensorFlow
	submitCmd.Flags().IntVar(&submitPSCount, "ps-count", 0, "TensorFlow: number of parameter servers")
	submitCmd.Flags().BoolVar(&submitChief, "chief", false, "TensorFlow: enable Chief worker")
	submitCmd.Flags().BoolVar(&submitEvaluator, "evaluator", false, "TensorFlow: enable Evaluator worker")

	// Framework-specific: MPI
	submitCmd.Flags().IntVar(&submitSlotsPerWorker, "slots-per-worker", 0, "MPI: slots per worker")
	submitCmd.Flags().BoolVar(&submitGPUTopology, "gpu-topology", false, "MPI: enable GPU topology (sets host networking, gpu-topology/gpu-topology-replica labels, and MPI annotation)")
	submitCmd.Flags().BoolVar(&submitMountsOnLauncher, "mounts-on-launcher", false, "MPI: mount volumes on launcher")

	// Dry-run
	submitCmd.Flags().BoolVar(&submitDryRun, "dry-run", false, "print CRD as JSON without submitting")

	_ = submitCmd.MarkFlagRequired("name")
	_ = submitCmd.MarkFlagRequired("image")

	rootCmd.AddCommand(submitCmd)
}
