package topnode

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/apis/utils"
	v1 "k8s.io/api/core/v1"
)

var gpushareTemplate = `
Name:     %v
Status:   %v
Role:     %v
Type:     %v
Address:  %v
%v
-----------------------------------------------------------------------------------------
`

type gpushare struct {
	node *v1.Node
	pods []*v1.Pod
	baseNode
}

func NewGPUShareNode(node *v1.Node, pods []*v1.Pod, index int, args ...interface{}) (Node, error) {
	return &gpushare{
		node: node,
		pods: pods,
		baseNode: baseNode{
			index:    index,
			node:     node,
			pods:     pods,
			nodeType: types.GPUShareNode,
		},
	}, nil
}

func (g *gpushare) CapacityResourceCount() int {
	val, ok := g.node.Status.Capacity[v1.ResourceName(types.GPUShareResourceName)]
	if !ok {
		return 0
	}
	return int(val.Value())
}

func (g *gpushare) AllocatableResourceCount() int {
	val, ok := g.node.Status.Allocatable[v1.ResourceName(types.GPUShareResourceName)]
	if !ok {
		return 0
	}
	return int(val.Value())
}

func (g *gpushare) UsedResourceCount() int {
	usedGPUMemory := 0
	for _, pod := range g.pods {
		if utils.IsCompletedPod(pod) {
			continue
		}
		usedGPUMemory += utils.GPUMemoryCountInPod(pod)
	}
	return usedGPUMemory
}

func (g *gpushare) gpuCount() int {
	val, ok := g.node.Status.Capacity[v1.ResourceName(types.GPUShareCountName)]
	if !ok {
		return 0
	}
	return int(val.Value())
}

func (g *gpushare) singleGPUMemory() int {
	singleGPUMemory := 0
	if g.gpuCount() != 0 {
		singleGPUMemory = g.CapacityResourceCount() / g.gpuCount()
	}
	return singleGPUMemory
}

func (g *gpushare) IsHealthy() bool {
	return g.AllocatableResourceCount() == g.CapacityResourceCount()
}

func (g *gpushare) convert2NodeInfo() types.GPUShareNodeInfo {
	singleGPUMemory := g.singleGPUMemory()
	deviceMap := map[string]types.GPUShareDeviceInfo{}
	podInfos := []types.GPUSharePodInfo{}
	gpushareInfo := types.GPUShareNodeInfo{
		CommonNodeInfo: types.CommonNodeInfo{
			Name:   g.Name(),
			IP:     g.IP(),
			Status: g.Status(),
			Type:   types.GPUShareNode,
		},
		GPUCount:        g.gpuCount(),
		UnHealthyGPUMem: g.CapacityResourceCount() - g.AllocatableResourceCount(),
		TotalGPUMem:     g.CapacityResourceCount(),
		UsedGPUMem:      g.UsedResourceCount(),
	}
	for _, pod := range g.pods {
		if utils.IsCompletedPod(pod) {
			continue
		}
		allocation := utils.GetPodAllocation(pod)
		if len(allocation) == 0 {
			continue
		}
		for gpuId, count := range allocation {
			_, ok := deviceMap[gpuId]
			if !ok {
				deviceMap[gpuId] = types.GPUShareDeviceInfo{
					ID:          gpuId,
					TotalGPUMem: singleGPUMemory,
				}
			}
			deviceInfo := deviceMap[gpuId]
			deviceInfo.UsedGPUMem = deviceInfo.UsedGPUMem + count
			deviceMap[gpuId] = deviceInfo
		}
		podInfos = append(podInfos, types.GPUSharePodInfo{
			Name:          pod.Name,
			Namespace:     pod.Namespace,
			RequestMemory: utils.GPUMemoryCountInPod(pod),
			Allocation:    allocation,
		})
	}
	devices := []types.GPUShareDeviceInfo{}
	for _, dev := range deviceMap {
		devices = append(devices, dev)
	}
	gpushareInfo.Devices = devices
	gpushareInfo.Pods = podInfos
	return gpushareInfo
}

func (g *gpushare) Convert2NodeInfo() interface{} {
	return g.convert2NodeInfo()
}

