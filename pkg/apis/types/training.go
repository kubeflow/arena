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
// limitations under the License

package types

// TrainingJobType defines the supporting training job type
type TrainingJobType string

const (
	// TFTrainingJob defines the tfjob
	TFTrainingJob TrainingJobType = "tfjob"
	// MPITrainingJob defines the mpijob
	MPITrainingJob TrainingJobType = "mpijob"
	// PytorchTrainingJob defines the pytorchjob
	PytorchTrainingJob TrainingJobType = "pytorchjob"
	// HorovodTrainingJob defines the horovod job
	HorovodTrainingJob TrainingJobType = "horovodjob"
	// VolcanoTrainingJob defines the volcano job
	VolcanoTrainingJob TrainingJobType = "volcanojob"
	// ETTrainingJob defines the etjob
	ETTrainingJob TrainingJobType = "etjob"
	// AllTrainingJob represents all job types
	AllTrainingJob TrainingJobType = ""
	// UnknownTrainingJob defines the unknown training
	UnknownTrainingJob TrainingJobType = "unknown"
)

// TrainingJobInfo stores training job information
type TrainingJobInfo struct {
	// The name of the training job
	Name string `json:"name"`
	// The namespace of the training job
	Namespace string `json:"namespace"`
	// The time of the training job
	Duration string `json:"duration"`
	// The status of the training Job
	Status TrainingJobStatus `json:"status"`

	// The training type of the training job
	Trainer TrainingJobType `json:"trainer"`
	// The tensorboard of the training job
	Tensorboard string `json:"tensorboard,omitempty"`

	// The name of the chief Instance
	ChiefName string `json:"chiefName" yaml:"chiefName"`

	// The instances under the training job
	Instances []TrainingJobInstance `json:"instances"`

	// The priority of the training job
	Priority string `json:"priority"`
}

// TrainingJobStatus defines all the kinds of JobStatus
type TrainingJobStatus string

const (
	// TrainingJobPending means the job is pending
	TrainingJobPending TrainingJobStatus = "PENDING"
	// TrainingJobRunning means the job is running
	TrainingJobRunning TrainingJobStatus = "RUNNING"
	// TrainingJobSucceeded means the job is Succeeded
	TrainingJobSucceeded TrainingJobStatus = "SUCCEEDED"
	// TrainingJobFailed means the job is failed
	TrainingJobFailed TrainingJobStatus = "FAILED"
)

// TrainingJobInstance defines the instance of training job
type TrainingJobInstance struct {
	// the status of of instance
	Status string `json:"status"`
	// the name of instance
	Name string `json:"name"`
	// the age of instance
	Age string `json:"age"`
	// the node instance runs on
	Node string `json:"node"`
	// the instance is chief or not
	IsChief bool `json:"chief" yaml:"chief"`
}
