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

import "errors"

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
	// SparkTrainingJob defines the spark job
	SparkTrainingJob TrainingJobType = "sparkjob"
	// DeepSpeedTrainingJob defines the deepspeed job
	DeepSpeedTrainingJob TrainingJobType = "deepspeedjob"
	// AllTrainingJob represents all job types
	AllTrainingJob TrainingJobType = ""
	// UnknownTrainingJob defines the unknown training
	UnknownTrainingJob TrainingJobType = "unknown"
)

type TrainingJobTypeInfo struct {
	Name      TrainingJobType
	Alias     string
	Shorthand string
}

var (
	ErrTrainingJobNotFound      = errors.New("training job not found,please use 'arena list' to make sure job is existed.")
	ErrNoPrivilegesToOperateJob = errors.New("you have no privileges to operate the job,because the owner of job is not you")
)

// ServingTypeMap collects serving job type and their alias
var TrainingTypeMap = map[TrainingJobType]TrainingJobTypeInfo{
	TFTrainingJob: {
		Name:      TFTrainingJob,
		Alias:     "Tensorflow",
		Shorthand: "tf",
	},
	MPITrainingJob: {
		Name:      MPITrainingJob,
		Alias:     "MPI",
		Shorthand: "mpi",
	},
	PytorchTrainingJob: {
		Name:      PytorchTrainingJob,
		Alias:     "Pytorch",
		Shorthand: "py",
	},
	HorovodTrainingJob: {
		Name:      HorovodTrainingJob,
		Alias:     "Horovod",
		Shorthand: "horovod",
	},
	VolcanoTrainingJob: {
		Name:      VolcanoTrainingJob,
		Alias:     "Volcano",
		Shorthand: "volcano",
	},
	ETTrainingJob: {
		Name:      ETTrainingJob,
		Alias:     "ElasticTraining",
		Shorthand: "et",
	},
	SparkTrainingJob: {
		Name:      SparkTrainingJob,
		Alias:     "Spark",
		Shorthand: "spark",
	},
	DeepSpeedTrainingJob: {
		Name:      DeepSpeedTrainingJob,
		Alias:     "DeepSpeed",
		Shorthand: "dp",
	},
}

// TrainingJobInfo stores training job information
type TrainingJobInfo struct {
	// The unique identity of the training job
	UUID string `json:"uuid" yaml:"uuid"`
	// The name of the training job
	Name string `json:"name" yaml:"name"`
	// The namespace of the training job
	Namespace string `json:"namespace" yaml:"namespace"`
	// The time of the training job
	Duration string `json:"duration" yaml:"duration"`
	// The status of the training Job
	Status TrainingJobStatus `json:"status" yaml:"status"`

	// The training type of the training job
	Trainer TrainingJobType `json:"trainer" yaml:"trainer"`
	// The tensorboard of the training job
	Tensorboard string `json:"tensorboard" yaml:"tensorboard"`

	// The name of the chief Instance
	ChiefName string `json:"chiefName" yaml:"chiefName"`

	// The instances under the training job
	Instances []TrainingJobInstance `json:"instances" yaml:"instances"`

	// The priority of the training job
	Priority string `json:"priority" yaml:"priority"`

	// RequestGPU stores the request gpus
	RequestGPU int64 `json:"requestGPUs" yaml:"requestGPUs"`

	// AllocatedGPU stores the allocated gpus
	AllocatedGPU int64 `json:"allocatedGPUs" yaml:"allocatedGPUs"`

	// CreationTimestamp stores the creation timestamp of job
	CreationTimestamp int64 `json:"creationTimestamp" yaml:"creationTimestamp"`

	// Model information associated with this job
	ModelName    string `json:"modelName"`
	ModelVersion string `json:"modelVersion"`
	ModelSource  string `json:"modelSource"`
}

// TrainingJobStatus defines all the kinds of JobStatus
type TrainingJobStatus string

const (
	// TrainingJobQueuing means the job is queuing
	TrainingJobQueuing TrainingJobStatus = "QUEUING"
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
	// IP defines the instance ip
	IP string `json:"ip" yaml:"ip"`
	// the status of of instance
	Status string `json:"status"`
	// the name of instance
	Name string `json:"name"`
	// the age of instance
	Age string `json:"age"`
	// the node instance runs on
	Node string `json:"node"`
	// NodeIP is store the node ip
	NodeIP string `json:"nodeIP" yaml:"nodeIP"`
	// the instance is chief or not
	IsChief bool `json:"chief" yaml:"chief"`
	// RequestGPUs is used to store request gpu count
	RequestGPUs int `json:"requestGPUs" yaml:"requestGPUs"`
	// GpuDutyCycle stores the gpu metrics
	GPUMetrics map[string]GpuMetric `json:"gpuMetrics" yaml:"gpuMetrics"`
	// CreationTimestamp returns the creation timestamp of instance
	CreationTimestamp int64 `json:"creationTimestamp" yaml:"creationTimestamp"`
}

const (
	RequestGPUsOfJobAnnoKey = "requestGPUsOfJobOwner"
)
