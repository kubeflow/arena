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

import (
	v1 "k8s.io/api/core/v1"
)

// Job defines some common behaviors for serving job and training job
type Job interface {
	// return the job name
	GetName() string
	// return the job namespace
	GetNamespace() string
	// return the type of job
	GetType() interface{}
	// return the time which the job has been runing
	GetAge() string
	// return all instances of the job
	GetAllPods() []v1.Pod
	// return the job status
	GetStatus() interface{}
}

// JobManager defines some common behaviors for serving job and training job
type JobManager interface {
	// return all jobs
	GetAllJobs() []interface{}
	// return target jobs with filter args
	GetTargetJobs(filterArgs interface{}) ([]interface{}, error)
}

// JobPrinter is return job information with some format like 'json' 'yaml' 'wide'
type JobPrinter interface {
	// return json format string for job information
	GetJsonFormatString() (string, error)
	// return yaml format string for job information
	GetYamlFormatString() (string, error)
	// return wide format string for job information
	GetWideFormatString() (string, error)
	// return the help information when there is more than one pod in job
	GetHelpInfo(obj ...interface{}) (string, error)
}

// JobInfo defines job information which user can get.
type JobPrinterInfo struct {
	// The name of the training job
	Name string `json:"name"`
	// The namespace of the  job
	Namespace string `json:"namespace"`
	// The age of the job
	Age string `json:"age"`
	// The status of the Job
	Status interface{} `json:"status"`
	// The training type of the job
	Type interface{} `json:"type"`
	// The instances under the job
	Instances []interface{} `json:"instances"`
}

// instance is used to return some simple information for pod
type Instance struct {
	// the status of of instance
	Status string `json:"status"`
	// the name of instance
	Name string `json:"name"`
	// the age of instance
	Age string `json:"age"`
	// the node instance runs on
	Node string `json:"node"`
}
