package types

type ServingJobType string

const (
	// TFServingJob defines the tensorflow serving job
	TFServingJob ServingJobType = "tf-serving"
	// TRTServingJob defines the tensorrt serving job
	TRTServingJob ServingJobType = "trt-serving"
	// KFServingJob defines the kfserving job
	KFServingJob ServingJobType = "kf-serving"
	// CustomServingJob defines the custom serving job
	CustomServingJob ServingJobType = "custom-serving"
	// AllServingJob represents all serving job type
	AllServingJob ServingJobType = ""
	// UnknownServingJob defines the unknown serving job
	UnknownServingJob ServingJobType = "unknown"
)

// ServingTypeMap collects serving job type and their alias
var ServingTypeMap = map[ServingJobType][]string{
	CustomServingJob: {
		"custom",
		"custom-serving",
	},
	KFServingJob: {
		"kf",
		"kfserving",
		"kf-serving",
	},
	TFServingJob: {
		"tf",
		"tfserving",
		"tf-serving",
		"tensorflow-serving",
	},
	TRTServingJob: {
		"trt",
		"trt-serving",
		"trtserving",
		"tensorrt-serving",
	},
}

// ServingJobInfo display serving job information
type ServingJobInfo struct {
	// Name specifies serving job name
	Name string `json:"name" yaml:"name"`
	// Namespace specifies serving job namespace
	Namespace string `json:"namespace" yaml:"namespace"`
	// Type specifies serving job type
	Type ServingJobType `json:"type" yaml:"type"`
	// Version specifies serving job version
	Version string `json:"version" yaml:"version"`
	// Age specifies the serving job age
	Age string `json:"age" yaml:"age"`
	// Desired specifies the desired instances
	Desired int `json:"desired" yaml:"desired"`
	// Available specifies the available instances
	Available int `json:"available" yaml:"available"`
	// Endpoints specifies the endpoints
	Endpoints []Endpoint `json:"endpoints" yaml:"endpoints"`
	// IPAddress specifies the ip address
	IPAddress string `json:"ip_address" yaml:"ip_address"`
	// Instances gives the instance informations
	Instances []ServingInstance `json:"instances" yaml:"instances"`
	// RequestGPU specifies the request gpus
	RequestGPU int `json:"request_gpu" yaml:"request_gpu"`
	// RequestGPUMemory specifies the request gpu memory,only for gpushare
	RequestGPUMemory int `json:"request_gpu_mem" yaml:"request_gpu_mem"`
}

type Endpoint struct {
	// Endpoint Name
	Name string `json:"name" yaml:"name"`
	// Port specifies endpoint port
	Port int `json:"port" yaml:"port"`
	// NodePort specifies the node port
	NodePort int `json:"node_port" yaml:"node_port"`
}

type ServingInstance struct {
	// Name gives the instance name
	Name string `json:"name" yaml:"name"`
	// Status gives the instance status
	Status string `json:"status" yaml:"status"`
	// Age gives the instance ge
	Age string `json:"age" yaml:"age"`
	// ReadyContainer represents the count of ready containers
	ReadyContainer int `json:"ready" yaml:"ready"`
	// TotalContainer represents the count of  total containers
	TotalContainer int `json:"total" yaml:"total"`
	// RestartCount represents the count of instance restarts
	RestartCount int `json:"restart" yaml:"restart"`
	// HostIP specifies host ip of instance
	NodeIP string `json:"node_ip" yaml:"node_ip"`
	// NodeName returns the node name
	NodeName string `json:"node_name" yaml:"node_name"`
	// IP returns the instance ip
	IP string `json:"ip" yaml:"ip"`
}

type CommonServingArgs struct {
	Name            string            `yaml:"servingName"`
	Version         string            `yaml:"servingVersion"`
	Namespace       string            `yaml:"-"`
	Type            ServingJobType    `yaml:"-"`
	Image           string            `yaml:"image"`
	ImagePullPolicy string            `yaml:"imagePullPolicy"` // --imagePullPolicy
	GPUCount        int               `yaml:"gpuCount"`        // --gpus
	GPUMemory       int               `yaml:"gpuMemory"`       // --gpumemory
	Cpu             string            `yaml:"cpu"`             // --cpu
	Memory          string            `yaml:"memory"`          // --memory
	Envs            map[string]string `yaml:"envs"`            // --envs
	Command         string            `yaml:"command"`         // --command
	Replicas        int               `yaml:"replicas"`        // --replicas
	EnableIstio     bool              `yaml:"enableIstio"`     // --enableIstio
	ExposeService   bool              `yaml:"exposeService"`   // --exposeService
	ModelDirs       map[string]string `yaml:"modelDirs"`
	HostVolumes     []DataDirVolume   `yaml:"hostVolumes"`   // --data-dir
	NodeSelectors   map[string]string `yaml:"nodeSelectors"` // --selector
	Tolerations     []string          `yaml:"tolerations"`   // --toleration
	Annotations     map[string]string `yaml:"annotations"`

	ModelServiceExists bool `yaml:"modelServiceExists"` // --modelServiceExists
}

type CustomServingArgs struct {
	Port              int `yaml:"port"`        // --port
	RestfulPort       int `yaml:"restApiPort"` // --restfulPort
	CommonServingArgs `yaml:",inline"`
}

type TensorFlowServingArgs struct {
	VersionPolicy          string `yaml:"versionPolicy"`   // --versionPolicy
	ModelConfigFile        string `yaml:"modelConfigFile"` // --modelConfigFile
	ModelConfigFileContent string `yaml:"modelConfigFileContent"`
	ModelName              string `yaml:"modelName"`   // --modelName
	ModelPath              string `yaml:"modelPath"`   // --modelPath
	Port                   int    `yaml:"port"`        // --port
	RestfulPort            int    `yaml:"restApiPort"` // --restfulPort
	CommonServingArgs      `yaml:",inline"`
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
