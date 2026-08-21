package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kubeflow/arena/pkg/provider"
	"github.com/kubeflow/arena/pkg/task"
)

func TestSubmitCmd_NoFrameworkArgPrintsHelp(t *testing.T) {
	// RunE guards the empty-args case with help (the framework is the first
	// positional argument); full stdout coverage lives in
	// TestSubmitSubcommands_ParentFallbackBehaviors.
	cmd := newSubmitCmd()
	err := cmd.RunE(cmd, nil)
	assert.NoError(t, err)
}

func TestSubmitCmd_RegisteredWithRootCmd(t *testing.T) {
	found := false
	for _, cmd := range NewRootCommand().Commands() {
		if cmd.Name() == "submit" {
			found = true
			break
		}
	}
	assert.True(t, found, "submit command should be registered with root command")
}

func TestSubmitCmd_HasRequiredFlags(t *testing.T) {
	cmd := newSubmitCmd()
	flagNames := []string{"name", "image", "workers", "gpus", "cpu", "memory"}
	for _, name := range flagNames {
		f := cmd.Flags().Lookup(name)
		require.NotNil(t, f, "flag %q should be registered", name)
	}
}

func TestSubmitCmd_HasFrameworkFlags(t *testing.T) {
	cmd := newSubmitCmd()
	f := cmd.Flags().Lookup("nproc-per-node")
	assert.NotNil(t, f, "nproc-per-node flag should be registered")

	f = cmd.Flags().Lookup("ps")
	assert.NotNil(t, f, "ps flag should be registered")

	f = cmd.Flags().Lookup("slots-per-worker")
	assert.NotNil(t, f, "slots-per-worker flag should be registered")
}

func TestSubmitCmd_HasDryRunFlag(t *testing.T) {
	cmd := newSubmitCmd()
	f := cmd.Flags().Lookup("dry-run")
	require.NotNil(t, f, "dry-run flag should be registered")
	assert.Equal(t, "false", f.DefValue)
}

