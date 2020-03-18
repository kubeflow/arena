// Copyright 2018 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package commands

import (
	"strconv"

	"k8s.io/api/core/v1"
)

// filter out the pods with GPU
func gpuPods(pods []v1.Pod) (podsWithGPU []v1.Pod) {
	for _, pod := range pods {
		if gpuInPod(pod) > 0 {
			podsWithGPU = append(podsWithGPU, pod)
		}
	}
	return podsWithGPU
}

// The way to get total GPU Count of Node: nvidia.com/gpu
func totalGpuInNode(node v1.Node) int64 {
	val, ok := node.Status.Capacity[NVIDIAGPUResourceName]

	if !ok {
		return gpuInNodeDeprecated(node)
	}

	return val.Value()
}

// The way to get allocatble GPU Count of Node: nvidia.com/gpu
func allocatableGpuInNode(node v1.Node) int64 {
	val, ok := node.Status.Allocatable[NVIDIAGPUResourceName]

	if !ok {
		return gpuInNodeDeprecated(node)
	}

	return val.Value()
}

// The way to get GPU Count of Node: alpha.kubernetes.io/nvidia-gpu
func gpuInNodeDeprecated(node v1.Node) int64 {
	val, ok := node.Status.Allocatable[DeprecatedNVIDIAGPUResourceName]

	if !ok {
		return 0
	}

	return val.Value()
}

func gpuInPod(pod v1.Pod) (gpuCount int64) {
	containers := pod.Spec.Containers
	for _, container := range containers {
		gpuCount += gpuInContainer(container)
	}

	return gpuCount
}

func gpuInActivePod(pod v1.Pod) (gpuCount float64) {
	if pod.Status.Phase == v1.PodSucceeded || pod.Status.Phase == v1.PodFailed {
		return 0
	}

	gpuFractionUsed := getGPUFractionUsedByPod(pod)
	if gpuFractionUsed > 0 {
		return gpuFractionUsed
	}

	return float64(gpuInPod(pod))
}

func getGPUFractionUsedByPod(pod v1.Pod) float64 {
	if pod.Annotations != nil {
		gpuFraction, GPUFractionErr := strconv.ParseFloat(pod.Annotations[runaiGPUFraction], 64)
		if GPUFractionErr == nil {
			return gpuFraction
		}
	}

	return 0
}

func gpuInContainer(container v1.Container) int64 {
	val, ok := container.Resources.Limits[NVIDIAGPUResourceName]

	if !ok {
		return gpuInContainerDeprecated(container)
	}

	return val.Value()
}

func gpuInContainerDeprecated(container v1.Container) int64 {
	val, ok := container.Resources.Limits[DeprecatedNVIDIAGPUResourceName]

	if !ok {
		return 0
	}

	return val.Value()
}

func sharedGPUsUsedInNode(nodeInfo NodeInfo) int64 {
	gpuIndexUsed := map[string]bool{}
	for _, pod := range nodeInfo.pods {
		if pod.Status.Phase == v1.PodSucceeded || pod.Status.Phase == v1.PodFailed {
			return 0
		}

		if pod.Annotations != nil {
			gpuIndex, found := pod.Annotations[runaiGPUIndex]
			if !found {
				continue
			}

			gpuIndexUsed[gpuIndex] = true
		}
	}

	return int64(len(gpuIndexUsed))
}
