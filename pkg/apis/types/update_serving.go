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
