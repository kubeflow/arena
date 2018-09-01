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
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/pkg/kubectl/metricsutil"
)

var (
	showDetails bool
)

type NodeInfo struct {
	node v1.Node
	pods []v1.Pod
}

func NewTopNodeCommand() *cobra.Command {

	var command = &cobra.Command{
		Use:   "node",
		Short: "Display Resource (GPU) usage of nodes.",
		Run: func(cmd *cobra.Command, args []string) {
			setupKubeconfig()
			client, err := initKubeClient()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			allPods, err := acquireAllActivePods(client)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			nd := newNodeDescriber(client, allPods)
			nodeInfos, err := nd.getAllNodeInfos()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			heapsterMetricsClient := nd.newHeapsterMetricsClient()
			displayTopNode(nodeInfos, heapsterMetricsClient)
		},
	}

	command.Flags().BoolVarP(&showDetails, "details", "d", false, "Display details")
	return command
}

type NodeDescriber struct {
	client  *kubernetes.Clientset
	allPods []v1.Pod
}

func newNodeDescriber(client *kubernetes.Clientset, pods []v1.Pod) *NodeDescriber {
	return &NodeDescriber{
		client:  client,
		allPods: pods,
	}
}

func (d *NodeDescriber) newHeapsterMetricsClient() *metricsutil.HeapsterMetricsClient {
	return metricsutil.DefaultHeapsterMetricsClient(d.client.CoreV1())
}

func (d *NodeDescriber) getAllNodeInfos() ([]NodeInfo, error) {
	nodeInfoList := []NodeInfo{}
	nodeList, err := d.client.CoreV1().Nodes().List(metav1.ListOptions{})

	if err != nil {
		return nodeInfoList, err
	}

	for _, node := range nodeList.Items {
		pods := d.getPodsFromNode(node)
		nodeInfoList = append(nodeInfoList, NodeInfo{
			node: node,
			pods: pods,
		})
	}

	return nodeInfoList, nil
}

func (d *NodeDescriber) getPodsFromNode(node v1.Node) []v1.Pod {
	pods := []v1.Pod{}
	for _, pod := range d.allPods {
		if pod.Spec.NodeName == node.Name {
			pods = append(pods, pod)
		}
	}

	return pods
}

func getNodeMetrics(heapsterMetricsClient *metricsutil.HeapsterMetricsClient) map[string]v1.ResourceList {
	nodeMetrics, err := heapsterMetricsClient.GetNodeMetrics("", "")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if len(nodeMetrics) == 0 {
		fmt.Println("node metrics not available yet")
		os.Exit(1)
	}
	res := make(map[string]v1.ResourceList, len(nodeMetrics))
	for _, m := range nodeMetrics {
		res[m.ObjectMeta.Name] = m.Usage
	}
	return res
}

func getPodMetrics(heapsterMetricsClient *metricsutil.HeapsterMetricsClient) map[string]v1.ResourceList {
	podMetrics, err := heapsterMetricsClient.GetPodMetrics("", "", true, labels.Everything())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if len(podMetrics) == 0 {
		fmt.Println("pod metrics not available yet")
		os.Exit(1)
	}
	res := make(map[string]v1.ResourceList, len(podMetrics))
	measuredResources := []v1.ResourceName{
		v1.ResourceCPU,
		v1.ResourceMemory,
	}
	for _, m := range podMetrics {
		for _, r := range measuredResources {
			if res[m.ObjectMeta.Name] == nil {
				res[m.ObjectMeta.Name] = make(map[v1.ResourceName]resource.Quantity)
			}
			res[m.ObjectMeta.Name][r], _ = resource.ParseQuantity("0")
		}
	}
	// add each container's metrics in one pod
	for _, m := range podMetrics {
		for _, c := range m.Containers {
			for _, r := range measuredResources {
				quantity := res[m.ObjectMeta.Name][r]
				quantity.Add(c.Usage[r])
				res[m.ObjectMeta.Name][r] = quantity
			}
		}
	}
	return res
}

func displayTopNode(nodes []NodeInfo, heapsterMetricsClient *metricsutil.HeapsterMetricsClient) {
	if showDetails {
		displayTopNodeDetails(nodes, heapsterMetricsClient)
	} else {
		displayTopNodeSummary(nodes, heapsterMetricsClient)
	}
}

