package types

type EvaluateJobType string

const (
	// EvaluateJob defines the tensorflow serving job
	EvaluateJob EvaluateJobType = "evaluatejob"
)

type EvaluateJobArgs struct {

	// Name stores the job name,match option --name
	Name string `yaml:"-"`

	// Namespace  stores the namespace of job,match option --namespace
	Namespace string `yaml:"-"`

	// NodeSelectors defines the node selectors,match option --selector
	NodeSelectors map[string]string `yaml:"nodeSelectors"`

	// Tolerations defines the tolerations which tolerates node taints
	// match option --toleration
	Tolerations []string `yaml:"tolerations"`

	// Image stores the docker image of job,match option --image
	Image string `yaml:"image"`

	// Envs stores the envs of container in job, match option --env
	Envs map[string]string `yaml:"envs"`

	WorkingDir string `yaml:"workingDir"`

	// Command stores the command of job
	Command string `yaml:"command"`

	// DataDirs stores the files(or directories) in k8s node which will map to containers
	// match option --data-dir
	DataDirs []DataDirVolume `yaml:"dataDirs"`

	// DataSources stores the kubernetes pvc names
	DataSources map[string]string `yaml:"dataSources"`

	// Annotations defines pod annotations of job,match option --annotation
	Annotations map[string]string `yaml:"annotations"`

	// Labels specify the job labels and it is work for pods
	Labels map[string]string `yaml:"labels"`

	// ImagePullSecrets stores image pull secrets,match option --image-pull-secrets
	ImagePullSecrets []string `yaml:"imagePullSecrets"`

	// HelmOptions stores the helm options
	HelmOptions []string `yaml:"-"`

	ModelName string `yaml:"modelName"` // --model-name

	ModelPath string `yaml:"modelPath"` // --model-path

	ModelVersion string `yaml:"modelVersion"` // --model-version

	MetricsPath string `yaml:"metricsPath"` // --metrics-path

	DatasetPath string `yaml:"datasetPath"` // --dataset-path

	Cpu string `yaml:"cpu"` // --cpu

	Memory string `yaml:"memory"` // --memory

	GPUCount int `yaml:"gpuCount"` // --gpus

	// for sync up source code
	SubmitSyncCodeArgs `yaml:",inline"`
}

type EvaluateJobInfo struct {
	UUID string `json:"uuid" yaml:"uuid"`

	JobID string `json:"jobId" yaml:"jobId"`

	Name string `json:"name" yaml:"name"`

	Namespace string `json:"namespace" yaml:"namespace"`

	ModelName string `json:"modelName" yaml:"modelName"`

	ModelPath string `json:"modelPath" yaml:"modelPath"`

	ModelVersion string `json:"modelVersion" yaml:"modelVersion"`

	MetricsPath string `json:"metricsPath" yaml:"metricsPath"`

	DatasetPath string `json:"datasetPath" yaml:"datasetPath"`

	// Information when was the last time the job was successfully scheduled.
	// +optional
	LastScheduleTime string `json:"lastScheduleTime" yaml:"lastScheduleTime"`

	// CreationTimestamp stores the creation timestamp of job
	CreationTimestamp string `json:"creationTimestamp" yaml:"creationTimestamp"`
}
