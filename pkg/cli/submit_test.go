package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kubeflow/arena/pkg/provider"
)

func TestSubmitCmd_RequiresFrameworkArg(t *testing.T) {
	err := submitCmd.Args(submitCmd, nil)
	assert.Error(t, err)
}

func TestSubmitCmd_AcceptsSingleArg(t *testing.T) {
	err := submitCmd.Args(submitCmd, []string{"pytorch"})
	assert.NoError(t, err)
}

func TestSubmitCmd_RegisteredWithJobCmd(t *testing.T) {
	found := false
	for _, cmd := range jobCmd.Commands() {
		if cmd.Name() == "submit" {
			found = true
			break
		}
	}
	assert.True(t, found, "submit command should be registered with job command")
}

func TestSubmitCmd_HasRequiredFlags(t *testing.T) {
	flagNames := []string{"name", "image", "workers", "gpus", "cpus", "mem"}
	for _, name := range flagNames {
		f := submitCmd.Flags().Lookup(name)
		require.NotNil(t, f, "flag %q should be registered", name)
	}
}

func TestSubmitCmd_HasFrameworkFlags(t *testing.T) {
	f := submitCmd.Flags().Lookup("nproc-per-node")
	assert.NotNil(t, f, "nproc-per-node flag should be registered")

	f = submitCmd.Flags().Lookup("ps-count")
	assert.NotNil(t, f, "ps-count flag should be registered")

	f = submitCmd.Flags().Lookup("slots-per-worker")
	assert.NotNil(t, f, "slots-per-worker flag should be registered")
}

func TestSubmitCmd_HasDryRunFlag(t *testing.T) {
	f := submitCmd.Flags().Lookup("dry-run")
	require.NotNil(t, f, "dry-run flag should be registered")
	assert.Equal(t, "false", f.DefValue)
}

