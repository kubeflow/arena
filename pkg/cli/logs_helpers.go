package cli

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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

// buildTFJobFallbackSelector constructs the label selector for the TFJob
// worker-0 pod, used as a fallback when the chief selector returns no pods
// (chief is optional in TFJob). Uses metav1.LabelSelector for consistent
// label value escaping.
func buildTFJobFallbackSelector(jobName string) string {
	return metav1.FormatLabelSelector(&metav1.LabelSelector{
		MatchLabels: map[string]string{
			constants.LabelJobName:      jobName,
			constants.LabelReplicaType:  constants.ReplicaRoleWorker,
			constants.LabelReplicaIndex: "0",
		},
	})
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
	names := make([]string, 0, len(pod.Spec.InitContainers)+len(pod.Spec.Containers))
	for _, c := range pod.Spec.InitContainers {
		names = append(names, c.Name)
	}
	for _, c := range pod.Spec.Containers {
		names = append(names, c.Name)
	}
	return names
}
