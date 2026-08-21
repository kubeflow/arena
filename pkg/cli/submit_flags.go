package cli

// Flag registration and flag→overrides-map collection for the submit
// command. All flag semantics (how a flag value lands in the Task) live in
// pkg/task.ApplyOverrides; this file only declares the CLI surface.

import (
	"github.com/spf13/cobra"

	"github.com/kubeflow/arena/pkg/constants"
)

var (
	submitName                 string
	submitImage                string
	submitWorkers              int
	submitGPUs                 int
	submitCPU                  string
	submitMemory               string
	submitEnvs                 []string
	submitData                 []string
	submitLabels               []string
	submitAnnotations          []string
	submitSelectors            []string
	submitTolerations          []string
	submitPriorityClass        string
	submitGang                 bool
	submitScheduler            string
	submitCleanTaskPolicy      string
	submitRunningTimeout       string
	submitTTLAfterFinished     string
	submitJobBackoffLimit      int
	submitImagePullPolicy      string
	submitImagePullSecret      []string
	submitServiceAccount       string
	submitJobRestartPolicy     string
	submitHostNetwork          bool
	submitHostIPC              bool
	submitHostPID              bool
	submitWorkingDir           string
	submitShell                string
	submitShareMemory          string
	submitDevice               []string
	submitGPUType              string
	submitTensorBoard          bool
	submitLogDir               string
	submitTBImage              string
	submitNprocPerNode         string
	submitPS                   int
	submitChief                bool
	submitEvaluator            bool
	submitSlotsPerWorker       int
	submitGPUTopology          bool
	submitMountsOnLauncher     bool
	submitAffinityPolicy       string
	submitAffinityConstraint   string
	submitAffinityTarget       string
	submitSuccessPolicy        string
	submitDryRun               bool
	submitQueue                bool
	submitDataDir              []string
	submitConfigFile           []string
	submitSyncMode             string
	submitSyncSource           string
	submitSyncImage            string
	submitPSCPU                string
	submitPSMemory             string
	submitPSGPUs               int
	submitChiefCPU             string
	submitChiefMemory          string
	submitEvaluatorCPU         string
	submitEvaluatorMemory      string
	submitWorkerCPU            string
	submitWorkerMemory         string
	submitPSCPULimit           string
	submitPSMemoryLimit        string
	submitChiefCPULimit        string
	submitChiefMemoryLimit     string
	submitEvaluatorCPULimit    string
	submitEvaluatorMemoryLimit string
	submitWorkerCPULimit       string
	submitWorkerMemoryLimit    string
	submitLogLevel             string
	submitIgnoredString        string
	submitIgnoredBool          bool
)

