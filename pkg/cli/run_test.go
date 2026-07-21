package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kubeflow/arena/pkg/provider"
	"github.com/kubeflow/arena/pkg/task"
)

func TestGetProvider(t *testing.T) {
	tests := []struct {
		name      string
		framework string
		wantType  string
		wantErr   bool
	}{
		{"pytorch", "pytorch", "PyTorchJob", false},
		{"tensorflow", "tensorflow", "TFJob", false},
		{"mpi", "mpi", "MPIJob", false},
		{"horovod", "horovod", "MPIJob", false},
		{"unsupported", "jax", "", true},
		{"empty", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := getProvider(tt.framework)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "unsupported framework")
				assert.Nil(t, p)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, p)
				assert.Equal(t, tt.wantType, p.GetJobType())
			}
		})
	}
}

func TestRunCmd_FileRequired(t *testing.T) {
	original := runFile
	defer func() { runFile = original }()

	runFile = ""
	err := runCmd.RunE(runCmd, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--file is required")
}

func TestRunCmd_InvalidFile(t *testing.T) {
	original := runFile
	defer func() { runFile = original }()

	runFile = "/nonexistent/path/to/file.yaml"
	err := runCmd.RunE(runCmd, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read file")
}

func TestRunCmd_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "invalid.yaml")
	err := os.WriteFile(tmpFile, []byte("not: valid: yaml: {{{"), 0644)
	require.NoError(t, err)

	original := runFile
	defer func() { runFile = original }()

	runFile = tmpFile
	err = runCmd.RunE(runCmd, nil)
	assert.Error(t, err)
}

func TestRunCmd_MissingRequiredFields(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "incomplete.yaml")
	content := `name: ""
image: some-image
run: echo hello
framework:
  name: pytorch
worker:
  replicas: 1
`
	err := os.WriteFile(tmpFile, []byte(content), 0644)
	require.NoError(t, err)

	original := runFile
	defer func() { runFile = original }()

	runFile = tmpFile
	err = runCmd.RunE(runCmd, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load task")
}

func TestRunCmd_UnsupportedFramework(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "unsupported.yaml")
	content := `name: test-job
image: some-image
run: echo hello
framework:
  name: jax
worker:
  replicas: 1
`
	err := os.WriteFile(tmpFile, []byte(content), 0644)
	require.NoError(t, err)

	original := runFile
	defer func() { runFile = original }()

	runFile = tmpFile
	err = runCmd.RunE(runCmd, nil)
	assert.Error(t, err)
}

func TestRunCmd_DryRun(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "valid.yaml")
	content := `name: dry-run-test
image: pytorch:1.13
run: python train.py
framework:
  name: pytorch
worker:
  replicas: 2
`
	err := os.WriteFile(tmpFile, []byte(content), 0644)
	require.NoError(t, err)

	originalFile := runFile
	originalDryRun := runDryRun
	defer func() {
		runFile = originalFile
		runDryRun = originalDryRun
	}()

	runFile = tmpFile
	runDryRun = true
	err = runCmd.RunE(runCmd, nil)
	assert.NoError(t, err)
}

func TestRunCmd_HasDryRunFlag(t *testing.T) {
	f := runCmd.Flags().Lookup("dry-run")
	require.NotNil(t, f, "dry-run flag should be registered")
	assert.Equal(t, "false", f.DefValue)
}

func TestRunCmd_HasSetFlag(t *testing.T) {
	f := runCmd.Flags().Lookup("set")
	require.NotNil(t, f, "set flag should be registered")
}

func TestRunCmd_RegisteredWithJobCmd(t *testing.T) {
	found := false
	for _, cmd := range jobCmd.Commands() {
		if cmd.Name() == "run" {
			found = true
			break
		}
	}
	assert.True(t, found, "run command should be registered with job command")
}

func TestRunCmd_FrameworkLabelInjected(t *testing.T) {
	tests := []struct {
		name      string
		framework string
	}{
		{"pytorch", "pytorch"},
		{"tensorflow", "tensorflow"},
		{"mpi", "mpi"},
		{"horovod", "horovod"},
		{"deepspeed", "deepspeed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "test.yaml")
			content := "name: test-" + tt.framework + "\nimage: some-image\nrun: echo hello\nframework:\n  name: " + tt.framework + "\nworker:\n  replicas: 1\n"
			err := os.WriteFile(tmpFile, []byte(content), 0644)
			require.NoError(t, err)

			tk, err := task.LoadFromFile(tmpFile)
			require.NoError(t, err)

			p, err := getProvider(tk.Framework.Name)
			require.NoError(t, err)

			// MPI provider requires an explicit APIVersion before BuildCRD
			if mpiP, ok := p.(*provider.MPIProvider); ok {
				mpiP.APIVersion = provider.MPIAPIVersionV2beta1
			}

			crd, err := p.BuildCRD(tk)
			require.NoError(t, err)

			setFrameworkLabel(crd, tk.Framework.Name)

			labels := crd.GetLabels()
			assert.Equal(t, tt.framework, labels[FrameworkLabel],
				"framework label should be %q for framework %q", tt.framework, tt.framework)
		})
	}
}