func TestSubmitCmd_NameAndImageRequired(t *testing.T) {
	cmd := newSubmitCmd()
	err := cmd.RunE(cmd, []string{"pytorch"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required flag")
}

func TestSubmitCmd_UnsupportedFramework(t *testing.T) {
	cmd := newSubmitCmd()
	submitName = "test-job"
	submitImage = "test-image:latest"

	err := cmd.RunE(cmd, []string{"jax"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported framework")
}

func TestSubmitCmd_ValidationFailsWithoutName(t *testing.T) {
	cmd := newSubmitCmd()
	require.NoError(t, cmd.Flags().Parse([]string{"--image", "some-image:latest"}))

	err := cmd.RunE(cmd, []string{"pytorch"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), `required flag(s) "name" not set`)
}

func TestSubmitCmd_ValidationFailsWithoutImage(t *testing.T) {
	cmd := newSubmitCmd()
	require.NoError(t, cmd.Flags().Parse([]string{"--name", "my-job"}))

	err := cmd.RunE(cmd, []string{"pytorch"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), `required flag(s) "image" not set`)
}

func TestBuildSubmitTask_PyTorchWorkersNMinusOne(t *testing.T) {
	tests := []struct {
		name             string
		framework        string
		workers          int
		expectedReplicas int
		expectNilWorker  bool
		expectMaster     bool
	}{
		{
			name:             "pytorch with 4 workers converts to 3",
			framework:        "pytorch",
			workers:          4,
			expectedReplicas: 3,
		},
		{
			name:             "pytorch with 2 workers converts to 1",
			framework:        "pytorch",
			workers:          2,
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
			name:             "tensorflow with 4 workers stays at 4",
			framework:        "tensorflow",
			workers:          4,
			expectedReplicas: 4,
		},
		{
			name:             "mpi with 3 workers stays at 3",
			framework:        "mpi",
			workers:          3,
			expectedReplicas: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetSubmitGlobals(t)
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
	resetSubmitGlobals(t)
	submitName = "test-job"
	submitImage = "test:latest"
	submitWorkers = 2
	submitChief = true
	submitEvaluator = true
	submitPS = 3

	// v1 per-role resource flags. These must flow through buildSubmitFlags
	// into task.ApplyOverrides and land on each role's Resources — a dropped
	// or misnamed flag key would silently break the v1 compat interface.
	submitCPU = "2"
	submitMemory = "4Gi"
	submitPSCPU = "500m"
	submitPSMemory = "1Gi"
	submitPSGPUs = 1
	submitChiefCPU = "1"
	submitChiefMemory = "2Gi"
	submitEvaluatorCPU = "250m"
	submitEvaluatorMemory = "512Mi"
	submitWorkerCPU = "8"
	submitWorkerMemory = "32Gi"

	// Role sections are created by task.ApplyOverrides from the
	// chief/evaluator/ps flag keys, not by buildSubmitTask.
	tk := buildSubmitTask("tensorflow", nil)
	require.NoError(t, task.ApplyOverrides(tk, buildSubmitFlags()))

	require.NotNil(t, tk.Chief, "chief should be set")
	require.NotNil(t, tk.Evaluator, "evaluator should be set")
	require.NotNil(t, tk.PS, "ps should be set")
	require.NotNil(t, tk.Worker, "worker should be set")
	assert.Equal(t, 3, tk.PS.Replicas)

	// Per-role resources land on their role and beat the generic
	// --cpu/--memory (v1 precedence).
	assert.Equal(t, "500m", tk.PS.Resources["cpu"])
	assert.Equal(t, "1Gi", tk.PS.Resources["memory"])
	assert.Equal(t, "1", tk.PS.Resources["nvidia.com/gpu"])
	assert.Equal(t, "1", tk.Chief.Resources["cpu"])
	assert.Equal(t, "2Gi", tk.Chief.Resources["memory"])
	assert.Equal(t, "250m", tk.Evaluator.Resources["cpu"])
	assert.Equal(t, "512Mi", tk.Evaluator.Resources["memory"])
	assert.Equal(t, "8", tk.Worker.Resources["cpu"])
	assert.Equal(t, "32Gi", tk.Worker.Resources["memory"])
}

func TestBuildSubmitTask_TrailingArgs(t *testing.T) {
	resetSubmitGlobals(t)
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
		{"horovod", "horovod"},
		{"Horovod", "horovod"},
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
	resetSubmitGlobals(t)
	submitTolerations = []string{"key1=value1:NoSchedule", "key2=value2:NoExecute"}

	flags := buildSubmitFlags()

	assert.Contains(t, flags, "toleration", "buildSubmitFlags should include tolerations")
	assert.Equal(t, submitTolerations, flags["toleration"])
}

func TestBuildSubmitFlags_EmptyTolerations(t *testing.T) {
	resetSubmitGlobals(t)
	submitTolerations = nil

	flags := buildSubmitFlags()

	assert.NotContains(t, flags, "toleration", "empty tolerations should not be included")
}

func TestBuildSubmitFlags(t *testing.T) {
	resetSubmitGlobals(t)

	submitName = "test-job"
	submitImage = "test:latest"
	submitGPUs = 2
	submitCPU = "4"
	submitMemory = "8Gi"
	submitEnvs = []string{"FOO=bar"}
	submitWorkers = 3

	flags := buildSubmitFlags()

	assert.Equal(t, "test-job", flags["name"])
	assert.Equal(t, 2, flags["gpus"])
	assert.Equal(t, "4", flags["cpu"])
	assert.Equal(t, "8Gi", flags["memory"])
	assert.Equal(t, []string{"FOO=bar"}, flags["env"])
}

func TestSubmitDeepSpeed(t *testing.T) {
	fw := normalizeFramework("deepspeed")
	if fw != "deepspeed" {
		t.Errorf("expected deepspeed, got %s", fw)
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
	resetSubmitGlobals(t)

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
	resetSubmitGlobals(t)

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
	assert.Contains(t, err.Error(), "apiversion is not set")

	// After setting APIVersion, BuildCRD succeeds with v2beta1
	mpiP.APIVersion = provider.MPIAPIVersionV2beta1
	crd, err := p.BuildCRD(tk)
	require.NoError(t, err)
	assert.Equal(t, "kubeflow.org/v2beta1", crd.GetAPIVersion())
}

func TestSubmitCmd_MPIVersionIntegration_DeepSpeed(t *testing.T) {
	// Verifies that deepspeed (MPI-family) also uses the detected version.
	resetSubmitGlobals(t)

	submitName = "deepspeed-v1"
	submitImage = "deepspeed:latest"
	submitWorkers = 2

	framework := normalizeFramework("deepspeed")
	require.Equal(t, "deepspeed", framework)
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
	resetSubmitGlobals(t)

	submitName = "horovod-v1"
	submitImage = "horovod:latest"
	submitWorkers = 3

	framework := normalizeFramework("horovod")
	require.Equal(t, "horovod", framework)

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

// resetSubmitGlobals rebinds every submit flag variable to its default by
// constructing a fresh submit command: pflag writes the default into the
// bound variable at registration time, so a fresh tree is a full reset.
func resetSubmitGlobals(t *testing.T) {
	t.Helper()
	newSubmitCmd()
}

func TestCRDReplicaSpecs_PyTorch(t *testing.T) {
	yamlPath := filepath.Join(examplesDir(t), "pytorch-simple.yaml")
	taskObj, err := task.LoadFromFile(yamlPath)
	require.NoError(t, err)

	assert.Equal(t, "pytorch-example", taskObj.Name)
	assert.Equal(t, "pytorch", taskObj.Framework.Name)
	assert.Equal(t, 2, taskObj.Worker.Replicas)

	p, err := getProvider(taskObj.Framework.Name)
	require.NoError(t, err)

	crd, err := p.BuildCRD(taskObj)
	require.NoError(t, err)
	assert.Equal(t, "PyTorchJob", crd.GetKind())
	assert.Equal(t, "kubeflow.org/v1", crd.GetAPIVersion())

	spec := crd.Object["spec"].(map[string]interface{})
	replicaSpecs := spec["pytorchReplicaSpecs"].(map[string]interface{})

	master := replicaSpecs["Master"].(map[string]interface{})
	assert.Equal(t, int64(1), master["replicas"])

	worker := replicaSpecs["Worker"].(map[string]interface{})
	assert.Equal(t, int64(2), worker["replicas"])

	// Full CRUD lifecycle on fake client
	ctx := context.Background()
	k8sClient := newFakeK8sClient(t)
	crd.SetNamespace("default")
	require.NoError(t, k8sClient.Create(ctx, crd))

	obj, err := k8sClient.Get(ctx, "PyTorchJob", "default", "pytorch-example")
	require.NoError(t, err)
	assert.Equal(t, "pytorch-example", obj.GetName())

	jobs, err := k8sClient.List(ctx, "PyTorchJob", "default", "")
	require.NoError(t, err)
	require.Len(t, jobs, 1)

	require.NoError(t, k8sClient.Delete(ctx, "PyTorchJob", "default", "pytorch-example"))
	_, err = k8sClient.Get(ctx, "PyTorchJob", "default", "pytorch-example")
	require.Error(t, err)
}

func TestCRDReplicaSpecs_TensorFlow(t *testing.T) {
	yamlPath := filepath.Join(examplesDir(t), "tensorflow-simple.yaml")
	taskObj, err := task.LoadFromFile(yamlPath)
	require.NoError(t, err)

	assert.Equal(t, "tensorflow", taskObj.Framework.Name)
	assert.Equal(t, 2, taskObj.Worker.Replicas)

	p, err := getProvider(taskObj.Framework.Name)
	require.NoError(t, err)

	crd, err := p.BuildCRD(taskObj)
	require.NoError(t, err)
	assert.Equal(t, "TFJob", crd.GetKind())

	spec := crd.Object["spec"].(map[string]interface{})
	replicaSpecs := spec["tfReplicaSpecs"].(map[string]interface{})

	worker := replicaSpecs["Worker"].(map[string]interface{})
	assert.Equal(t, int64(2), worker["replicas"])
}

func TestCRDReplicaSpecs_MPI(t *testing.T) {
	yamlPath := filepath.Join(examplesDir(t), "mpi-simple.yaml")
	taskObj, err := task.LoadFromFile(yamlPath)
	require.NoError(t, err)

	assert.Equal(t, "mpi", taskObj.Framework.Name)
	assert.Equal(t, 2, taskObj.Worker.Replicas)

	p, err := getProvider(taskObj.Framework.Name)
	require.NoError(t, err)

	mpiP, ok := p.(*provider.MPIProvider)
	require.True(t, ok, "expected MPIProvider for mpi framework")
	mpiP.APIVersion = "v1"

	crd, err := p.BuildCRD(taskObj)
	require.NoError(t, err)
	assert.Equal(t, "MPIJob", crd.GetKind())

	spec := crd.Object["spec"].(map[string]interface{})
	assert.Equal(t, int64(1), spec["slotsPerWorker"])

	replicaSpecs := spec["mpiReplicaSpecs"].(map[string]interface{})

	launcher := replicaSpecs["Launcher"].(map[string]interface{})
	assert.Equal(t, int64(1), launcher["replicas"])

	worker := replicaSpecs["Worker"].(map[string]interface{})
	assert.Equal(t, int64(2), worker["replicas"])
}

func TestApplyOverrides_Flags(t *testing.T) {
	taskObj := &task.Task{
		Framework: task.Framework{Name: "pytorch"},
		Worker:    &task.Worker{Replicas: 1},
	}

	flags := map[string]interface{}{
		"name":           "my-pytorch-job",
		"image":          "pytorch/pytorch:2.1",
		"run":            "python train.py --lr 0.001",
		"workers":        4,
		"gpus":           2,
		"cpu":            "4",
		"memory":         "16Gi",
		"framework":      "pytorch",
		"nproc-per-node": "auto",
	}
	require.NoError(t, task.ApplyOverrides(taskObj, flags))

	taskObj.SetDefaults()
	require.NoError(t, task.Validate(taskObj))

	assert.Equal(t, "my-pytorch-job", taskObj.Name)
	assert.Equal(t, "pytorch/pytorch:2.1", taskObj.Image)
	assert.Equal(t, "python train.py --lr 0.001", taskObj.Run)
	assert.Equal(t, 4, taskObj.Worker.Replicas)
	assert.Equal(t, "2", taskObj.Worker.Resources["nvidia.com/gpu"])
	assert.Equal(t, "auto", taskObj.Framework.Options.NprocPerNode)

	p, err := getProvider("pytorch")
	require.NoError(t, err)
	crd, err := p.BuildCRD(taskObj)
	require.NoError(t, err)
	assert.Equal(t, "PyTorchJob", crd.GetKind())
	assert.Equal(t, "my-pytorch-job", crd.GetName())

	ctx := context.Background()
	k8sClient := newFakeK8sClient(t)
	crd.SetNamespace("default")
	require.NoError(t, k8sClient.Create(ctx, crd))

	obj, err := k8sClient.Get(ctx, "PyTorchJob", "default", "my-pytorch-job")
	require.NoError(t, err)
	assert.Equal(t, "my-pytorch-job", obj.GetName())
}

func TestApplyOverrides_Namespace(t *testing.T) {
	taskObj := &task.Task{
		Name:      "ns-override",
		Image:     "pytorch:1.13",
		Framework: task.Framework{Name: "pytorch"},
		Worker:    &task.Worker{Replicas: 1},
	}

	flags := map[string]interface{}{
		"namespace": "custom-ns",
	}
	require.NoError(t, task.ApplyOverrides(taskObj, flags))

	assert.Equal(t, "custom-ns", taskObj.Namespace)
}

func TestValidationRejectsInvalidTasks(t *testing.T) {
	tests := []struct {
		name string
		yaml string
		err  string
	}{
		{
			name: "zero replicas",
			yaml: `
name: test
image: pytorch:1.13
run: echo test
framework:
  name: pytorch
worker:
  replicas: 0
`,
			err: "worker.replicas must be > 0",
		},
		{
			name: "invalid cleanPodPolicy",
			yaml: `
name: test
image: pytorch:1.13
run: echo test
framework:
  name: pytorch
worker:
  replicas: 1
lifecycle:
  clean_pod_policy: InvalidPolicy
`,
			err: "invalid clean_pod_policy",
		},
		{
			name: "invalid restart",
			yaml: `
name: test
image: pytorch:1.13
run: echo test
framework:
  name: pytorch
worker:
  replicas: 1
restart: BadPolicy
`,
			err: "invalid restart",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := task.LoadFromBytes([]byte(tt.yaml))
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.err)
		})
	}
}

func TestValidationRejectsInvalidNprocPerNode(t *testing.T) {
	yamlData := `
name: test
image: pytorch:1.13
run: echo test
framework:
  name: pytorch
  options:
    nproc_per_node: "not-a-number"
worker:
  replicas: 1
`
	_, err := task.LoadFromBytes([]byte(yamlData))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nproc_per_node")
}

func TestProviderRejectsWrongFramework(t *testing.T) {
	tk := &task.Task{
		Name:      "wrong",
		Image:     "pytorch:1.13",
		Framework: task.Framework{Name: "pytorch"},
		Worker:    &task.Worker{Replicas: 1},
	}

	tests := []struct {
		name     string
		provider provider.Provider
		err      string
	}{
		{"TensorFlow rejects PyTorch", &provider.TensorFlowProvider{}, "tensorflow"},
		{"MPI rejects PyTorch", &provider.MPIProvider{}, "mpi"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.provider.BuildCRD(tk)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.err)
		})
	}
}

func TestSubmitCmd_DryRunCapturesOutput(t *testing.T) {
	cmd := newSubmitCmd()
	require.NoError(t, cmd.Flags().Parse([]string{
		"--name", "dryrun-test",
		"--image", "pytorch:2.1",
		"--workers", "2",
		"--gpus", "1",
		"--dry-run",
	}))

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	// Trailing args "python train.py" become the run command (required by validation).
	err := cmd.RunE(cmd, []string{"pytorch", "python", "train.py"})

	require.NoError(t, err, "dry-run should succeed without a cluster")
	output := buf.String()
	require.NotEmpty(t, output, "dry-run should produce JSON output")

	var crd map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(output), &crd),
		"output should be valid JSON")
	assert.Equal(t, "PyTorchJob", crd["kind"])
	assert.Equal(t, "dryrun-test", crd["metadata"].(map[string]interface{})["name"])
}

// The generated CRD must wire the shared volume into both containers: the
// sync init container (writer) and the main container (reader).
func TestSubmitCompat_SyncVolumeSharedWithMainContainer(t *testing.T) {
	cmd := newSubmitCmd()
	v1Args := []string{
		"--name", "sync-volume-e2e",
		"--image", "pytorch:2.1",
		"--sync-mode", "git",
		"--sync-source", "https://github.com/kubeflow/arena.git",
		"--dry-run",
	}
	require.NoError(t, cmd.Flags().Parse(v1Args))

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	err := cmd.RunE(cmd, []string{"pytorch", "python", "train.py"})

	output := buf.String()
	require.NoError(t, err, "sync invocation should succeed in dry-run mode")

	var crd map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(output), &crd), "dry-run should print valid JSON, got: %s", output)

	spec := crd["spec"].(map[string]interface{})
	replicaSpecs := spec["pytorchReplicaSpecs"].(map[string]interface{})
	master := replicaSpecs["Master"].(map[string]interface{})
	podSpec := master["template"].(map[string]interface{})["spec"].(map[string]interface{})

	mountPaths := func(container map[string]interface{}) map[string]string {
		result := map[string]string{}
		for _, m := range container["volumeMounts"].([]interface{}) {
			mm := m.(map[string]interface{})
			result[mm["name"].(string)] = mm["mountPath"].(string)
		}
		return result
	}

	containers := podSpec["containers"].([]interface{})
	mainMounts := mountPaths(containers[0].(map[string]interface{}))
	assert.Equal(t, "/root/code", mainMounts["code-sync"],
		"main container should mount the shared code-sync volume at $workingDir/code")

	initContainers := podSpec["initContainers"].([]interface{})
	require.Len(t, initContainers, 1)
	initMounts := mountPaths(initContainers[0].(map[string]interface{}))
	assert.Equal(t, "/root/code", initMounts["code-sync"],
		"sync init container should write into the shared code-sync volume")
}

