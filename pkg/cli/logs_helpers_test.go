package cli

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPodBelongsToJob(t *testing.T) {
	tests := []struct {
		name      string
		pod       *corev1.Pod
		jobName   string
		wantMatch bool
	}{
		{
			name: "pod belongs to job",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"training.kubeflow.org/job-name": "my-job"},
				},
			},
			jobName:   "my-job",
			wantMatch: true,
		},
		{
			name: "pod with wrong job name",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"training.kubeflow.org/job-name": "other-job"},
				},
			},
			jobName:   "my-job",
			wantMatch: false,
		},
		{
			name: "pod with no labels",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{},
			},
			jobName:   "my-job",
			wantMatch: false,
		},
		{
			name: "pod with unrelated labels",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "nginx"},
				},
			},
			jobName:   "my-job",
			wantMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := podBelongsToJob(tt.pod, tt.jobName)
			if got != tt.wantMatch {
				t.Errorf("podBelongsToJob() = %v, want %v", got, tt.wantMatch)
			}
		})
	}
}

func TestContainerExists(t *testing.T) {
	pod := &corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "pytorch"},
				{Name: "tensorboard"},
			},
		},
	}

	if !containerExists(pod, "pytorch") {
		t.Error("containerExists() should return true for existing container")
	}
	if !containerExists(pod, "tensorboard") {
		t.Error("containerExists() should return true for existing container")
	}
	if containerExists(pod, "nonexistent") {
		t.Error("containerExists() should return false for non-existent container")
	}
}

func TestGetAvailableContainers(t *testing.T) {
	pod := &corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "pytorch"},
				{Name: "tensorboard"},
				{Name: "sidecar"},
			},
		},
	}

	containers := getAvailableContainers(pod)
	if len(containers) != 3 {
		t.Errorf("getAvailableContainers() returned %d containers, want 3", len(containers))
	}
	if containers[0] != "pytorch" || containers[1] != "tensorboard" || containers[2] != "sidecar" {
		t.Errorf("getAvailableContainers() = %v, want [pytorch, tensorboard, sidecar]", containers)
	}
}
