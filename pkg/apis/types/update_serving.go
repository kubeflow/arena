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

type CommonUpdateServingArgs struct {
	Name          string            `yaml:"servingName"`
	Version       string            `yaml:"servingVersion"`
	Namespace     string            `yaml:"-"`
	Type          ServingJobType    `yaml:"-"`
	Image         string            `yaml:"image"`
	GPUCount      int               `yaml:"gpuCount"`      // --gpus
	GPUMemory     int               `yaml:"gpuMemory"`     // --gpumemory
	GPUCore       int               `yaml:"gpuCore"`       // --gpucore
	Cpu           string            `yaml:"cpu"`           // --cpu
	Memory        string            `yaml:"memory"`        // --memory
	Replicas      int               `yaml:"replicas"`      // --replicas
	Envs          map[string]string `yaml:"envs"`          // --envs
	Annotations   map[string]string `yaml:"annotations"`   // --annotation
	Labels        map[string]string `yaml:"labels"`        // --label
	NodeSelectors map[string]string `yaml:"nodeSelectors"` // --selector
	Tolerations   []TolerationArgs  `yaml:"tolerations"`   // --toleration
	Shell         string            `yaml:"shell"`         // --shell
	Command       string            `yaml:"command"`       // --command
	ModelDirs     map[string]string `yaml:"modelDirs"`     // --data
}

type UpdateTensorFlowServingArgs struct {
	ModelConfigFile         string `yaml:"modelConfigFile"`      // --model-config-file
	MonitoringConfigFile    string `yaml:"monitoringConfigFile"` // --monitoring-config-file
	ModelName               string `yaml:"modelName"`            // --model-name
	ModelPath               string `yaml:"modelPath"`            // --model-path
	CommonUpdateServingArgs `yaml:",inline"`
}

type UpdateTritonServingArgs struct {
	ModelRepository         string `yaml:"modelRepository"` // --model-repository
	AllowMetrics            bool   `yaml:"allowMetrics"`    // --allow-metrics
	CommonUpdateServingArgs `yaml:",inline"`
}

type UpdateCustomServingArgs struct {
	CommonUpdateServingArgs `yaml:",inline"`
}

type UpdateKServeArgs struct {
	ModelFormat             *ModelFormat `yaml:"modelFormat"`                    // --model-format
	Runtime                 string       `yaml:"runtime"`                        // --runtime
	StorageUri              string       `yaml:"storageUri"`                     // --storageUri
	RuntimeVersion          string       `yaml:"runtimeVersion"`                 // --runtime-version
	ProtocolVersion         string       `yaml:"protocolVersion"`                // --protocol-version
	MinReplicas             int          `yaml:"minReplicas"`                    // --min-replicas
	MaxReplicas             int          `yaml:"maxReplicas"`                    // --max-replicas
	ScaleTarget             int          `yaml:"scaleTarget"`                    // --scale-target
	ScaleMetric             string       `yaml:"scaleMetric"`                    // --scale-metric
	ContainerConcurrency    int64        `yaml:"containerConcurrency"`           // --container-concurrency
	TimeoutSeconds          int64        `yaml:"timeout"`                        // --timeout
	CanaryTrafficPercent    int64        `yaml:"canaryTrafficPercent,omitempty"` // --canary-traffic-percent
	Port                    int          `yaml:"port"`                           // --port
	CommonUpdateServingArgs `yaml:",inline"`
}

type UpdateDistributedServingArgs struct {
	Workers                 int    `yaml:"workers"`         // --workers
	MasterCpu               string `yaml:"masterCPU"`       // --master-cpu
	WorkerCpu               string `yaml:"workerCPU"`       // --worker-cpu
	MasterGPUCount          int    `yaml:"masterGPUCount"`  // master-gpus
	WorkerGPUCount          int    `yaml:"workerGPUCount"`  // worker-gpus
	MasterMemory            string `yaml:"masterMemory"`    // master-memory
	WorkerMemory            string `yaml:"workerMemory"`    // worker-memory
	MasterGPUMemory         int    `yaml:"masterGPUMemory"` // master-gpumemory
	WorkerGPUMemory         int    `yaml:"workerGPUMemory"` // worker-gpumemory
	MasterGPUCore           int    `yaml:"masterGPUCore"`   // master-gpucore
	WorkerGPUCore           int    `yaml:"workerGPUCore"`   // worker-gpucore
	MasterCommand           string `yaml:"masterCommand"`   // master-command
	WorkerCommand           string `yaml:"workerCommand"`   // worker-command
	CommonUpdateServingArgs `yaml:",inline"`
}
