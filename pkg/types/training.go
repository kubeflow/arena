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

package types

type TrainingJobPrinterInfo struct {
	JobPrinterInfo `yaml:",inline" json:",inline"`
	// The tensorboard of the training job
	Tensorboard string `json:"tensorboard,omitempty"`
	// The name of the chief Instance
	ChiefName string `json:"chiefName" yaml:"chiefName"`
	// The priority of the training job
	Priority string `json:"priority"`
}

// all the kinds of JobStatus
type JobStatus string

const (
	// JobPending means the job is pending
	JobPending JobStatus = "PENDING"
	// JobRunning means the job is running
	JobRunning JobStatus = "RUNNING"
	// JobSucceeded means the job is Succeeded
	JobSucceeded JobStatus = "SUCCEEDED"
	// JobFailed means the job is failed
	JobFailed JobStatus = "FAILED"
)

type TrainingInstance struct {
	Instance `yaml:",inline" json:",inline"`
	// the instance is chief or not
	IsChief bool `json:"chief" yaml:"chief"`
}
