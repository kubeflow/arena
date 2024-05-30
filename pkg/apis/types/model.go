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

import "fmt"

// ModelJobType defines the supporting model job type
type ModelJobType string

const (
	// ModelProfileJob defines the model profile job
	ModelProfileJob ModelJobType = "profile"
	// ModelOptimizeJob defines the model optimize job
	ModelOptimizeJob ModelJobType = "optimize"
	// ModelBenchmarkJob defines the model benchmark job
	ModelBenchmarkJob ModelJobType = "benchmark"
	// ModelEvaluateJob defines the model evaluate job
	ModelEvaluateJob ModelJobType = "evaluate"
	// AllModelJob defines all model job
	AllModelJob ModelJobType = ""
	// UnknownModelJob defines the unknown model job
	UnknownModelJob ModelJobType = "unknown"
)

type ModelTypeInfo struct {
	Name      ModelJobType
	Alias     string
	Shorthand string
}

// ModelTypeMap collects model job type and their alias
var ModelTypeMap = map[ModelJobType]ModelTypeInfo{
	ModelProfileJob: {
		Name:      ModelProfileJob,
		Alias:     "Profile",
		Shorthand: "profile",
	},
	ModelOptimizeJob: {
		Name:      ModelOptimizeJob,
		Alias:     "Optimize",
		Shorthand: "optimize",
	},
	ModelBenchmarkJob: {
		Name:      ModelBenchmarkJob,
		Alias:     "Benchmark",
		Shorthand: "benchmark",
	},
	ModelEvaluateJob: {
		Name:      ModelEvaluateJob,
		Alias:     "Evaluate",
		Shorthand: "evaluate",
	},
}

// ModelJobStatus defines all the kinds of JobStatus
type ModelJobStatus string

const (
	// ModelJobPending means the job is pending
	ModelJobPending ModelJobStatus = "PENDING"
	// ModelJobRunning means the job is running
	ModelJobRunning ModelJobStatus = "RUNNING"
	// ModelJobComplete means the job is complete
	ModelJobComplete ModelJobStatus = "COMPLETE"
	// ModelJobFailed means the job is failed
	ModelJobFailed ModelJobStatus = "FAILED"
	// ModelJobUnknown means the job status is unknown
	ModelJobUnknown ModelJobStatus = "UNKNOWN"
)

type ModelJobInfo struct {
	// The unique identity of the model job
	UUID string `json:"uuid" yaml:"uuid"`

	// The name of the model job
	Name string `json:"name" yaml:"name"`

	// The namespace of the model job
	Namespace string `json:"namespace" yaml:"namespace"`

	// The time of the model job
	Duration string `json:"duration" yaml:"duration"`

	// Age specifies the model job age
	Age string `json:"age" yaml:"age"`

	// The status of the model Job
	Status string `json:"status" yaml:"status"`

	// The model type of the model job
	Type string `json:"type" yaml:"type"`

	// The instances under the model job
	Instances []ModelJobInstance `json:"instances" yaml:"instances"`

	// RequestCPUs GPU count of the Job
	RequestCPUs int64 `json:"requestCPUs" yaml:"requestCPUs"`

	// RequestGPUs stores the request gpus
	RequestGPUs int64 `json:"requestGPUs" yaml:"requestGPUs"`

	// RequestGPUMemory stores the request gpus
	RequestGPUMemory int64 `json:"requestGPUMemory" yaml:"requestGPUMemory"`

	// RequestGPUCore stores the request gpus core
	RequestGPUCore int64 `json:"requestGPUCore" yaml:"requestGPUCore"`

	// CreationTimestamp stores the creation timestamp of job
	CreationTimestamp int64 `json:"creationTimestamp" yaml:"creationTimestamp"`

	// CreationTimestamp stores the job parameters
	Params map[string]string `json:"params" yaml:"params"`
}

type ModelJobInstance struct {
	// Name gives the instance name
	Name string `json:"name" yaml:"name"`
	// Status gives the instance status
	Status string `json:"status" yaml:"status"`
	// Age gives the instance ge
	Age string `json:"age" yaml:"age"`
	// ReadyContainer represents the count of ready containers
	ReadyContainer int `json:"readyContainers" yaml:"readyContainers"`
	// TotalContainer represents the count of  total containers
	TotalContainer int `json:"totalContainers" yaml:"totalContainers"`
	// RestartCount represents the count of instance restarts
	RestartCount int `json:"restartCount" yaml:"restartCount"`
	// HostIP specifies host ip of instance
	NodeIP string `json:"nodeIP" yaml:"nodeIP"`
	// NodeName returns the node name
	NodeName string `json:"nodeName" yaml:"nodeName"`
	// IP returns the instance ip
	IP string `json:"ip" yaml:"ip"`
	// RequestGPU returns the request gpus
	RequestGPUs float64 `json:"requestGPUs" yaml:"requestGPUs"`
	// RequestGPUMemory returns the request gpu memory
	RequestGPUMemory int `json:"requestGPUMemory" yaml:"requestGPUMemory"`
	// RequestGPUCore returns the request gpu core
	RequestGPUCore int `json:"requestGPUCore" yaml:"requestGPUCore"`
	// CreationTimestamp returns the creation timestamp of instance
	CreationTimestamp int64 `json:"creationTimestamp" yaml:"creationTimestamp"`
}

