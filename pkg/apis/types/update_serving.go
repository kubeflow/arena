package types

type CommonUpdateServingArgs struct {
	Name        string            `yaml:"servingName"`
	Version     string            `yaml:"servingVersion"`
	Namespace   string            `yaml:"-"`
	Type        ServingJobType    `yaml:"-"`
	Image       string            `yaml:"image"`
	GPUCount    int               `yaml:"gpuCount"`    // --gpus
	GPUMemory   int               `yaml:"gpuMemory"`   // --gpumemory
	Cpu         string            `yaml:"cpu"`         // --cpu
	Memory      string            `yaml:"memory"`      // --memory
	Replicas    int               `yaml:"replicas"`    // --replicas
	Envs        map[string]string `yaml:"envs"`        // --envs
	Annotations map[string]string `yaml:"annotations"` // --annotation
	Labels      map[string]string `yaml:"labels"`      // --label
	Shell       string            `yaml:"shell"`       // --shell
	Command     string            `yaml:"command"`     // --command
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
