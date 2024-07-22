// Copyright 2018 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License
package types

type SubmitTFJobArgs struct {
	// TFNodeSelectors assigns tfjob node selectors
	TFNodeSelectors map[string]map[string]string `yaml:"tfNodeSelectors"`
	// Port defines the defaut port if workerPort and PSPort are not set
	Port int
	// WorkerImage assigns worker image,match option --worker-image
	WorkerImage string `yaml:"workerImage"`
	// WorkerPort stores worker port,match option --work-port
	WorkerPort int `yaml:"workerPort"`
	// PSPort stores the ps port,match option --ps-port
	PSPort int `yaml:"psPort"`
	// PSCount stores the ps count,--ps-count
	PSCount int `yaml:"ps"`
	// PSImage stores the ps image,--ps-image
	PSImage string `yaml:"psImage"`
	// WorkerCpu stores the cpu of job worker,match option --worker-cpu
	WorkerCpu string `yaml:"workerCPU"`
	// WorkerCpuLimit stores the cpu limit of job worker,match option --worker-cpu-limit
	WorkerCpuLimit string `yaml:"workerCPULimit"`
	//WorkerNodeSelectors map[string]string `yaml:"workerNodeSelectors"` // --worker-selector
	// WorkerMemory stores woker memory,match option --worker-memory
	WorkerMemory string `yaml:"workerMemory"`
	// WorkerMemoryLimit stores woker memory limit,match option --worker-memory-limit
	WorkerMemoryLimit string `yaml:"workerMemoryLimit"`
	// PSCpu stores ps cpu,match option --ps-cpu
	PSCpu string `yaml:"psCPU"`
	// PSCpuLimit stores ps cpu limit,match option --ps-cpu-limit
	PSCpuLimit string `yaml:"psCPULimit"`
	// PSGpu stores ps gpu,match option --ps-gpus
	PSGpu int `yaml:"psGPU"` // --ps-gpus
	// PSMemory stores the ps memory,match option --ps-memory
	PSMemory string `yaml:"psMemory"`
	// PSMemoryLimit stores the ps memory limit,match option --ps-memory-limit
	PSMemoryLimit string `yaml:"psMemoryLimit"`
	// CleanPodPolicy stores the cleaning pod policy,match option --clean-task-policy
	CleanPodPolicy string `yaml:"cleanPodPolicy"`
	// UseChief stores the using chief or not,match option --chief
	UseChief bool `yaml:",omitempty"` // --chief
	// ChiefCount stores the chief count of job,match option --chief-count
	ChiefCount int `yaml:"chief"`
	// UseEvaluator is used to enable evaluator or not,match option --evaluator
	UseEvaluator bool `yaml:",omitempty"`
	// ChiefPort stores the chief port,match option --chief-port
	ChiefPort int `yaml:"chiefPort"`
	//ChiefNodeSelectors map[string]string `yaml:"chiefNodeSelectors"` // --chief-selector
	// ChiefCpu stores the chief pod cpu,match option --chief-cpu
	ChiefCpu string `yaml:"chiefCPU"`
	// ChiefCpuLimit stores the chief pod cpu limit,match option --chief-cpu-limit
	ChiefCpuLimit string `yaml:"chiefCPULimit"`
	// ChiefMemory stores the chief pod memory,match option --chief-memory
	ChiefMemory string `yaml:"chiefMemory"`
	// ChiefMemoryLimit stores the chief pod memory limit,match option --chief-memory-limit
	ChiefMemoryLimit string `yaml:"chiefMemoryLimit"`
	// EvaluatorCpu stores the evaluator pod cpu,match option --evaluator-cpu
	EvaluatorCpu string `yaml:"evaluatorCPU"`
	// EvaluatorCpuLimit stores the evaluator pod cpu limit,match option --evaluator-cpu-limit
	EvaluatorCpuLimit string `yaml:"evaluatorCPULimit"`
	//EvaluatorNodeSelectors map[string]string `yaml:"evaluatorNodeSelectors"` // --evaluator-selector
	// EvaluatorMemory stores the evaluator pod memory,match option --evaluator-memory
	EvaluatorMemory string `yaml:"evaluatorMemory"` // --evaluatorMemory
	// EvaluatorMemoryLimit stores the evaluator pod memory limit,match option --evaluator-memory-limit
	EvaluatorMemoryLimit string `yaml:"evaluatorMemoryLimit"` // --evaluatorMemoryLimit
	// EvaluatorCount stores the evaluator pod count,match option --evaluator-count
	EvaluatorCount int `yaml:"evaluator"`
	// HasGangScheduler determines if it has gang scheduler
	HasGangScheduler bool `yaml:"hasGangScheduler"`
	// ActiveDeadlineSeconds Specifies the duration (in seconds) since startTime during which the job can remain active
	// before it is terminated
	ActiveDeadlineSeconds int64 `yaml:"activeDeadlineSeconds,omitempty"`
	// StartingDeadlineSeconds Specifies the duration (in seconds) since startTime during which the job can remain pending
	// before it is terminated
	StartingDeadlineSeconds int64 `yaml:"startingDeadlineSeconds,omitempty"`
	// Defines the TTL for cleaning up finished TFJobs. Defaults to infinite.
	TTLSecondsAfterFinished int32 `yaml:"ttlSecondsAfterFinished,omitempty"`
	// ShareMemory Specifies the shared memory size
	ShareMemory string `yaml:"shareMemory"`
	// for common args
	CommonSubmitArgs `yaml:",inline"`

	// SubmitTensorboardArgs stores tensorboard information
	SubmitTensorboardArgs `yaml:",inline"`

	// SubmitSyncCodeArgs stores syncing code information
	SubmitSyncCodeArgs `yaml:",inline"`

	// TFRuntime stores the runtime
	TFRuntime `yaml:"-"`

	// TrainingOperatorCRD compatible with training-operator crd.
	TrainingOperatorCRD bool `yaml:"trainingOperatorCRD,omitempty"`
}

// SubmitTensorboardArgs is used to store tensorborad information
type SubmitTensorboardArgs struct {
	UseTensorboard   bool   `yaml:"useTensorboard"`             // --tensorboard
	TensorboardImage string `yaml:"tensorboardImage,omitempty"` // --tensorboardImage
	TrainingLogdir   string `yaml:"trainingLogdir"`             // --logdir
	HostLogPath      string `yaml:"hostLogPath"`
	IsLocalLogging   bool   `yaml:"isLocalLogging"`
}

// Customized runtime for tf training training
type TFRuntime interface {
	// check the tfjob args
	Check(tf *SubmitTFJobArgs) (err error)
	// transform the tfjob
	Transform(tf *SubmitTFJobArgs) (err error)
	Runtime
}

type Runtime interface {
	// get the chart
	GetChartName() string
	// defines the runtime is default or not
	IsDefault() bool
}
