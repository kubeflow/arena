package cli

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kubeflow/arena/pkg/log"
	"github.com/kubeflow/arena/pkg/task"
)

// v1RenamedFlags are v1 flag names that must be registered on the submit
// command.
var v1RenamedFlags = []string{
	"cpu",
	"memory",
	"scheduler",
	"logdir",
	"clean-task-policy",
	"running-timeout",
	"share-memory",
	"job-restart-policy",
	"job-backoff-limit",
	"retry",
	"ps",
	"gputopology",
	"hostNetwork",
	"hostIPC",
	"hostPID",
}

// wrongV2Names are v2-invented flag names that diverged from v1 and must not
// exist on the submit command.
var wrongV2Names = []string{
	"cpus",
	"mem",
	"scheduler-name",
	"tensorboard-logdir",
	"clean-pod-policy",
	"active-deadline",
	"shm",
	"restart",
	"backoff-limit",
	"ps-count",
	"gpu-topology",
	"host-network",
	"host-ipc",
	"host-pid",
	"priority-class-name",
	// -limit variants that never existed in v1
	"cpu-limit",
	"memory-limit",
	"gpus-limit",
	"ps-gpus-limit",
}

func TestSubmitCompat_V1FlagNamesRegistered(t *testing.T) {
	cmd := newSubmitCmd()
	for _, name := range v1RenamedFlags {
		f := cmd.Flags().Lookup(name)
		require.NotNil(t, f, "v1 flag --%s should be registered", name)
	}
}

func TestSubmitCompat_WrongV2FlagNamesRemoved(t *testing.T) {
	cmd := newSubmitCmd()
	for _, name := range wrongV2Names {
		assert.Nil(t, cmd.Flags().Lookup(name),
			"--%s must not be registered; the v1 name is the correct name", name)
	}
}

func TestSubmitCompat_V1LimitVariantsRegistered(t *testing.T) {
	cmd := newSubmitCmd()
	// v1 TFJob exposed -limit variants for per-role resources. They bind to their own variables and override only limits; requests come from the request flag.
	limits := map[string]string{
		"ps-cpu-limit":           "ps-cpu",
		"ps-memory-limit":        "ps-memory",
		"chief-cpu-limit":        "chief-cpu",
		"chief-memory-limit":     "chief-memory",
		"evaluator-cpu-limit":    "evaluator-cpu",
		"evaluator-memory-limit": "evaluator-memory",
		"worker-cpu-limit":       "worker-cpu",
		"worker-memory-limit":    "worker-memory",
	}
	for limit, base := range limits {
		f := cmd.Flags().Lookup(limit)
		require.NotNil(t, f, "v1 flag --%s should be registered", limit)
		assert.NotNil(t, cmd.Flags().Lookup(base), "--%s should also be registered", base)
	}
}

