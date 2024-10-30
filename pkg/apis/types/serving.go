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

// ServingJobType defines the serving job type
// name must like shorthand + "-serving"
type ServingJobType string

const (
	// TFServingJob defines the tensorflow serving job
	TFServingJob ServingJobType = "tf-serving"
	// TRTServingJob defines the tensorrt serving job
	TRTServingJob ServingJobType = "trt-serving"
	// KFServingJob defines the kfserving job
	KFServingJob ServingJobType = "kf-serving"
	// KServeJob defines the kserve job
	KServeJob ServingJobType = "kserve"
	// SeldonServingJob defines the seldon core job
	SeldonServingJob ServingJobType = "seldon-serving"
	// TritonServingJob defines the nvidia triton server job
	TritonServingJob ServingJobType = "triton-serving"
	// CustomServingJob defines the custom serving job
	CustomServingJob ServingJobType = "custom-serving"
	// DistributedServingJob defines the distributed serving job
	DistributedServingJob ServingJobType = "distributed-serving"
	// AllServingJob represents all serving job type
	AllServingJob ServingJobType = ""
	// UnknownServingJob defines the unknown serving job
	UnknownServingJob ServingJobType = "unknown"
)

type ServingTypeInfo struct {
	Name      ServingJobType
	Alias     string
	Shorthand string
}

// ServingTypeMap collects serving job type and their alias
var ServingTypeMap = map[ServingJobType]ServingTypeInfo{
	CustomServingJob: {
		Name:      CustomServingJob,
		Alias:     "Custom",
		Shorthand: "custom",
	},
	KFServingJob: {
		Name:      KFServingJob,
		Alias:     "KFServing",
		Shorthand: "kf",
	},
	KServeJob: {
		Name:      KServeJob,
		Alias:     "KServe",
		Shorthand: "kserve",
	},
	TFServingJob: {
		Name:      TFServingJob,
		Alias:     "Tensorflow",
		Shorthand: "tf",
	},
	TRTServingJob: {
		Name:      TRTServingJob,
		Alias:     "Tensorrt",
		Shorthand: "trt",
	},
	TritonServingJob: {
		Name:      TritonServingJob,
		Alias:     "Triton",
		Shorthand: "Triton",
	},
	SeldonServingJob: {
		Name:      SeldonServingJob,
		Alias:     "Seldon",
		Shorthand: "seldon",
	},
	DistributedServingJob: {
		Name:      DistributedServingJob,
		Alias:     "Distributed",
		Shorthand: "distributed",
	},
}

// ServingJobInfo display serving job information
type ServingJobInfo struct {
	// UUID specifies the unique identity of the serving job
	UUID string `json:"uuid" yaml:"uuid"`
	// Name specifies serving job name
	Name string `json:"name" yaml:"name"`
	// Namespace specifies serving job namespace
	Namespace string `json:"namespace" yaml:"namespace"`
	// Type specifies serving job type
	Type string `json:"type" yaml:"type"`
	// Version specifies serving job version
	Version string `json:"version" yaml:"version"`
	// Age specifies the serving job age
	Age string `json:"age" yaml:"age"`
	// Desired specifies the desired instances
	Desired int `json:"desiredInstances" yaml:"desiredInstances"`
	// Available specifies the available instances
	Available int `json:"availableInstances" yaml:"availableInstances"`
	// Endpoints specifies the endpoints
	Endpoints []Endpoint `json:"endpoints" yaml:"endpoints"`
	// IPAddress specifies the ip address
	IPAddress string `json:"ip" yaml:"ip"`
	// Instances gives the instance informations
	Instances []ServingInstance `json:"instances" yaml:"instances"`
	// RequestCPUs specifies the request cpus
	RequestCPUs float64 `json:"requestCPUs" yaml:"requestCPUs"`
	// RequestGPUs specifies the request gpus
	RequestGPUs float64 `json:"requestGPUs" yaml:"requestGPUs"`
	// RequestGPUMemory specifies the request gpu memory,only for gpushare
	RequestGPUMemory int `json:"requestGPUMemory" yaml:"requestGPUMemory"`
	// RequestGPUMemory specifies the request gpu core,only for gpushare
	RequestGPUCore int `json:"requestGPUCore" yaml:"requestGPUCore"`
	// CreationTimestamp stores the creation timestamp of job
	CreationTimestamp int64 `json:"creationTimestamp" yaml:"creationTimestamp"`
}

type Endpoint struct {
	// Endpoint Name
	Name string `json:"name" yaml:"name"`
	// Port specifies endpoint port
	Port int `json:"port" yaml:"port"`
	// NodePort specifies the node port
	NodePort int `json:"nodePort" yaml:"nodePort"`
}

type ServingInstance struct {
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
	// RequestGPUMemory specifies the request gpu core,only for gpushare
	RequestGPUCore int `json:"requestGPUCore" yaml:"requestGPUCore"`
	// CreationTimestamp returns the creation timestamp of instance
	CreationTimestamp int64 `json:"creationTimestamp" yaml:"creationTimestamp"`
}