func (g *gpushare) WideFormat() string {
	role := strings.Join(g.Role(), ",")
	if role == "" {
		role = "<none>"
	}
	nodeInfo := g.convert2NodeInfo()
	deviceInfos := map[string]types.GPUShareDeviceInfo{}
	lines := []string{"", "Instances:", "  NAMESPACE\tNAME\tGPU_MEM(Requested)\tGPU_MEM(Allocated)"}
	lines = append(lines, "  ---------\t----\t------------------\t------------------")
	for _, dev := range nodeInfo.Devices {
		deviceInfos[dev.ID] = dev
	}
	for _, podInfo := range nodeInfo.Pods {
		items := []string{}
		for i := 0; i < g.gpuCount(); i++ {
			gpuId := fmt.Sprintf("%v", i)
			count, ok := podInfo.Allocation[gpuId]
			if !ok {
				continue
			}
			items = append(items, fmt.Sprintf("%v(%vGiB)", gpuId, count))
		}
		lines = append(lines, fmt.Sprintf("  %v\t%v\t%vGiB\t%v", podInfo.Namespace, podInfo.Name, podInfo.RequestMemory, strings.Join(items, ",")))
	}
	if len(lines) == 4 {
		lines = []string{}
	}
	lines = append(lines, []string{"", "GPUs:", "  INDEX\tMEMORY(Used/Total)\tPERCENT"}...)
	lines = append(lines, "  -----\t------------------\t-------")
	for i := 0; i < g.gpuCount(); i++ {
		percent := float32(0)
		gpuId := fmt.Sprintf("%v", i)
		devInfo, ok := deviceInfos[gpuId]
		if !ok {
			gpuMem := g.singleGPUMemory()
			lines = append(lines, fmt.Sprintf("  %v\t%v/%v(GiB)\t%.0f%%", gpuId, 0, gpuMem, percent))
			continue
		}
		if devInfo.TotalGPUMem != 0 {
			percent = float32(devInfo.UsedGPUMem) / float32(devInfo.TotalGPUMem) * 100
		}
		lines = append(lines, fmt.Sprintf("  %v\t%v/%v(GiB)\t%.1f%%", devInfo.ID, devInfo.UsedGPUMem, devInfo.TotalGPUMem, percent))
	}
	unhealthyGPUMems := fmt.Sprintf("%v/%v", 0, g.CapacityResourceCount())
	if g.CapacityResourceCount()-g.AllocatableResourceCount() != 0 {
		percent := float32(0)
		gpuMems := g.CapacityResourceCount() - g.AllocatableResourceCount()
		if g.CapacityResourceCount() != 0 {
			percent = float32(gpuMems) / float32(g.CapacityResourceCount()) * 100
		}
		unhealthyGPUMems = fmt.Sprintf("%v/%v", gpuMems, g.CapacityResourceCount())
		if percent != float32(0) {
			unhealthyGPUMems = fmt.Sprintf("%v/%v(%.1f%%)", gpuMems, g.CapacityResourceCount(), percent)
		}
	}
	usedGPUMem := fmt.Sprintf("%v/%v", nodeInfo.UsedGPUMem, nodeInfo.TotalGPUMem)
	if nodeInfo.UsedGPUMem != 0 {
		percent := float32(0)
		if g.CapacityResourceCount() != 0 {
			percent = float32(nodeInfo.UsedGPUMem) / float32(nodeInfo.TotalGPUMem) * 100
		}
		if percent != float32(0) {
			usedGPUMem = fmt.Sprintf("%v/%v(%.1f%%)", nodeInfo.UsedGPUMem, nodeInfo.TotalGPUMem, percent)
		}
	}
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("  Total(Memory/GiB): %v,Allocated(Memory/GiB): %v,Unhealthy(Memory/GiB): %v",
		nodeInfo.TotalGPUMem,
		usedGPUMem,
		unhealthyGPUMems,
	))
	lines = append(lines, "")
	return fmt.Sprintf(strings.TrimRight(gpushareTemplate, "\n"),
		nodeInfo.Name,
		nodeInfo.Status,
		role,
		nodeInfo.Type,
		nodeInfo.IP,
		strings.Join(lines, "\n"),
	)
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
	PrintLine(w, "===================================== GPUShareNode ======================================")
	totalGPUMem := 0
	unhealthyGPUMem := 0
	usedGPUMem := 0
	unhealthyPercent := float32(0)
	usedPercent := float32(0)
	for _, node := range nodes {
		nodeInfo := node.Convert2NodeInfo().(types.GPUShareNodeInfo)
		totalGPUMem += nodeInfo.TotalGPUMem
		usedGPUMem += nodeInfo.UsedGPUMem
		nodeUnhealthy := node.CapacityResourceCount() - node.AllocatableResourceCount()
		unhealthyGPUMem += nodeUnhealthy
		PrintLine(w, node.WideFormat())
	}
	if totalGPUMem != 0 {
		usedPercent = float32(usedGPUMem) / float32(totalGPUMem) * 100
	}
	PrintLine(w, fmt.Sprintf("Allocated/Total GPU Memory In Cluster: %v/%v(%.1f%%)", usedGPUMem, totalGPUMem, usedPercent))
	if unhealthyGPUMem != 0 {
		if totalGPUMem != 0 {
			unhealthyPercent = float32(unhealthyGPUMem) / float32(totalGPUMem) * 100
		}
		PrintLine(w, fmt.Sprintf("Unhealthy/Total GPU Memory In Cluster: %v/%v(%.1f%%)", unhealthyGPUMem, totalGPUMem, unhealthyPercent))
	}
	PrintLine(w, "")
}

func displayGPUShareNodeSummary(w *tabwriter.Writer, nodes []Node, isUnhealthy, showNodeType bool) {
	for _, node := range nodes {
		nodeInfo := node.Convert2NodeInfo().(types.GPUShareNodeInfo)
		items := []string{}
		items = append(items, node.Name())
		items = append(items, node.IP())
		role := nodeInfo.Role
		if role == "" {
			role = "<none>"
		}
		items = append(items, role)
		items = append(items, node.Status())
		items = append(items, fmt.Sprintf("%v", nodeInfo.GPUCount))
		used := float32(0)
		totalGPUMem := nodeInfo.TotalGPUMem
		usedGPUMem := nodeInfo.UsedGPUMem
		if totalGPUMem != 0 && nodeInfo.GPUCount != 0 {
			used = float32(nodeInfo.GPUCount*usedGPUMem) / float32(totalGPUMem)
		}
		items = append(items, fmt.Sprintf("%.1f", used))
		if showNodeType {
			for _, typeInfo := range types.NodeTypeSlice {
				if typeInfo.Name == types.GPUShareNode {
					items = append(items, typeInfo.Alias)
				}
			}
		}
		if isUnhealthy && totalGPUMem != 0 {
			unhealthy := nodeInfo.GPUCount * nodeInfo.UnHealthyGPUMem / nodeInfo.TotalGPUMem
			items = append(items, fmt.Sprintf("%v", unhealthy))
		}
		PrintLine(w, items...)
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
		nodeType:            types.GPUShareNode,
		key:                 "gpuShareNodes",
		builder:             NewGPUShareNode,
		canBuildNode:        IsGPUShareNode,
		displayNodesDetails: displayGPUShareNodeDetails,
		displayNodesSummary: displayGPUShareNodeSummary,
	}
}