// registerSubmitCommonFlags registers the framework-agnostic submit flags on
// cmd. Required marking of --name/--image is left to the caller: only the
// per-framework subcommands enforce them at cobra level, the parent's flag
// set is the v1-compat fallback parser and defers missing values to task
// validation.
func registerSubmitCommonFlags(cmd *cobra.Command) {
	f := cmd.Flags()

	// Required flags
	f.StringVar(&submitName, "name", "", "job name")
	f.StringVar(&submitImage, "image", "", "container image")

	// Worker configuration
	f.IntVar(&submitWorkers, "workers", 1, "number of worker replicas")

	// Resource flags
	f.IntVar(&submitGPUs, "gpus", 0, "number of GPUs per worker")
	f.StringVar(&submitCPU, "cpu", "", "CPU request (e.g. 500m, 2)")
	f.StringVar(&submitMemory, "memory", "", "memory request (e.g. 1Gi, 512Mi)")

	// Environment and data
	f.StringArrayVarP(&submitEnvs, "env", "e", nil, "environment variable (key=value, repeatable)")
	f.StringArrayVarP(&submitData, "data", "d", nil, "data volume (name:path:pvc, repeatable)")
	f.StringArrayVar(&submitDataDir, "data-dir", nil, "host path volume (name:path:hostpath, repeatable)")
	f.StringArrayVar(&submitConfigFile, "config-file", nil, "configmap volume (name:path:configmap, repeatable)")
	f.StringArrayVarP(&submitLabels, "label", "l", nil, "label (key=value, repeatable)")
	f.StringArrayVarP(&submitAnnotations, "annotation", "a", nil, "annotation (key=value, repeatable)")

	// Scheduling
	f.StringArrayVar(&submitSelectors, "selector", nil, "node selector (key=value, repeatable)")
	f.StringArrayVar(&submitTolerations, "toleration", nil, "toleration (key=value:effect, repeatable)")
	f.StringVarP(&submitPriorityClass, "priority", "p", "", "priority class name")
	f.BoolVar(&submitGang, "gang", false, "enable gang scheduling")
	f.StringVar(&submitScheduler, "scheduler", "", "custom scheduler name")
	f.StringVar(&submitAffinityPolicy, "affinity-policy", "", "affinity policy")
	f.StringVar(&submitAffinityConstraint, "affinity-constraint", "", "affinity constraint")
	f.StringVar(&submitAffinityTarget, "affinity-target", "", "affinity target (pod or node)")
	f.BoolVar(&submitQueue, "queue", false, "enables the feature to queue jobs after they are scheduled (Kube-queue needs to be pre-installed https://github.com/kube-queue/kube-queue)")

	// Lifecycle
	f.StringVar(&submitCleanTaskPolicy, "clean-task-policy", "Running", "clean pod policy (None, Running, All); pass \"\" to use the operator default")
	f.StringVar(&submitRunningTimeout, "running-timeout", "", "running timeout (e.g. 2h, 7d)")
	f.StringVar(&submitTTLAfterFinished, "ttl-after-finished", "", "TTL after finished (e.g. 7d)")
	f.IntVar(&submitJobBackoffLimit, "job-backoff-limit", 0, "max restart count for the job")
	f.IntVar(&submitJobBackoffLimit, "retry", 0, "times to retry the job (same as --job-backoff-limit)")

	// Runtime
	f.StringVar(&submitImagePullPolicy, "image-pull-policy", "", "image pull policy (Always, IfNotPresent, Never)")
	f.StringArrayVar(&submitImagePullSecret, "image-pull-secret", nil, "image pull secret name (repeatable)")
	f.StringVar(&submitServiceAccount, "service-account", "", "service account name")
	f.StringVar(&submitJobRestartPolicy, "job-restart-policy", "", "restart policy (Always, OnFailure, Never)")
	f.BoolVar(&submitHostNetwork, "hostNetwork", false, "use host network")
	f.BoolVar(&submitHostIPC, "hostIPC", false, "use host IPC namespace")
	f.BoolVar(&submitHostPID, "hostPID", false, "use host PID namespace")

	// Task
	f.StringVar(&submitWorkingDir, "working-dir", "", "working directory in container")
	f.StringVar(&submitShell, "shell", "", "shell to use (default /bin/sh)")
	f.StringVar(&submitShareMemory, "share-memory", "2Gi", "shared memory size (default 2Gi like v1; pass \"\" to disable)")
	f.StringArrayVar(&submitDevice, "device", nil, "extended resource (name=count, repeatable)")
	f.StringVar(&submitGPUType, "gpu-type", "", "GPU type (sets node selector nvidia.com/gpu.product)")

	// Logging / TensorBoard
	f.BoolVar(&submitTensorBoard, "tensorboard", false, "enable TensorBoard sidecar (TensorBoard has no built-in authentication)")
	f.StringVar(&submitLogDir, "logdir", "/training_logs", "TensorBoard log directory (default /training_logs like v1)")
	f.StringVar(&submitTBImage, "tensorboard-image", "", "TensorBoard container image")

	// Dry-run
	f.BoolVar(&submitDryRun, "dry-run", false, "print CRD as JSON (default) or YAML (-o yaml) without submitting")
	registerOutputFlag(cmd)

	_ = cmd.RegisterFlagCompletionFunc("clean-task-policy", completeStaticChoices(
		"None\tDo not clean pods", "Running\tClean running pods", "All\tClean all pods"))
	_ = cmd.RegisterFlagCompletionFunc("image-pull-policy", completeStaticChoices(
		"Always\tAlways pull", "IfNotPresent\tPull if not present", "Never\tNever pull"))
	_ = cmd.RegisterFlagCompletionFunc("job-restart-policy", completeStaticChoices(
		"Always\tAlways restart", "OnFailure\tRestart on failure", "Never\tNever restart"))
}

// registerPyTorchSubmitFlags registers the PyTorch-only submit flags on cmd.
func registerPyTorchSubmitFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&submitNprocPerNode, "nproc-per-node", "", "PyTorch: processes per node (auto, gpu, cpu, or int)")
	_ = cmd.RegisterFlagCompletionFunc("nproc-per-node", completeStaticChoices(
		"auto\tAuto-detect", "gpu\tOne per GPU", "cpu\tOne per CPU"))
}

