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

package types

const (
	CPUResourceName = "cpu"
)

const (
	// defines the nvidia resource name
	NvidiaGPUResourceName = "nvidia.com/gpu"
)

const (
	GPUShareResourceName        = "aliyun.com/gpu-mem"
	GPUCoreShareResourceName    = "aliyun.com/gpu-core.percentage"
	GPUShareCountName           = "aliyun.com/gpu-count"
	GPUShareEnvGPUID            = "ALIYUN_COM_GPU_MEM_IDX"
	GPUShareAllocationLabel     = "scheduler.framework.gpushare.allocation"
	GPUCoreShareAllocationLabel = "gpushare.alibabacloud.com/core-percentage"
	GPUShareNodeLabels          = "gpushare=true,cgpu=true,ack.node.gpu.schedule=share,ack.node.gpu.schedule=cgpu"
)

const (
	AliyunGPUResourceName      = "aliyun.com/gpu"
	GPUTopologyAllocationLabel = "topology.kubernetes.io/gpu-group"
	GPUTopologyVisibleGPULabel = "topology.kubernetes.io/gpu-visible"
	GPUTopologyNodeLabels      = "ack.node.gpu.schedule=topology"
)

const (
	MultiTenantIsolationLabel = "arena.kubeflow.org/isolate-user"
	UserNameIdLabel           = "arena.kubeflow.org/uid"
	UserNameNameLabel         = "arena.kubeflow.org/username"
	SSHSecretName             = "arena.kubeflow.org/ssh-secret"
)
