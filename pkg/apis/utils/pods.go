// Copyright 2024 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"encoding/json"

	"github.com/kubeflow/arena/pkg/apis/types"
	v1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
)

func CPUCountInPod(pod *v1.Pod) float64 {
	total := 0.0
	for _, count := range ResourceInContainers(pod, types.CPUResourceName) {
		total += count.(float64)
	}
	return total
}

func GPUCountInPod(pod *v1.Pod) int {
	total := int64(0)
	for _, count := range ResourceInContainers(pod, types.NvidiaGPUResourceName) {
		c := count.(int64)
		total += c
	}
	return int(total)
}

func AliyunGPUCountInPod(pod *v1.Pod) int {
	total := int64(0)
	for _, count := range ResourceInContainers(pod, types.AliyunGPUResourceName) {
		c := count.(int64)
		total += c
	}
	return int(total)
}

func ResourceInContainers(pod *v1.Pod, resourceName string) map[int]interface{} {
	total := make(map[int]interface{})
	containers := pod.Spec.Containers
	for index, container := range containers {
		if val, ok := container.Resources.Limits[v1.ResourceName(resourceName)]; ok && int(val.Value()) != 0 {
			total[index] = val.Value()
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
	total := int64(0)
	for _, count := range ResourceInContainers(pod, types.GPUShareResourceName) {
		c := count.(int64)
		total += c
	}
	return int(total)
}

func GPUCoreCountInPod(pod *v1.Pod) int {
	total := int64(0)
	for _, count := range ResourceInContainers(pod, types.GPUCoreShareResourceName) {
		c := count.(int64)
		total += c
	}
	return int(total)
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

func GetContainerGPUCoreAllocation(pod *v1.Pod) map[int]map[string]int {
	allocation := map[int]map[string]int{}
	alloc := getPodAnnotation(pod, types.GPUCoreShareAllocationLabel)
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

func GetPodGPUCoreAllocation(pod *v1.Pod) map[string]int {
	result := map[string]int{}
	allocation := GetContainerGPUCoreAllocation(pod)
	if len(allocation) != 0 {
		for _, g := range allocation {
			for gpuId, count := range g {
				result[gpuId] = result[gpuId] + count
			}
		}
		return result
	}
	return result
}

func AcquireAllActivePods(client *kubernetes.Clientset) ([]*v1.Pod, error) {
	allPods := []*v1.Pod{}

	fieldSelector, err := fields.ParseSelector("status.phase!=" + string(v1.PodSucceeded) + ",status.phase!=" + string(v1.PodFailed))
	if err != nil {
		return allPods, err
	}
	nodeNonTerminatedPodsList, err := client.CoreV1().Pods(metav1.NamespaceAll).List(context.TODO(), metav1.ListOptions{FieldSelector: fieldSelector.String()})
	if err != nil {
		return allPods, err
	}

	for _, pod := range nodeNonTerminatedPodsList.Items {
		allPods = append(allPods, pod.DeepCopy())
	}
	return allPods, nil
}

func AcquireAllActivePodsOfNode(client *kubernetes.Clientset, nodeName string) ([]*v1.Pod, error) {
	allPods := []*v1.Pod{}
	selector := fmt.Sprintf("spec.nodeName=%v,status.phase!=%v,status.phase!=%v", nodeName, string(v1.PodSucceeded), string(v1.PodFailed))

	fieldSelector, err := fields.ParseSelector(selector)
	if err != nil {
		return allPods, err
	}
	nodeNonTerminatedPodsList, err := client.CoreV1().Pods(metav1.NamespaceAll).List(context.TODO(), metav1.ListOptions{FieldSelector: fieldSelector.String()})
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

func GetPendingTimeOfPod(pod *v1.Pod) time.Duration {
	if pod.Status.Phase == v1.PodPending {
		if pod.CreationTimestamp.IsZero() {
			return time.Duration(0)
		}
		return metav1.Now().Sub(pod.CreationTimestamp.Time)
	}
	if pod.Status.StartTime == nil {
		return metav1.Now().Sub(pod.CreationTimestamp.Time)
	}
	return pod.Status.StartTime.Sub(pod.CreationTimestamp.Time)
}

func GetRunningTimeOfPod(pod *v1.Pod) time.Duration {
	if pod.Status.Phase == v1.PodPending || pod.Status.Phase == v1.PodUnknown {
		return time.Duration(0)
	}
	var startTime *metav1.Time
	var endTime *metav1.Time
	// get pod start time
	allContainerStatuses := []v1.ContainerStatus{}
	allContainerStatuses = append(allContainerStatuses, pod.Status.InitContainerStatuses...)
	allContainerStatuses = append(allContainerStatuses, pod.Status.ContainerStatuses...)
	startTime, endTime = getStartTimeAndEndTime(allContainerStatuses)
	if startTime == nil && pod.Status.StartTime != nil {
		startTime = pod.Status.StartTime
	}
	if startTime == nil {
		startTime = &pod.CreationTimestamp
	}
	if pod.Status.Phase == v1.PodRunning || endTime == nil {
		return metav1.Now().Sub(startTime.Time)
	}
	return endTime.Sub(startTime.Time)
}

func getStartTimeAndEndTime(containerStatuses []v1.ContainerStatus) (*metav1.Time, *metav1.Time) {
	startTimes := []*metav1.Time{}
	endTimes := []*metav1.Time{}
	for _, containerStatus := range containerStatuses {
		switch {
		case containerStatus.State.Running != nil:
			startTimes = append(startTimes, &containerStatus.State.Running.DeepCopy().StartedAt)
		case containerStatus.State.Terminated != nil:
			startTimes = append(startTimes, &containerStatus.State.Terminated.DeepCopy().StartedAt)
			endTimes = append(endTimes, &containerStatus.State.Terminated.DeepCopy().FinishedAt)
		}
	}
	sort.Slice(startTimes, func(i, j int) bool {
		return startTimes[j].After(startTimes[i].Time)
	})
	sort.Slice(endTimes, func(i, j int) bool {
		return endTimes[i].After(endTimes[j].Time)
	})
	var startTime *metav1.Time
	var endTime *metav1.Time
	if len(startTimes) != 0 {
		startTime = startTimes[0]
	}
	if len(endTimes) != 0 {
		endTime = endTimes[0]
	}
	return startTime, endTime
}

func GetDurationOfPod(pod *v1.Pod) time.Duration {
	if pod.Status.Phase == v1.PodPending {
		return GetPendingTimeOfPod(pod)
	}
	return GetRunningTimeOfPod(pod)
}