type CommonServingArgs struct {
	Name               string            `yaml:"servingName"`
	Version            string            `yaml:"servingVersion"`
	Namespace          string            `yaml:"-"`
	Type               ServingJobType    `yaml:"-"`
	Image              string            `yaml:"image"`
	ImagePullPolicy    string            `yaml:"imagePullPolicy"`     // --imagePullPolicy
	GPUCount           int               `yaml:"gpuCount"`            // --gpus
	GPUMemory          int               `yaml:"gpuMemory"`           // --gpumemory
	GPUCore            int               `yaml:"gpuCore"`             // --gpucore
	Devices            map[string]string `yaml:"devices"`             // --device
	Cpu                string            `yaml:"cpu"`                 // --cpu
	Memory             string            `yaml:"memory"`              // --memory
	Envs               map[string]string `yaml:"envs"`                // --envs
	EnvsFromSecret     map[string]string `yaml:"envsFromSecret"`      // --env-from-secret
	Shell              string            `yaml:"shell"`               // --shell
	Command            string            `yaml:"command"`             // --command
	Replicas           int               `yaml:"replicas"`            // --replicas
	EnableIstio        bool              `yaml:"enableIstio"`         // --enableIstio
	ExposeService      bool              `yaml:"exposeService"`       // --exposeService
	ModelDirs          map[string]string `yaml:"modelDirs"`           // --data
	DataSubpathExprs   map[string]string `yaml:"dataSubPathExprs"`    // --data-subpath-expr
	TempDirSubpathExpr map[string]string `yaml:"tempDirSubPathExprs"` // --temp-dir-subpath-expr
	TempDirs           map[string]string `yaml:"tempDirs"`            // --temp-dir
	ShareMemory        string            `yaml:"shareMemory"`         // --share-memory

	ImagePullSecrets []string          `yaml:"imagePullSecrets"` //--image-pull-secrets
	HostVolumes      []DataDirVolume   `yaml:"dataDirs"`         // --data-dir
	NodeSelectors    map[string]string `yaml:"nodeSelectors"`    // --selector
	Tolerations      []TolerationArgs  `yaml:"tolerations"`      // --toleration
	Annotations      map[string]string `yaml:"annotations"`
	Labels           map[string]string `yaml:"labels"` // --label
	// ConfigFiles stores the config file which is existed in client host node
	// and map it to container,match option --config-file
	ConfigFiles map[string]map[string]ConfigFileInfo `yaml:"configFiles"`
	// HelmOptions stores the helm options
	HelmOptions []string `yaml:"-"`

	ModelServiceExists bool `yaml:"modelServiceExists"` // --modelServiceExists

	ModelName    string `yaml:"modelName"`    // --model-name
	ModelVersion string `yaml:"modelVersion"` // --model-version
}

type CustomServingArgs struct {
	Port                       int      `yaml:"port"`                       // --port
	RestfulPort                int      `yaml:"restApiPort"`                // --restfulPort
	MetricsPort                int      `yaml:"metricsPort"`                // --metrics-port
	MaxSurge                   string   `yaml:"maxSurge"`                   // --maxSurge
	MaxUnavailable             string   `yaml:"maxUnavailable"`             // --maxUnavailable
	LivenessProbeAction        string   `yaml:"livenessProbeAction"`        // --liveness-probe-action
	LivenessProbeActionOption  []string `yaml:"livenessProbeActionOption"`  // --liveness-probe-action-option
	LivenessProbeOption        []string `yaml:"livenessProbeOption"`        // --liveness-probe-option
	ReadinessProbeAction       string   `yaml:"readinessProbeAction"`       // --readiness-probe-action
	ReadinessProbeActionOption []string `yaml:"readinessProbeActionOption"` // --readiness-probe-action-option
	ReadinessProbeOption       []string `yaml:"readinessProbeOption"`       // --readiness-probe-option
	StartupProbeAction         string   `yaml:"startupProbeAction"`         // --startup-probe-action
	StartupProbeActionOption   []string `yaml:"startupProbeActionOption"`   // --startup-probe-action-option
	StartupProbeOption         []string `yaml:"startupProbeOption"`         // --startup-probe-option
	CommonServingArgs          `yaml:",inline"`
}

type TensorFlowServingArgs struct {
	VersionPolicy        string `yaml:"versionPolicy"`        // --version-policy
	ModelConfigFile      string `yaml:"modelConfigFile"`      // --model-config-file
	MonitoringConfigFile string `yaml:"monitoringConfigFile"` // --monitoring-config-file
	ModelPath            string `yaml:"modelPath"`            // --model-path
	Port                 int    `yaml:"port"`                 // --port
	RestfulPort          int    `yaml:"restApiPort"`          // --restful-port
	CommonServingArgs    `yaml:",inline"`
}

