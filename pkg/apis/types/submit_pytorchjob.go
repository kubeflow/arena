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

type SubmitPyTorchJobArgs struct {
	Cpu    string `yaml:"cpu"`    // --cpu
	Memory string `yaml:"memory"` // --memory
	// for common args
	CommonSubmitArgs `yaml:",inline"`

	// for tensorboard
	SubmitTensorboardArgs `yaml:",inline"`

	// for sync up source code
	SubmitSyncCodeArgs `yaml:",inline"`

	// worker init pytorch image, default "alpine:3.10";
	// TODO jiaqianjing: user can set init-pytorch container image by param "--worker-init-pytorch-image"
	// WorkerInitPytorchImage string `yaml: workerInitPytorchImage`

	// clean-task-policy
	CleanPodPolicy string `yaml:"cleanPodPolicy"`

	// ActiveDeadlineSeconds Specifies the duration (in seconds) since startTime during which the job can remain active
	// before it is terminated
	ActiveDeadlineSeconds int64 `yaml:"activeDeadlineSeconds,omitempty"`

	// Defines the TTL for cleaning up finished PytorchJobs. Defaults to infinite.
	TTLSecondsAfterFinished int32 `yaml:"ttlSecondsAfterFinished,omitempty"`

	// TrainingOperatorCRD compatible with training-operator crd.
	TrainingOperatorCRD bool `yaml:"trainingOperatorCRD,omitempty"`

	// ShareMemory Specifies the shared memory size
	ShareMemory string `yaml:"shareMemory"`

	// Number of workers per node, supported values: [auto, cpu, gpu, int].
	// For more, https://github.com/pytorch/pytorch/blob/26f7f470df64d90e092081e39507e4ac751f55d6/torch/distributed/run.py#L629-L658.
	NprocPerNode string `yaml:"nprocPerNode,omitempty"`
}