func TestSubmitCompat_V1CommandEndToEndDryRun(t *testing.T) {
	cmd := newSubmitCmd()

	// A representative v1 invocation using only v1 flag names.
	v1Args := []string{
		"--name", "v1-compat-job",
		"--image", "pytorch:2.1",
		"--workers", "2",
		"--cpu", "4",
		"--memory", "8Gi",
		"-p", "high",
		"--share-memory", "2Gi",
		"--running-timeout", "2h",
		"--retry", "3",
		"--sync-mode", "git",
		"--sync-source", "https://github.com/kubeflow/arena.git",
		"--dry-run",
	}
	require.NoError(t, cmd.Flags().Parse(v1Args))

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	err := cmd.RunE(cmd, []string{"pytorch", "python", "train.py"})

	output := buf.String()
	require.NoError(t, err, "v1-style invocation should succeed in dry-run mode")

	var crd map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(output), &crd), "dry-run should print valid JSON, got: %s", output)
	assert.Equal(t, "PyTorchJob", crd["kind"])

	spec := crd["spec"].(map[string]interface{})
	runPolicy := spec["runPolicy"].(map[string]interface{})
	assert.Equal(t, float64(7200), runPolicy["activeDeadlineSeconds"],
		"--running-timeout 2h should map to runPolicy.activeDeadlineSeconds")
	assert.Equal(t, float64(3), runPolicy["backoffLimit"],
		"--retry 3 should map to runPolicy.backoffLimit")

	replicaSpecs := spec["pytorchReplicaSpecs"].(map[string]interface{})
	master := replicaSpecs["Master"].(map[string]interface{})
	podSpec := master["template"].(map[string]interface{})["spec"].(map[string]interface{})

	assert.Equal(t, "high", podSpec["priorityClassName"],
		"-p should set the priority class name (v1 semantics)")

	containers := podSpec["containers"].([]interface{})
	container := containers[0].(map[string]interface{})
	resources := container["resources"].(map[string]interface{})
	requests := resources["requests"].(map[string]interface{})
	assert.Equal(t, "4", requests["cpu"], "--cpu should map to the CPU request")
	assert.Equal(t, "8Gi", requests["memory"], "--memory should map to the memory request")

	initContainers := podSpec["initContainers"].([]interface{})
	require.Len(t, initContainers, 1, "git sync should produce one init container")
	init := initContainers[0].(map[string]interface{})
	assert.Contains(t, init["name"], "arena-sync")
}

