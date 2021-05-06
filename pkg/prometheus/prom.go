package prometheus

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/kubeflow/arena/pkg/apis/types"
	log "github.com/sirupsen/logrus"
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
		v = math.Trunc(v/(1024*1024*1024)) * (1024 * 1024 * 1024)
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

func GetPodsGpuInfo(client *kubernetes.Clientset, podNames []string) (JobGpuMetric, error) {
	jobMetric := &JobGpuMetric{}
	query := fmt.Sprintf(types.POD_METRIC_TMP, strings.Join(types.GPU_METRIC_LIST, "|"), strings.Join(podNames, "|"))
	gpuMetrics, err := QueryPrometheusMetrics(client, query)
	if err != nil {
		return nil, err
	}
	if gpuMetrics == nil {
		return nil, fmt.Errorf("failed to get gpu metrics,because the result of querying is null")
	}
	for _, metric := range gpuMetrics {
		jobMetric.SetPodMetric(metric)
	}
	return *jobMetric, nil
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

// {__name__=~"nvidia_gpu_duty_cycle|nvidia_gpu_memory_used_bytes|nvidia_gpu_memory_total_bytes", pod_name=~"tf-distributed-test-ps-0|tf-distributed-test-worker-0"}

func GetNodeGPUMetrics(client *kubernetes.Clientset, nodeNames []string) (map[string]types.NodeGpuMetric, error) {

	query := fmt.Sprintf(types.NODE_METRIC_TMP, strings.Join(types.GPU_METRIC_LIST, "|"), strings.Join(nodeNames, "|"))
	gpuMetrics, err := QueryPrometheusMetrics(client, query)
	if err != nil {
		return nil, err
	}
	if gpuMetrics == nil {
		return nil, fmt.Errorf("failed to get node gpu metrics,because the result of querying is null")
	}
	return generateNodeGPUMetrics(gpuMetrics), nil
}

func generateNodeGPUMetrics(metrics []types.GpuMetricInfo) map[string]types.NodeGpuMetric {
	nodeMetrics := map[string]types.NodeGpuMetric{}
	shareModeUsedGPUMemory := map[string]map[string][]float64{}
	for _, metric := range metrics {
		if metric.PodNamespace == "" {
			continue
		}
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
		if shareModeUsedGPUMemory[metric.NodeName] == nil {
			shareModeUsedGPUMemory[metric.NodeName] = map[string][]float64{}
		}
		if shareModeUsedGPUMemory[metric.NodeName][metric.Id] == nil {
			shareModeUsedGPUMemory[metric.NodeName][metric.Id] = []float64{}
		}
		switch metric.MetricName {
		case "nvidia_gpu_duty_cycle":
			nodeMetrics[metric.NodeName][metric.Id].GpuDutyCycle = v
		case "nvidia_gpu_memory_used_bytes":
			nodeMetrics[metric.NodeName][metric.Id].GpuMemoryUsed = v
			if metric.AllocateMode == "share" {
				shareModeUsedGPUMemory[metric.NodeName][metric.Id] = append(shareModeUsedGPUMemory[metric.NodeName][metric.Id], v)
			}
		case "nvidia_gpu_memory_total_bytes":
			v = math.Trunc(v/(1024*1024*1024)) * (1024 * 1024 * 1024)
			nodeMetrics[metric.NodeName][metric.Id].GpuMemoryTotal = v
		}
		if metric.PodName != "" {
			podName := fmt.Sprintf("%v/%v", metric.PodNamespace, metric.PodName)
			nodeMetrics[metric.NodeName][metric.Id].PodNames = append(nodeMetrics[metric.NodeName][metric.Id].PodNames, podName)
		}
	}
	for nodeName, allUsedGPUMemory := range shareModeUsedGPUMemory {
		for gpuId, usedGPUMemory := range allUsedGPUMemory {
			if len(usedGPUMemory) > 0 {
				total := float64(0)
				for _, v := range usedGPUMemory {
					total += v
				}
				nodeMetrics[nodeName][gpuId].GpuMemoryUsed = total
			}
		}
	}
	return nodeMetrics
}
