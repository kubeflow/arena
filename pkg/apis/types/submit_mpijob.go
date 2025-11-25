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

type SubmitMPIJobArgs struct {
	Cpu    string `yaml:"cpu"`    // --cpu
	Memory string `yaml:"memory"` // --memory
	// for common args
	CommonSubmitArgs `yaml:",inline"`

	// for tensorboard
	SubmitTensorboardArgs `yaml:",inline"`

	// for sync up source code
	SubmitSyncCodeArgs `yaml:",inline"`

	// enable gpu topology scheduling
	GPUTopology        bool   `yaml:"gputopology"`
	GPUTopologyReplica string `yaml:"gputopologyreplica"`
	MountsOnLauncher   bool   `yaml:"mountsOnLauncher"`

	// clean-task-policy
	CleanPodPolicy string `yaml:"cleanPodPolicy"`

	// slot count for every worker
	SlotsPerWorker int `yaml:"slotsPerWorker"`
}
