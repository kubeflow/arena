package utils

import (
	"fmt"
	"strings"

	"encoding/json"

	"github.com/kubeflow/arena/pkg/apis/types"
	v1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
)

func GPUCountInPod(pod *v1.Pod) int {
	total := 0
	for _, count := range ResourceInContainers(pod, types.NvidiaGPUResourceName) {
		total += count
	}
	return total
}

func AliyunGPUCountInPod(pod *v1.Pod) int {
	total := 0
	for _, count := range ResourceInContainers(pod, types.AliyunGPUResourceName) {
		total += count
	}
	return total
}

func ResourceInContainers(pod *v1.Pod, resourceName string) map[int]int {
	total := map[int]int{}
	containers := pod.Spec.Containers
	for index, container := range containers {
		if val, ok := container.Resources.Limits[v1.ResourceName(resourceName)]; ok && int(val.Value()) != 0 {
			total[index] = int(val.Value())
		}
	}
	return total
}

// IsCompletedPod determines if the pod is completed or not
func IsCompletedPod(pod *v1.Pod) bool {
	if pod.DeletionTimestamp != nil {
		return true
	}
	if pod.Status.Phase == v1.PodSucceeded || pod.Status.Phase == v1.PodFailed {
		return true
	}
	return false
}

// DefinePodPhaseStatus returns the pod status display in kubectl
func DefinePodPhaseStatus(pod v1.Pod) (string, int, int, int) {
	restarts := 0
	totalContainers := len(pod.Spec.Containers)
	readyContainers := 0

	reason := string(pod.Status.Phase)
	if pod.Status.Reason != "" {
		reason = pod.Status.Reason
	}
	initializing := false
	for i := range pod.Status.InitContainerStatuses {
		container := pod.Status.InitContainerStatuses[i]
		restarts += int(container.RestartCount)
		switch {
		case container.State.Terminated != nil && container.State.Terminated.ExitCode == 0:
			continue
		case container.State.Terminated != nil:
			// initialization is failed
			if len(container.State.Terminated.Reason) == 0 {
				if container.State.Terminated.Signal != 0 {
					reason = fmt.Sprintf("Init:Signal:%d", container.State.Terminated.Signal)
				} else {
					reason = fmt.Sprintf("Init:ExitCode:%d", container.State.Terminated.ExitCode)
				}
			} else {
				reason = "Init:" + container.State.Terminated.Reason
			}
			initializing = true
		case container.State.Waiting != nil && len(container.State.Waiting.Reason) > 0 && container.State.Waiting.Reason != "PodInitializing":
			reason = "Init:" + container.State.Waiting.Reason
			initializing = true
		default:
			reason = fmt.Sprintf("Init:%d/%d", i, len(pod.Spec.InitContainers))
			initializing = true
		}
		break
	}
	if !initializing {
		restarts = 0
		hasRunning := false
		for i := len(pod.Status.ContainerStatuses) - 1; i >= 0; i-- {
			container := pod.Status.ContainerStatuses[i]

			restarts += int(container.RestartCount)
			if container.State.Waiting != nil && container.State.Waiting.Reason != "" {
				reason = container.State.Waiting.Reason
			} else if container.State.Terminated != nil && container.State.Terminated.Reason != "" {
				reason = container.State.Terminated.Reason
			} else if container.State.Terminated != nil && container.State.Terminated.Reason == "" {
				if container.State.Terminated.Signal != 0 {
					reason = fmt.Sprintf("Signal:%d", container.State.Terminated.Signal)
				} else {
					reason = fmt.Sprintf("ExitCode:%d", container.State.Terminated.ExitCode)
				}
			} else if container.Ready && container.State.Running != nil {
				hasRunning = true
				readyContainers++
			}
		}

		// change pod status back to "Running" if there is at least one container still reporting as "Running" status
		if reason == "Completed" && hasRunning {
			reason = "Running"
		}
	}

	if pod.DeletionTimestamp != nil && pod.Status.Reason == "NodeLost" {
		reason = "Unknown"
	} else if pod.DeletionTimestamp != nil {
		reason = "Terminating"
	}
	return reason, totalContainers, restarts, readyContainers
}

func GPUMemoryCountInPod(pod *v1.Pod) int {
	total := 0
	for _, count := range ResourceInContainers(pod, types.GPUShareResourceName) {
		total += count
	}
	return total
}

func GetContainerAllocation(pod *v1.Pod) map[int]map[string]int {
	allocation := map[int]map[string]int{}
	alloc := getPodAnnotation(pod, types.GPUShareAllocationLabel)
	if alloc == "" {
		return allocation
	}
	err := json.Unmarshal([]byte(alloc), &allocation)
	if err != nil {
		return allocation
	}
	return allocation
}

func getPodAnnotation(pod *v1.Pod, key string) (ids string) {
	if pod.ObjectMeta.Annotations == nil {
		return ids
	}
	if value, ok := pod.ObjectMeta.Annotations[key]; ok {
		ids = value
		return ids
	}
	return ids
}

func GetPodAllocation(pod *v1.Pod) map[string]int {
	result := map[string]int{}
	allocation := GetContainerAllocation(pod)
	if len(allocation) != 0 {
		for _, g := range allocation {
			for gpuId, count := range g {
				result[gpuId] = result[gpuId] + count
			}
		}
		return result
	}
	gpuIndex := getPodAnnotation(pod, types.GPUShareEnvGPUID)
	if gpuIndex == "" {
		return result
	}
	count := GPUMemoryCountInPod(pod)
	result[gpuIndex] = count
	return result
}

func AcquireAllActivePods(client *kubernetes.Clientset) ([]*v1.Pod, error) {
	allPods := []*v1.Pod{}

	fieldSelector, err := fields.ParseSelector("status.phase!=" + string(v1.PodSucceeded) + ",status.phase!=" + string(v1.PodFailed))
	if err != nil {
		return allPods, err
	}
	nodeNonTerminatedPodsList, err := client.CoreV1().Pods(metav1.NamespaceAll).List(metav1.ListOptions{FieldSelector: fieldSelector.String()})
	if err != nil {
		return allPods, err
	}

	for _, pod := range nodeNonTerminatedPodsList.Items {
		allPods = append(allPods, pod.DeepCopy())
	}
	return allPods, nil
}

func GetPodGPUTopologyAllocation(pod *v1.Pod) []string {
	topoAllocation := getPodAnnotation(pod, types.GPUTopologyAllocationLabel)
	return strings.Split(topoAllocation, ",")
}

func GetPodGPUTopologyVisibleGPUs(pod *v1.Pod) []string {
	visibleGPUs := getPodAnnotation(pod, types.GPUTopologyVisibleGPULabel)
	return strings.Split(visibleGPUs, ",")
}
