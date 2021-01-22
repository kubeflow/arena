package prometheus

import (
	"encoding/json"
	"fmt"
	"github.com/kubeflow/arena/pkg/apis/config"
	"github.com/kubeflow/arena/pkg/arenacache"
	"strconv"
	"strings"
	"time"

	"github.com/kubeflow/arena/pkg/apis/types"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type JobGpuMetric map[string]types.PodGpuMetric

func (m *JobGpuMetric) SetPodMetric(metric types.GpuMetricInfo) {
	v, err := strconv.ParseFloat(metric.Value, 64)
	if err != nil {
		return
	}
	metricMap := *m
	if _, ok := metricMap[metric.PodName]; !ok {
		metricMap[metric.PodName] = types.PodGpuMetric{}
	}

	podMetric := metricMap[metric.PodName]
	if _, ok := podMetric[metric.Id]; !ok {
		podMetric[metric.Id] = &types.GpuMetric{}
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

func (m JobGpuMetric) GetPodMetrics(podName string) types.PodGpuMetric {
	metricMap := m
	if podMetrics, ok := metricMap[podName]; ok {
		return podMetrics
	}
	return nil
}

func GpuMonitoringInstalled(client *kubernetes.Clientset) bool {
	server := GetPrometheusServer(client)
	if server == nil {
		return false
	}
	log.Debugf("get prometheus service: %v", server.Service)
	return true
	//gpuDeviceMetrics, _ := QueryMetricByPrometheus(client, server, "nvidia_gpu_num_devices")
	//return len(gpuDeviceMetrics) > 0
}

func GetPodsGpuInfo(client *kubernetes.Clientset, server *types.PrometheusServer, podNames []string) (JobGpuMetric, error) {
	jobMetric := &JobGpuMetric{}

	gpuMetrics, err := QueryMetricByPrometheus(client, server, fmt.Sprintf(types.POD_METRIC_TMP, strings.Join(types.GPU_METRIC_LIST, "|"), strings.Join(podNames, "|")))
	if err != nil {
		return nil, err
	}
	for _, metric := range gpuMetrics {
		jobMetric.SetPodMetric(metric)
	}
	return *jobMetric, nil
}

func QueryMetricByPrometheus(client *kubernetes.Clientset, server *types.PrometheusServer, query string) ([]types.GpuMetricInfo, error) {
	var gpuMetric []types.GpuMetricInfo

	svcClient := client.CoreV1()
	log.Debugf("query: %v", query)
	req := svcClient.Services(server.Service.Namespace).ProxyGet(server.Protocol, server.Service.Name, server.Port, server.Path, map[string]string{
		"query": query,
		"time":  strconv.FormatInt(time.Now().Unix(), 10),
	})
	log.Debugf("Query prometheus for by %s in ns %s", query, server.Service.Namespace)
	metric, err := req.DoRaw()
	if err != nil {
		log.Debugf("Query prometheus failed due to err %v", err)
		log.Debugf("Query prometheus failed due to result %s", string(metric))
	}
	var metricResponse *types.PrometheusMetric
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
		gpuMetric = append(gpuMetric, types.GpuMetricInfo{
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

func getMetricAverage(metrics []types.GpuMetricInfo) float64 {
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

// GetPrometheusServer get the matched prometheus server from the supported prometheus server
func GetPrometheusServer(client *kubernetes.Clientset) *types.PrometheusServer {
	for _, s := range types.SUPPORT_PROMETHEUS_SERVERS {
		svc := getPrometheusService(client, s.ServiceLabels)
		if svc == nil {
			continue
		}
		s.Service = svc
		return s
	}
	return nil
}

func getPrometheusService(client *kubernetes.Clientset, label string) *v1.Service {
	// find the prometheus server from all namespaces
	serviceList, err := listSvc(client, label)
	if err != nil {
		log.Debugf("Failed to get PrometheusServiceName: %v", err)
		return nil
	}
	if len(serviceList.Items) == 0 {
		log.Debugf("not found k8s services which own labels:[%v]", label)
		return nil
	}
	return serviceList.Items[0].DeepCopy()
}

func listSvc(client *kubernetes.Clientset, label string) (*v1.ServiceList, error) {
	if config.GetArenaConfiger().IsDaemonMode() {
		svcList := &v1.ServiceList{}
		return svcList, arenacache.GetCacheClient().ListResources(svcList, metav1.NamespaceAll, metav1.ListOptions{
			LabelSelector: label,
		})
	}

	return client.CoreV1().Services(metav1.NamespaceAll).List(metav1.ListOptions{
		LabelSelector: label,
	})
}

// {__name__=~"nvidia_gpu_duty_cycle|nvidia_gpu_memory_used_bytes|nvidia_gpu_memory_total_bytes", pod_name=~"tf-distributed-test-ps-0|tf-distributed-test-worker-0"}

func GetNodeGPUMetrics(client *kubernetes.Clientset, server *types.PrometheusServer, nodeNames []string) (map[string]types.NodeGpuMetric, error) {

	gpuMetrics, err := QueryMetricByPrometheus(client, server, fmt.Sprintf(types.NODE_METRIC_TMP, strings.Join(types.GPU_METRIC_LIST, "|"), strings.Join(nodeNames, "|")))
	if err != nil {
		return nil, err
	}
	return generateNodeGPUMetrics(gpuMetrics), nil
}

func generateNodeGPUMetrics(metrics []types.GpuMetricInfo) map[string]types.NodeGpuMetric {
	nodeMetrics := map[string]types.NodeGpuMetric{}
	for _, metric := range metrics {
		v, err := strconv.ParseFloat(metric.Value, 64)
		if err != nil {
			log.Debugf("failed to parse gpu duty cycle,reason: %v", err)
			continue
		}
		if nodeMetrics[metric.NodeName] == nil {
			nodeMetrics[metric.NodeName] = types.NodeGpuMetric{}
		}

		if nodeMetrics[metric.NodeName][metric.Id] == nil {
			nodeMetrics[metric.NodeName][metric.Id] = &types.AdvancedGpuMetric{
				Id:       metric.Id,
				UUID:     metric.GPUUID,
				PodNames: []string{},
			}
		}
		switch metric.MetricName {
		case "nvidia_gpu_duty_cycle":
			nodeMetrics[metric.NodeName][metric.Id].GpuDutyCycle = v
		case "nvidia_gpu_memory_used_bytes":
			nodeMetrics[metric.NodeName][metric.Id].GpuMemoryUsed = v
		case "nvidia_gpu_memory_total_bytes":
			nodeMetrics[metric.NodeName][metric.Id].GpuMemoryTotal = v
		}
		if metric.PodName != "" {
			podName := fmt.Sprintf("%v/%v", metric.PodNamespace, metric.PodName)
			nodeMetrics[metric.NodeName][metric.Id].PodNames = append(nodeMetrics[metric.NodeName][metric.Id].PodNames, podName)
		}
	}
	return nodeMetrics
}
