package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kubeflow/arena/pkg/constants"
)

func TestLogsCmd_ArgsValidator(t *testing.T) {
	assert.Error(t, logsCmd.Args(logsCmd, []string{}))
	assert.NoError(t, logsCmd.Args(logsCmd, []string{"my-job"}))
	assert.Error(t, logsCmd.Args(logsCmd, []string{"a", "b"}))
}

func TestLogsCmd_FollowFlag(t *testing.T) {
	f := logsCmd.Flags().Lookup("follow")
	require.NotNil(t, f, "expected --follow flag to be registered")
	assert.Equal(t, "f", f.Shorthand)
	assert.Equal(t, "false", f.DefValue)
}

func TestLogsCmd_TailFlag(t *testing.T) {
	f := logsCmd.Flags().Lookup("tail")
	require.NotNil(t, f, "expected --tail flag to be registered")
	assert.Equal(t, "-1", f.DefValue)
}

func TestLogsCmd_RegisteredOnJob(t *testing.T) {
	found := false
	for _, cmd := range jobCmd.Commands() {
		if cmd.Use == "logs <name>" {
			found = true
			break
		}
	}
	assert.True(t, found, "logs command should be registered on job command")
}

func TestLogsCmd_RunE_FailsWithInvalidKubeconfig(t *testing.T) {
	orig := kubeconfig
	defer func() { kubeconfig = orig }()

	kubeconfig = "/nonexistent/kubeconfig"
	err := logsCmd.RunE(logsCmd, []string{"my-job"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to build config")
}

func TestLogsCmd_HasCorrectUse(t *testing.T) {
	assert.Equal(t, "logs <name>", logsCmd.Use)
	assert.NotEmpty(t, logsCmd.Short)
}

func TestLogsCmd_RunE_RequiresKubeconfig(t *testing.T) {
	orig := kubeconfig
	defer func() { kubeconfig = orig }()

	t.Setenv("KUBECONFIG", "/nonexistent/env-kubeconfig")
	kubeconfig = ""
	err := logsCmd.RunE(logsCmd, []string{"my-job"})
	// Without a valid kubeconfig, client creation or REST config should fail.
	assert.Error(t, err)
}

func TestLogsCmd_PodFlag(t *testing.T) {
	f := logsCmd.Flags().Lookup("pod")
	require.NotNil(t, f, "expected --pod flag to be registered")
	assert.Equal(t, "", f.DefValue)
}

func TestLogsCmd_ContainerFlag(t *testing.T) {
	f := logsCmd.Flags().Lookup("container")
	require.NotNil(t, f, "expected --container flag to be registered")
	assert.Equal(t, "", f.DefValue)
}

func TestPodBelongsToJob(t *testing.T) {
	tests := []struct {
		name    string
		pod     *corev1.Pod
		jobName string
		want    bool
	}{
		{
			name: "matching job-name label",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-job-master-0",
					Labels: map[string]string{
						constants.LabelJobName: "my-job",
					},
				},
			},
			jobName: "my-job",
			want:    true,
		},
		{
			name: "mismatched job-name label",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: "other-job-master-0",
					Labels: map[string]string{
						constants.LabelJobName: "other-job",
					},
				},
			},
			jobName: "my-job",
			want:    false,
		},
		{
			name: "nil labels",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: "bare-pod",
				},
			},
			jobName: "my-job",
			want:    false,
		},
		{
			name: "empty labels map",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "labeled-pod",
					Labels: map[string]string{},
				},
			},
			jobName: "my-job",
			want:    false,
		},
		{
			name: "label present but empty value",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: "empty-label-pod",
					Labels: map[string]string{
						constants.LabelJobName: "",
					},
				},
			},
			jobName: "my-job",
			want:    false,
		},
		{
			name: "other labels present but not job-name",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: "partial-labels-pod",
					Labels: map[string]string{
						constants.LabelReplicaType:  constants.ReplicaRoleMaster,
						constants.LabelReplicaIndex: "0",
					},
				},
			},
			jobName: "my-job",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := podBelongsToJob(tt.pod, tt.jobName)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestContainerExists(t *testing.T) {
	tests := []struct {
		name          string
		pod           *corev1.Pod
		containerName string
		want          bool
	}{
		{
			name: "regular container found",
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "pytorch"},
						{Name: "tensorboard"},
					},
				},
			},
			containerName: "pytorch",
			want:          true,
		},
		{
			name: "init container found",
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{
						{Name: "arena-sync-0"},
					},
					Containers: []corev1.Container{
						{Name: "pytorch"},
					},
				},
			},
			containerName: "arena-sync-0",
			want:          true,
		},
		{
			name: "container not found",
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "pytorch"},
					},
				},
			},
			containerName: "sidecar",
			want:          false,
		},
		{
			name: "empty pod spec",
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{},
			},
			containerName: "pytorch",
			want:          false,
		},
		{
			name: "second regular container",
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "main"},
						{Name: "sidecar"},
						{Name: "monitor"},
					},
				},
			},
			containerName: "sidecar",
			want:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containerExists(tt.pod, tt.containerName)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetAvailableContainers(t *testing.T) {
	tests := []struct {
		name string
		pod  *corev1.Pod
		want []string
	}{
		{
			name: "regular containers only",
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "pytorch"},
						{Name: "tensorboard"},
					},
				},
			},
			want: []string{"pytorch", "tensorboard"},
		},
		{
			name: "init and regular containers",
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{
						{Name: "arena-sync-0"},
					},
					Containers: []corev1.Container{
						{Name: "pytorch"},
					},
				},
			},
			want: []string{"arena-sync-0", "pytorch"},
		},
		{
			name: "no containers",
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{},
			},
			want: nil,
		},
		{
			name: "multiple init containers",
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{
						{Name: "arena-sync-0"},
						{Name: "arena-sync-1"},
					},
					Containers: []corev1.Container{
						{Name: "main"},
					},
				},
			},
			want: []string{"arena-sync-0", "arena-sync-1", "main"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getAvailableContainers(tt.pod)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetProvider_LogPodSelector(t *testing.T) {
	tests := []struct {
		name      string
		framework string
		jobName   string
		want      string
		wantErr   bool
	}{
		{
			name:      "pytorch selects master",
			framework: constants.FrameworkPyTorch,
			jobName:   "my-job",
			want:      constants.LabelJobName + "=my-job," + constants.LabelReplicaType + "=" + constants.ReplicaRoleMaster,
		},
		{
			name:      "tensorflow selects chief",
			framework: constants.FrameworkTensorFlow,
			jobName:   "tf-training",
			want:      constants.LabelJobName + "=tf-training," + constants.LabelReplicaType + "=" + constants.ReplicaRoleChief,
		},
		{
			name:      "mpi selects launcher",
			framework: constants.FrameworkMPI,
			jobName:   "mpi-run",
			want:      constants.LabelJobName + "=mpi-run," + constants.LabelReplicaType + "=" + constants.ReplicaRoleLauncher,
		},
		{
			name:      "horovod maps to mpi launcher",
			framework: constants.FrameworkHorovod,
			jobName:   "horovod-job",
			want:      constants.LabelJobName + "=horovod-job," + constants.LabelReplicaType + "=" + constants.ReplicaRoleLauncher,
		},
		{
			name:      "deepspeed maps to mpi launcher",
			framework: constants.FrameworkDeepSpeed,
			jobName:   "ds-job",
			want:      constants.LabelJobName + "=ds-job," + constants.LabelReplicaType + "=" + constants.ReplicaRoleLauncher,
		},
		{
			name:      "unsupported framework returns error",
			framework: "unsupported",
			wantErr:   true,
		},
		{
			name:      "ray not yet implemented",
			framework: constants.FrameworkRay,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := getProvider(tt.framework)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, p)

			got := p.GetLogPodSelector(tt.jobName)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFrameworkToKind(t *testing.T) {
	tests := []struct {
		framework string
		want      string
	}{
		{constants.FrameworkPyTorch, constants.KindPyTorchJob},
		{constants.FrameworkTensorFlow, constants.KindTFJob},
		{constants.FrameworkMPI, constants.KindMPIJob},
		{constants.FrameworkHorovod, constants.KindMPIJob},
		{constants.FrameworkDeepSpeed, constants.KindMPIJob},
		{"tf", constants.KindTFJob},
		{"unknown", ""},
	}

	for _, tt := range tests {
		t.Run(tt.framework, func(t *testing.T) {
			got := frameworkToKind(tt.framework)
			assert.Equal(t, tt.want, got)
		})
	}
}

// Additional edge case tests

func TestPodBelongsToJob_AdditionalCases(t *testing.T) {
	tests := []struct {
		name    string
		pod     *corev1.Pod
		jobName string
		want    bool
	}{
		{
			name: "job name with special characters",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-job-123_test",
					Labels: map[string]string{
						constants.LabelJobName: "my-job-123_test",
					},
				},
			},
			jobName: "my-job-123_test",
			want:    true,
		},
		{
			name: "case sensitivity check",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-job",
					Labels: map[string]string{
						constants.LabelJobName: "My-Job",
					},
				},
			},
			jobName: "my-job",
			want:    false, // case sensitive
		},
		{
			name: "multiple labels with job-name",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: "multi-label-pod",
					Labels: map[string]string{
						constants.LabelJobName:      "my-job",
						constants.LabelReplicaType:  constants.ReplicaRoleWorker,
						constants.LabelReplicaIndex: "1",
						"custom-label":              "value",
					},
				},
			},
			jobName: "my-job",
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := podBelongsToJob(tt.pod, tt.jobName)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestContainerExists_AdditionalCases(t *testing.T) {
	tests := []struct {
		name          string
		pod           *corev1.Pod
		containerName string
		want          bool
	}{
		{
			name: "ephemeral container",
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "main"},
					},
					EphemeralContainers: []corev1.EphemeralContainer{
						{EphemeralContainerCommon: corev1.EphemeralContainerCommon{Name: "debug"}},
					},
				},
			},
			containerName: "debug",
			want:          false, // ephemeral containers not checked
		},
		{
			name: "case sensitivity",
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "PyTorch"},
					},
				},
			},
			containerName: "pytorch",
			want:          false, // case sensitive
		},
		{
			name: "empty container name",
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: ""},
					},
				},
			},
			containerName: "",
			want:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containerExists(tt.pod, tt.containerName)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetProvider_LogPodSelector_SelectorFormat(t *testing.T) {
	// Verify the selector format matches expected Kubernetes label selector syntax
	p, err := getProvider(constants.FrameworkPyTorch)
	require.NoError(t, err)

	selector := p.GetLogPodSelector("test-job")

	// Should be a valid label selector string
	assert.Contains(t, selector, constants.LabelJobName+"=test-job")
	assert.Contains(t, selector, constants.LabelReplicaType+"="+constants.ReplicaRoleMaster)

	// Verify it's comma-separated (standard label selector format)
	parts := splitSelector(selector)
	assert.Len(t, parts, 2, "selector should have exactly 2 label requirements")
}