func TestSubmitCompat_UniformV1Defaults(t *testing.T) {
	cmd := newSubmitCmd()
	f := cmd.Flags().Lookup("share-memory")
	require.NotNil(t, f)
	assert.Equal(t, "2Gi", f.DefValue)
	f = cmd.Flags().Lookup("clean-task-policy")
	require.NotNil(t, f)
	assert.Equal(t, "Running", f.DefValue)
	f = cmd.Flags().Lookup("logdir")
	require.NotNil(t, f)
	assert.Equal(t, "/training_logs", f.DefValue)
}

func TestSubmitCompat_DefaultsReachTheTask(t *testing.T) {
	resetSubmitGlobals(t)
	tk := buildSubmitTask("pytorch", nil)
	flags := buildSubmitFlags()
	require.NoError(t, task.ApplyOverrides(tk, flags))

	assert.Equal(t, "Running", tk.Lifecycle.CleanPodPolicy)
	require.NotNil(t, tk.Logging.TensorBoard)
	assert.Equal(t, "/training_logs", tk.Logging.TensorBoard.LogDir)

	foundSHM := false
	for _, s := range tk.Storages {
		if s.SHM != "" {
			assert.Equal(t, "2Gi", s.SHM)
			foundSHM = true
		}
	}
	assert.True(t, foundSHM, "default 2Gi shm storage should be applied")
}

func TestSubmitCompat_DefaultsOptOutViaEmptyValue(t *testing.T) {
	cmd := newSubmitCmd()
	require.NoError(t, cmd.Flags().Parse([]string{"--share-memory", "", "--clean-task-policy", "", "--logdir", ""}))
	flags := buildSubmitFlags()
	assert.NotContains(t, flags, "share-memory")
	assert.NotContains(t, flags, "clean-task-policy")
	assert.NotContains(t, flags, "logdir")
}