// registerTFSubmitFlags registers the TensorFlow-only submit flags on cmd:
// role toggles, success policy, and the v1 per-role resource flags.
func registerTFSubmitFlags(cmd *cobra.Command) {
	f := cmd.Flags()

	f.IntVar(&submitPS, "ps", 0, "TensorFlow: number of parameter servers")
	f.BoolVar(&submitChief, "chief", false, "TensorFlow: enable Chief worker")
	f.BoolVar(&submitEvaluator, "evaluator", false, "TensorFlow: enable Evaluator worker")
	f.StringVar(&submitSuccessPolicy, "success-policy", "", "success policy (ChiefWorker, AllWorkers, TF only). ChiefWorker is an alias for the default \"\"")
	_ = cmd.RegisterFlagCompletionFunc("success-policy", completeStaticChoices(
		"ChiefWorker\tChief and worker succeed", "AllWorkers\tAll workers succeed"))

	f.StringVar(&submitPSCPU, "ps-cpu", "", "TensorFlow: CPU for parameter servers (e.g. 500m, 2)")
	f.StringVar(&submitPSMemory, "ps-memory", "", "TensorFlow: memory for parameter servers (e.g. 1Gi)")
	f.IntVar(&submitPSGPUs, "ps-gpus", 0, "TensorFlow: GPUs per parameter server")
	f.StringVar(&submitChiefCPU, "chief-cpu", "", "TensorFlow: CPU for the Chief role (e.g. 500m, 2)")
	f.StringVar(&submitChiefMemory, "chief-memory", "", "TensorFlow: memory for the Chief role (e.g. 1Gi)")
	f.StringVar(&submitEvaluatorCPU, "evaluator-cpu", "", "TensorFlow: CPU for the Evaluator role (e.g. 500m, 2)")
	f.StringVar(&submitEvaluatorMemory, "evaluator-memory", "", "TensorFlow: memory for the Evaluator role (e.g. 1Gi)")

	// v1 TFJob -limit variants. v2 defaults to Guaranteed QoS (one value for
	// requests and limits); an explicit -limit value overrides only the limit,
	// reproducing v1's independent request/limit flags.
	f.StringVar(&submitPSCPULimit, "ps-cpu-limit", "", "TensorFlow: CPU limit for parameter servers (requests stay at --ps-cpu)")
	f.StringVar(&submitPSMemoryLimit, "ps-memory-limit", "", "TensorFlow: memory limit for parameter servers (requests stay at --ps-memory)")
	f.StringVar(&submitChiefCPULimit, "chief-cpu-limit", "", "TensorFlow: CPU limit for the Chief role (requests stay at --chief-cpu)")
	f.StringVar(&submitChiefMemoryLimit, "chief-memory-limit", "", "TensorFlow: memory limit for the Chief role (requests stay at --chief-memory)")
	f.StringVar(&submitEvaluatorCPULimit, "evaluator-cpu-limit", "", "TensorFlow: CPU limit for the Evaluator role (requests stay at --evaluator-cpu)")
	f.StringVar(&submitEvaluatorMemoryLimit, "evaluator-memory-limit", "", "TensorFlow: memory limit for the Evaluator role (requests stay at --evaluator-memory)")
}

// registerMPISubmitFlags registers the MPI-family-only submit flags on cmd
// (mpi, horovod and deepspeed all render the MPIJob CRD).
func registerMPISubmitFlags(cmd *cobra.Command) {
	f := cmd.Flags()
	f.IntVar(&submitSlotsPerWorker, "slots-per-worker", 0, "MPI: slots per worker")
	f.BoolVar(&submitGPUTopology, "gputopology", false, "MPI: enable GPU topology (sets host networking, gpu-topology/gpu-topology-replica labels, and MPI annotation)")
	f.BoolVar(&submitMountsOnLauncher, "mounts-on-launcher", false, "MPI: mount volumes on launcher")
}

