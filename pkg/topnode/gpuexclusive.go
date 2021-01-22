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
Name:          %v
Status:        %v
Role:          %v
Type:          %v
Address:       %v
Description:
%v
%v

-----------------------------------------------------------------------------------------
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
	return utils.DataUnitTransfer("GiB", "bytes", allocatedGPUMemory)
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
		podInfos = append(podInfos, types.GPUExclusivePodInfo{
			Name:       pod.Name,
			Namespace:  pod.Namespace,
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
	return fmt.Sprintf(strings.TrimRight(gpuExclusiveTemplate, "\n"),
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
	podLines := []string{"", "Instances:", "  NAMESPACE\tNAME\tGPU(Requested)"}
	podLines = append(podLines, "  ---------\t----\t--------------")
	for _, podInfo := range nodeInfo.PodInfos {
		podLines = append(podLines, fmt.Sprintf("  %v\t%v\t%v", podInfo.Namespace, podInfo.Name, podInfo.RequestGPU))
	}
	if len(podLines) == 4 {
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
	return lines
}

func (g *gpuexclusive) displayDeviceInfoUnderMetrics(lines []string, nodeInfo types.GPUExclusiveNodeInfo) []string {
	deviceLines := []string{"", "GPUs:", "  INDEX\tMEMORY(Total)\tMEMORY(Allocated)\tMEMORY(Used)\tDUTY_CYCLE"}
	deviceLines = append(deviceLines, "  -----\t-------------\t-----------------\t------------\t----------")
	deviceMap := map[string]*types.AdvancedGpuMetric{}
	for _, dev := range g.gpuMetrics {
		deviceMap[dev.Id] = dev
	}
	totalGPUMemory := float64(0)
	totalAllocatedGPUMemory := float64(0)
	totalUsedGPUMemory := float64(0)
	podMap := map[string]bool{}
	for _, pod := range g.pods {
		if utils.GPUCountInPod(pod) == 0 {
			continue
		}
		podMap[fmt.Sprintf("%v/%v", pod.Namespace, pod.Name)] = true
	}
	for i := 0; i < nodeInfo.TotalGPUs; i++ {
		gpuId := fmt.Sprintf("%v", i)
		devInfo, ok := deviceMap[gpuId]
		if !ok {
			continue
		}
		totalGPUMemory += devInfo.GpuMemoryTotal
		allocatedGPUMemory := float64(0)
		idle := true
		for _, podName := range devInfo.PodNames {
			if podMap[podName] == true {
				idle = false
				break
			}
		}
		if idle == false {
			allocatedGPUMemory = devInfo.GpuMemoryTotal
			totalAllocatedGPUMemory += allocatedGPUMemory
		}
		totalUsedGPUMemory += devInfo.GpuMemoryUsed
		deviceLines = append(deviceLines, fmt.Sprintf("  %v\t%.1f GiB\t%.1f GiB\t%.1f GiB\t%.1f%%",
			devInfo.Id,
			utils.DataUnitTransfer("bytes", "GiB", devInfo.GpuMemoryTotal),
			utils.DataUnitTransfer("bytes", "GiB", allocatedGPUMemory),
			utils.DataUnitTransfer("bytes", "GiB", devInfo.GpuMemoryUsed),
			devInfo.GpuDutyCycle),
		)
	}
	if len(deviceLines) == 4 {
		deviceLines = []string{"", "GPUs:"}
	} else {
		deviceLines = append(deviceLines, "")
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
	/*
		deviceLines = append(deviceLines, "GPU Summary:", "  GPU(Total)\tGPU MEMORY(Total)\tGPU(Allocated)\tGPU MEMORY(Allocated)\tGPU(Unhealthy)")
		deviceLines = append(deviceLines, "  ----------\t-----------------\t--------------\t---------------------\t--------------")
		deviceLines = append(deviceLines, fmt.Sprintf("  %v\t%.1fGiB\t%v\t%.1fGiB\t%v",
			nodeInfo.TotalGPUs,
			utils.DataUnitTransfer("bytes", "GiB", totalGPUMemory),
			nodeInfo.AllocatedGPUs,
			utils.DataUnitTransfer("bytes", "GiB", totalAllocatedGPUMemory),
			g.getUnhealthyGPUs(),
		))
	*/
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
	PrintLine(w, "===================================== GPU MODE: Exclusive ===============================")
	totalGPUs := 0
	totalUnhealthyGPUs := 0
	totalAllocatedGPUs := 0
	//unhealthyPercent := float64(0)
	//allocatedPercent := float64(0)
	for _, node := range nodes {
		nodeInfo := node.Convert2NodeInfo().(types.GPUExclusiveNodeInfo)
		totalGPUs += nodeInfo.TotalGPUs
		totalAllocatedGPUs += nodeInfo.AllocatedGPUs
		totalUnhealthyGPUs += nodeInfo.UnhealthyGPUs
		PrintLine(w, node.WideFormat())
	}
	/*
		if totalGPUs != 0 {
			allocatedPercent = float64(totalAllocatedGPUs) / float64(totalGPUs) * 100
		}
			PrintLine(w, fmt.Sprintf("Allocated/Total GPUs of nodes with gpu exclusive mode In Cluster: %v/%v(%.1f%%)", totalAllocatedGPUs, totalGPUs, allocatedPercent))
			if totalUnhealthyGPUs != 0 {
				if totalGPUs != 0 {
					unhealthyPercent = float64(totalUnhealthyGPUs) / float64(totalGPUs) * 100
				}
				PrintLine(w, fmt.Sprintf("Unhealthy/Total GPUs of nodes with gpu exclusive mode In Cluster: %v/%v(%.1f%%)", totalUnhealthyGPUs, totalGPUs, unhealthyPercent))
			}
	*/
	PrintLine(w, "")
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

func NewGPUExclusiveNodeProcesser() NodeProcesser {
	return &nodeProcesser{
		nodeType:            types.GPUExclusiveNode,
		key:                 "gpuExclusiveNodes",
		builder:             NewGPUExclusiveNode,
		canBuildNode:        IsGPUExclusiveNode,
		displayNodesDetails: displayGPUExclusiveNodeDetails,
		displayNodesSummary: displayGPUExclusiveNodeSummary,
	}
}