func TestRunCmd_DryRunJSONOutput(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "valid.yaml")
	content := "name: dry-json-test\nimage: pytorch:1.13\nrun: python train.py\nframework:\n  name: pytorch\nworker:\n  replicas: 2\n"
	err := os.WriteFile(tmpFile, []byte(content), 0644)
	require.NoError(t, err)

	tk, err := task.LoadFromFile(tmpFile)
	require.NoError(t, err)

	p, err := getProvider("pytorch")
	require.NoError(t, err)

	crd, err := p.BuildCRD(tk)
	require.NoError(t, err)

	setFrameworkLabel(crd, "pytorch")

	data, err := json.MarshalIndent(crd.Object, "", "  ")
	require.NoError(t, err)

	var parsed map[string]interface{}
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	assert.Equal(t, "PyTorchJob", parsed["kind"])
	metadata := parsed["metadata"].(map[string]interface{})
	assert.Equal(t, "dry-json-test", metadata["name"])

	labels, ok := metadata["labels"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "pytorch", labels[FrameworkLabel])
}

func TestRunCmd_NamespaceSetOnCRD(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "ns.yaml")
	content := "name: ns-test\nimage: pytorch:1.13\nrun: echo hello\nframework:\n  name: pytorch\nworker:\n  replicas: 1\n"
	err := os.WriteFile(tmpFile, []byte(content), 0644)
	require.NoError(t, err)

	tk, err := task.LoadFromFile(tmpFile)
	require.NoError(t, err)

	p := &provider.PyTorchProvider{}
	crd, err := p.BuildCRD(tk)
	require.NoError(t, err)

	crd.SetNamespace("my-namespace")
	assert.Equal(t, "my-namespace", crd.GetNamespace())
}

// resetRunFlags resets run flag variables to their defaults for test isolation.
func resetRunFlags(t *testing.T) {
	t.Helper()
	runFile = ""
	runDryRun = false
	runSetExprs = nil
}

// writeTestYAML creates a temporary YAML file for run command tests.
func writeTestYAML(t *testing.T, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.yaml")
	err := os.WriteFile(tmpFile, []byte(content), 0644)
	require.NoError(t, err)
	return tmpFile
}

const testRunYAML = `name: base-job
image: pytorch:1.13
run: python train.py
framework:
  name: pytorch
worker:
  replicas: 2
  resources:
    cpu: "2"
    memory: "4Gi"
`

func TestRunCmd_NoOverrides_TaskMatchesFile(t *testing.T) {
	resetRunFlags(t)

	tmpFile := writeTestYAML(t, testRunYAML)
	runFile = tmpFile
	runDryRun = true

	err := runCmd.RunE(runCmd, nil)
	assert.NoError(t, err)
}

func TestRunCmd_SetOverrideName(t *testing.T) {
	resetRunFlags(t)

	tmpFile := writeTestYAML(t, testRunYAML)
	runFile = tmpFile
	runDryRun = true
	runSetExprs = []string{"name=overridden-name"}

	err := runCmd.RunE(runCmd, nil)
	assert.NoError(t, err)

	// Verify merged value reaches the Task struct
	yamlData, err := os.ReadFile(tmpFile)
	require.NoError(t, err)
	mergedData, err := task.ApplySetOverrides(yamlData, runSetExprs)
	require.NoError(t, err)
	tk, err := task.LoadFromBytes(mergedData)
	require.NoError(t, err)
	assert.Equal(t, "overridden-name", tk.Name)
}

func TestRunCmd_SetOverrideWorkers(t *testing.T) {
	resetRunFlags(t)

	tmpFile := writeTestYAML(t, testRunYAML)
	runFile = tmpFile
	runDryRun = true
	runSetExprs = []string{"worker.replicas=8"}

	err := runCmd.RunE(runCmd, nil)
	assert.NoError(t, err)

	yamlData, err := os.ReadFile(tmpFile)
	require.NoError(t, err)
	mergedData, err := task.ApplySetOverrides(yamlData, runSetExprs)
	require.NoError(t, err)
	tk, err := task.LoadFromBytes(mergedData)
	require.NoError(t, err)
	assert.Equal(t, 8, tk.Worker.Replicas)
}

