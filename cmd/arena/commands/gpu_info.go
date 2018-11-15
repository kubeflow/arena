package commands

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"encoding/json"
	"fmt"
	"strings"
	log "github.com/sirupsen/logrus"
	v12 "k8s.io/api/core/v1"
	"strconv"
	"time"
)

const KUBE_SYSTEM_NAMESPACE = "kube-system"
const PROMETHEUS_SCHEME = "http"
const PROMETHEUS_SVC_LABEL = "kubernetes.io/name=Prometheus"
//
type PrometheusMetric struct {
	Status string `json:"status,inline"`
	Data PrometheusMetricData `json:"data,omitempty"`
}

type PrometheusMetricData struct {
	Result []PrometheusMetricResult `json:"result"`
	ResultType string `json:"resultType"`
}

type PrometheusMetricResult struct {
	Metric map[string]string          `json:"metric"`
	Value []PrometheusMetricValue `json:"value"`
}

type PrometheusMetricValue interface{}

type GpuMetricInfo struct {
	MetricName string
	Value string
	Time float64
	PodName string
	PodNamespace string
	ContainerName string
	NodeName string
	GPUUID string
}


const PROMETHEUS_METRIC_QUERL_TPM = `avg (%s{pod_name=~"%s", container_name!=""}) by (pod_name)`

type JobGpuMetric struct {
	GpuDutyCycle float64
	GpuMemoryUsed float64
	GpuMemoryTotal float64
}

func GpuMonitoringInstalled(client *kubernetes.Clientset) bool {
	prometheusServiceName := GetPrometheusServiceName(client)
	if (prometheusServiceName == "") {
		return false
	}
	gpuDeviceMetrics, _ := GetGpuInfo(client, prometheusServiceName, "nvidia_gpu_num_devices")
	return len(gpuDeviceMetrics) > 0
}



func GetJobGpuMetric(client *kubernetes.Clientset, job TrainingJob) (*JobGpuMetric, error) {
	jobStatus := job.GetStatus()
	runningPods := []string{}
	if jobStatus == "RUNNING" {
		pods := job.AllPods()
		for _, pod := range pods {
			if pod.Status.Phase == v12.PodPending {
				continue
			}
			runningPods = append(runningPods, pod.Name)
		}
	}

	prometheusServiceName := GetPrometheusServiceName(client)
	if (prometheusServiceName == "") {
		return nil, nil
	}
	gpuDutyCycleMetrics, _ := GetGpuInfo(client, prometheusServiceName, fmt.Sprintf(PROMETHEUS_METRIC_QUERL_TPM, "nvidia_gpu_duty_cycle", strings.Join(runningPods, "|")))
	gpuMemoryUsedMetrics, _ := GetGpuInfo(client, prometheusServiceName, fmt.Sprintf(PROMETHEUS_METRIC_QUERL_TPM, "nvidia_gpu_memory_used_bytes", strings.Join(runningPods, "|")) )
	gpuMemoryTotaldMetrics, _ := GetGpuInfo(client, prometheusServiceName, fmt.Sprintf(PROMETHEUS_METRIC_QUERL_TPM, "nvidia_gpu_memory_total_bytes", strings.Join(runningPods, "|")) )
	return &JobGpuMetric {
		GpuDutyCycle: getMetricAverage(gpuDutyCycleMetrics),
		GpuMemoryUsed: getMetricAverage(gpuMemoryUsedMetrics),
		GpuMemoryTotal: getMetricAverage(gpuMemoryTotaldMetrics),
	}, nil
}

func GetGpuInfo(client *kubernetes.Clientset, prometheusServiceName string, query string) ([]GpuMetricInfo, error) {
	var gpuMetric []GpuMetricInfo

	svcClient := client.CoreV1()
	req :=  svcClient.Services(KUBE_SYSTEM_NAMESPACE).ProxyGet(PROMETHEUS_SCHEME, prometheusServiceName, "9090", "api/v1/query", map[string]string{
		"query": query,
		"time": strconv.FormatInt(time.Now().Unix(), 10),
	})
	metric, _ := req.DoRaw()
	var metricResponse *PrometheusMetric
	err := json.Unmarshal(metric, &metricResponse)
	if err != nil {
		log.Errorf("failed to unmarshall heapster response: %v", err)
		return gpuMetric, fmt.Errorf("failed to unmarshall heapster response: %v", err)
	}
	if metricResponse.Status != "success" {
		log.Errorf("failed to query prometheus, status: %s", metricResponse.Status)
		return gpuMetric, fmt.Errorf("failed to query prometheus, status: %s", metricResponse.Status)
	}
	if len(metricResponse.Data.Result) == 0 {
		log.Errorf("gpu metric is not exist in prometheus for query  %s", query)
		return gpuMetric, fmt.Errorf("gpu metric is not exist in prometheusfor query  %s", query)
	}
	for _, m := range metricResponse.Data.Result {
		gpuMetric = append(gpuMetric, GpuMetricInfo{
			MetricName: m.Metric["__name__"],
			PodNamespace: m.Metric["namespace_name"],
			NodeName: m.Metric["node_name"],
			PodName: m.Metric["pod_name"],
			ContainerName: m.Metric["container_name"],
			GPUUID: m.Metric["uuid"],
			Value: m.Value[1].(string),
			Time: m.Value[0].(float64),
		})
	}
	return gpuMetric, nil
}

func getMetricAverage(metrics []GpuMetricInfo) float64 {
	var result float64
	result = 0
	for _, metric := range metrics {
		v, _ := strconv.ParseFloat(metric.Value, 64)
		result = result + v
	}
	if result == 0 {
		return result
	}
	result = result / float64(len(metrics))
	return result
}

func GetPrometheusServiceName(client *kubernetes.Clientset) string {
	services, err := client.CoreV1().Services(KUBE_SYSTEM_NAMESPACE).List(v1.ListOptions{
		LabelSelector: PROMETHEUS_SVC_LABEL,
	})
	if err != nil {
		return ""
	}
	if len(services.Items) == 0 {
		return ""
	}
	prometheusService := services.Items[0]
	return prometheusService.Name

}