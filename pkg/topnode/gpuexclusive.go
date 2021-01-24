package topnode

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/apis/utils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

var GPUExclusiveNodeDescription = `
  1.This node is enabled gpu exclusive mode.
  2.Pods can request resource 'nvidia.com/gpu' to use gpu exclusive feature on this node
`

var gpuExclusiveTemplate = `
Name:    %v
Status:  %v
Role:    %v
Type:    %v
Address: %v
Description:
%v
%v
`

var gpuExclusiveSummary = `
GPU Summary:
  Total GPUs:           %v
  Allocated GPUs:       %v
  Unhealthy GPUs:       %v
  Total GPU Memory:     %.1f GiB
  Allocated GPU Memory: %.1f GiB
  Used GPU Memory:      %.1f GiB
`

type gpuexclusive struct {
	node       *v1.Node
	pods       []*v1.Pod
	gpuMetrics types.NodeGpuMetric
	baseNode
}

func NewGPUExclusiveNode(client *kubernetes.Clientset, node *v1.Node, index int, args buildNodeArgs) (Node, error) {
	pods, err := listNodePods(client, node.Name)
	if err != nil {
		return nil, err
	}
	return &gpuexclusive{
		node:       node,
		pods:       pods,
		gpuMetrics: getGPUMetricsByNodeName(node.Name, args.nodeGPUMetrics),
		baseNode: baseNode{
			index:    index,
			node:     node,
			pods:     pods,
			nodeType: types.GPUExclusiveNode,
		},
	}, nil
}

func (g *gpuexclusive) gpuMetricsIsEnabled() bool {
	return len(g.gpuMetrics) != 0
}

func (g *gpuexclusive) getTotalGPUs() int {
	if len(g.gpuMetrics) != 0 {
		return len(g.gpuMetrics)
	}
	val, ok := g.node.Status.Capacity[v1.ResourceName(types.NvidiaGPUResourceName)]
	if !ok {
		return 0
	}
	return int(val.Value())
}

func (g *gpuexclusive) getAllocatedGPUs() int {
	allocatedGPUs := 0
	for _, pod := range g.pods {
		if utils.IsCompletedPod(pod) {
			continue
		}
		allocation := utils.GPUCountInPod(pod)
		allocatedGPUs += allocation
	}
	return allocatedGPUs
}