func splitSelector(selector string) []string {
	if selector == "" {
		return nil
	}
	result := []string{}
	for _, part := range splitByComma(selector) {
		trimmed := trimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func splitByComma(s string) []string {
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == ',' {
			result = append(result, s[start:i])
			start = i + 1
		}
	}
	result = append(result, s[start:])
	return result
}

func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}

func TestGetProvider_AllFrameworks(t *testing.T) {
	// Test that all supported frameworks can get a provider
	frameworks := []string{
		constants.FrameworkPyTorch,
		constants.FrameworkTensorFlow,
		constants.FrameworkMPI,
		constants.FrameworkHorovod,
		constants.FrameworkDeepSpeed,
	}

	for _, fw := range frameworks {
		t.Run(fw, func(t *testing.T) {
			p, err := getProvider(fw)
			require.NoError(t, err, "framework %s should have a provider", fw)
			require.NotNil(t, p)

			// Verify the provider can generate a log selector
			selector := p.GetLogPodSelector("test-job")
			assert.NotEmpty(t, selector)
			assert.Contains(t, selector, constants.LabelJobName+"=test-job")
		})
	}
}

func TestLogsCmd_BufferConfiguration(t *testing.T) {
	// Verify that the logs command exists and has proper configuration
	// The actual buffer size is set in the RunE function, but we can verify
	// the command structure is correct
	assert.Equal(t, "logs <name>", logsCmd.Use)
	assert.NotNil(t, logsCmd.RunE)

	// Verify all expected flags are present
	flags := []string{"follow", "tail", "pod", "container"}
	for _, flagName := range flags {
		f := logsCmd.Flags().Lookup(flagName)
		require.NotNil(t, f, "expected flag %s to be registered", flagName)
	}
}
