package utils

import (
	"github.com/kubeflow/arena/pkg/apis/types"
	v1 "k8s.io/api/core/v1"
)

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

func IsHorovodPod(name, ns string, pod *v1.Pod) bool {
	if pod.Labels["release"] != name {
		return false
	}
	if pod.Labels["app"] != "tf-horovod" {
		return false
	}
	if pod.Namespace != ns {
		return false
	}
	return true
}

func IsVolcanoPod(name, ns string, pod *v1.Pod) bool {
	if pod.Labels["release"] != name {
		return false
	}
	if pod.Labels["app"] != string(types.VolcanoTrainingJob) {
		return false
	}
	if pod.Namespace != ns {
		return false
	}
	return true
}

func IsETPod(name, ns string, pod *v1.Pod) bool {
	if pod.Labels["release"] != name {
		return false
	}
	if pod.Labels["app"] != string(types.ETTrainingJob) {
		return false
	}
	if pod.Labels[etLabelGroupName] != "kai.alibabacloud.com" {
		return false
	}
	if pod.Namespace != ns {
		return false
	}
	return true
}

func IsSparkPod(name, ns string, item *v1.Pod) bool {
	if item.Labels["release"] != name {
		return false
	}
	if item.Labels["app"] != "sparkjob" {
		return false
	}
	if item.Namespace != ns {
		return false
	}
	return true
}