func (g *gpuexclusive) getTotalGPUMemory() float64 {
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

func (g *gpuexclusive) getAllocatedGPUMemory() float64 {
	if !g.gpuMetricsIsEnabled() {
		return float64(0)
	}
	allocatedGPUMemory := float64(0)
	allocatedGPUs := 0
	for _, pod := range g.pods {
		if utils.IsCompletedPod(pod) {
			continue
		}
		allocation := utils.GPUCountInPod(pod)
		allocatedGPUs += allocation
	}
	for _, metric := range g.gpuMetrics {
		allocatedGPUMemory = float64(allocatedGPUs) * metric.GpuMemoryTotal
	}
	return allocatedGPUMemory
}

func (g *gpuexclusive) getUsedGPUMemory() float64 {
	// can not to detect gpu memory if no gpu metrics data
	if !g.gpuMetricsIsEnabled() {
		return float64(0)
	}
	usedGPUMemory := float64(0)
	for _, gpuMetric := range g.gpuMetrics {
		usedGPUMemory += gpuMetric.GpuMemoryUsed
	}
	return usedGPUMemory
}

func (g *gpuexclusive) getDutyCycle() float64 {
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

func (g *gpuexclusive) getUnhealthyGPUs() int {
	totalGPUs := g.getTotalGPUs()
	allocatableGPUs, ok := g.node.Status.Allocatable[v1.ResourceName(types.NvidiaGPUResourceName)]
	if !ok {
		return 0
	}
	if totalGPUs <= 0 {
		return 0
	}
	return totalGPUs - int(allocatableGPUs.Value())
}

func (g *gpuexclusive) getTotalGPUMemoryOfDevice(id string) float64 {
	if metric, ok := g.gpuMetrics[id]; ok {
		return metric.GpuMemoryTotal
	}
	return 0
}

func (g *gpuexclusive) convert2NodeInfo() types.GPUExclusiveNodeInfo {
	podInfos := []types.GPUExclusivePodInfo{}
	metrics := []*types.AdvancedGpuMetric{}
	for _, metric := range g.gpuMetrics {
		metrics = append(metrics, metric)
	}
	gpuExclusiveInfo := types.GPUExclusiveNodeInfo{
		CommonNodeInfo: types.CommonNodeInfo{
			Name:        g.Name(),
			IP:          g.IP(),
			Status:      g.Status(),
			Type:        types.GPUExclusiveNode,
			Description: GPUExclusiveNodeDescription,
		},
		CommonGPUNodeInfo: types.CommonGPUNodeInfo{
			TotalGPUs:     g.getTotalGPUs(),
			AllocatedGPUs: g.getAllocatedGPUs(),
			UnhealthyGPUs: g.getUnhealthyGPUs(),
			GPUMetrics:    metrics,
		},
	}
	for _, pod := range g.pods {
		if utils.IsCompletedPod(pod) {
			continue
		}
		gpuCount := utils.GPUCountInPod(pod)
		if gpuCount == 0 {
			continue
		}
		status, _, _, _ := utils.DefinePodPhaseStatus(*pod)
		podInfos = append(podInfos, types.GPUExclusivePodInfo{
			Name:       pod.Name,
			Namespace:  pod.Namespace,
			Status:     status,
			RequestGPU: gpuCount,
		})
	}
	gpuExclusiveInfo.PodInfos = podInfos
	gpuExclusiveInfo.GPUMetrics = metrics
	return gpuExclusiveInfo
}

func (g *gpuexclusive) AllDevicesAreHealthy() bool {
	return g.getUnhealthyGPUs() == 0
}

func (g *gpuexclusive) Convert2NodeInfo() interface{} {
	return g.convert2NodeInfo()
}

func (g *gpuexclusive) WideFormat() string {
	role := strings.Join(g.Role(), ",")
	if role == "" {
		role = "<none>"
	}
	nodeInfo := g.convert2NodeInfo()
	lines := []string{}
	lines = g.displayPodInfos(lines, nodeInfo)
	lines = g.displayDeviceInfos(lines, nodeInfo)
	return fmt.Sprintf(gpuExclusiveTemplate,
		nodeInfo.Name,
		nodeInfo.Status,
		role,
		nodeInfo.Type,
		nodeInfo.IP,
		strings.Trim(nodeInfo.Description, "\n"),
		strings.Join(lines, "\n"),
	)
}

func (g *gpuexclusive) displayPodInfos(lines []string, nodeInfo types.GPUExclusiveNodeInfo) []string {
	title := []string{"  NAMESPACE", "NAME", "STATUS", "GPU(Requested)"}
	splitLine := []string{"  ---------", "----", "------", "--------------"}
	if g.gpuMetricsIsEnabled() {
		title = append(title, "GPU(Allocated)")
		splitLine = append(splitLine, "--------------")
	}
	podLines := []string{"Instances:", strings.Join(title, "\t")}
	podLines = append(podLines, strings.Join(splitLine, "\t"))
	for _, podInfo := range nodeInfo.PodInfos {
		if g.gpuMetricsIsEnabled() {
			gpus := []string{}
			for _, dev := range g.gpuMetrics {
				for _, podName := range dev.PodNames {
					if podName == fmt.Sprintf("%v/%v", podInfo.Namespace, podInfo.Name) {
						gpus = append(gpus, fmt.Sprintf("gpu%v", dev.Id))
						break
					}
				}
			}
			allocatedGPUs := strings.Join(gpus, ",")
			if allocatedGPUs == "" {
				allocatedGPUs = "N/A"
			}
			podLines = append(podLines, fmt.Sprintf("  %v\t%v\t%v\t%v\t%v", podInfo.Namespace, podInfo.Name, podInfo.Status, podInfo.RequestGPU, allocatedGPUs))
			continue
		}
		podLines = append(podLines, fmt.Sprintf("  %v\t%v\t%v\t%v", podInfo.Namespace, podInfo.Name, podInfo.Status, podInfo.RequestGPU))
	}
	if len(podLines) == 3 {
		podLines = []string{}
	}
	lines = append(lines, podLines...)
	return lines
}

func (g *gpuexclusive) displayDeviceInfos(lines []string, nodeInfo types.GPUExclusiveNodeInfo) []string {
	if !g.gpuMetricsIsEnabled() {
		return g.displayDeviceUnderNoGPUMetric(lines, nodeInfo)
	}
	return g.displayDeviceInfoUnderMetrics(lines, nodeInfo)
}

func (g *gpuexclusive) displayDeviceUnderNoGPUMetric(lines []string, nodeInfo types.GPUExclusiveNodeInfo) []string {
	deviceLines := []string{"GPU Summary:"}
	deviceLines = append(deviceLines, fmt.Sprintf("  Total GPUs:     %v", nodeInfo.TotalGPUs))
	deviceLines = append(deviceLines, fmt.Sprintf("  Allocated GPUs: %v", nodeInfo.AllocatedGPUs))
	deviceLines = append(deviceLines, fmt.Sprintf("  Unhealthy GPUs: %v", nodeInfo.UnhealthyGPUs))
	lines = append(lines, deviceLines...)
	return lines
}

func (g *gpuexclusive) displayDeviceInfoUnderMetrics(lines []string, nodeInfo types.GPUExclusiveNodeInfo) []string {
	deviceLines := []string{"GPUs:", "  INDEX\tMEMORY(Total)\tMEMORY(Allocated)\tMEMORY(Used)\tDUTY_CYCLE"}
	deviceLines = append(deviceLines, "  -----\t-------------\t-----------------\t------------\t----------")
	deviceMap := map[string]*types.AdvancedGpuMetric{}
	for _, dev := range g.gpuMetrics {
		deviceMap[dev.Id] = dev
	}
	podMap := map[string]bool{}
	for _, pod := range g.pods {
		if pod.Status.Phase != v1.PodRunning {
			continue
		}
		if utils.GPUCountInPod(pod) <= 0 {
			continue
		}
		podMap[fmt.Sprintf("%v/%v", pod.Namespace, pod.Name)] = true
	}
	totalGPUMemory := float64(0)
	totalAllocatedGPUMemory := g.getAllocatedGPUMemory()
	totalUsedGPUMemory := float64(0)
	for i := 0; i < nodeInfo.TotalGPUs; i++ {
		gpuId := fmt.Sprintf("%v", i)
		devInfo, ok := deviceMap[gpuId]
		if !ok {
			continue
		}
		totalGPUMemory += devInfo.GpuMemoryTotal
		allocatedGPUMemory := float64(0)
		names := []string{}
		for _, name := range deviceMap[gpuId].PodNames {
			if _, ok := podMap[name]; ok {
				names = append(names, name)
			}
		}
		if len(names) != 0 {
			allocatedGPUMemory = devInfo.GpuMemoryTotal
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
		deviceLines = []string{"GPUs:"}
	}
	deviceLines = append(deviceLines,
		fmt.Sprintf(strings.Trim(gpuExclusiveSummary, "\n"),
			nodeInfo.TotalGPUs,
			nodeInfo.AllocatedGPUs,
			g.getUnhealthyGPUs(),
			utils.DataUnitTransfer("bytes", "GiB", totalGPUMemory),
			utils.DataUnitTransfer("bytes", "GiB", totalAllocatedGPUMemory),
			utils.DataUnitTransfer("bytes", "GiB", totalUsedGPUMemory),
		))
	lines = append(lines, deviceLines...)
	return lines
}

func IsGPUExclusiveNode(node *v1.Node) bool {
	val, ok := node.Status.Allocatable[v1.ResourceName(types.NvidiaGPUResourceName)]
	if !ok {
		return false
	}
	return int(val.Value()) > 0
}

/*
format is like:

===================================== GPUExclusiveNode ==================================

Name:          cn-shanghai.192.168.7.182
Status:        Ready
Role:          <none>
Type:          GPUExclusive
Address:       192.168.7.182
TotalGPUs:     1
UsedGPUs:      1/1(100.0%)
UnhealthyGPUs: 0/1

Instances:
  NAMESPACE  NAME                          GPU(Requested)
  ---------  ----                          --------------
  default    tf-standalone-test-1-chief-0  1
-----------------------------------------------------------------------------------------
Allocated/Total GPUs In Cluster: 1/1(100.0%)

===================================== End ===============================================
*/
func displayGPUExclusiveNodeDetails(w *tabwriter.Writer, nodes []Node) {
	if len(nodes) == 0 {
		return
	}
	totalGPUs := 0
	totalUnhealthyGPUs := 0
	totalAllocatedGPUs := 0
	for _, node := range nodes {
		nodeInfo := node.Convert2NodeInfo().(types.GPUExclusiveNodeInfo)
		totalGPUs += nodeInfo.TotalGPUs
		totalAllocatedGPUs += nodeInfo.AllocatedGPUs
		totalUnhealthyGPUs += nodeInfo.UnhealthyGPUs
		PrintLine(w, node.WideFormat())
	}
}

func displayGPUExclusiveNodeSummary(w *tabwriter.Writer, nodes []Node, isUnhealthy, showNodeType bool) (int, int, int) {
	totalGPUs := 0
	allocatedGPUs := 0
	unhealthyGPUs := 0
	for _, node := range nodes {
		nodeInfo := node.Convert2NodeInfo().(types.GPUExclusiveNodeInfo)
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
				if typeInfo.Name == types.GPUExclusiveNode {
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

func displayGPUExclusiveNodesCustomSummary(w *tabwriter.Writer, nodes []Node) {
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
		nodeInfo := node.Convert2NodeInfo().(types.GPUExclusiveNodeInfo)
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
	PrintLine(w, "Allocated/Total GPUs of nodes which own resource nvidia.com/gpu In Cluster:")
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
		PrintLine(w, "Unhealthy/Total GPUs of nodes which own resource nvidia.com/gpu In Cluster:")
		PrintLine(w, fmt.Sprintf("%v/%v (%.1f%%)", unhealthyGPUs, totalGPUs, unhealthyPercent))
	}
}

func NewGPUExclusiveNodeProcesser() NodeProcesser {
	return &nodeProcesser{
		nodeType:                  types.GPUExclusiveNode,
		key:                       "gpuExclusiveNodes",
		builder:                   NewGPUExclusiveNode,
		canBuildNode:              IsGPUExclusiveNode,
		displayNodesDetails:       displayGPUExclusiveNodeDetails,
		displayNodesSummary:       displayGPUExclusiveNodeSummary,
		displayNodesCustomSummary: displayGPUExclusiveNodesCustomSummary,
	}
}