func TestRunCmd_SetOverrideGPUs(t *testing.T) {
	resetRunFlags(t)

	tmpFile := writeTestYAML(t, testRunYAML)
	runFile = tmpFile
	runDryRun = true
	runSetExprs = []string{"worker.resources.gpu=4"}

	err := runCmd.RunE(runCmd, nil)
	assert.NoError(t, err)

	yamlData, err := os.ReadFile(tmpFile)
	require.NoError(t, err)
	mergedData, err := task.ApplySetOverrides(yamlData, runSetExprs)
	require.NoError(t, err)
	tk, err := task.LoadFromBytes(mergedData)
	require.NoError(t, err)
	assert.Equal(t, "4", tk.Worker.Resources["gpu"])
}

func TestRunCmd_SetOverrideEnvs(t *testing.T) {
	resetRunFlags(t)

	tmpFile := writeTestYAML(t, testRunYAML)
	runFile = tmpFile
	runDryRun = true
	runSetExprs = []string{"envs.MY_VAR=hello", "envs.OTHER=world"}

	err := runCmd.RunE(runCmd, nil)
	assert.NoError(t, err)

	yamlData, err := os.ReadFile(tmpFile)
	require.NoError(t, err)
	mergedData, err := task.ApplySetOverrides(yamlData, runSetExprs)
	require.NoError(t, err)
	tk, err := task.LoadFromBytes(mergedData)
	require.NoError(t, err)
	assert.Equal(t, "hello", tk.Envs["MY_VAR"].Value)
	assert.Equal(t, "world", tk.Envs["OTHER"].Value)
}

func TestRunCmd_SetMultipleOverrides(t *testing.T) {
	resetRunFlags(t)

	tmpFile := writeTestYAML(t, testRunYAML)
	runFile = tmpFile
	runDryRun = true
	runSetExprs = []string{
		"name=multi-override",
		"image=pytorch:2.0",
		"worker.replicas=16",
		"worker.resources.gpu=8",
		"worker.resources.cpu=16",
		"worker.resources.memory=64Gi",
		"scheduling.gang.enabled=true",
	}

	err := runCmd.RunE(runCmd, nil)
	assert.NoError(t, err)

	yamlData, err := os.ReadFile(tmpFile)
	require.NoError(t, err)
	mergedData, err := task.ApplySetOverrides(yamlData, runSetExprs)
	require.NoError(t, err)
	tk, err := task.LoadFromBytes(mergedData)
	require.NoError(t, err)
	assert.Equal(t, "multi-override", tk.Name)
	assert.Equal(t, "pytorch:2.0", tk.Image)
	assert.Equal(t, 16, tk.Worker.Replicas)
	assert.Equal(t, "8", tk.Worker.Resources["gpu"])
	assert.Equal(t, "16", tk.Worker.Resources["cpu"])
	assert.Equal(t, "64Gi", tk.Worker.Resources["memory"])
	assert.True(t, tk.Scheduling.Gang.Enabled)
}

