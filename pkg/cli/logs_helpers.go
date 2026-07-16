package cli

import (
	corev1 "k8s.io/api/core/v1"

	"github.com/kubeflow/arena/pkg/constants"
)

// podBelongsToJob checks if a pod belongs to a job by examining the
// training.kubeflow.org/job-name label set by the training operator.
func podBelongsToJob(pod *corev1.Pod, jobName string) bool {
	labels := pod.Labels
	if labels == nil {
		return false
	}
	return labels[constants.LabelJobName] == jobName
}

// containerExists checks if a container exists in the pod spec.
func containerExists(pod *corev1.Pod, containerName string) bool {
	for _, c := range pod.Spec.InitContainers {
		if c.Name == containerName {
			return true
		}
	}
	for _, c := range pod.Spec.Containers {
		if c.Name == containerName {
			return true
		}
	}
	return false
}

// getAvailableContainers returns a list of container names in the pod.
func getAvailableContainers(pod *corev1.Pod) []string {
	var names []string
	for _, c := range pod.Spec.InitContainers {
		names = append(names, c.Name)
	}
	for _, c := range pod.Spec.Containers {
		names = append(names, c.Name)
	}
	return names
}