func TestSubmitCompat_V1FlagsBindToVariables(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		parseErr bool
		check    func(t *testing.T)
	}{
		{
			name:  "--cpu sets the cpu variable",
			args:  []string{"--cpu", "4"},
			check: func(t *testing.T) { assert.Equal(t, "4", submitCPU) },
		},
		{
			name:  "--memory sets the memory variable",
			args:  []string{"--memory", "8Gi"},
			check: func(t *testing.T) { assert.Equal(t, "8Gi", submitMemory) },
		},
		{
			name:  "--scheduler sets the scheduler variable",
			args:  []string{"--scheduler", "volcano"},
			check: func(t *testing.T) { assert.Equal(t, "volcano", submitScheduler) },
		},
		{
			// Distinct from the resetSubmitFlags default (/training_logs) so the
			// subtest proves parsing, not the pre-loaded default.
			name:  "--logdir sets the logdir variable",
			args:  []string{"--logdir", "/custom/logs"},
			check: func(t *testing.T) { assert.Equal(t, "/custom/logs", submitLogDir) },
		},
		{
			name:  "--clean-task-policy sets the clean task policy variable",
			args:  []string{"--clean-task-policy", "None"},
			check: func(t *testing.T) { assert.Equal(t, "None", submitCleanTaskPolicy) },
		},
		{
			name:  "--running-timeout sets the running timeout variable",
			args:  []string{"--running-timeout", "2h22m"},
			check: func(t *testing.T) { assert.Equal(t, "2h22m", submitRunningTimeout) },
		},
		{
			name:  "--job-backoff-limit sets the backoff variable",
			args:  []string{"--job-backoff-limit", "3"},
			check: func(t *testing.T) { assert.Equal(t, 3, submitJobBackoffLimit) },
		},
		{
			name:  "--retry sets the backoff variable",
			args:  []string{"--retry", "5"},
			check: func(t *testing.T) { assert.Equal(t, 5, submitJobBackoffLimit) },
		},
		{
			name:  "retry and job-backoff-limit share one variable (last wins)",
			args:  []string{"--retry", "5", "--job-backoff-limit", "2"},
			check: func(t *testing.T) { assert.Equal(t, 2, submitJobBackoffLimit) },
		},
		{
			// Distinct from the resetSubmitFlags default (2Gi) so the subtest
			// proves parsing, not the pre-loaded default.
			name:  "--share-memory sets the share memory variable",
			args:  []string{"--share-memory", "4Gi"},
			check: func(t *testing.T) { assert.Equal(t, "4Gi", submitShareMemory) },
		},
		{
			name:  "--ps sets the ps count variable",
			args:  []string{"--ps", "2"},
			check: func(t *testing.T) { assert.Equal(t, 2, submitPS) },
		},
		{
			name:  "--gputopology sets the gpu topology variable",
			args:  []string{"--gputopology"},
			check: func(t *testing.T) { assert.True(t, submitGPUTopology) },
		},
		{
			name:  "--hostNetwork sets the host network variable",
			args:  []string{"--hostNetwork"},
			check: func(t *testing.T) { assert.True(t, submitHostNetwork) },
		},
		{
			name:  "--hostIPC sets the host ipc variable",
			args:  []string{"--hostIPC"},
			check: func(t *testing.T) { assert.True(t, submitHostIPC) },
		},
		{
			name:  "--hostPID sets the host pid variable",
			args:  []string{"--hostPID"},
			check: func(t *testing.T) { assert.True(t, submitHostPID) },
		},
		{
			name:  "--job-restart-policy sets the restart variable",
			args:  []string{"--job-restart-policy", "OnFailure"},
			check: func(t *testing.T) { assert.Equal(t, "OnFailure", submitJobRestartPolicy) },
		},
		{
			name:  "--ps-cpu-limit sets the ps cpu limit variable",
			args:  []string{"--ps-cpu-limit", "2"},
			check: func(t *testing.T) { assert.Equal(t, "2", submitPSCPULimit) },
		},
		{
			name:  "--worker-memory-limit sets the worker memory limit variable",
			args:  []string{"--worker-memory-limit", "16Gi"},
			check: func(t *testing.T) { assert.Equal(t, "16Gi", submitWorkerMemoryLimit) },
		},
		{
			name:     "--cpu-limit is not a v1 flag and must not parse",
			args:     []string{"--cpu-limit", "2"},
			check:    func(t *testing.T) { t.Error("should not reach here") },
			parseErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newSubmitCmd()
			err := cmd.Flags().Parse(tt.args)
			if tt.parseErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			tt.check(t)
		})
	}
}

func TestSubmitCompat_PriorityIsV1ClassName(t *testing.T) {
	t.Run("--priority takes a class name like v1", func(t *testing.T) {
		cmd := newSubmitCmd()
		require.NoError(t, cmd.Flags().Parse([]string{"--priority", "high-priority"}))
		assert.Equal(t, "high-priority", submitPriorityClass)
	})

	t.Run("-p shorthand is preserved", func(t *testing.T) {
		cmd := newSubmitCmd()
		require.NoError(t, cmd.Flags().Parse([]string{"-p", "high"}))
		assert.Equal(t, "high", submitPriorityClass)
	})

	t.Run("--priority is a string flag", func(t *testing.T) {
		f := newSubmitCmd().Flags().Lookup("priority")
		require.NotNil(t, f)
		assert.Equal(t, "string", f.Value.Type())
	})
}

func TestSubmitCompat_QueueIsV1Bool(t *testing.T) {
	t.Run("--queue is a bool flag like v1", func(t *testing.T) {
		f := newSubmitCmd().Flags().Lookup("queue")
		require.NotNil(t, f)
		assert.Equal(t, "bool", f.Value.Type())
	})

	t.Run("bare --queue parses like a v1 invocation", func(t *testing.T) {
		cmd := newSubmitCmd()
		require.NoError(t, cmd.Flags().Parse([]string{"--queue"}))
		assert.True(t, submitQueue)
	})

	t.Run("--queue suspends the job for kube-queue", func(t *testing.T) {
		cmd := newSubmitCmd()
		require.NoError(t, cmd.Flags().Parse([]string{"--queue"}))

		tk := buildSubmitTask("pytorch", nil)
		flags := buildSubmitFlags()
		require.NoError(t, task.ApplyOverrides(tk, flags))

		require.NotNil(t, tk.Lifecycle.Suspend, "--queue should set suspend")
		assert.True(t, *tk.Lifecycle.Suspend)
	})
}

func TestSubmitCompat_RepeatableFlagsAreStringArrays(t *testing.T) {
	cmd := newSubmitCmd()
	names := []string{
		"env", "data", "data-dir", "config-file",
		"label", "annotation", "selector", "toleration",
		"device", "image-pull-secret",
	}
	for _, name := range names {
		f := cmd.Flags().Lookup(name)
		require.NotNil(t, f, "--%s should be registered", name)
		assert.Equal(t, "stringArray", f.Value.Type(),
			"--%s must be a stringArray flag (v1 semantics, no comma splitting)", name)
	}
}

