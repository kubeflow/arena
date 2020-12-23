package types

import (
	v1 "k8s.io/api/core/v1"
)

const PROMETHEUS_INSTALL_DOC_URL = "https://github.com/kubeflow/arena/blob/master/docs/userguide/9-top-job-gpu-metric.md"
const KUBE_SYSTEM_NAMESPACE = "kube-system"
const PROMETHEUS_SCHEME = "http"
const PROMETHEUS_SVC_LABEL = "kubernetes.io/name=Prometheus"
const POD_METRIC_TMP = `{__name__=~"%s", pod_name=~"%s"}`
const NODE_METRIC_TMP = `{__name__=~"%s", node_name=~"%s"}`
const KUBEFLOW_NAMESPACE = "kubeflow"

var GPU_METRIC_LIST = []string{"nvidia_gpu_duty_cycle", "nvidia_gpu_memory_used_bytes", "nvidia_gpu_memory_total_bytes"}

// PrometheusServer is used to define prometheus server
type PrometheusServer struct {
	Name          string
	ServiceLabels string
	Protocol      string
	Port          string
	Path          string
	MetricList    []string
	Service       *v1.Service
}

var SUPPORT_PROMETHEUS_SERVERS = []*PrometheusServer{
	// aliyun prometheus
	{
		Name:          "arms-prometheus-server",
		ServiceLabels: "kubernetes.io/name=prometheus-server",
		Protocol:      "http",
		Port:          "9090",
		Path:          "api/v1/query",
		MetricList: []string{
			"nvidia_gpu_duty_cycle",
			"nvidia_gpu_memory_used_bytes",
			"nvidia_gpu_memory_total_bytes",
		},
	},
	// aliyun prometheus
	{
		Name:          "arms-prometheus-admin",
		ServiceLabels: "kubernetes.io/name=prometheus-admin",
		Protocol:      "http",
		Port:          "9335",
		Path:          "api/v1/query",
		MetricList: []string{
			"nvidia_gpu_duty_cycle",
			"nvidia_gpu_memory_used_bytes",
			"nvidia_gpu_memory_total_bytes",
		},
	},
	{
		Name:          "default",
		ServiceLabels: "kubernetes.io/name=prometheus",
		Protocol:      "http",
		Port:          "9090",
		Path:          "api/v1/query",
		MetricList: []string{
			"nvidia_gpu_duty_cycle",
			"nvidia_gpu_memory_used_bytes",
			"nvidia_gpu_memory_total_bytes",
		},
	},
}

type PrometheusMetric struct {
	Status string               `json:"status,inline"`
	Data   PrometheusMetricData `json:"data,omitempty"`
}

type PrometheusMetricData struct {
	Result     []PrometheusMetricResult `json:"result"`
	ResultType string                   `json:"resultType"`
}

type PrometheusMetricResult struct {
	Metric map[string]string       `json:"metric"`
	Value  []PrometheusMetricValue `json:"value"`
}

type PrometheusMetricValue interface{}

type GpuMetricInfo struct {
	MetricName    string
	Value         string
	Time          float64
	PodName       string
	PodNamespace  string
	ContainerName string
	NodeName      string
	GPUUID        string
	Id            string
}

type JobGpuMetric map[string]PodGpuMetric

type PodGpuMetric map[string]*GpuMetric

// key of map is device id
type NodeGpuMetric map[string]*AdvancedGpuMetric

type GpuMetric struct {
	GpuDutyCycle   float64 `json:"gpuDutyCycle" yaml:"gpuDutyCycle"`
	GpuMemoryUsed  float64 `json:"usedGPUMemory" yaml:"usedGPUMemory"`
	GpuMemoryTotal float64 `json:"totalGPUMemory" yaml:"totalGPUMemory"`
}

type AdvancedGpuMetric struct {
	Id             string  `json:"id" yaml:"id"`
	UUID           string  `json:"uuid" yaml:"uuid"`
	Status         string  `json:"status" yaml:"status"`
	GpuDutyCycle   float64 `json:"gpuDutyCycle" yaml:"gpuDutyCycle"`
	GpuMemoryUsed  float64 `json:"usedGPUMemory" yaml:"usedGPUMemory"`
	GpuMemoryTotal float64 `json:"totalGPUMemory" yaml:"totalGPUMemory"`
}
