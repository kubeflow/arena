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
	"github.com/kubeflow/arena/pkg/apis/types"
	corev1 "k8s.io/api/core/v1"
)

func IsTensorFlowPod(name, ns string, pod *corev1.Pod) bool {
	// check the release name is matched tfjob name
	createdBy := pod.Labels["createdBy"]
	if createdBy != "" && createdBy == "Cron" {
		if pod.Labels["tf-job-name"] != name {
			return false
		}
	} else {
		if pod.Labels["release"] != name {
			return false
		}
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
	case pod.Labels[OperatorNameLabel] == "tfjob-controller":
		return true
	}
	return false
}
func IsPyTorchPod(name, ns string, pod *corev1.Pod) bool {
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
	switch {
	case pod.Labels[labelPyTorchGroupName] == "kubeflow.org":
		return true
	case pod.Labels[OperatorNameLabel] == "pytorchjob-controller":
		return true
	}
	return false
}

func IsRayJobPod(name, ns string, pod *corev1.Pod) bool {
	// determine whether the pod is a ray job pod
	if pod.Labels["job-name"] == name {
		return true
	}
	// determine whether the pod is a ray cluster pod
	if pod.Labels["release"] != name {
		return false
	}
	if pod.Labels["app"] != string(types.RayJob) {
		return false
	}
	if pod.Namespace != ns {
		return false
	}
	return true
}

func IsMPIPod(name, ns string, pod *corev1.Pod) bool {
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

func IsHorovodPod(name, ns string, pod *corev1.Pod) bool {
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

func IsVolcanoPod(name, ns string, pod *corev1.Pod) bool {
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

func IsETPod(name, ns string, pod *corev1.Pod) bool {
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

func IsSparkPod(name, ns string, item *corev1.Pod) bool {
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

func IsDeepSpeedPod(name, ns string, pod *corev1.Pod) bool {
	if pod.Labels["release"] != name {
		return false
	}
	if pod.Labels["app"] != string(types.DeepSpeedTrainingJob) {
		return false
	}
	if pod.Labels[deepspeedGroupName] != "kai.alibabacloud.com" {
		return false
	}
	if pod.Namespace != ns {
		return false
	}
	return true
}