func displayTopNodeSummary(nodeInfos []NodeInfo, heapsterMetricsClient *metricsutil.HeapsterMetricsClient) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	var (
		totalGPUsInCluster     int64
		allocatedGPUsInCluster int64
		totalCPUsInCluster     int64
		usedCPUsInCluster      int64
		totalMemoryInCluster   int64
		usedMemoryInCluster    int64
	)
	nodeMetrics := getNodeMetrics(heapsterMetricsClient)

	fmt.Fprintf(w, "NAME\tIPADDRESS\tROLE\tGPU(Total)\tGPU(Allocated)\tCPU(Total)\tCPU(Usage)\tMemory(Total)\tMemory(Usage)\n")
	for _, nodeInfo := range nodeInfos {
		totalGPU, allocatedGPU := calculateNodeGPU(nodeInfo)
		totalCPU, usedCPU := calculateNodeCPU(nodeInfo, nodeMetrics)
		totalMemory, usedMemory := calculateNodeMemory(nodeInfo, nodeMetrics)
		totalGPUsInCluster += totalGPU
		allocatedGPUsInCluster += allocatedGPU
		totalCPUsInCluster += totalCPU
		usedCPUsInCluster += usedCPU
		totalMemoryInCluster += totalMemory
		usedMemoryInCluster += usedMemory

		address := "unknown"
		if len(nodeInfo.node.Status.Addresses) > 0 {
			// address = nodeInfo.node.Status.Addresses[0].Address
			for _, addr := range nodeInfo.node.Status.Addresses {
				if addr.Type == v1.NodeInternalIP {
					address = addr.Address
					break
				}
			}
		}

		role := strings.Join(findNodeRoles(&nodeInfo.node), ",")
		if len(role) == 0 {
			role = "<none>"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n", nodeInfo.node.Name,
			address,
			role,
			strconv.FormatInt(totalGPU, 10),
			strconv.FormatInt(allocatedGPU, 10),
			resource.NewMilliQuantity(totalCPU, resource.DecimalSI).String(),
			resource.NewMilliQuantity(usedCPU, resource.DecimalSI).String(),
			fmt.Sprintf("%vMi", totalMemory/(1024*1024)),
			fmt.Sprintf("%vMi", usedMemory/(1024*1024)))
	}
	fmt.Fprintf(w, "--------------------------------------------------------------------------------------------------------------------------------------\n")
	fmt.Fprintf(w, "Allocated/Total GPUs In Cluster:\n")
	log.Debugf("gpu: %s, allocated GPUs %s", strconv.FormatInt(totalGPUsInCluster, 10),
		strconv.FormatInt(allocatedGPUsInCluster, 10))
	var gpuUsage float64 = 0
	if totalGPUsInCluster > 0 {
		gpuUsage = float64(allocatedGPUsInCluster) / float64(totalGPUsInCluster) * 100
	}
	fmt.Fprintf(w, "%s/%s (%d%%)\t\n",
		strconv.FormatInt(allocatedGPUsInCluster, 10),
		strconv.FormatInt(totalGPUsInCluster, 10),
		int64(gpuUsage))
	fmt.Fprintf(w, "Used/Total CPUs In Cluster:\n")
	log.Debugf("cpu: %s, used CPUs %s", resource.NewMilliQuantity(totalCPUsInCluster, resource.DecimalSI).String(),
		resource.NewMilliQuantity(usedCPUsInCluster, resource.DecimalSI).String())
	var cpuUsage float64 = 0
	if totalCPUsInCluster > 0 {
		cpuUsage = float64(usedCPUsInCluster) / float64(totalCPUsInCluster) * 100
	}
	fmt.Fprintf(w, "%s/%s (%d%%)\t\n",
		resource.NewMilliQuantity(usedCPUsInCluster, resource.DecimalSI).String(),
		resource.NewMilliQuantity(totalCPUsInCluster, resource.DecimalSI).String(),
		int64(cpuUsage))
	fmt.Fprintf(w, "Used/Total Memory In Cluster:\n")
	log.Debugf("Memory: %s, used Memory %s", fmt.Sprintf("%vMi", usedMemoryInCluster/(1024*1024)),
		fmt.Sprintf("%vMi", totalMemoryInCluster/(1024*1024)))
	var memoryUsage float64 = 0
	if totalMemoryInCluster > 0 {
		memoryUsage = float64(usedMemoryInCluster) / float64(totalMemoryInCluster) * 100
	}
	fmt.Fprintf(w, "%s/%s (%d%%)\t\n",
		fmt.Sprintf("%vMi", usedMemoryInCluster/(1024*1024)),
		fmt.Sprintf("%vMi", totalMemoryInCluster/(1024*1024)),
		int64(memoryUsage))

	_ = w.Flush()
}

