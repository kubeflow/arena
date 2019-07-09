// Copyright 2018 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package commands

import (
	"fmt"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	metricsapi "k8s.io/metrics/pkg/apis/metrics"
	metricsV1beta1api "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	metricsclientset "k8s.io/metrics/pkg/client/clientset_generated/clientset"
)

func getMetricsClient() (*metricsclientset.Clientset, error) {
	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	metricsClient, err := metricsclientset.NewForConfig(restConfig)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return metricsClient, nil
}

// introduced from `kubectl top node` command
func getNodeMetricsFromMetricsAPI(metricsClient metricsclientset.Interface, resourceName string, selector labels.Selector) (*metricsapi.NodeMetricsList, error) {
	var err error
	versionedMetrics := &metricsV1beta1api.NodeMetricsList{}
	mc := metricsClient.MetricsV1beta1()
	nm := mc.NodeMetricses()
	if resourceName != "" {
		m, err := nm.Get(resourceName, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		versionedMetrics.Items = []metricsV1beta1api.NodeMetrics{*m}
	} else {
		versionedMetrics, err = nm.List(metav1.ListOptions{LabelSelector: selector.String()})
		if err != nil {
			return nil, err
		}
	}
	metrics := &metricsapi.NodeMetricsList{}
	err = metricsV1beta1api.Convert_v1beta1_NodeMetricsList_To_metrics_NodeMetricsList(versionedMetrics, metrics, nil)
	if err != nil {
		return nil, err
	}
	return metrics, nil
}

func getNodeMetricsList() (*metricsapi.NodeMetricsList, error) {
	metricsClient, err := getMetricsClient()
	if err != nil {
		return nil, err
	}
	selector := labels.Everything()
	metricsList, err := getNodeMetricsFromMetricsAPI(metricsClient, "", selector)
	if err != nil {
		return nil, err
	}
	return metricsList, nil
}

func getNodeMetrics() (map[string]v1.ResourceList, error) {
	metricsList, err := getNodeMetricsList()
	if err != nil {
		return nil, err
	}
	res := make(map[string]v1.ResourceList, len(metricsList.Items))
	for _, m := range metricsList.Items {
		res[m.ObjectMeta.Name] = m.Usage
	}
	return res, nil
}

// introduced from `kubectl top pod` command
func getPodMetricsFromMetricsAPI(metricsClient metricsclientset.Interface, namespace, resourceName string, allNamespaces bool, selector labels.Selector) (*metricsapi.PodMetricsList, error) {
	var err error
	ns := metav1.NamespaceAll
	if !allNamespaces {
		ns = namespace
	}
	versionedMetrics := &metricsV1beta1api.PodMetricsList{}
	if resourceName != "" {
		m, err := metricsClient.MetricsV1beta1().PodMetricses(ns).Get(resourceName, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		versionedMetrics.Items = []metricsV1beta1api.PodMetrics{*m}
	} else {
		versionedMetrics, err = metricsClient.MetricsV1beta1().PodMetricses(ns).List(metav1.ListOptions{LabelSelector: selector.String()})
		if err != nil {
			return nil, err
		}
	}
	metrics := &metricsapi.PodMetricsList{}
	err = metricsV1beta1api.Convert_v1beta1_PodMetricsList_To_metrics_PodMetricsList(versionedMetrics, metrics, nil)
	if err != nil {
		return nil, err
	}
	return metrics, nil
}

func getPodMetricsList() (*metricsapi.PodMetricsList, error) {
	metricsClient, err := getMetricsClient()
	if err != nil {
		return nil, err
	}
	selector := labels.Everything()
	metricsList, err := getPodMetricsFromMetricsAPI(metricsClient, "", "", true, selector)
	if err != nil {
		return nil, err
	}
	return metricsList, nil
}

func getPodMetrics() (map[string]v1.ResourceList, error) {
	podMetricsList, err := getPodMetricsList()
	if err != nil {
		return nil, err
	}
	res := make(map[string]v1.ResourceList, len(podMetricsList.Items))
	measuredResources := []v1.ResourceName{
		v1.ResourceCPU,
		v1.ResourceMemory,
	}
	for _, m := range podMetricsList.Items {
		for _, r := range measuredResources {
			if res[m.ObjectMeta.Name] == nil {
				res[m.ObjectMeta.Name] = make(map[v1.ResourceName]resource.Quantity)
			}
			res[m.ObjectMeta.Name][r], _ = resource.ParseQuantity("0")
		}
	}
	// add each container's metrics in one pod
	for _, m := range podMetricsList.Items {
		for _, c := range m.Containers {
			for _, r := range measuredResources {
				quantity := res[m.ObjectMeta.Name][r]
				quantity.Add(c.Usage[r])
				res[m.ObjectMeta.Name][r] = quantity
			}
		}
	}
	return res, nil
}

// calculate the CPU count of each node
func calculateNodeCPU(nodeInfo NodeInfo, nodeMetrics map[string]v1.ResourceList) (int64, int64) {
	usedCPU := nodeMetrics[nodeInfo.node.Name][v1.ResourceCPU]
	return cpuInNode(nodeInfo.node), usedCPU.MilliValue()
}

// calculate the Memory count of each node
func calculateNodeMemory(nodeInfo NodeInfo, nodeMetrics map[string]v1.ResourceList) (int64, int64) {
	usedMemory := nodeMetrics[nodeInfo.node.Name][v1.ResourceMemory]
	return memoryInNode(nodeInfo.node), usedMemory.Value()
}

// CPU usage in a pod
func cpuInPod(pod v1.Pod, podMetrics map[string]v1.ResourceList) int64 {
	usedCPU := podMetrics[pod.Name][v1.ResourceCPU]
	return usedCPU.MilliValue()
}

// memory usage in a pod
func memoryInPod(pod v1.Pod, podMetrics map[string]v1.ResourceList) int64 {
	usedMemory := podMetrics[pod.Name][v1.ResourceMemory]
	return usedMemory.Value()
}

// The way to get CPU Count of Node
func cpuInNode(node v1.Node) int64 {
	val, ok := node.Status.Capacity[v1.ResourceCPU]
	if !ok {
		return 0
	}
	return val.MilliValue()
}

// The way to get Memory Count of Node
func memoryInNode(node v1.Node) int64 {
	val, ok := node.Status.Capacity[v1.ResourceMemory]
	if !ok {
		return 0
	}
	return val.Value()
}

// filter out the pods which use resources
func resourcePods(pods []v1.Pod, podMetrics map[string]v1.ResourceList) (podsWithResource []v1.Pod) {
	for _, pod := range pods {
		if gpuInPod(pod) > 0 || cpuInPod(pod, podMetrics) > 0 || memoryInPod(pod, podMetrics) > 0 {
			podsWithResource = append(podsWithResource, pod)
		}
	}
	return podsWithResource
}
