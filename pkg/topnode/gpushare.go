package topnode

import (
	"fmt"
	"math"
	"strings"
	"text/tabwriter"

	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/apis/utils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

var GPUShareNodeDescription = `
  1.This node is enabled gpu sharing mode.
  2.Pods can request resource 'aliyun.com/gpu-mem' to use gpu sharing feature on this node 
`

var gpushareTemplate = `
Name:    %v
Status:  %v
Role:    %v
Type:    %v
Address: %v
Description:
%v
%v
`

var gpushareSummary = `
GPU Summary:
  Total GPUs:           %v
  Allocated GPUs:       %v
  Unhealthy GPUs:       %v
  Total GPU Memory:     %.1f GiB
  Allocated GPU Memory: %.1f GiB
  Used GPU Memory:      %.1f GiB
`

type gpushare struct {
	node       *v1.Node
	pods       []*v1.Pod
	gpuMetrics types.NodeGpuMetric
	baseNode
}

func NewGPUShareNode(client *kubernetes.Clientset, node *v1.Node, index int, args buildNodeArgs) (Node, error) {
	pods := getNodePods(node, args.pods)
	return &gpushare{
		node:       node,
		pods:       pods,
		gpuMetrics: getGPUMetricsByNodeName(node.Name, args.nodeGPUMetrics),
		baseNode: baseNode{
			index:    index,
			node:     node,
			pods:     pods,
			nodeType: types.GPUShareNode,
		},
	}, nil
}

func (g *gpushare) gpuMetricsIsEnabled() bool {
	return len(g.gpuMetrics) != 0
}

func (g *gpushare) getTotalGPUs() float64 {
	if len(g.gpuMetrics) != 0 {
		return float64(len(g.gpuMetrics))
	}
	val, ok := g.node.Status.Capacity[v1.ResourceName(types.GPUShareCountName)]
	if !ok {
		return 0
	}
	return float64(val.Value())
}

func (g *gpushare) getAllocatedGPUs() float64 {
	total := float64(0)
	for _, pod := range g.pods {
		if utils.IsCompletedPod(pod) {
			continue
		}
		allocation := utils.GetPodAllocation(pod)
		for key, allocatedGPUMemory := range allocation {
			totalGPUMemory := g.getTotalGPUMemoryOfDevice(key)
			if totalGPUMemory == 0 {
				continue
			}
			totalGPUMemory = utils.DataUnitTransfer("bytes", "GiB", totalGPUMemory)
			total += float64(allocatedGPUMemory) / totalGPUMemory
		}
	}
	return math.Round(total*10) / 10
}

func (g *gpushare) getTotalGPUMemory() float64 {
	totalGPUMemory := float64(0)
	for _, metric := range g.gpuMetrics {
		totalGPUMemory += metric.GpuMemoryTotal
	}
	// if gpu metric is enable,return the value given by prometheus
	if totalGPUMemory != 0 {
		return totalGPUMemory
	}
	val, ok := g.node.Status.Capacity[v1.ResourceName(types.GPUShareResourceName)]
	if !ok {
		return float64(0)
	}
	return utils.DataUnitTransfer("GiB", "bytes", float64(val.Value()))
}

func (g *gpushare) getAllocatedGPUMemory() float64 {
	allocatedGPUMemory := float64(0)
	for _, pod := range g.pods {
		if utils.IsCompletedPod(pod) {
			continue
		}
		allocation := utils.GetPodAllocation(pod)
		for _, gpuMem := range allocation {
			allocatedGPUMemory += float64(gpuMem)
		}
	}
	return utils.DataUnitTransfer("GiB", "bytes", allocatedGPUMemory)
}

func (g *gpushare) getUsedGPUMemory() float64 {
	if !g.gpuMetricsIsEnabled() {
		return float64(0)
	}
	usedGPUMemory := float64(0)
	for _, gpuMetric := range g.gpuMetrics {
		usedGPUMemory += gpuMetric.GpuMemoryUsed
	}
	return usedGPUMemory
}

func (g *gpushare) getDutyCycle() float64 {
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

func (g *gpushare) getUnhealthyGPUs() float64 {
	totalGPUs := g.getTotalGPUs()
	totalGPUMemory, ok := g.node.Status.Capacity[v1.ResourceName(types.GPUShareResourceName)]
	if !ok {
		return 0
	}
	allocatableGPUMemory, ok := g.node.Status.Allocatable[v1.ResourceName(types.GPUShareResourceName)]
	if !ok {
		return 0
	}
	if totalGPUs <= 0 {
		return 0
	}
	if totalGPUMemory.Value() <= 0 {
		return totalGPUs
	}
	unhealthyGPUMemory := totalGPUMemory.Value() - allocatableGPUMemory.Value()
	return float64(int64(totalGPUs) * unhealthyGPUMemory / totalGPUMemory.Value())
}

func (g *gpushare) getTotalGPUMemoryOfDevice(id string) float64 {
	if metric, ok := g.gpuMetrics[id]; ok {
		return metric.GpuMemoryTotal
	}
	totalGPUs := g.getTotalGPUs()
	totalGPUMemory := g.getTotalGPUMemory()
	if totalGPUs > 0 {
		return totalGPUMemory / float64(totalGPUs)
	}
	return 0
}

func (g *gpushare) GPUMetricsIsEnabled() bool {
	return len(g.gpuMetrics) != 0
}

func (g *gpushare) convert2NodeInfo() types.GPUShareNodeInfo {
	podInfos := []types.GPUSharePodInfo{}
	metrics := []*types.AdvancedGpuMetric{}
	for _, metric := range g.gpuMetrics {
		metrics = append(metrics, metric)
	}
	gpushareInfo := types.GPUShareNodeInfo{
		CommonNodeInfo: types.CommonNodeInfo{
			Name:        g.Name(),
			Description: GPUShareNodeDescription,
			IP:          g.IP(),
			Status:      g.Status(),
			Type:        types.GPUShareNode,
		},
		TotalGPUMemory:     g.getTotalGPUMemory(),
		AllocatedGPUMemory: g.getAllocatedGPUMemory(),
		CommonGPUNodeInfo: types.CommonGPUNodeInfo{
			AllocatedGPUs: g.getAllocatedGPUs(),
			TotalGPUs:     g.getTotalGPUs(),
			UnhealthyGPUs: g.getUnhealthyGPUs(),
			GPUMetrics:    metrics,
		},
	}
	// build devices
	deviceMap := map[string]types.GPUShareNodeDevice{}
	for i := 0; i < int(g.getTotalGPUs()); i++ {
		gpuId := fmt.Sprintf("%v", i)
		deviceMap[gpuId] = types.GPUShareNodeDevice{
			Id:             gpuId,
			TotalGPUMemory: g.getTotalGPUMemoryOfDevice(gpuId),
		}
	}
	for _, pod := range g.pods {
		if utils.IsCompletedPod(pod) {
			continue
		}
		allocation := utils.GetPodAllocation(pod)
		if len(allocation) == 0 {
			continue
		}
		for gpuId, gpuMem := range allocation {
			_, ok := deviceMap[gpuId]
			if !ok {
				deviceMap[gpuId] = types.GPUShareNodeDevice{
					Id:             gpuId,
					TotalGPUMemory: g.getTotalGPUMemoryOfDevice(gpuId),
				}
			}
			deviceInfo := deviceMap[gpuId]
			deviceInfo.AllocatedGPUMemory += utils.DataUnitTransfer("GiB", "bytes", float64(gpuMem))
			deviceMap[gpuId] = deviceInfo
		}
		status, _, _, _ := utils.DefinePodPhaseStatus(*pod)
		podInfos = append(podInfos, types.GPUSharePodInfo{
			Name:          pod.Name,
			Namespace:     pod.Namespace,
			Status:        status,
			RequestMemory: utils.GPUMemoryCountInPod(pod),
			Allocation:    allocation,
		})
	}
	devices := []types.GPUShareNodeDevice{}
	for _, dev := range deviceMap {
		devices = append(devices, dev)
	}
	gpushareInfo.Devices = devices
	gpushareInfo.PodInfos = podInfos
	gpushareInfo.GPUMetrics = metrics
	return gpushareInfo
}

func (g *gpushare) Convert2NodeInfo() interface{} {
	return g.convert2NodeInfo()
}

func (g *gpushare) AllDevicesAreHealthy() bool {
	return g.getUnhealthyGPUs() == 0
}

func (g *gpushare) WideFormat() string {
	role := strings.Join(g.Role(), ",")
	if role == "" {
		role = "<none>"
	}
	nodeInfo := g.convert2NodeInfo()
	lines := []string{}
	lines = g.displayPodInfos(lines, nodeInfo)
	lines = g.displayDeviceInfos(lines, nodeInfo)
	lines = append(lines, "")
	return fmt.Sprintf(gpushareTemplate,
		nodeInfo.Name,
		nodeInfo.Status,
		role,
		nodeInfo.Type,
		nodeInfo.IP,
		strings.Trim(nodeInfo.Description, "\n"),
		strings.Join(lines, "\n"),
	)
}

func (g *gpushare) displayPodInfos(lines []string, nodeInfo types.GPUShareNodeInfo) []string {
	podLines := []string{"Instances:", "  NAMESPACE\tNAME\tGPU_MEM(Requested)\tGPU_MEM(Allocated)"}
	podLines = append(podLines, "  ---------\t----\t------------------\t------------------")
	deviceMap := map[string]types.GPUShareNodeDevice{}
	for _, dev := range nodeInfo.Devices {
		deviceMap[dev.Id] = dev
	}
	for _, podInfo := range nodeInfo.PodInfos {
		items := []string{}
		for i := 0; i < int(nodeInfo.TotalGPUs); i++ {
			gpuId := fmt.Sprintf("%v", i)
			count, ok := podInfo.Allocation[gpuId]
			if !ok {
				continue
			}
			items = append(items, fmt.Sprintf("gpu%v(%.1fGiB)", gpuId, float64(count)))
		}
		podLines = append(podLines, fmt.Sprintf("  %v\t%v\t%.1f GiB\t%v", podInfo.Namespace, podInfo.Name, float64(podInfo.RequestMemory), strings.Join(items, ",")))
	}
	if len(podLines) == 3 {
		podLines = []string{}
	}
	lines = append(lines, podLines...)
	return lines
}

func (g *gpushare) displayDeviceInfos(lines []string, nodeInfo types.GPUShareNodeInfo) []string {
	if !g.GPUMetricsIsEnabled() {
		return g.displayDeviceUnderNoGPUMetric(lines, nodeInfo)
	}
	return g.displayDeviceUnderGPUMetric(lines, nodeInfo)
}

func (g *gpushare) displayDeviceUnderNoGPUMetric(lines []string, nodeInfo types.GPUShareNodeInfo) []string {
	deviceLines := []string{"GPUs:", "  INDEX\tMEMORY(Total)\tMEMORY(Allocated)\tPERCENT"}
	deviceLines = append(deviceLines, "  -----\t-------------\t-----------------\t-------")
	deviceMap := map[string]types.GPUShareNodeDevice{}
	for _, dev := range nodeInfo.Devices {
		deviceMap[dev.Id] = dev
	}
	for i := 0; i < int(nodeInfo.TotalGPUs); i++ {
		percent := float64(0)
		gpuId := fmt.Sprintf("%v", i)
		devInfo, ok := deviceMap[gpuId]
		if !ok {
			continue
		}
		if devInfo.TotalGPUMemory != 0 {
			percent = float64(devInfo.AllocatedGPUMemory) / float64(devInfo.TotalGPUMemory) * 100
		}
		deviceLines = append(deviceLines, fmt.Sprintf("  %v\t%.1f GiB\t%.1f GiB\t%.1f%%",
			devInfo.Id,
			utils.DataUnitTransfer("bytes", "GiB", devInfo.TotalGPUMemory),
			utils.DataUnitTransfer("bytes", "GiB", devInfo.AllocatedGPUMemory),
			percent,
		))
	}
	if len(deviceLines) == 3 {
		deviceLines = []string{}
	}
	deviceLines = append(deviceLines, "GPU Summary:")
	deviceLines = append(deviceLines, fmt.Sprintf("  Total GPUs: %.1f", nodeInfo.TotalGPUs))
	deviceLines = append(deviceLines, fmt.Sprintf("  Allocated GPUs: %.1f", nodeInfo.AllocatedGPUs))
	deviceLines = append(deviceLines, fmt.Sprintf("  Unhealthy GPUs: %.1f", g.getUnhealthyGPUs()))
	deviceLines = append(deviceLines, fmt.Sprintf("  Total GPU Memory: %.1f GiB", utils.DataUnitTransfer("bytes", "GiB", nodeInfo.TotalGPUMemory)))
	deviceLines = append(deviceLines, fmt.Sprintf("  Allocated GPU Memory: %.1f GiB", utils.DataUnitTransfer("bytes", "GiB", nodeInfo.AllocatedGPUMemory)))
	lines = append(lines, deviceLines...)
	return lines
}

func (g *gpushare) displayDeviceUnderGPUMetric(lines []string, nodeInfo types.GPUShareNodeInfo) []string {
	deviceLines := []string{"GPUs:", "  INDEX\tMEMORY(Total)\tMEMORY(Allocated)\tMEMORY(Used)\tDUTY_CYCLE"}
	deviceLines = append(deviceLines, "  -----\t-------------\t-----------------\t------------\t----------")
	deviceMap := map[string]*types.AdvancedGpuMetric{}
	totalUsedGPUMemory := float64(0)
	for _, dev := range g.gpuMetrics {
		deviceMap[dev.Id] = dev
	}
	for i := 0; i < int(nodeInfo.TotalGPUs); i++ {
		gpuId := fmt.Sprintf("%v", i)
		devInfo, ok := deviceMap[gpuId]
		if !ok {
			continue
		}
		allocatedGPUMemory := float64(0)
		for _, dev := range nodeInfo.Devices {
			if dev.Id == gpuId {
				allocatedGPUMemory = dev.AllocatedGPUMemory
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
		fmt.Sprintf(strings.Trim(gpushareSummary, "\n"),
			nodeInfo.TotalGPUs,
			nodeInfo.AllocatedGPUs,
			g.getUnhealthyGPUs(),
			utils.DataUnitTransfer("bytes", "GiB", nodeInfo.TotalGPUMemory),
			utils.DataUnitTransfer("bytes", "GiB", nodeInfo.AllocatedGPUMemory),
			utils.DataUnitTransfer("bytes", "GiB", totalUsedGPUMemory),
		))
	lines = append(lines, deviceLines...)
	return lines
}

/*
format is like:
===================================== GPUShareNode ======================================

Name:     cn-shanghai.192.168.7.183
Status:   Ready
Role:     <none>
Type:     GPUShare
Address:  192.168.7.183

Instances:
  NAMESPACE  NAME                                                       GPU_MEM(Requested)  GPU_MEM(Allocated)
  ---------  ----                                                       ------------------  ------------------
  default    binpack-0                                                  3                   GPU3->3
  default    fast-style-transfer-alpha-custom-serving-754c5ff685-vzjmt  5                   GPU3->5
  default    multi-gpushare-f4rgv                                       8                   GPU0->2,GPU1->2,GPU2->2,GPU3->2
  default    multi-gpushare-qcsqq                                       8                   GPU0->2,GPU1->2,GPU2->2,GPU3->2
  default    multi-gpushare-vz6xc                                       8                   GPU0->2,GPU1->2,GPU2->2,GPU3->2

GPUs:
  INDEX  MEMORY(Used/Total)  PERCENT
  -----  ------------------  -------
  GPU0   6/15(GiB)           40.0%
  GPU1   6/15(GiB)           40.0%
  GPU2   6/15(GiB)           40.0%
  GPU3   14/15(GiB)          93.3%

  Total(Memory/GiB): 60,Allocated(Memory/GiB): 32/60(53.3%),Unhealthy(Memory/GiB): 0/60

-----------------------------------------------------------------------------------------
Allocated/Total GPU Memory In Cluster: 32/60(53.3%)

===================================== End ===============================================
*/
func displayGPUShareNodeDetails(w *tabwriter.Writer, nodes []Node) {
	if len(nodes) == 0 {
		return
	}
	totalGPUMemory := float64(0)
	totalGPUs := float64(0)
	allocatedGPUs := float64(0)
	allocatedGPUMemory := float64(0)
	unhealthyGPUs := float64(0)
	for _, node := range nodes {
		nodeInfo := node.Convert2NodeInfo().(types.GPUShareNodeInfo)
		totalGPUs += nodeInfo.TotalGPUs
		totalGPUMemory += nodeInfo.TotalGPUMemory
		allocatedGPUs += nodeInfo.AllocatedGPUs
		allocatedGPUMemory += nodeInfo.AllocatedGPUMemory
		unhealthyGPUs += nodeInfo.UnhealthyGPUs
		PrintLine(w, node.WideFormat())
	}
}

func displayGPUShareNodeSummary(w *tabwriter.Writer, nodes []Node, isUnhealthy, showNodeType bool) (float64, float64, float64) {
	totalGPUs := float64(0)
	allocatedGPUs := float64(0)
	unhealthyGPUs := float64(0)
	for _, node := range nodes {
		nodeInfo := node.Convert2NodeInfo().(types.GPUShareNodeInfo)
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
		items = append(items, fmt.Sprintf("%.1f", nodeInfo.TotalGPUs))
		items = append(items, fmt.Sprintf("%.1f", nodeInfo.AllocatedGPUs))
		if showNodeType {
			for _, typeInfo := range types.NodeTypeSlice {
				if typeInfo.Name == types.GPUShareNode {
					items = append(items, typeInfo.Alias)
				}
			}
		}
		if isUnhealthy && nodeInfo.TotalGPUs != 0 {
			items = append(items, fmt.Sprintf("%.1f", nodeInfo.UnhealthyGPUs))
		}
		PrintLine(w, items...)
	}
	return totalGPUs, allocatedGPUs, unhealthyGPUs
}

func displayGPUShareNodesCustomSummary(w *tabwriter.Writer, nodes []Node) {
	if len(nodes) == 0 {
		return
	}
	header := []string{"NAME", "IPADDRESS", "ROLE", "STATUS", "GPUs(Allocated/Total)", "GPU_MEMORY(Allocated/Total)"}
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
	totalGPUs := float64(0)
	totalGPUMemory := float64(0)
	allocatedGPUMemory := float64(0)
	allocatedGPUs := float64(0)
	unhealthyGPUs := float64(0)
	for _, node := range nodes {
		nodeInfo := node.Convert2NodeInfo().(types.GPUShareNodeInfo)
		totalGPUs += nodeInfo.TotalGPUs
		allocatedGPUs += nodeInfo.AllocatedGPUs
		unhealthyGPUs += nodeInfo.UnhealthyGPUs
		totalGPUMemory += nodeInfo.TotalGPUMemory
		allocatedGPUMemory += nodeInfo.AllocatedGPUMemory
		items := []string{}
		items = append(items, node.Name())
		items = append(items, node.IP())
		role := nodeInfo.Role
		if role == "" {
			role = "<none>"
		}
		items = append(items, role)
		items = append(items, node.Status())
		items = append(items, fmt.Sprintf("%.1f/%.1f", nodeInfo.AllocatedGPUs, nodeInfo.TotalGPUs))
		items = append(items, fmt.Sprintf("%.1f/%.1f GiB",
			utils.DataUnitTransfer("bytes", "GiB", nodeInfo.AllocatedGPUMemory),
			utils.DataUnitTransfer("bytes", "GiB", nodeInfo.TotalGPUMemory)))
		if isUnhealthy {
			items = append(items, fmt.Sprintf("%.1f", nodeInfo.UnhealthyGPUs))
		}
		PrintLine(w, items...)
	}
	PrintLine(w, "-----------------------------------------------------------------------------------------------------------")
	// 1. print the utilization of gpus
	PrintLine(w, "Allocated/Total GPUs of nodes that own resource aliyun.com/gpu-mem In Cluster:")
	allocatedPercent := float64(0)
	if totalGPUs != 0 {
		allocatedPercent = float64(allocatedGPUs) / float64(totalGPUs) * 100
	}
	PrintLine(w, fmt.Sprintf("%.1f/%.1f (%.1f%%)", allocatedGPUs, totalGPUs, allocatedPercent))
	// 2. print the utilization of gpu memory
	PrintLine(w, "Allocated/Total GPU Memory of nodes that own resource aliyun.com/gpu-mem In Cluster:")
	allocatedGPUMemoryPercent := float64(0)
	if totalGPUMemory != 0 {
		allocatedGPUMemoryPercent = allocatedGPUMemory / totalGPUMemory * 100
	}
	PrintLine(w, fmt.Sprintf("%.1f/%.1f GiB(%.1f%%)",
		utils.DataUnitTransfer("bytes", "GiB", allocatedGPUMemory),
		utils.DataUnitTransfer("bytes", "GiB", totalGPUMemory),
		allocatedGPUMemoryPercent,
	))
	unhealthyPercent := float64(0)
	if totalGPUs != 0 {
		unhealthyPercent = float64(unhealthyGPUs) / float64(totalGPUs) * 100
	}
	if unhealthyGPUs != 0 {
		PrintLine(w, "Unhealthy/Total GPUs of nodes that own resource aliyun.com/gpu-mem In Cluster:")
		PrintLine(w, fmt.Sprintf("%.1f/%.1f (%.1f%%)", unhealthyGPUs, totalGPUs, unhealthyPercent))
	}
}

func IsGPUShareNode(node *v1.Node) bool {
	labels := strings.Split(types.GPUShareNodeLabels, ",")
	for _, label := range labels {
		gpushareLey := strings.Split(label, "=")[0]
		gpushareVal := strings.Split(label, "=")[1]
		if val, ok := node.Labels[gpushareLey]; ok && val == gpushareVal {
			return true
		}
	}
	return false
}

func NewGPUShareNodeProcesser() NodeProcesser {
	return &nodeProcesser{
		nodeType:                  types.GPUShareNode,
		key:                       "gpuShareNodes",
		builder:                   NewGPUShareNode,
		canBuildNode:              IsGPUShareNode,
		displayNodesDetails:       displayGPUShareNodeDetails,
		displayNodesSummary:       displayGPUShareNodeSummary,
		displayNodesCustomSummary: displayGPUShareNodesCustomSummary,
	}
}
