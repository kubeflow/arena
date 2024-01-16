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

package training

import "errors"

const (
	CHART_PKG_LOC = "CHARTREPO"
	// GPUResourceName is the extended name of the GPU resource since v1.8
	// this uses the device plugin mechanism
	NVIDIAGPUResourceName = "nvidia.com/gpu"
	// GPUShareResourceName is the gpushare resource name
	GPUShareResourceName = "aliyun.com/gpu-mem"

	DeprecatedNVIDIAGPUResourceName = "alpha.kubernetes.io/nvidia-gpu"

	masterLabelRole = "node-role.kubernetes.io/master"

	gangSchdName = "kube-batchd"

	// labelNodeRolePrefix is a label prefix for node roles
	// It's copied over to here until it's merged in core: https://github.com/kubernetes/kubernetes/pull/39112
	labelNodeRolePrefix = "node-role.kubernetes.io/"

	// nodeLabelRole specifies the role of a node
	nodeLabelRole = "kubernetes.io/role"

	aliyunENIAnnotation = "k8s.aliyun.com/eni"

	requestGPUsOfJobAnnoKey = "requestGPUsOfJobOwner"

	spotInstanceJobStatusAnnotation = "job-supervisor.kube-ai.io/job-status"

	// TrainingReplicaTypeLabel training-operator replica type label
	TrainingReplicaTypeLabel = "training.kubeflow.org/replica-type"
	// TrainingReplicaIndexLabel training-operator replica index label
	TrainingReplicaIndexLabel = "training.kubeflow.org/replica-index"
)

var (
	errNotFoundOperator = errors.New("the server could not find the requested resource")
)