func displayTopNodeDetails(nodeInfos []NodeInfo, heapsterMetricsClient *metricsutil.HeapsterMetricsClient) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	var (
		totalGPUsInCluster     int64
		allocatedGPUsInCluster int64
		totalCPUsInCluster     int64
		usedCPUsInCluster      int64
		totalMemoryInCluster   int64
		usedMemoryInCluster    int64
	)

	nodeMetrics := getNodeMetrics(heapsterMetricsClient)
	podMetrics := getPodMetrics(heapsterMetricsClient)
	fmt.Fprintf(w, "\n")
	for _, nodeInfo := range nodeInfos {
		totalGPU, allocatedGPU := calculateNodeGPU(nodeInfo)
		totalCPU, usedCPU := calculateNodeCPU(nodeInfo, nodeMetrics)
		totalMemory, usedMemory := calculateNodeMemory(nodeInfo, nodeMetrics)
		totalGPUsInCluster += totalGPU
		allocatedGPUsInCluster += allocatedGPU
		totalCPUsInCluster += totalCPU
		usedCPUsInCluster += usedCPU
		totalMemoryInCluster += totalMemory
		usedMemoryInCluster += usedMemory

		address := "unknown"
		if len(nodeInfo.node.Status.Addresses) > 0 {
			address = nodeInfo.node.Status.Addresses[0].Address
		}

		role := strings.Join(findNodeRoles(&nodeInfo.node), ",")
		if len(role) == 0 {
			role = "<none>"
		}

		fmt.Fprintf(w, "\n")
		fmt.Fprintf(w, "NAME:\t%s\n", nodeInfo.node.Name)
		fmt.Fprintf(w, "IPADDRESS:\t%s\n", address)
		fmt.Fprintf(w, "ROLE:\t%s\n", role)

		pods := resourcePods(nodeInfo.pods, podMetrics)
		if len(pods) > 0 {
			fmt.Fprintf(w, "\n")
			fmt.Fprintf(w, "NAMESPACE\tNAME\tGPU REQUESTS\tGPU LIMITS\tCPU USAGE\tMEM USAGE\n")
			for _, pod := range pods {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", pod.Namespace,
					pod.Name,
					strconv.FormatInt(gpuInPod(pod), 10),
					strconv.FormatInt(gpuInPod(pod), 10),
					resource.NewMilliQuantity(cpuInPod(pod, podMetrics), resource.DecimalSI).String(),
					fmt.Sprintf("%vMi", memoryInPod(pod, podMetrics)/(1024*1024)))
			}
			fmt.Fprintf(w, "\n")
		}
		var gpuUsageInNode float64 = 0
		if totalGPU > 0 {
			gpuUsageInNode = float64(allocatedGPU) / float64(totalGPU) * 100
		} else {
			fmt.Fprintf(w, "\n")
		}
		var cpuUsageInNode float64 = 0
		if totalCPU > 0 {
			cpuUsageInNode = float64(usedCPU) / float64(totalCPU) * 100
		} else {
			fmt.Fprintf(w, "\n")
		}
		var memoryUsageInNode float64 = 0
		if totalMemory > 0 {
			memoryUsageInNode = float64(usedMemory) / float64(totalMemory) * 100
		} else {
			fmt.Fprintf(w, "\n")
		}

		fmt.Fprintf(w, "Total GPUs In Node %s:\t%s \t\n", nodeInfo.node.Name, strconv.FormatInt(totalGPU, 10))
		fmt.Fprintf(w, "Allocated GPUs In Node %s:\t%s (%d%%)\t\n", nodeInfo.node.Name, strconv.FormatInt(allocatedGPU, 10), int64(gpuUsageInNode))
		log.Debugf("gpu: %s, allocated GPUs %s", strconv.FormatInt(totalGPU, 10),
			strconv.FormatInt(allocatedGPU, 10))

		fmt.Fprintf(w, "Total CPUs In Node %s:\t%s \t\n", nodeInfo.node.Name, resource.NewMilliQuantity(totalCPU, resource.DecimalSI).String())
		fmt.Fprintf(w, "Used CPUs In Node %s:\t%s (%d%%)\t\n", nodeInfo.node.Name, resource.NewMilliQuantity(usedCPU, resource.DecimalSI).String(), int64(cpuUsageInNode))
		log.Debugf("cpu: %s, allocated CPUs %s", resource.NewMilliQuantity(totalCPU, resource.DecimalSI).String(),
			resource.NewMilliQuantity(usedCPU, resource.DecimalSI).String())
		fmt.Fprintf(w, "Total Memory In Node %s:\t%s \t\n", nodeInfo.node.Name, fmt.Sprintf("%vMi", totalMemory/(1024*1024)))
		fmt.Fprintf(w, "Used Memory In Node %s:\t%s (%d%%)\t\n", nodeInfo.node.Name, fmt.Sprintf("%vMi", usedMemory/(1024*1024)), int64(memoryUsageInNode))
		log.Debugf("memory: %s, allocated Memory %s", fmt.Sprintf("%vMi", totalMemory/(1024*1024)),
			fmt.Sprintf("%vMi", usedMemory/(1024*1024)))
		fmt.Fprintf(w, "--------------------------------------------------------------------------------------------------------------------------------------\n")
	}
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "Allocated/Total GPUs In Cluster:\t")
	log.Debugf("gpu: %s, allocated GPUs %s", strconv.FormatInt(totalGPUsInCluster, 10),
		strconv.FormatInt(allocatedGPUsInCluster, 10))

	var gpuUsage float64 = 0
	if totalGPUsInCluster > 0 {
		gpuUsage = float64(allocatedGPUsInCluster) / float64(totalGPUsInCluster) * 100
	}
	fmt.Fprintf(w, "%s/%s (%d%%)\t\n",
		strconv.FormatInt(allocatedGPUsInCluster, 10),
		strconv.FormatInt(totalGPUsInCluster, 10),
		int64(gpuUsage))

	fmt.Fprintf(w, "Used/Total CPUs In Cluster:\t")
	log.Debugf("cpu: %s, used CPUs %s", resource.NewMilliQuantity(totalCPUsInCluster, resource.DecimalSI).String(),
		resource.NewMilliQuantity(usedCPUsInCluster, resource.DecimalSI).String())
	var cpuUsage float64 = 0
	if totalCPUsInCluster > 0 {
		cpuUsage = float64(usedCPUsInCluster) / float64(totalCPUsInCluster) * 100
	}
	fmt.Fprintf(w, "%s/%s (%d%%)\t\n",
		resource.NewMilliQuantity(usedCPUsInCluster, resource.DecimalSI).String(),
		resource.NewMilliQuantity(totalCPUsInCluster, resource.DecimalSI).String(),
		int64(cpuUsage))
	fmt.Fprintf(w, "Used/Total Memory In Cluster:\t")
	log.Debugf("memory: %s, used Memory %s", fmt.Sprintf("%vMi", totalMemoryInCluster/(1024*1024)),
		fmt.Sprintf("%vMi", usedMemoryInCluster/(1024*1024)))
	var memoryUsage float64 = 0
	if totalMemoryInCluster > 0 {
		memoryUsage = float64(usedMemoryInCluster) / float64(totalMemoryInCluster) * 100
	}
	fmt.Fprintf(w, "%s/%s (%d%%)\t\n",
		fmt.Sprintf("%vMi", usedMemoryInCluster/(1024*1024)),
		fmt.Sprintf("%vMi", totalMemoryInCluster/(1024*1024)),
		int64(memoryUsage))

	_ = w.Flush()
}

