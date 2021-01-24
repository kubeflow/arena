package prometheus

import (
	"context"
	"os"
	"reflect"
	"sync"
	"time"

	"encoding/json"
	"fmt"
	"strconv"

	"github.com/kubeflow/arena/pkg/apis/config"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/prometheus/common/model"
	log "github.com/sirupsen/logrus"

	"github.com/kubeflow/arena/pkg/arenacache"
	"github.com/prometheus/client_golang/api"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var globalPrometheusClient promv1.API
var prometheusOnce sync.Once

func GetPrometheusClient() promv1.API {
	prometheusOnce.Do(func() {
		globalPrometheusClient = getPrometheusClient(config.GetArenaConfiger().GetConfigsFromConfigFile())
	})
	return globalPrometheusClient
}

// address format should like: http://123.123.123.123:9000
func getPrometheusAddress(configs map[string]string) string {
	if os.Getenv("PROMETHEUS_ADDRESS") != "" {
		return os.Getenv("PROMETHEUS_ADDRESS")
	}
	for name, value := range configs {
		if name == "prometheus_address" {
			return value
		}
	}
	return ""
}

func getPrometheusClient(configs map[string]string) promv1.API {
	address := getPrometheusAddress(configs)
	if address == "" {
		log.Debugf("not set prometheus address at env PROMETHEUS_ADDRESS or the arena configuration file,skip to create client")
		return nil
	}
	client, err := api.NewClient(api.Config{
		Address: address,
	})
	if err != nil {
		log.Debugf("failed to create prometheus client,reason: %v", err)
		return nil
	}

	return promv1.NewAPI(client)
}

func QueryPrometheusMetrics(client *kubernetes.Clientset, query string) ([]types.GpuMetricInfo, error) {
	v1api := GetPrometheusClient()
	if v1api != nil {
		return queryPrometheusMetricsByAddress(query)
	}
	return queryPrometheusMetricsProxyByAPIServer(client, query)
}

// queryPrometheusMetricsByAddress is used when the prometheus server address has been specified in env PROMETHEUS_ADDRESS
// or arena configuration file.
func queryPrometheusMetricsByAddress(query string) ([]types.GpuMetricInfo, error) {
	log.Debugf("the prom sql is %v", query)
	gpuMetrics := []types.GpuMetricInfo{}
	v1api := GetPrometheusClient()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	result, warnings, err := v1api.Query(ctx, query, time.Now())
	if err != nil {
		log.Debugf("Error querying Prometheus by %v: %v\n", query, err)
		return nil, err
	}
	if len(warnings) > 0 {
		log.Debugf("Warnings: %v", warnings)
	}
	log.Debugf("get metrics: %v", result.String())
	switch result.Type() {
	case model.ValVector:
		vectorVal := result.(model.Vector)
		for _, v := range vectorVal {
			gpuMetric := types.GpuMetricInfo{
				Time:  float64(v.Timestamp),
				Value: v.Value.String(),
			}
			for labelKey, labelVal := range v.Metric {
				switch string(labelKey) {
				case "__name__":
					gpuMetric.MetricName = string(labelVal)
				case "namespace_name":
					gpuMetric.PodNamespace = string(labelVal)
				case "node_name":
					gpuMetric.NodeName = string(labelVal)
				case "pod_name":
					gpuMetric.PodName = string(labelVal)
				case "container_name":
					gpuMetric.ContainerName = string(labelVal)
				case "uuid":
					gpuMetric.GPUUID = string(labelVal)
				case "minor_number":
					gpuMetric.Id = string(labelVal)
				}
			}
			gpuMetrics = append(gpuMetrics, gpuMetric)
		}
	//case model.ValScalar:
	//	scalarVal := val.(*model.Scalar)

	default:
		return nil, fmt.Errorf("failed to get metrics, unknown metric type %v,we want model.Vector", reflect.TypeOf(result.Type()))
	}
	return gpuMetrics, nil
}

// queryPrometheusMetricsProxyByAPIServer is used to query metrics proxy by k8s api server.
func queryPrometheusMetricsProxyByAPIServer(client *kubernetes.Clientset, query string) ([]types.GpuMetricInfo, error) {
	gpuMetric := []types.GpuMetricInfo{}
	server := getPrometheusServer(client)
	if server == nil {
		log.Debugf("the prometheus is not installed,skip to get the gpu metrics")
		return gpuMetric, nil
	}
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

func prometheusInstalled(client *kubernetes.Clientset) bool {
	server := getPrometheusServer(client)
	if server == nil {
		return false
	}
	log.Debugf("get prometheus service: %v", server.Service)
	return true
	//gpuDeviceMetrics, _ := QueryMetricByPrometheus(client, server, "nvidia_gpu_num_devices")
	//return len(gpuDeviceMetrics) > 0
}

func GetPrometheusServer(client *kubernetes.Clientset) *types.PrometheusServer {
	return getPrometheusServer(client)
}

// GetPrometheusServer get the matched prometheus server from the supported prometheus server
func getPrometheusServer(client *kubernetes.Clientset) *types.PrometheusServer {
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