func TestRunCmd_SetOverrideInvalidValue(t *testing.T) {
	resetRunFlags(t)

	tmpFile := writeTestYAML(t, testRunYAML)
	runFile = tmpFile
	runDryRun = true
	runSetExprs = []string{"restart=InvalidPolicy"}

	err := runCmd.RunE(runCmd, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")
}

func TestRunCmd_SetOverrideInvalidSyntax(t *testing.T) {
	resetRunFlags(t)

	tmpFile := writeTestYAML(t, testRunYAML)
	runFile = tmpFile
	runDryRun = true
	runSetExprs = []string{"=nokey"}

	err := runCmd.RunE(runCmd, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to apply --set")
}

func TestRunCmd_SetOverrideDeeplyNested(t *testing.T) {
	resetRunFlags(t)

	yaml := `name: deep-test
image: pytorch:2.1
run: python train.py
framework:
  name: pytorch
worker:
  replicas: 1
scheduling:
  affinity:
    policy: spread
    constraint: preferred
    target: pod
`
	tmpFile := writeTestYAML(t, yaml)
	runFile = tmpFile
	runDryRun = true
	runSetExprs = []string{"scheduling.affinity.policy=binpack"}

	err := runCmd.RunE(runCmd, nil)
	assert.NoError(t, err)

	yamlData, err := os.ReadFile(tmpFile)
	require.NoError(t, err)
	mergedData, err := task.ApplySetOverrides(yamlData, runSetExprs)
	require.NoError(t, err)
	tk, err := task.LoadFromBytes(mergedData)
	require.NoError(t, err)
	assert.Equal(t, "binpack", tk.Scheduling.Affinity.Policy)
	assert.Equal(t, "preferred", tk.Scheduling.Affinity.Constraint)
}

func TestRunCmd_DryRunWithSetOverridesEndToEnd(t *testing.T) {
	resetRunFlags(t)

	tmpFile := writeTestYAML(t, testRunYAML)
	runFile = tmpFile
	runDryRun = true
	runSetExprs = []string{"name=e2e-override", "worker.resources.gpu=2"}

	err := runCmd.RunE(runCmd, nil)
	assert.NoError(t, err)

	yamlData, err := os.ReadFile(tmpFile)
	require.NoError(t, err)
	mergedData, err := task.ApplySetOverrides(yamlData, runSetExprs)
	require.NoError(t, err)
	tk, err := task.LoadFromBytes(mergedData)
	require.NoError(t, err)
	assert.Equal(t, "e2e-override", tk.Name)
	assert.Equal(t, "2", tk.Worker.Resources["gpu"])
}

func TestRunCmd_SetOverrideValueAssertion(t *testing.T) {
	resetRunFlags(t)

	tmpFile := writeTestYAML(t, testRunYAML)

	yamlData, err := os.ReadFile(tmpFile)
	require.NoError(t, err)

	mergedData, err := task.ApplySetOverrides(yamlData, []string{"worker.replicas=8", "name=asserted-name"})
	require.NoError(t, err)

	tk, err := task.LoadFromBytes(mergedData)
	require.NoError(t, err)

	assert.Equal(t, "asserted-name", tk.Name)
	assert.Equal(t, 8, tk.Worker.Replicas)
}

func TestRunCmd_MPIVersionIntegration_V1FromCluster(t *testing.T) {
	// Simulates the run command flow when the cluster has MPIJob CRD with
	// storage version v1. Verifies that the generated CR uses kubeflow.org/v1.
	yamlContent := `name: mpi-v1-integration
image: openmpi:4.1
run: mpirun -np 4 ./train
framework:
  name: mpi
worker:
  replicas: 4
`
	tmpFile := writeTestYAML(t, yamlContent)

	tk, err := task.LoadFromFile(tmpFile)
	require.NoError(t, err)

	p, err := getProvider(tk.Framework.Name)
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
	_, hasLauncherPolicy := spec["launcherCreationPolicy"]
	assert.False(t, hasLauncherPolicy, "v1 CR should not have launcherCreationPolicy")
}

func TestRunCmd_MPIVersionIntegration_V2beta1Default(t *testing.T) {
	// Simulates the run command flow in dry-run mode (no cluster).
	// With the removal of the default fallback, APIVersion must be explicitly set.
	yamlContent := `name: mpi-v2beta1-default
image: openmpi:4.1
run: mpirun -np 4 ./train
framework:
  name: mpi
worker:
  replicas: 2
`
	tmpFile := writeTestYAML(t, yamlContent)

	tk, err := task.LoadFromFile(tmpFile)
	require.NoError(t, err)

	p, err := getProvider(tk.Framework.Name)
	require.NoError(t, err)

	mpiP, ok := p.(*provider.MPIProvider)
	require.True(t, ok, "expected MPIProvider for mpi framework")
	assert.Empty(t, mpiP.APIVersion, "APIVersion should be empty (unset) before BuildCRD")

	// Empty APIVersion must now produce an error (no default fallback)
	_, err = p.BuildCRD(tk)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "APIVersion must be set")

	// After setting APIVersion, BuildCRD succeeds with v2beta1
	mpiP.APIVersion = provider.MPIAPIVersionV2beta1
	crd, err := p.BuildCRD(tk)
	require.NoError(t, err)

	assert.Equal(t, "MPIJob", crd.GetKind())
	assert.Equal(t, "kubeflow.org/v2beta1", crd.GetAPIVersion())

	// v2beta1 CR should have v2beta1-only fields
	spec := crd.Object["spec"].(map[string]interface{})
	_, hasSSHAuth := spec["sshAuthMountPath"]
	assert.True(t, hasSSHAuth, "v2beta1 CR should have sshAuthMountPath")
}

func TestRunCmd_MPIVersionIntegration_HorovodUsesDetectedVersion(t *testing.T) {
	// Verifies that horovod (MPI-family) also uses the detected version.
	yamlContent := `name: horovod-v1
image: horovod:latest
run: mpirun train
framework:
  name: horovod
worker:
  replicas: 2
`
	tmpFile := writeTestYAML(t, yamlContent)

	tk, err := task.LoadFromFile(tmpFile)
	require.NoError(t, err)

	assert.True(t, isMPIFamily(tk.Framework.Name))

	p, err := getProvider(tk.Framework.Name)
	require.NoError(t, err)

	mpiP, ok := p.(*provider.MPIProvider)
	require.True(t, ok, "expected MPIProvider for horovod framework")
	mpiP.APIVersion = "v1"

	crd, err := p.BuildCRD(tk)
	require.NoError(t, err)

	assert.Equal(t, "kubeflow.org/v1", crd.GetAPIVersion())
}

func TestRunCmd_SetOverrideQuotedKeyValueAssertion(t *testing.T) {
	resetRunFlags(t)

	tmpFile := writeTestYAML(t, testRunYAML)

	yamlData, err := os.ReadFile(tmpFile)
	require.NoError(t, err)

	mergedData, err := task.ApplySetOverrides(yamlData, []string{"worker.resources.'nvidia.com/gpu'=4"})
	require.NoError(t, err)

	tk, err := task.LoadFromBytes(mergedData)
	require.NoError(t, err)

	assert.Equal(t, "4", tk.Worker.Resources["nvidia.com/gpu"])
}

func TestRunCmd_SetOverrideTypeMismatch(t *testing.T) {
	resetRunFlags(t)

	tmpFile := writeTestYAML(t, testRunYAML)
	runFile = tmpFile
	runDryRun = true
	runSetExprs = []string{"worker.replicas=notanint"}

	err := runCmd.RunE(runCmd, nil)
	assert.Error(t, err)
	// The --set parser stores "notanint" as a string; YAML unmarshal into int field fails
	// during LoadFromBytes, which wraps the error.
	assert.Contains(t, err.Error(), "failed to load task")
}

func TestRunCmd_SetOverrideWithNamespace(t *testing.T) {
	resetRunFlags(t)

	// Add namespace to the YAML
	yamlWithNS := "namespace: my-ns\n" + testRunYAML
	tmpFile := writeTestYAML(t, yamlWithNS)
	runFile = tmpFile
	runDryRun = true
	runSetExprs = []string{"name=ns-override"}

	err := runCmd.RunE(runCmd, nil)
	assert.NoError(t, err)

	// Verify both the override and the namespace reach the Task
	yamlData, err := os.ReadFile(tmpFile)
	require.NoError(t, err)
	mergedData, err := task.ApplySetOverrides(yamlData, runSetExprs)
	require.NoError(t, err)
	tk, err := task.LoadFromBytes(mergedData)
	require.NoError(t, err)
	assert.Equal(t, "ns-override", tk.Name)
	assert.Equal(t, "my-ns", tk.Namespace)
}

func TestRunCmd_SetOverrideMPIJobWithSet(t *testing.T) {
	// Verifies that --set overrides work correctly on MPI jobs and that
	// the merged YAML produces a valid Task that can be used with MPIProvider.
	yamlContent := `name: mpi-set-test
image: openmpi:4.1
run: mpirun -np 4 ./train
framework:
  name: mpi
worker:
  replicas: 4
`
	yamlData := []byte(yamlContent)
	mergedData, err := task.ApplySetOverrides(yamlData, []string{
		"name=mpi-overridden",
		"worker.replicas=8",
	})
	require.NoError(t, err)

	tk, err := task.LoadFromBytes(mergedData)
	require.NoError(t, err)
	assert.Equal(t, "mpi-overridden", tk.Name)
	assert.Equal(t, 8, tk.Worker.Replicas)
	assert.Equal(t, "mpi", tk.Framework.Name)

	// Verify the merged task produces a valid MPI CRD
	p, err := getProvider(tk.Framework.Name)
	require.NoError(t, err)
	mpiP, ok := p.(*provider.MPIProvider)
	require.True(t, ok)
	mpiP.APIVersion = provider.MPIAPIVersionV2beta1

	crd, err := p.BuildCRD(tk)
	require.NoError(t, err)
	assert.Equal(t, "MPIJob", crd.GetKind())
	assert.Equal(t, "kubeflow.org/v2beta1", crd.GetAPIVersion())
	assert.Equal(t, "mpi-overridden", crd.GetName())
}
