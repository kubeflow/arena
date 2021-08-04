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
	// SeldonServingJob defines the seldon core job
	SeldonServingJob ServingJobType = "seldon-serving"
	// TritonServingJob defines the nvidia triton server job
	TritonServingJob ServingJobType = "triton-serving"
	// CustomServingJob defines the custom serving job
	CustomServingJob ServingJobType = "custom-serving"
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
	// CreationTimestamp returns the creation timestamp of instance
	CreationTimestamp int64 `json:"creationTimestamp" yaml:"creationTimestamp"`
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
	Labels          map[string]string `yaml:"labels"` // --label

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

type SeldonServingArgs struct {
	Implementation    string `yaml:"implementation"` // --implementation
	ModelUri          string `yaml:"modelUri"`       // --modelUri
	CommonServingArgs `yaml:",inline"`
}

type TritonServingArgs struct {
	ModelRepository   string `yaml:"modelRepository"` // --model-repository
	MetricsPort       int    `yaml:"metricsPort"`     // --metrics-port
	HttpPort          int    `yaml:"httpPort"`        // --http-port
	GrpcPort          int    `yaml:"grpcPort"`        // --grpc-port
	AllowMetrics      bool   `yaml:"allowMetrics"`    // --allow-metrics
	CommonServingArgs `yaml:",inline"`
}