func TestSubmitCompat_CommaValuesNotSplit(t *testing.T) {
	cmd := newSubmitCmd()
	require.NoError(t, cmd.Flags().Parse([]string{
		"--env", "A=1,2",
		"--data", "ds:/data",
	}))

	assert.Equal(t, []string{"A=1,2"}, submitEnvs,
		"--env value with a comma must stay one entry like v1")
	assert.Equal(t, []string{"ds:/data"}, submitData)
}

func TestSubmitCompat_LimitFlagsHaveOwnVariables(t *testing.T) {
	cmd := newSubmitCmd()
	require.NoError(t, cmd.Flags().Parse([]string{
		"--ps-cpu", "1", "--ps-cpu-limit", "2",
		"--worker-cpu", "1", "--worker-cpu-limit", "3",
	}))
	assert.Equal(t, "1", submitPSCPU)
	assert.Equal(t, "2", submitPSCPULimit)
	assert.Equal(t, "1", submitWorkerCPU)
	assert.Equal(t, "3", submitWorkerCPULimit)
}

// v1OfficialMissingFlags are the v1-official submit flags (visible in
// `arena v0.15.4 submit <type> -h`) that arena-v2 had not registered.
var v1OfficialMissingFlags = []string{
	"config", "loglevel", "rdma",
	// per-role flags: accepted, warned, ignored until RoleConfig grows
	// selector/annotation/image/port fields
	"ps-selector", "worker-selector", "chief-selector", "evaluator-selector", "launcher-selector",
	"worker-annotation", "launcher-annotation",
	"ps-image", "worker-image",
	"ps-port", "worker-port", "chief-port", "ssh-port",
	"ps-affinity-policy", "ps-affinity-constraint",
	"worker-affinity-policy", "worker-affinity-constraint",
	// flags with no v2 equivalent
	"pprof", "trace", "helm-binary", "arena-namespace",
	"model-name", "model-source", "starting-timeout", "role-sequence", "ssh-secret",
}

func TestSubmitCompat_V1OfficialFlagsRegistered(t *testing.T) {
	cmd := newSubmitCmd()
	for _, name := range v1OfficialMissingFlags {
		f := cmd.Flags().Lookup(name)
		require.NotNil(t, f, "v1 official flag --%s should be registered", name)
		assert.True(t, f.Hidden, "--%s is deprecated and must be hidden from help", name)
	}
}

func TestSubmitCompat_ConfigBindsKubeconfig(t *testing.T) {
	oldKubeconfig := kubeconfig
	t.Cleanup(func() { kubeconfig = oldKubeconfig })

	cmd := newSubmitCmd()
	require.NoError(t, cmd.Flags().Parse([]string{"--config", "/tmp/kube-config"}))
	assert.Equal(t, "/tmp/kube-config", kubeconfig)
}

func TestSubmitCompat_DeprecatedFlagsParse(t *testing.T) {
	cmd := newSubmitCmd()
	require.NoError(t, cmd.Flags().Parse([]string{
		"--rdma",
		"--ps-selector", "gpu=a100",
		"--worker-image", "tf:2.15",
		"--model-name", "my-model",
	}))
	// Deprecated no-op flags must not disturb the task build path.
	tk := buildSubmitTask("pytorch", nil)
	flags := buildSubmitFlags()
	require.NoError(t, task.ApplyOverrides(tk, flags))
	assert.Empty(t, tk.Annotations, "no-op flags must not leak into the task")
}

func TestApplyV1LogLevel(t *testing.T) {
	// applyV1LogLevel mutates klog verbosity on flag.CommandLine; restore the
	// default afterwards so a mid-test failure cannot leak state.
	t.Cleanup(func() {
		if err := log.SetVerbosity(flag.CommandLine, 0); err != nil {
			t.Errorf("failed to restore verbosity to 0: %v", err)
		}
	})
	vFlag := flag.CommandLine.Lookup("v")
	require.NotNil(t, vFlag, "klog -v flag should be registered on flag.CommandLine")

	require.NoError(t, applyV1LogLevel(""))
	require.NoError(t, applyV1LogLevel("debug"))
	assert.Equal(t, "2", vFlag.Value.String(), "debug should map to verbosity 2")
	require.NoError(t, applyV1LogLevel("info"))
	assert.Equal(t, "1", vFlag.Value.String(), "info should map to verbosity 1")
	require.NoError(t, applyV1LogLevel("warn"))
	assert.Equal(t, "0", vFlag.Value.String(), "warn should map to verbosity 0")
	require.NoError(t, applyV1LogLevel("error"))
	assert.Equal(t, "0", vFlag.Value.String(), "error should map to verbosity 0")
	err := applyV1LogLevel("bogus")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must be debug, info, warn, or error")
}