type CommonModelArgs struct {
	Name            string `yaml:"name"`            // --name
	Namespace       string `yaml:"namespace"`       // --namespace
	ModelConfigFile string `yaml:"modelConfigFile"` // --model-config-file
	ModelName       string `yaml:"modelName"`       // --model-name
	ModelPath       string `yaml:"modelPath"`       // --model-path
	Inputs          string `yaml:"inputs"`          // --inputs
	Outputs         string `yaml:"outputs"`         // --outputs

	Image           string `yaml:"image"`           // --image
	ImagePullPolicy string `yaml:"imagePullPolicy"` // --image-pull-policy
	// ImagePullSecrets stores image pull secrets,match option --image-pull-secrets
	ImagePullSecrets []string `yaml:"imagePullSecrets"`

	GPUCount  int    `yaml:"gpuCount"`  // --gpus
	GPUMemory int    `yaml:"gpuMemory"` // --gpumemory
	GPUCore   int    `yaml:"gpuCore"`   // --gpucore
	Cpu       string `yaml:"cpu"`       // --cpu
	Memory    string `yaml:"memory"`    // --memory

	// DataSet stores the kubernetes pvc names
	DataSet map[string]string `yaml:"dataset"` // --data
	// DataDirs stores the files(or directories) in k8s node which will map to containers
	DataDirs []DataDirVolume `yaml:"dataDirs"` // --data-dir

	Envs          map[string]string `yaml:"envs"`          // --env
	NodeSelectors map[string]string `yaml:"nodeSelectors"` // --selector
	Tolerations   []TolerationArgs  `yaml:"tolerations"`   // --toleration
	Annotations   map[string]string `yaml:"annotations"`   // --annotation
	Labels        map[string]string `yaml:"labels"`        // --label

	Shell   string `yaml:"shell"` // --shell
	Command string `yaml:"command"`

	Type ModelJobType `yaml:"type"`
	// HelmOptions stores the helm options
	HelmOptions []string `yaml:"-"`
}

type ModelProfileArgs struct {
	ReportPath       string `yaml:"reportPath"`       // --report-path
	UseTensorboard   bool   `yaml:"useTensorboard"`   // --tensorboard
	TensorboardImage string `yaml:"tensorboardImage"` // --tensorboardImage

	CommonModelArgs `yaml:",inline"`
}

type ModelOptimizeArgs struct {
	Optimizer       string `yaml:"optimizer"`    // --optimizer
	TargetDevice    string `yaml:"targetDevice"` // --target-device
	ExportPath      string `yaml:"exportPath"`   // --export-path
	CommonModelArgs `yaml:",inline"`
}

type ModelBenchmarkArgs struct {
	Concurrency     int    `yaml:"concurrency"` // --concurrency
	Requests        int    `yaml:"requests"`    // --requests
	Duration        int    `yaml:"duration"`    // --duration (seconds)
	ReportPath      string `yaml:"reportPath"`  // --report-path
	CommonModelArgs `yaml:",inline"`
}

type ModelEvaluateArgs struct {
	ModelPlatform   string `yaml:"modelPlatform"` // --model-platform
	DatasetPath     string `yaml:"datasetPath"`   // --dataset-path
	ReportPath      string `yaml:"reportPath"`    // --report-path
	BatchSize       int    `yaml:"batchSize"`     // --batch-size
	CommonModelArgs `yaml:",inline"`
	// for sync up source code
	SubmitSyncCodeArgs `yaml:",inline"`
}

// Model Management
type RegisteredModel struct {
	Name                 string                  `json:"name"`
	CreationTimestamp    int64                   `json:"creation_timestamp,omitempty"`
	LastUpdatedTimestamp int64                   `json:"last_updated_timestamp,omitempty"`
	Description          string                  `json:"description,omitempty"`
	LatestVersions       []*ModelVersion         `json:"latest_versions,omitempty"`
	Tags                 []*RegisteredModelTag   `json:"tags,omitempty"`
	Aliases              []*RegisteredModelAlias `json:"aliases,omitempty"`
}

type RegisteredModelTag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (t RegisteredModelTag) String() string {
	return fmt.Sprintf("%s=%s", t.Key, t.Value)
}

type RegisteredModelAlias struct {
	Alias   string `json:"alias"`
	Version string `json:"version"`
}

type ModelVersion struct {
	Name                 string             `json:"name"`
	Version              string             `json:"version,omitempty"`
	CreationTimestamp    int64              `json:"creation_timestamp,omitempty"`
	LastUpdatedTimestamp int64              `json:"last_updated_timestamp,omitempty"`
	Description          string             `json:"description,omitempty"`
	UserId               string             `json:"user_id,omitempty"`
	CurrentStage         string             `json:"current_stage,omitempty"`
	Source               string             `json:"source,omitempty"`
	RunId                string             `json:"run_id,omitempty"`
	Status               ModelVersionStatus `json:"status,omitempty"`
	StatusMessage        string             `json:"status_message,omitempty"`
	Tags                 []*ModelVersionTag `json:"tags,omitempty"`
	RunLink              string             `json:"run_link,omitempty"`
	Aliases              []string           `json:"aliases,omitempty"`
}

type ModelVersionStatus string

const (
	PENDING_REGISTRATION ModelVersionStatus = "PENDING_REGISTRATION"
	FAILED_REGISTRATION  ModelVersionStatus = "FAILED_REGISTRATION"
	READY                ModelVersionStatus = "READY"
)

type ModelVersionTag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