// calculate the GPU count of each node
func calculateNodeGPU(nodeInfo NodeInfo) (totalGPU int64, allocatedGPU int64) {
	node := nodeInfo.node
	totalGPU = gpuInNode(node)
	// allocatedGPU = gpuInPod()

	for _, pod := range nodeInfo.pods {
		allocatedGPU += gpuInPod(pod)
	}

	return totalGPU, allocatedGPU
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

func cpuInPod(pod v1.Pod, podMetrics map[string]v1.ResourceList) int64 {
	usedCPU :=  podMetrics[pod.Name][v1.ResourceCPU]
	return usedCPU.MilliValue()
}

func memoryInPod(pod v1.Pod, podMetrics map[string]v1.ResourceList) int64 {
	usedMemory :=  podMetrics[pod.Name][v1.ResourceMemory]
	return usedMemory.Value()
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

func isMasterNode(node v1.Node) bool {
	if _, ok := node.Labels[masterLabelRole]; ok {
		return true
	}

	return false
}

// findNodeRoles returns the roles of a given node.
// The roles are determined by looking for:
// * a node-role.kubernetes.io/<role>="" label
// * a kubernetes.io/role="<role>" label
func findNodeRoles(node *v1.Node) []string {
	roles := sets.NewString()
	for k, v := range node.Labels {
		switch {
		case strings.HasPrefix(k, labelNodeRolePrefix):
			if role := strings.TrimPrefix(k, labelNodeRolePrefix); len(role) > 0 {
				roles.Insert(role)
			}

		case k == nodeLabelRole && v != "":
			roles.Insert(v)
		}
	}
	return roles.List()
}
