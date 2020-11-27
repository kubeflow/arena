package utils

import (
	"fmt"

	"github.com/kubeflow/arena/pkg/apis/types"
	v1 "k8s.io/api/core/v1"
)

const (
	// tf-operator added labels for pods and servers.
	labelGroupName         = "group-name"
	labelGroupNameV1alpha2 = "group_name"

	// pytorchjob
	labelPyTorchGroupName = "group-name"
)

// GetTrainingJobTypes returns the supported training job types
func GetTrainingJobTypes() []types.TrainingJobType {
	return []types.TrainingJobType{
		types.MPITrainingJob,
		types.TFTrainingJob,
		types.PytorchTrainingJob,
	}
}

// TransferTrainingJobType returns the training job type
func TransferTrainingJobType(jobType string) types.TrainingJobType {
	if jobType == "" {
		return types.AllTrainingJob
	}
	switch jobType {
	case "tfjob", "tf":
		return types.TFTrainingJob
	case "pytorchjob", "pytorch", "py":
		return types.PytorchTrainingJob
	case "mpijob", "mpi":
		return types.MPITrainingJob
	}
	return types.UnknownTrainingJob
}

func GetLogLevel() []types.LogLevel {
	return []types.LogLevel{
		types.LogDebug,
		types.LogError,
		types.LogInfo,
		types.LogWarning,
	}
}
func TransferLogLevel(loglevel string) types.LogLevel {
	for _, knownLogLevel := range GetLogLevel() {
		if types.LogLevel(loglevel) == knownLogLevel {
			return knownLogLevel
		}
	}
	return types.LogUnknown
}
func GetFormatStyle() []types.FormatStyle {
	return []types.FormatStyle{
		types.JsonFormat,
		types.WideFormat,
		types.YamlFormat,
	}
}

func TransferPrintFormat(format string) types.FormatStyle {
	for _, knownFormat := range GetFormatStyle() {
		if types.FormatStyle(format) == knownFormat {
			return knownFormat
		}
	}
	return types.UnknownFormat
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

func IsTensorFlowPod(name, ns string, pod *v1.Pod) bool {
	// check the release name is matched tfjob name
	if pod.Labels["release"] != name {
		return false
	}
	// check the job type is tfjob
	if pod.Labels["app"] != string(types.TFTrainingJob) {
		return false
	}
	// check the namespace
	if pod.Namespace != ns {
		return false
	}
	// check the group name
	switch {
	case pod.Labels[labelGroupName] == "kubeflow.org":
		return true
	case pod.Labels[labelGroupNameV1alpha2] == "kubeflow.org":
		return true
	}
	return false
}
func IsPyTorchPod(name, ns string, pod *v1.Pod) bool {
	// check the release name is matched pytorchjob name
	if pod.Labels["release"] != name {
		return false
	}
	// check the job type is pytorchjob
	if pod.Labels["app"] != string(types.PytorchTrainingJob) {
		return false
	}
	// check the namespace
	if pod.Namespace != ns {
		return false
	}
	// check the group name
	if pod.Labels[labelPyTorchGroupName] != "kubeflow.org" {
		return false
	}
	return true
}

func IsMPIPod(name, ns string, pod *v1.Pod) bool {
	// check the release name is matched mpijob name
	if pod.Labels["release"] != name {
		return false
	}
	// check the job type is mpijob
	if pod.Labels["app"] != string(types.MPITrainingJob) {
		return false
	}

	if pod.Labels["group_name"] != "kubeflow.org" {
		return false
	}
	if pod.Namespace != ns {
		return false
	}
	return true
}
