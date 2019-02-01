package commands

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const PROMETHEUS_INSTALL_DOC_URL = "https://github.com/kubeflow/arena/blob/master/docs/userguide/9-top-job-gpu-metric.md"
const KUBE_SYSTEM_NAMESPACE = "kube-system"
const PROMETHEUS_SCHEME = "http"
const PROMETHEUS_SVC_LABEL = "kubernetes.io/name=Prometheus"
const POD_METRIC_TMP = `{__name__=~"%s", pod_name=~"%s"}`
const KUBEFLOW_NAMESPACE = "kubeflow"

var GPU_METRIC_LIST = []string{"nvidia_gpu_duty_cycle", "nvidia_gpu_memory_used_bytes", "nvidia_gpu_memory_total_bytes"}

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

type GpuMetric struct {
	GpuDutyCycle   float64
	GpuMemoryUsed  float64
	GpuMemoryTotal float64
}

func (m *JobGpuMetric) SetPodMetric(metric GpuMetricInfo) {
	v, err := strconv.ParseFloat(metric.Value, 64)
	if err != nil {
		return
	}
	metricMap := *m
	if _, ok := metricMap[metric.PodName]; !ok {
		metricMap[metric.PodName] = PodGpuMetric{}
	}

	podMetric := metricMap[metric.PodName]
	if _, ok := podMetric[metric.Id]; !ok {
		podMetric[metric.Id] = &GpuMetric{}
	}
	podGPUMetric := podMetric[metric.Id]
	switch metric.MetricName {
	case "nvidia_gpu_duty_cycle":
		podGPUMetric.GpuDutyCycle = v
	case "nvidia_gpu_memory_used_bytes":
		podGPUMetric.GpuMemoryUsed = v
	case "nvidia_gpu_memory_total_bytes":
		podGPUMetric.GpuMemoryTotal = v
	}
}

func (m JobGpuMetric) GetPodMetrics(podName string) PodGpuMetric {
	metricMap := m
	if podMetrics, ok := metricMap[podName]; ok {
		return podMetrics
	}
	return nil
}

func GpuMonitoringInstalled(client *kubernetes.Clientset) bool {
	prometheusServiceName, namespace := GetPrometheusServiceName(client)
	if prometheusServiceName == "" {
		return false
	}
	gpuDeviceMetrics, _ := QueryMetricByPrometheus(client, prometheusServiceName, namespace, "nvidia_gpu_num_devices")
	return len(gpuDeviceMetrics) > 0
}

func GetJobGpuMetric(client *kubernetes.Clientset, job TrainingJob) (jobMetric JobGpuMetric, err error) {
	runningPods := []string{}
	jobStatus := job.GetStatus()
	if jobStatus == "RUNNING" {
		pods := job.AllPods()
		for _, pod := range pods {
			if pod.Status.Phase == corev1.PodPending {
				continue
			}
			runningPods = append(runningPods, pod.Name)
		}
	}
	prometheusServiceName, namespace := GetPrometheusServiceName(client)
	if prometheusServiceName == "" {
		return
	}
	podsMetrics, err := GetPodsGpuInfo(client, prometheusServiceName, namespace, runningPods)
	return podsMetrics, err
}

func GetPodsGpuInfo(client *kubernetes.Clientset, prometheusServiceName string, namespace string, podNames []string) (JobGpuMetric, error) {
	jobMetric := &JobGpuMetric{}

	gpuMetrics, err := QueryMetricByPrometheus(client, prometheusServiceName, namespace, fmt.Sprintf(POD_METRIC_TMP, strings.Join(GPU_METRIC_LIST, "|"), strings.Join(podNames, "|")))
	if err != nil {
		return nil, err
	}
	for _, metric := range gpuMetrics {
		jobMetric.SetPodMetric(metric)
	}
	return *jobMetric, nil
}

func QueryMetricByPrometheus(client *kubernetes.Clientset, prometheusServiceName string, namespace string, query string) ([]GpuMetricInfo, error) {
	var gpuMetric []GpuMetricInfo

	svcClient := client.CoreV1()
	req := svcClient.Services(namespace).ProxyGet(PROMETHEUS_SCHEME, prometheusServiceName, "9090", "api/v1/query", map[string]string{
		"query": query,
		"time":  strconv.FormatInt(time.Now().Unix(), 10),
	})
	log.Debugf("Query prometheus for by %s in ns %s", query, namespace)
	metric, err := req.DoRaw()
	if err != nil {
		log.Debugf("Query prometheus failed due to err %v", err)
		log.Debugf("Query prometheus failed due to result %s", string(metric))
	}
	var metricResponse *PrometheusMetric
	err = json.Unmarshal(metric, &metricResponse)
	log.Debugf("Prometheus metric:%v", metricResponse)
	if err != nil {
		log.Errorf("failed to unmarshall heapster response: %v", err)
		return gpuMetric, fmt.Errorf("failed to unmarshall heapster response: %v", err)
	}
	if metricResponse.Status != "success" {
		log.Errorf("failed to query prometheus, status: %s", metricResponse.Status)
		return gpuMetric, fmt.Errorf("failed to query prometheus, status: %s", metricResponse.Status)
	}
	if len(metricResponse.Data.Result) == 0 {
		log.Debugf("gpu metric is not exist in prometheus for query  %s", query)
		return gpuMetric, nil
	}
	for _, m := range metricResponse.Data.Result {
		gpuMetric = append(gpuMetric, GpuMetricInfo{
			MetricName:    m.Metric["__name__"],
			PodNamespace:  m.Metric["namespace_name"],
			NodeName:      m.Metric["node_name"],
			PodName:       m.Metric["pod_name"],
			ContainerName: m.Metric["container_name"],
			GPUUID:        m.Metric["uuid"],
			Id:            m.Metric["minor_number"],
			Value:         m.Value[1].(string),
			Time:          m.Value[0].(float64),
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

/**
* Get Prometheus from different namespaces
 */
func GetPrometheusServiceName(client *kubernetes.Clientset) (name string, ns string) {

	// 1. Locate the service in the current namespace
	name = getPrometheusServiceName(client, namespace)
	if len(name) > 0 {
		return name, namespace
	}

	// 2. Locate the service in the arena namespace
	name = getPrometheusServiceName(client, arenaNamespace)
	if len(name) > 0 {
		return name, arenaNamespace
	}

	// 3. Locate the service in the arena namespace
	name = getPrometheusServiceName(client, KUBEFLOW_NAMESPACE)
	if len(name) > 0 {
		return name, KUBEFLOW_NAMESPACE
	}

	// 4. Locate the service in the current namespace
	name = getPrometheusServiceName(client, KUBE_SYSTEM_NAMESPACE)
	if len(name) > 0 {
		return name, KUBE_SYSTEM_NAMESPACE
	}

	return "", ""
}

func getPrometheusServiceName(client *kubernetes.Clientset, namespace string) (name string) {
	services, err := client.CoreV1().Services(namespace).List(metav1.ListOptions{
		LabelSelector: PROMETHEUS_SVC_LABEL,
	})
	if err != nil {
		log.Debugf("Failed to get PrometheusServiceName from %s due to %v", namespace, err)
	} else if len(services.Items) > 0 {
		return services.Items[0].Name
	}

	return ""
}

func SortMapKeys(podMetric PodGpuMetric) []string {
	var keys []string
	for k, _ := range podMetric {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
