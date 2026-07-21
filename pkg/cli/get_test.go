package cli

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"

	"github.com/kubeflow/arena/pkg/client"
	"github.com/kubeflow/arena/pkg/task"
)

func TestGetCmd_HasDetailsFlag(t *testing.T) {
	flag := getCmd.Flags().Lookup("details")
	assert.NotNil(t, flag, "get command should have --details flag")
	assert.Equal(t, "false", flag.DefValue, "--details default should be false")
}

func TestGetCmd_DetailsFlagDescription(t *testing.T) {
	flag := getCmd.Flags().Lookup("details")
	assert.NotNil(t, flag)
	assert.Equal(t, "show job configuration details", flag.Usage)
}

func TestJobInfo_ConfigurationField(t *testing.T) {
	// Verify JobInfo has a Configuration field that is nil by default
	info := &client.JobInfo{
		Status: client.JobStatus{
			Name:      "test-job",
			Namespace: "default",
			Status:    "Running",
			Replicas:  1,
			Ready:     1,
			Age:       "1m",
		},
	}
	assert.Nil(t, info.Configuration, "Configuration should be nil when not set")
}

func TestJobInfo_ConfigurationFieldSet(t *testing.T) {
	config := &task.Task{
		Name:  "test-job",
		Image: "pytorch:2.1",
		Run:   "python train.py",
		Framework: task.Framework{
			Name: "pytorch",
		},
	}
	info := &client.JobInfo{
		Status: client.JobStatus{
			Name: "test-job",
		},
		Configuration: config,
	}
	assert.NotNil(t, info.Configuration)
	assert.Equal(t, "test-job", info.Configuration.Name)
	assert.Equal(t, "pytorch:2.1", info.Configuration.Image)
}

func TestPodDisplayStatus(t *testing.T) {
	tests := []struct {
		name     string
		pod      corev1.Pod
		expected string
	}{
		{
			name: "running phase with container running",
			pod: corev1.Pod{
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
					ContainerStatuses: []corev1.ContainerStatus{
						{
							State: corev1.ContainerState{
								Running: &corev1.ContainerStateRunning{},
							},
						},
					},
				},
			},
			expected: "Running",
		},
		{
			name: "pending phase with init container ImagePullBackOff",
			pod: corev1.Pod{
				Status: corev1.PodStatus{
					Phase: corev1.PodPending,
					InitContainerStatuses: []corev1.ContainerStatus{
						{
							State: corev1.ContainerState{
								Waiting: &corev1.ContainerStateWaiting{
									Reason: "ImagePullBackOff",
								},
							},
						},
					},
				},
			},
			expected: "ImagePullBackOff",
		},
		{
			name: "pending phase with container ContainerCreating",
			pod: corev1.Pod{
				Status: corev1.PodStatus{
					Phase: corev1.PodPending,
					ContainerStatuses: []corev1.ContainerStatus{
						{
							State: corev1.ContainerState{
								Waiting: &corev1.ContainerStateWaiting{
									Reason: "ContainerCreating",
								},
							},
						},
					},
				},
			},
			expected: "ContainerCreating",
		},
		{
			name: "running phase with container CrashLoopBackOff",
			pod: corev1.Pod{
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
					ContainerStatuses: []corev1.ContainerStatus{
						{
							State: corev1.ContainerState{
								Waiting: &corev1.ContainerStateWaiting{
									Reason: "CrashLoopBackOff",
								},
							},
						},
					},
				},
			},
			expected: "CrashLoopBackOff",
		},
		{
			name: "succeeded phase",
			pod: corev1.Pod{
				Status: corev1.PodStatus{
					Phase: corev1.PodSucceeded,
					ContainerStatuses: []corev1.ContainerStatus{
						{
							State: corev1.ContainerState{
								Terminated: &corev1.ContainerStateTerminated{
									ExitCode: 0,
								},
							},
						},
					},
				},
			},
			expected: "Succeeded",
		},
		{
			name: "failed phase",
			pod: corev1.Pod{
				Status: corev1.PodStatus{
					Phase: corev1.PodFailed,
					ContainerStatuses: []corev1.ContainerStatus{
						{
							State: corev1.ContainerState{
								Terminated: &corev1.ContainerStateTerminated{
									ExitCode: 1,
								},
							},
						},
					},
				},
			},
			expected: "Failed",
		},
		{
			name: "pending phase with no waiting containers",
			pod: corev1.Pod{
				Status: corev1.PodStatus{
					Phase: corev1.PodPending,
				},
			},
			expected: "Pending",
		},
		{
			name: "init container waiting takes priority over regular container",
			pod: corev1.Pod{
				Status: corev1.PodStatus{
					Phase: corev1.PodPending,
					InitContainerStatuses: []corev1.ContainerStatus{
						{
							State: corev1.ContainerState{
								Waiting: &corev1.ContainerStateWaiting{
									Reason: "PodInitializing",
								},
							},
						},
					},
					ContainerStatuses: []corev1.ContainerStatus{
						{
							State: corev1.ContainerState{
								Running: &corev1.ContainerStateRunning{},
							},
						},
					},
				},
			},
			expected: "PodInitializing",
		},
		{
			name: "waiting container with empty reason falls back to phase",
			pod: corev1.Pod{
				Status: corev1.PodStatus{
					Phase: corev1.PodPending,
					ContainerStatuses: []corev1.ContainerStatus{
						{
							State: corev1.ContainerState{
								Waiting: &corev1.ContainerStateWaiting{
									Reason: "",
								},
							},
						},
					},
				},
			},
			expected: "Pending",
		},
		{
			name: "first waiting container reason wins with multiple containers",
			pod: corev1.Pod{
				Status: corev1.PodStatus{
					Phase: corev1.PodPending,
					ContainerStatuses: []corev1.ContainerStatus{
						{
							State: corev1.ContainerState{
								Waiting: &corev1.ContainerStateWaiting{
									Reason: "ContainerCreating",
								},
							},
						},
						{
							State: corev1.ContainerState{
								Waiting: &corev1.ContainerStateWaiting{
									Reason: "CrashLoopBackOff",
								},
							},
						},
					},
				},
			},
			expected: "ContainerCreating",
		},
		{
			name: "unknown phase falls back to phase string",
			pod: corev1.Pod{
				Status: corev1.PodStatus{
					Phase: corev1.PodUnknown,
				},
			},
			expected: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := podDisplayStatus(&tt.pod)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetRealPods_ReturnsNilOnBadConfig(t *testing.T) {
	orig := kubeconfig
	defer func() { kubeconfig = orig }()

	kubeconfig = "/nonexistent/kubeconfig"

	pods := getRealPods(context.Background(), "default", "app=test")
	assert.Nil(t, pods, "getRealPods should return nil pods when config loading fails (errors are logged)")
}