type TensorRTServingArgs struct {
	ModelStore        string `yaml:"modelStore"`   // --modelStore
	MetricsPort       int    `yaml:"metricsPort"`  // --metricsPort
	HttpPort          int    `yaml:"httpPort"`     // --httpPort
	GrpcPort          int    `yaml:"grpcPort"`     // --grpcPort
	AllowMetrics      bool   `yaml:"allowMetrics"` // --allowMetrics
	CommonServingArgs `yaml:",inline"`
}

type KFServingArgs struct {
	Port              int    `yaml:"port"`          // --port
	ModelType         string `yaml:"modelType"`     // --modelType
	CanaryPercent     int    `yaml:"canaryPercent"` // --canaryTrafficPercent
	StorageUri        string `yaml:"storageUri"`    // --storageUri
	CommonServingArgs `yaml:",inline"`
}

type KServeArgs struct {
	ModelFormat          *ModelFormat      `yaml:"modelFormat"`                    // --model-format
	Runtime              string            `yaml:"runtime"`                        // --runtime
	StorageUri           string            `yaml:"storageUri"`                     // --storageUri
	RuntimeVersion       string            `yaml:"runtimeVersion"`                 // --runtime-version
	ProtocolVersion      string            `yaml:"protocolVersion"`                // --protocol-version
	MinReplicas          int               `yaml:"minReplicas"`                    // --min-replicas
	MaxReplicas          int               `yaml:"maxReplicas"`                    // --max-replicas
	ScaleTarget          int               `yaml:"scaleTarget"`                    // --scale-target
	ScaleMetric          string            `yaml:"scaleMetric"`                    // --scale-metric
	ContainerConcurrency int64             `yaml:"containerConcurrency"`           // --container-concurrency
	TimeoutSeconds       int64             `yaml:"timeout"`                        // --timeout
	CanaryTrafficPercent int64             `yaml:"canaryTrafficPercent,omitempty"` // --canary-traffic-percent
	Port                 int               `yaml:"port"`                           // --port
	EnablePrometheus     bool              `yaml:"enablePrometheus,omitempty"`     // --enable-prometheus
	MetricsPort          int               `yaml:"metricsPort,omitempty"`          // --metrics-port
	SecurityContext      map[string]string `yaml:"securityContext,omitempty"`      // --security-context
	CommonServingArgs    `yaml:",inline"`
}

type SeldonServingArgs struct {
	Implementation    string `yaml:"implementation"` // --implementation
	ModelUri          string `yaml:"modelUri"`       // --modelUri
	CommonServingArgs `yaml:",inline"`
}

type TritonServingArgs struct {
	Backend           string   `yaml:"backend"`         // --backend
	ModelRepository   string   `yaml:"modelRepository"` // --model-repository
	MetricsPort       int      `yaml:"metricsPort"`     // --metrics-port
	HttpPort          int      `yaml:"httpPort"`        // --http-port
	GrpcPort          int      `yaml:"grpcPort"`        // --grpc-port
	AllowMetrics      bool     `yaml:"allowMetrics"`    // --allow-metrics
	LoadModels        []string `yaml:"loadModels"`      // --load-model
	ExtendCommand     string   `yaml:"extendCommand"`   // --extend-command
	CommonServingArgs `yaml:",inline"`
}

type DistributedServingArgs struct {
	Masters           int    `yaml:"masters"`         // --masters
	Workers           int    `yaml:"workers"`         // --workers
	MasterCpu         string `yaml:"masterCpus"`      // --master-cpu
	WorkerCpu         string `yaml:"workerCpus"`      // --worker-cpu
	MasterGPUCount    int    `yaml:"masterGpus"`      // --master-gpus
	WorkerGPUCount    int    `yaml:"workerGpus"`      // --worker-gpus
	MasterMemory      string `yaml:"masterMemory"`    // --master-memory
	WorkerMemory      string `yaml:"workerMemory"`    // --worker-memory
	MasterGPUMemory   int    `yaml:"masterGPUMemory"` // --master-gpumemory
	WorkerGPUMemory   int    `yaml:"workerGPUMemory"` // --worker-gpumemory
	MasterGPUCore     int    `yaml:"masterGPUCore"`   // --master-gpucore
	WorkerGPUCore     int    `yaml:"workerGPUCore"`   // --worker-gpucore
	MasterCommand     string `yaml:"masterCommand"`   // --master-command
	WorkerCommand     string `yaml:"workerCommand"`   // --worker-command
	InitBackend       string `yaml:"initBackend"`     // --init-backend
	CustomServingArgs `yaml:",inline"`
}

type ModelFormat struct {
	// Name of the model format.
	// +required
	Name string `yaml:"name"`
	// Version of the model format.
	// Used in validating that a predictor is supported by a runtime.
	// Can be "major", "major.minor" or "major.minor.patch".
	// +optional
	Version *string `yaml:"version,omitempty"`
}
