package topnode

import (
	"encoding/json"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/apis/utils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

var GPUTopologyNodeDescription = `
  1.This node is enabled gpu topology mode.
  2.Pods can request resource 'aliyun.com/gpu' to use gpu topology feature on this node
`

var gpuTopologyTemplate = `
Name:    %v
Status:  %v
Role:    %v
Type:    %v
Address: %v
Description:
%v
%v
`

var gpuTopologySummary = `
GPU Summary:
  Total GPUs:           %v
  Allocated GPUs:       %v
  Unhealthy GPUs:       %v
  Total GPU Memory:     %.1f GiB
  Allocated GPU Memory: %.1f GiB
  Used GPU Memory:      %.1f GiB
`

type gputopo struct {
	node       *v1.Node
	pods       []*v1.Pod
	configmap  *v1.ConfigMap
	gpuMetrics types.NodeGpuMetric
	baseNode
}

func NewGPUTopologyNode(client *kubernetes.Clientset, node *v1.Node, index int, args buildNodeArgs) (Node, error) {
	pods := getNodePods(node, args.pods)
	var configmap *v1.ConfigMap
	for _, c := range args.configmaps {
		if val, ok := c.Labels["nodename"]; ok && val == node.Name {
			configmap = c.DeepCopy()
			break
		}
	}
	if configmap == nil {
		return nil, fmt.Errorf("not found configmap for building gpu topology node")
	}
	return &gputopo{
		node:       node,
		pods:       pods,
		configmap:  configmap,
		gpuMetrics: getGPUMetricsByNodeName(node.Name, args.nodeGPUMetrics),
		baseNode: baseNode{
			index:    index,
			node:     node,
			pods:     pods,
			nodeType: types.GPUTopologyNode,
		},
	}, nil
}

func (g *gputopo) gpuMetricsIsEnabled() bool {
	return len(g.gpuMetrics) != 0
}

func (g *gputopo) getTotalGPUs() int {
	if len(g.gpuMetrics) != 0 {
		return len(g.gpuMetrics)
	}
	val, ok := g.node.Status.Capacity[v1.ResourceName(types.AliyunGPUResourceName)]
	if !ok {
		return 0
	}
	return int(val.Value())
}

func (g *gputopo) getAllocatedGPUs() int {
	allocatedGPUs := 0
	for _, pod := range g.pods {
		if utils.IsCompletedPod(pod) {
			continue
		}
		allocation := utils.AliyunGPUCountInPod(pod)
		allocatedGPUs += allocation
	}
	return allocatedGPUs
}

func (g *gputopo) getTotalGPUMemory() float64 {
	totalGPUMemory := float64(0)
	for _, metric := range g.gpuMetrics {
		totalGPUMemory += metric.GpuMemoryTotal
	}
	// if gpu metric is enable,return the value given by prometheus
	if totalGPUMemory != 0 {
		return totalGPUMemory
	}
	return float64(0)
}

func (g *gputopo) getAllocatedGPUMemory() float64 {
	if !g.gpuMetricsIsEnabled() {
		return float64(0)
	}
	allocatedGPUMemory := float64(0)
	allocatedGPUs := map[string]bool{}
	for _, pod := range g.pods {
		if utils.IsCompletedPod(pod) {
			continue
		}
		allocation := utils.GetPodGPUTopologyAllocation(pod)
		for _, gpuId := range allocation {
			allocatedGPUs[gpuId] = true
		}
	}
	for key, metric := range g.gpuMetrics {
		if allocatedGPUs[key] {
			allocatedGPUMemory += metric.GpuMemoryTotal
		}
	}
	return utils.DataUnitTransfer("GiB", "bytes", allocatedGPUMemory)
}

func (g *gputopo) getUsedGPUMemory() float64 {
	if !g.gpuMetricsIsEnabled() {
		return float64(0)
	}
	usedGPUMemory := float64(0)
	for _, gpuMetric := range g.gpuMetrics {
		usedGPUMemory += gpuMetric.GpuMemoryUsed
	}
	return usedGPUMemory
}

func (g *gputopo) getDutyCycle() float64 {
	if !g.gpuMetricsIsEnabled() {
		return float64(0)
	}
	dutyCycle := float64(0)
	totalGPUs := float64(0)
	for _, gpuMetric := range g.gpuMetrics {
		totalGPUs += float64(1)
		dutyCycle += gpuMetric.GpuDutyCycle
	}
	if totalGPUs == 0 {
		return float64(0)
	}
	return dutyCycle / totalGPUs
}

func (g *gputopo) getUnhealthyGPUs() int {
	totalGPUs := g.getTotalGPUs()
	allocatableGPUs, ok := g.node.Status.Allocatable[v1.ResourceName(types.AliyunGPUResourceName)]
	if !ok {
		return 0
	}
	if totalGPUs <= 0 {
		return 0
	}
	return totalGPUs - int(allocatableGPUs.Value())
}

func (g *gputopo) getTotalGPUMemoryOfDevice(id string) float64 {
	if metric, ok := g.gpuMetrics[id]; ok {
		return metric.GpuMemoryTotal
	}
	return 0
}

func (g *gputopo) convert2NodeInfo() types.GPUTopologyNodeInfo {
	podInfos := []types.GPUTopologyPodInfo{}
	// 1.initilize the common node information
	gpuTopologyNodeInfo := types.GPUTopologyNodeInfo{
		CommonNodeInfo: types.CommonNodeInfo{
			Name:        g.Name(),
			IP:          g.IP(),
			Status:      g.Status(),
			Type:        types.GPUTopologyNode,
			Description: GPUTopologyNodeDescription,
		},
		CommonGPUNodeInfo: types.CommonGPUNodeInfo{
			TotalGPUs:     g.getTotalGPUs(),
			AllocatedGPUs: g.getAllocatedGPUs(),
			UnhealthyGPUs: g.getUnhealthyGPUs(),
		},
	}
	// 2.parse devices from field "devices" of configmap
	deviceMap := map[string]types.GPUTopologyNodeDevice{}
	if val, ok := g.configmap.Data["devices"]; ok {
		devicesFromConfigmap := map[string]string{}
		json.Unmarshal([]byte(val), &devicesFromConfigmap)
		for id, health := range devicesFromConfigmap {
			healthy := false
			if health == "Healthy" {
				healthy = true
			}
			deviceMap[id] = types.GPUTopologyNodeDevice{
				Id:      id,
				Healthy: healthy,
				Status:  "idle",
			}
		}
	}
	// 3.build pod informations
	for _, pod := range g.pods {
		if utils.IsCompletedPod(pod) {
			continue
		}
		gpuCount := utils.AliyunGPUCountInPod(pod)
		if gpuCount == 0 {
			continue
		}
		allocation := utils.GetPodGPUTopologyAllocation(pod)
		visibleGPUs := utils.GetPodGPUTopologyVisibleGPUs(pod)
		for _, gpuId := range allocation {
			devInfo, ok := deviceMap[gpuId]
			if !ok {
				continue
			}
			devInfo.Status = "using"
			deviceMap[gpuId] = devInfo
		}
		status, _, _, _ := utils.DefinePodPhaseStatus(*pod)
		podInfos = append(podInfos, types.GPUTopologyPodInfo{
			Name:        pod.Name,
			Namespace:   pod.Namespace,
			Status:      status,
			RequestGPU:  gpuCount,
			Allocation:  allocation,
			VisibleGPUs: visibleGPUs,
		})
	}
	// 4.parse gpu topology from field "bandwith" and field "linkType" of configmap
	topology := types.GPUTopology{
		LinkMatrix:      [][]string{},
		BandwidthMatrix: [][]float32{},
	}
	if val, ok := g.configmap.Data["linkType"]; ok {
		json.Unmarshal([]byte(val), &topology.LinkMatrix)
	}
	if val, ok := g.configmap.Data["bandwith"]; ok {
		json.Unmarshal([]byte(val), &topology.BandwidthMatrix)
	}
	gpuTopologyNodeInfo.PodInfos = podInfos
	gpuTopologyNodeInfo.GPUTopology = topology
	devices := []types.GPUTopologyNodeDevice{}
	for _, dev := range deviceMap {
		devices = append(devices, dev)
	}
	metrics := []*types.AdvancedGpuMetric{}
	for _, metric := range g.gpuMetrics {
		metrics = append(metrics, metric)
	}
	gpuTopologyNodeInfo.GPUMetrics = metrics
	gpuTopologyNodeInfo.Devices = devices
	return gpuTopologyNodeInfo
}

func (g *gputopo) Convert2NodeInfo() interface{} {
	return g.convert2NodeInfo()
}

func (g *gputopo) AllDevicesAreHealthy() bool {
	return g.getUnhealthyGPUs() == 0
}

func (g *gputopo) WideFormat() string {
	role := strings.Join(g.Role(), ",")
	if role == "" {
		role = "<none>"
	}
	nodeInfo := g.convert2NodeInfo()
	lines := []string{}
	lines = g.displayPodInfos(lines, nodeInfo)
	if len(nodeInfo.GPUTopology.LinkMatrix) != 0 {
		header := []string{"  "}
		lines = append(lines, "LinkTypeMatrix:")
		for index, _ := range nodeInfo.Devices {
			header = append(header, fmt.Sprintf("GPU%v", index))
		}
		lines = append(lines, strings.Join(header, "\t"))
		for row, links := range nodeInfo.GPUTopology.LinkMatrix {
			linkLine := []string{fmt.Sprintf("  GPU%v", row)}
			for _, link := range links {
				linkLine = append(linkLine, link)
			}
			lines = append(lines, strings.Join(linkLine, "\t"))
		}
	}
	if len(nodeInfo.GPUTopology.BandwidthMatrix) != 0 {
		header := []string{"  "}
		lines = append(lines, "BandwidthMatrix:")
		for index, _ := range nodeInfo.Devices {
			header = append(header, fmt.Sprintf("GPU%v", index))
		}
		lines = append(lines, strings.Join(header, "\t"))
		for row, bandwidths := range nodeInfo.GPUTopology.BandwidthMatrix {
			bandwidthLine := []string{fmt.Sprintf("  GPU%v", row)}
			for _, bandwidth := range bandwidths {
				bandwidthLine = append(bandwidthLine, fmt.Sprintf("%v", bandwidth))
			}
			lines = append(lines, strings.Join(bandwidthLine, "\t"))
		}
	}
	lines = g.displayDeviceInfos(lines, nodeInfo)
	return fmt.Sprintf(gpuTopologyTemplate,
		nodeInfo.Name,
		nodeInfo.Status,
		role,
		nodeInfo.Type,
		nodeInfo.IP,
		strings.Trim(nodeInfo.Description, "\n"),
		strings.Join(lines, "\n"),
	)
}

func (g *gputopo) displayPodInfos(lines []string, nodeInfo types.GPUTopologyNodeInfo) []string {
	podLines := []string{"Instances:", "  NAMESPACE\tNAME\tGPU(Requested)\tGPU(Allocated)"}
	podLines = append(podLines, "  ---------\t----\t--------------\t--------------")
	for _, podInfo := range nodeInfo.PodInfos {
		if len(podInfo.Allocation) == 0 {
			continue
		}
		podLines = append(podLines, fmt.Sprintf("  %v\t%v\t%v\t%v",
			podInfo.Namespace,
			podInfo.Name,
			podInfo.RequestGPU,
			strings.Join(podInfo.Allocation, ","),
		))
	}
	if len(podLines) == 3 {
		podLines = []string{}
	}
	lines = append(lines, podLines...)
	return lines
}

func (g *gputopo) displayDeviceInfos(lines []string, nodeInfo types.GPUTopologyNodeInfo) []string {
	if !g.gpuMetricsIsEnabled() {
		return g.displayDeviceInfoUnderNoMetrics(lines, nodeInfo)
	}
	return g.displayDeviceInfoUnderMetrics(lines, nodeInfo)
}

func (g *gputopo) displayDeviceInfoUnderNoMetrics(lines []string, nodeInfo types.GPUTopologyNodeInfo) []string {
	deviceLines := []string{"GPUs:", "  INDEX\tSTATUS\tHEALTHY"}
	deviceLines = append(deviceLines, "  ----\t------\t-------")
	unhealthyGPUs := 0
	allocatedGPUs := 0
	for _, dev := range nodeInfo.Devices {
		if !dev.Healthy {
			unhealthyGPUs++
		}
		if dev.Status == "using" {
			allocatedGPUs++
		}
		deviceLines = append(deviceLines, fmt.Sprintf("  GPU%v\t%v\t%v", dev.Id, dev.Status, dev.Healthy))
	}
	if len(deviceLines) == 3 {
		deviceLines = []string{}
	}
	deviceLines = append(deviceLines, "GPU Summary:")
	deviceLines = append(deviceLines, fmt.Sprintf("  Total GPUs: %v", nodeInfo.TotalGPUs))
	deviceLines = append(deviceLines, fmt.Sprintf("  Allocated GPUs: %v", nodeInfo.AllocatedGPUs))
	deviceLines = append(deviceLines, fmt.Sprintf("  Unhealthy GPUs: %v", g.getUnhealthyGPUs()))
	lines = append(lines, deviceLines...)
	return lines
}

func (g *gputopo) displayDeviceInfoUnderMetrics(lines []string, nodeInfo types.GPUTopologyNodeInfo) []string {
	deviceLines := []string{"GPUs:", "  INDEX\tMEMORY(Total)\tMEMORY(Allocated)\tMEMORY(Used)\tDUTY_CYCLE"}
	deviceLines = append(deviceLines, "  -----\t-------------\t-----------------\t------------\t----------")
	deviceMap := map[string]*types.AdvancedGpuMetric{}
	for _, dev := range g.gpuMetrics {
		deviceMap[dev.Id] = dev
	}
	totalGPUMemory := float64(0)
	totalAllocatedGPUMemory := float64(0)
	totalUsedGPUMemory := float64(0)
	for i := 0; i < nodeInfo.TotalGPUs; i++ {
		gpuId := fmt.Sprintf("%v", i)
		devInfo, ok := deviceMap[gpuId]
		if !ok {
			continue
		}
		totalGPUMemory += devInfo.GpuMemoryTotal
		allocatedGPUMemory := float64(0)
		for _, dev := range nodeInfo.Devices {
			if dev.Id == gpuId && dev.Status == "using" {
				allocatedGPUMemory = devInfo.GpuMemoryTotal
				totalAllocatedGPUMemory += allocatedGPUMemory
			}
		}
		usedGPUMemory := float64(0)
		gpuDutyCycle := float64(0)
		// we do not display the use gpu memory and gpu dutycycle when allocated gpu memory is 0
		if allocatedGPUMemory != 0 {
			totalUsedGPUMemory += devInfo.GpuMemoryUsed
			usedGPUMemory = devInfo.GpuMemoryUsed
			gpuDutyCycle = devInfo.GpuDutyCycle
		}
		deviceLines = append(deviceLines, fmt.Sprintf("  %v\t%.1f GiB\t%.1f GiB\t%.1f GiB\t%.1f%%",
			devInfo.Id,
			utils.DataUnitTransfer("bytes", "GiB", devInfo.GpuMemoryTotal),
			utils.DataUnitTransfer("bytes", "GiB", allocatedGPUMemory),
			utils.DataUnitTransfer("bytes", "GiB", usedGPUMemory),
			gpuDutyCycle),
		)
	}
	if len(deviceLines) == 3 {
		deviceLines = []string{}
	}
	deviceLines = append(deviceLines,
		fmt.Sprintf(strings.Trim(gpuTopologySummary, "\n"),
			nodeInfo.TotalGPUs,
			nodeInfo.AllocatedGPUs,
			g.getUnhealthyGPUs(),
			utils.DataUnitTransfer("bytes", "GiB", totalGPUMemory),
			utils.DataUnitTransfer("bytes", "GiB", totalAllocatedGPUMemory),
			utils.DataUnitTransfer("bytes", "GiB", totalUsedGPUMemory),
		))
	return lines
}

func IsGPUTopologyNode(node *v1.Node) bool {
	labels := strings.Split(types.GPUTopologyNodeLabels, ",")
	for _, label := range labels {
		topologyKey := strings.Split(label, "=")[0]
		topologyVal := strings.Split(label, "=")[1]
		if val, ok := node.Labels[topologyKey]; ok && val == topologyVal {
			return true
		}
	}
	return false
}

/*
format is like:

===================================== GPUTopologyNode ===================================

Name:          cn-shanghai.192.168.7.186
Status:        Ready
Role:          <none>
Type:          GPUTopology
Address:       192.168.7.186
TotalGPUs:     4
UsedGPUs:      0
UnhealthyGPUs: 0

Instances:
  NAMESPACE  NAME                               GPU(Requested)  GPU(Allocated)
  ---------  ----                               --------------  --------------
  default    nginx-deployment-6687789574-6mn2z  2
  default    nginx-deployment-6687789574-wd7vf  2

GPUs:
  INDEX  STATUS  HEALTHY
  ----   ------  -------
  GPU0   idle    true
  GPU1   idle    true
  GPU2   idle    true
  GPU3   idle    true

  Total: 4,Allocated: 0/4,Unhealthy: 0/4

LinkTypeMatrix:
        GPU0  GPU1  GPU2  GPU3
  GPU0  N-A   NV2   NV1   NV2
  GPU1  NV2   N-A   NV2   NV1
  GPU2  NV1   NV2   N-A   NV1
  GPU3  NV2   NV1   NV1   N-A

BandwidthMatrix:
        GPU0    GPU1    GPU2    GPU3
  GPU0  738.42  96.44   48.37   96.19
  GPU1  96.26   744.05  96.23   48.36
  GPU2  48.37   96.24   744.76  48.37
  GPU3  96.23   48.37   48.37   744.76
-----------------------------------------------------------------------------------------
Allocated/Total GPUs In Cluster: 4/4(100.0%)

===================================== End ===============================================
*/
func displayGPUTopologyNodeDetails(w *tabwriter.Writer, nodes []Node) {
	if len(nodes) == 0 {
		return
	}
	totalGPUs := 0
	totalUnhealthyGPUs := 0
	totalAllocatedGPUs := 0
	for _, node := range nodes {
		nodeInfo := node.Convert2NodeInfo().(types.GPUTopologyNodeInfo)
		totalGPUs += nodeInfo.TotalGPUs
		totalAllocatedGPUs += nodeInfo.AllocatedGPUs
		totalUnhealthyGPUs += nodeInfo.UnhealthyGPUs
		PrintLine(w, node.WideFormat())
	}
}

func displayGPUTopologyNodeSummary(w *tabwriter.Writer, nodes []Node, isUnhealthy, showNodeType bool) (int, int, int) {
	totalGPUs := 0
	allocatedGPUs := 0
	unhealthyGPUs := 0
	for _, node := range nodes {
		nodeInfo := node.Convert2NodeInfo().(types.GPUTopologyNodeInfo)
		totalGPUs += nodeInfo.TotalGPUs
		allocatedGPUs += nodeInfo.AllocatedGPUs
		unhealthyGPUs += nodeInfo.UnhealthyGPUs
		items := []string{}
		items = append(items, node.Name())
		items = append(items, node.IP())
		role := nodeInfo.Role
		if role == "" {
			role = "<none>"
		}
		items = append(items, role)
		items = append(items, node.Status())
		items = append(items, fmt.Sprintf("%v", nodeInfo.TotalGPUs))
		items = append(items, fmt.Sprintf("%v", nodeInfo.AllocatedGPUs))
		if showNodeType {
			for _, typeInfo := range types.NodeTypeSlice {
				if typeInfo.Name == types.GPUTopologyNode {
					items = append(items, typeInfo.Alias)
				}
			}
		}
		if isUnhealthy {
			items = append(items, fmt.Sprintf("%v", nodeInfo.UnhealthyGPUs))
		}
		PrintLine(w, items...)
	}
	return totalGPUs, allocatedGPUs, unhealthyGPUs
}

func displayGPUTopologyNodesCustomSummary(w *tabwriter.Writer, nodes []Node) {
	if len(nodes) == 0 {
		return
	}
	header := []string{"NAME", "IPADDRESS", "ROLE", "STATUS", "GPU(Total)", "GPU(Allocated)"}
	isUnhealthy := false
	for _, node := range nodes {
		if !node.AllDevicesAreHealthy() {
			isUnhealthy = true
		}
	}
	if isUnhealthy {
		header = append(header, "UNHEALTHY")
	}
	PrintLine(w, header...)
	totalGPUs := 0
	allocatedGPUs := 0
	unhealthyGPUs := 0
	for _, node := range nodes {
		nodeInfo := node.Convert2NodeInfo().(types.GPUTopologyNodeInfo)
		totalGPUs += nodeInfo.TotalGPUs
		allocatedGPUs += nodeInfo.AllocatedGPUs
		unhealthyGPUs += nodeInfo.UnhealthyGPUs
		items := []string{}
		items = append(items, node.Name())
		items = append(items, node.IP())
		role := nodeInfo.Role
		if role == "" {
			role = "<none>"
		}
		items = append(items, role)
		items = append(items, node.Status())
		items = append(items, fmt.Sprintf("%v", nodeInfo.TotalGPUs))
		items = append(items, fmt.Sprintf("%v", nodeInfo.AllocatedGPUs))
		if isUnhealthy {
			items = append(items, fmt.Sprintf("%v", nodeInfo.UnhealthyGPUs))
		}
		PrintLine(w, items...)
	}
	PrintLine(w, "---------------------------------------------------------------------------------------------------")
	PrintLine(w, "Allocated/Total GPUs of nodes which own resource aliyun.com/gpu In Cluster:")
	allocatedPercent := float64(0)
	if totalGPUs != 0 {
		allocatedPercent = float64(allocatedGPUs) / float64(totalGPUs) * 100
	}
	unhealthyPercent := float64(0)
	if totalGPUs != 0 {
		unhealthyPercent = float64(unhealthyGPUs) / float64(totalGPUs) * 100
	}
	PrintLine(w, fmt.Sprintf("%v/%v (%.1f%%)", allocatedGPUs, totalGPUs, allocatedPercent))
	if unhealthyGPUs != 0 {
		PrintLine(w, "Unhealthy/Total GPUs of nodes which own resource aliyun.com/gpu In Cluster:")
		PrintLine(w, fmt.Sprintf("%v/%v (%.1f%%)", unhealthyGPUs, totalGPUs, unhealthyPercent))
	}
}

func NewGPUTopologyNodeProcesser() NodeProcesser {
	return &nodeProcesser{
		nodeType:                  types.GPUTopologyNode,
		key:                       "gpuTopologyNodes",
		builder:                   NewGPUTopologyNode,
		canBuildNode:              IsGPUTopologyNode,
		displayNodesDetails:       displayGPUTopologyNodeDetails,
		displayNodesSummary:       displayGPUTopologyNodeSummary,
		displayNodesCustomSummary: displayGPUTopologyNodesCustomSummary,
	}
}