// registerSubmitFrameworkFlags registers the framework-specific submit flags
// on cmd: PyTorch flags for pytorch, TensorFlow flags for tensorflow, and
// the MPI-family flags for mpi/horovod/deepspeed (all three render MPIJob).
func registerSubmitFrameworkFlags(cmd *cobra.Command, framework string) {
	switch {
	case framework == constants.FrameworkPyTorch:
		registerPyTorchSubmitFlags(cmd)
	case framework == constants.FrameworkTensorFlow:
		registerTFSubmitFlags(cmd)
	case isMPIFamily(framework):
		registerMPISubmitFlags(cmd)
	}
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
	if submitCPU != "" {
		flags["cpu"] = submitCPU
	}
	if submitMemory != "" {
		flags["memory"] = submitMemory
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
	if submitPriorityClass != "" {
		flags["priority"] = submitPriorityClass
	}
	if submitGang {
		flags["gang"] = submitGang
	}
	if submitScheduler != "" {
		flags["scheduler"] = submitScheduler
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
	if submitQueue {
		flags["queue"] = submitQueue
	}

	// Lifecycle
	if submitCleanTaskPolicy != "" {
		flags["clean-task-policy"] = submitCleanTaskPolicy
	}
	if submitRunningTimeout != "" {
		flags["running-timeout"] = submitRunningTimeout
	}
	if submitTTLAfterFinished != "" {
		flags["ttl-after-finished"] = submitTTLAfterFinished
	}
	if submitJobBackoffLimit > 0 {
		flags["job-backoff-limit"] = submitJobBackoffLimit
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
	if submitJobRestartPolicy != "" {
		flags["job-restart-policy"] = submitJobRestartPolicy
	}
	if submitHostNetwork {
		flags["hostNetwork"] = submitHostNetwork
	}
	if submitHostIPC {
		flags["hostIPC"] = submitHostIPC
	}
	if submitHostPID {
		flags["hostPID"] = submitHostPID
	}

	// Task
	if submitWorkingDir != "" {
		flags["working-dir"] = submitWorkingDir
	}
	if submitShell != "" {
		flags["shell"] = submitShell
	}
	if submitShareMemory != "" {
		flags["share-memory"] = submitShareMemory
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
	if submitLogDir != "" {
		flags["logdir"] = submitLogDir
	}
	if submitTBImage != "" {
		flags["tensorboard-image"] = submitTBImage
	}

	// v1 code sync (applied by task.ApplyOverrides)
	if submitSyncMode != "" {
		flags["sync-mode"] = submitSyncMode
	}
	if submitSyncSource != "" {
		flags["sync-source"] = submitSyncSource
	}
	if submitSyncImage != "" {
		flags["sync-image"] = submitSyncImage
	}

	// Framework-specific
	if submitNprocPerNode != "" {
		flags["nproc-per-node"] = submitNprocPerNode
	}
	if submitPS > 0 {
		flags["ps"] = submitPS
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
		flags["gputopology"] = submitGPUTopology
	}
	if submitMountsOnLauncher {
		flags["mounts-on-launcher"] = submitMountsOnLauncher
	}

	// v1 per-role resources (applied by task.ApplyOverrides after the roles
	// are created, so they override the generic --cpu/--memory)
	if submitPSCPU != "" {
		flags["ps-cpu"] = submitPSCPU
	}
	if submitPSMemory != "" {
		flags["ps-memory"] = submitPSMemory
	}
	if submitPSGPUs > 0 {
		flags["ps-gpus"] = submitPSGPUs
	}
	if submitChiefCPU != "" {
		flags["chief-cpu"] = submitChiefCPU
	}
	if submitChiefMemory != "" {
		flags["chief-memory"] = submitChiefMemory
	}
	if submitEvaluatorCPU != "" {
		flags["evaluator-cpu"] = submitEvaluatorCPU
	}
	if submitEvaluatorMemory != "" {
		flags["evaluator-memory"] = submitEvaluatorMemory
	}
	if submitWorkerCPU != "" {
		flags["worker-cpu"] = submitWorkerCPU
	}
	if submitWorkerMemory != "" {
		flags["worker-memory"] = submitWorkerMemory
	}
	if submitPSCPULimit != "" {
		flags["ps-cpu-limit"] = submitPSCPULimit
	}
	if submitPSMemoryLimit != "" {
		flags["ps-memory-limit"] = submitPSMemoryLimit
	}
	if submitChiefCPULimit != "" {
		flags["chief-cpu-limit"] = submitChiefCPULimit
	}
	if submitChiefMemoryLimit != "" {
		flags["chief-memory-limit"] = submitChiefMemoryLimit
	}
	if submitEvaluatorCPULimit != "" {
		flags["evaluator-cpu-limit"] = submitEvaluatorCPULimit
	}
	if submitEvaluatorMemoryLimit != "" {
		flags["evaluator-memory-limit"] = submitEvaluatorMemoryLimit
	}
	if submitWorkerCPULimit != "" {
		flags["worker-cpu-limit"] = submitWorkerCPULimit
	}
	if submitWorkerMemoryLimit != "" {
		flags["worker-memory-limit"] = submitWorkerMemoryLimit
	}

	return flags
}

// registerSubmitCompatFlags registers the v1 flags that need custom wiring:
// sync, per-role worker resources, and the deprecated no-op compat flags.
func registerSubmitCompatFlags(cmd *cobra.Command) {
	f := cmd.Flags()

	// v1 -limit variants for the worker role (the TF per-role -limit variants
	// live in registerTFSubmitFlags).
	f.StringVar(&submitWorkerCPULimit, "worker-cpu-limit", "", "CPU limit for workers (requests stay at --worker-cpu)")
	f.StringVar(&submitWorkerMemoryLimit, "worker-memory-limit", "", "memory limit for workers (requests stay at --worker-memory)")

	// Sync code (v1 --sync-mode/--sync-source/--sync-image).
	f.StringVar(&submitSyncMode, "sync-mode", "", "sync mode: git, rsync, or hdfs")
	f.StringVar(&submitSyncSource, "sync-source", "", "sync source (repo URL for git, host::path for rsync, URI for hdfs)")
	f.StringVar(&submitSyncImage, "sync-image", "", "container image used by the sync init container")

	// Per-role worker resources (v1 flags mapped to v2 role blocks).
	f.StringVar(&submitWorkerCPU, "worker-cpu", "", "CPU for workers, overrides the generic resource flags (e.g. 500m, 2)")
	f.StringVar(&submitWorkerMemory, "worker-memory", "", "memory for workers, overrides the generic resource flags (e.g. 1Gi)")

	// --config: v1 name for the kubeconfig path; v2's global --kubeconfig is
	// the equivalent, so this binds the same variable and still works.
	f.StringVar(&kubeconfig, "config", "", "path to kubeconfig file (v1 compat)")
	_ = f.MarkDeprecated("config", "please use --kubeconfig instead")

	// --loglevel: v1 log level mapped onto v2 verbosity (debug->2, info->1,
	// warn/error->0) in runSubmit via applyV1LogLevel.
	f.StringVar(&submitLogLevel, "loglevel", "", "log level: debug, info, warn, or error (v1 compat)")
	_ = f.MarkDeprecated("loglevel", "please use --verbose instead")

	f.BoolVar(&submitIgnoredBool, "rdma", false, "enable RDMA (v1 compat)")
	_ = f.MarkDeprecated("rdma", "arena-v2 requests RDMA hardware via --device <resource>=<count>, e.g. --device rdma/hca=1")

	// TODO(role-config): bind these per-role flags to real RoleConfig fields
	// once it grows selector/annotation/image/port support. They are accepted
	// so v1 scripts do not hard-fail, and warn that they are ignored.
	for _, name := range []string{
		"ps-selector", "worker-selector", "chief-selector", "evaluator-selector", "launcher-selector",
		"worker-annotation", "launcher-annotation",
		"ps-image", "worker-image",
		"ps-port", "worker-port", "chief-port", "ssh-port",
		"ps-affinity-policy", "ps-affinity-constraint",
		"worker-affinity-policy", "worker-affinity-constraint",
	} {
		f.StringVar(&submitIgnoredString, name, "", "per-role setting (v1 compat)")
		_ = f.MarkDeprecated(name, "not compatible with arena-v2 yet (requires per-role RoleConfig extension); ignored")
	}

	// v1 flags with no v2 equivalent and no behavioral effect.
	f.BoolVar(&submitIgnoredBool, "pprof", false, "enable cpu profile (v1 compat)")
	_ = f.MarkDeprecated("pprof", "has no effect in arena-v2")
	f.BoolVar(&submitIgnoredBool, "trace", false, "enable trace (v1 compat)")
	_ = f.MarkDeprecated("trace", "has no effect in arena-v2")
	f.BoolVar(&submitIgnoredBool, "helm-binary", false, "use helm binary to submit job (v1 compat)")
	_ = f.MarkDeprecated("helm-binary", "has no effect in arena-v2")
	for _, name := range []string{
		"arena-namespace", "model-name", "model-source",
		"starting-timeout", "role-sequence", "ssh-secret",
	} {
		f.StringVar(&submitIgnoredString, name, "", "(v1 compat)")
		_ = f.MarkDeprecated(name, "has no effect in arena-v2")
	}
}