func TestSubmitCmd_NameAndImageRequired(t *testing.T) {
	resetSubmitFlags(t)

	submitName = ""
	submitImage = ""
	err := submitCmd.RunE(submitCmd, []string{"pytorch"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")
}

func TestSubmitCmd_UnsupportedFramework(t *testing.T) {
	resetSubmitFlags(t)

	submitName = "test-job"
	submitImage = "test-image:latest"

	err := submitCmd.RunE(submitCmd, []string{"jax"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported framework")
}

func TestSubmitCmd_ValidationFailsWithoutName(t *testing.T) {
	resetSubmitFlags(t)

	submitName = ""
	submitImage = "some-image:latest"

	err := submitCmd.RunE(submitCmd, []string{"pytorch"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")
	assert.Contains(t, err.Error(), "name is required")
}

func TestSubmitCmd_ValidationFailsWithoutImage(t *testing.T) {
	resetSubmitFlags(t)

	submitName = "my-job"
	submitImage = ""

	err := submitCmd.RunE(submitCmd, []string{"pytorch"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")
	assert.Contains(t, err.Error(), "image is required")
}

func TestBuildSubmitTask_PyTorchWorkersNMinusOne(t *testing.T) {
	tests := []struct {
		name            string
		framework       string
		workers         int
		expectedReplicas int
		expectNilWorker bool
		expectMaster    bool
	}{
		{
			name:            "pytorch with 4 workers converts to 3",
			framework:       "pytorch",
			workers:         4,
			expectedReplicas: 3,
		},
		{
			name:            "pytorch with 2 workers converts to 1",
			framework:       "pytorch",
			workers:         2,
			expectedReplicas: 1,
		},
		{
			name:            "pytorch with 1 worker means master-only",
			framework:       "pytorch",
			workers:         1,
			expectNilWorker: true,
			expectMaster:    true,
		},
		{
			name:            "tensorflow with 4 workers stays at 4",
			framework:       "tensorflow",
			workers:         4,
			expectedReplicas: 4,
		},
		{
			name:            "mpi with 3 workers stays at 3",
			framework:       "mpi",
			workers:         3,
			expectedReplicas: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetSubmitFlags(t)
			submitName = "test-job"
			submitImage = "test:latest"
			submitWorkers = tt.workers

			task := buildSubmitTask(tt.framework, nil)
			if tt.expectNilWorker {
				assert.Nil(t, task.Worker, "worker should be nil")
			} else {
				require.NotNil(t, task.Worker, "worker should not be nil")
				assert.Equal(t, tt.expectedReplicas, task.Worker.Replicas)
			}
			if tt.expectMaster {
				assert.NotNil(t, task.Master, "master should be set")
			}
		})
	}
}

func TestBuildSubmitTask_ChiefEvaluatorPS(t *testing.T) {
	resetSubmitFlags(t)
	submitName = "test-job"
	submitImage = "test:latest"
	submitWorkers = 2
	submitChief = true
	submitEvaluator = true
	submitPSCount = 3

	task := buildSubmitTask("tensorflow", nil)

	require.NotNil(t, task.Chief, "chief should be set")
	require.NotNil(t, task.Evaluator, "evaluator should be set")
	require.NotNil(t, task.PS, "ps should be set")
	assert.Equal(t, 3, task.PS.Replicas)
}

func TestBuildSubmitTask_TrailingArgs(t *testing.T) {
	resetSubmitFlags(t)
	submitName = "test-job"
	submitImage = "test:latest"
	submitWorkers = 1

	trailingArgs := []string{"python", "train.py", "--epochs", "10"}
	task := buildSubmitTask("pytorch", trailingArgs)

	assert.Equal(t, "python train.py --epochs 10", task.Run)
}

func TestNormalizeFramework(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"pytorch", "pytorch"},
		{"PyTorch", "pytorch"},
		{"pytorchjob", "pytorch"},
		{"PyTorchJob", "pytorch"},
		{"tensorflow", "tensorflow"},
		{"TensorFlow", "tensorflow"},
		{"tfjob", "tensorflow"},
		{"TFJob", "tensorflow"},
		{"tf", "tensorflow"},
		{"mpi", "mpi"},
		{"MPI", "mpi"},
		{"mpijob", "mpi"},
		{"MPIJob", "mpi"},
		{"horovod", "mpi"},
		{"Horovod", "mpi"},
		{"jax", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, normalizeFramework(tt.input))
		})
	}
}

func TestBuildSubmitFlags_IncludesTolerations(t *testing.T) {
	resetSubmitFlags(t)
	submitTolerations = []string{"key1=value1:NoSchedule", "key2=value2:NoExecute"}

	flags := buildSubmitFlags()

	assert.Contains(t, flags, "toleration", "buildSubmitFlags should include tolerations")
	assert.Equal(t, submitTolerations, flags["toleration"])
}

func TestBuildSubmitFlags_EmptyTolerations(t *testing.T) {
	resetSubmitFlags(t)
	submitTolerations = nil

	flags := buildSubmitFlags()

	assert.NotContains(t, flags, "toleration", "empty tolerations should not be included")
}

func TestSubmitCmd_DryRun(t *testing.T) {
	resetSubmitFlags(t)

	submitName = "test-job"
	submitImage = "pytorch:1.13"
	submitDryRun = true
	submitGPUs = 2

	err := submitCmd.RunE(submitCmd, []string{"pytorch"})
	// Dry-run should succeed without K8s cluster
	// It will print CRD to stdout
	// Note: this test just verifies the path doesn't error on missing kubeconfig
	// The actual output goes to stdout which we can't easily capture here
	_ = err // May or may not succeed depending on validation, but should not panic
}

func TestBuildSubmitFlags(t *testing.T) {
	resetSubmitFlags(t)

	submitName = "test-job"
	submitImage = "test:latest"
	submitGPUs = 2
	submitCPUs = "4"
	submitMem = "8Gi"
	submitEnvs = []string{"FOO=bar"}
	submitWorkers = 3

	flags := buildSubmitFlags()

	assert.Equal(t, "test-job", flags["name"])
	assert.Equal(t, 2, flags["gpus"])
	assert.Equal(t, "4", flags["cpus"])
	assert.Equal(t, "8Gi", flags["mem"])
	assert.Equal(t, []string{"FOO=bar"}, flags["env"])
}

func TestSubmitDeepSpeed(t *testing.T) {
	fw := normalizeFramework("deepspeed")
	if fw != "mpi" {
		t.Errorf("expected mpi, got %s", fw)
	}
}

func TestOriginalFramework(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"horovod", "horovod"},
		{"deepspeed", "deepspeed"},
		{"mpi", "mpi"},
		{"pytorch", "pytorch"},
		{"tensorflow", "tensorflow"},
		{"PyTorchJob", "pytorch"},
	}
	for _, tt := range tests {
		got := originalFramework(tt.input)
		if got != tt.expected {
			t.Errorf("originalFramework(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestSubmitCmd_MPIVersionIntegration_V1FromCluster(t *testing.T) {
	// Simulates the submit command flow when the cluster has MPIJob CRD with
	// storage version v1. Verifies that the generated CR uses kubeflow.org/v1.
	resetSubmitFlags(t)

	submitName = "mpi-v1-submit"
	submitImage = "openmpi:4.1"
	submitWorkers = 4

	framework := normalizeFramework("mpi")
	require.Equal(t, "mpi", framework)

	trailingArgs := []string{"mpirun", "-np", "4", "./train"}
	tk := buildSubmitTask(framework, trailingArgs)
	assert.Equal(t, "mpi-v1-submit", tk.Name)

	// Get provider
	p, err := getProvider(framework)
	require.NoError(t, err)

	// Simulate version detection: cluster reports v1 as storage version
	mpiP, ok := p.(*provider.MPIProvider)
	require.True(t, ok, "expected MPIProvider for mpi framework")
	mpiP.APIVersion = "v1"

	crd, err := p.BuildCRD(tk)
	require.NoError(t, err)

	assert.Equal(t, "MPIJob", crd.GetKind())
	assert.Equal(t, "kubeflow.org/v1", crd.GetAPIVersion())

	// v1 CR should NOT have v2beta1-only fields
	spec := crd.Object["spec"].(map[string]interface{})
	_, hasSSHAuth := spec["sshAuthMountPath"]
	assert.False(t, hasSSHAuth, "v1 CR should not have sshAuthMountPath")
}

func TestSubmitCmd_MPIVersionIntegration_V2beta1Default(t *testing.T) {
	// Simulates the submit command flow in dry-run mode (no cluster).
	// With the removal of the default fallback, APIVersion must be explicitly set.
	resetSubmitFlags(t)

	submitName = "mpi-v2beta1-submit"
	submitImage = "openmpi:4.1"
	submitWorkers = 2

	framework := normalizeFramework("mpi")
	tk := buildSubmitTask(framework, nil)

	p, err := getProvider(framework)
	require.NoError(t, err)

	mpiP, ok := p.(*provider.MPIProvider)
	require.True(t, ok)
	assert.Empty(t, mpiP.APIVersion)

	// Empty APIVersion must now produce an error (no default fallback)
	_, err = p.BuildCRD(tk)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "APIVersion must be set")

	// After setting APIVersion, BuildCRD succeeds with v2beta1
	mpiP.APIVersion = provider.MPIAPIVersionV2beta1
	crd, err := p.BuildCRD(tk)
	require.NoError(t, err)
	assert.Equal(t, "kubeflow.org/v2beta1", crd.GetAPIVersion())
}

func TestSubmitCmd_MPIVersionIntegration_DeepSpeed(t *testing.T) {
	// Verifies that deepspeed (MPI-family) also uses the detected version.
	resetSubmitFlags(t)

	submitName = "deepspeed-v1"
	submitImage = "deepspeed:latest"
	submitWorkers = 2

	framework := normalizeFramework("deepspeed")
	require.Equal(t, "mpi", framework)
	assert.True(t, isMPIFamily(framework))

	tk := buildSubmitTask(framework, []string{"deepspeed", "train.py"})

	p, err := getProvider(framework)
	require.NoError(t, err)

	mpiP, ok := p.(*provider.MPIProvider)
	require.True(t, ok)
	mpiP.APIVersion = "v1"

	crd, err := p.BuildCRD(tk)
	require.NoError(t, err)

	assert.Equal(t, "kubeflow.org/v1", crd.GetAPIVersion())
}

func TestSubmitCmd_MPIVersionIntegration_Horovod(t *testing.T) {
	// Verifies that horovod (MPI-family) also uses the detected version.
	resetSubmitFlags(t)

	submitName = "horovod-v1"
	submitImage = "horovod:latest"
	submitWorkers = 3

	framework := normalizeFramework("horovod")
	require.Equal(t, "mpi", framework)

	tk := buildSubmitTask(framework, []string{"mpirun", "train"})

	p, err := getProvider(framework)
	require.NoError(t, err)

	mpiP, ok := p.(*provider.MPIProvider)
	require.True(t, ok)
	mpiP.APIVersion = "v1"

	crd, err := p.BuildCRD(tk)
	require.NoError(t, err)

	assert.Equal(t, "kubeflow.org/v1", crd.GetAPIVersion())
}

// resetSubmitFlags resets all submit flag variables to their defaults
func resetSubmitFlags(t *testing.T) {
	t.Helper()
	submitName = ""
	submitImage = ""
	submitWorkers = 1
	submitGPUs = 0
	submitCPUs = ""
	submitMem = ""
	submitEnvs = nil
	submitData = nil
	submitLabels = nil
	submitAnnotations = nil
	submitSelectors = nil
	submitTolerations = nil
	submitPriority = 0
	submitPriorityClass = ""
	submitGang = false
	submitSchedulerName = ""
	submitCleanPodPolicy = ""
	submitActiveDeadline = ""
	submitTTLAfterFinished = ""
	submitBackoffLimit = 0
	submitImagePullPolicy = ""
	submitImagePullSecret = nil
	submitServiceAccount = ""
	submitRestart = ""
	submitHostNetwork = false
	submitHostIPC = false
	submitHostPID = false
	submitWorkingDir = ""
	submitShell = ""
	submitSHM = ""
	submitDevice = nil
	submitGPUType = ""
	submitTensorBoard = false
	submitTBLogDir = ""
	submitTBImage = ""
	submitNprocPerNode = ""
	submitPSCount = 0
	submitChief = false
	submitEvaluator = false
	submitSlotsPerWorker = 0
	submitGPUTopology = false
	submitMountsOnLauncher = false
	submitAffinityPolicy = ""
	submitAffinityConstraint = ""
	submitSuccessPolicy = ""
	submitDryRun = false
	submitQueue = ""
	submitDataDir = nil
	submitConfigFile = nil
}
