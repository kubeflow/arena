package topnode

import (
	"encoding/json"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/apis/utils"
	v1 "k8s.io/api/core/v1"
)

var gpuTopologyTemplate = `
Name:          %v
Status:        %v
Role:          %v
Type:          %v
Address:       %v
TotalGPUs:     %v
UsedGPUs:      %v 
UnhealthyGPUs: %v
%v
-----------------------------------------------------------------------------------------
`

type gputopo struct {
	node      *v1.Node
	pods      []*v1.Pod
	configmap *v1.ConfigMap
	baseNode
}

func NewGPUTopologyNode(node *v1.Node, pods []*v1.Pod, index int, args ...interface{}) (Node, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("build gpu topology node needs configmap")
	}
	configmaps, ok := args[0].([]*v1.ConfigMap)
	if !ok {
		return nil, fmt.Errorf("build gpu topology node needs configmap,but the given arg is not  []*v1.ConfigMap")
	}
	var configmap *v1.ConfigMap
	for _, c := range configmaps {
		if val, ok := c.Labels["nodename"]; ok && val == node.Name {
			configmap = c.DeepCopy()
			break
		}
	}
	if configmap == nil {
		return nil, fmt.Errorf("not found configmap for building gpu topology node")
	}
	return &gputopo{
		node:      node,
		pods:      pods,
		configmap: configmap,
		baseNode: baseNode{
			index:    index,
			node:     node,
			pods:     pods,
			nodeType: types.GPUTopologyNode,
		},
	}, nil
}

func (g *gputopo) CapacityResourceCount() int {
	val, ok := g.node.Status.Capacity[v1.ResourceName(types.AliyunGPUResourceName)]
	if !ok {
		return 0
	}
	return int(val.Value())
}

func (g *gputopo) AllocatableResourceCount() int {
	val, ok := g.node.Status.Allocatable[v1.ResourceName(types.AliyunGPUResourceName)]
	if !ok {
		return 0
	}
	return int(val.Value())
}

func (g *gputopo) UsedResourceCount() int {
	usedGPUMemory := 0
	for _, pod := range g.pods {
		if utils.IsCompletedPod(pod) {
			continue
		}
		usedGPUMemory += utils.AliyunGPUCountInPod(pod)
	}
	return usedGPUMemory
}

func (g *gputopo) IsHealthy() bool {
	return g.AllocatableResourceCount() == g.CapacityResourceCount()
}

func (g *gputopo) convert2NodeInfo() types.GPUTopologyNodeInfo {
	podInfos := []types.GPUTopologyPodInfo{}
	// 1.initilize the common node information
	gpuTopologyNodeInfo := types.GPUTopologyNodeInfo{
		CommonNodeInfo: types.CommonNodeInfo{
			Name:   g.Name(),
			IP:     g.IP(),
			Status: g.Status(),
			Type:   types.GPUTopologyNode,
		},
		UnHealthyGPUs: g.CapacityResourceCount() - g.AllocatableResourceCount(),
		TotalGPUs:     g.CapacityResourceCount(),
		UsedGPUs:      g.UsedResourceCount(),
	}
	// 2.parse devices from field "devices" of configmap
	devices := map[string]types.GPUTopologyDeviceInfo{}
	if val, ok := g.configmap.Data["devices"]; ok {
		devMap := map[string]string{}
		json.Unmarshal([]byte(val), &devMap)
		for id, healthy := range devMap {
			devices[id] = types.GPUTopologyDeviceInfo{
				Index:   id,
				Status:  "idle",
				Healthy: healthy == "Healthy",
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
			devInfo, ok := devices[gpuId]
			if !ok {
				continue
			}
			devInfo.Status = "using"
			devices[gpuId] = devInfo
		}
		podInfos = append(podInfos, types.GPUTopologyPodInfo{
			Name:        pod.Name,
			Namespace:   pod.Namespace,
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
	for _, dev := range devices {
		if gpuTopologyNodeInfo.Devices == nil {
			gpuTopologyNodeInfo.Devices = []types.GPUTopologyDeviceInfo{}
		}
		gpuTopologyNodeInfo.Devices = append(gpuTopologyNodeInfo.Devices, dev)
	}
	return gpuTopologyNodeInfo
}

func (g *gputopo) Convert2NodeInfo() interface{} {
	return g.convert2NodeInfo()
}

func (g *gputopo) WideFormat() string {
	role := strings.Join(g.Role(), ",")
	if role == "" {
		role = "<none>"
	}
	nodeInfo := g.convert2NodeInfo()
	lines := []string{"", "Instances:", "  NAMESPACE\tNAME\tGPU(Requested)\tGPU(Allocated)"}
	lines = append(lines, "  ---------\t----\t--------------\t--------------")
	for _, podInfo := range nodeInfo.PodInfos {
		if len(podInfo.Allocation) == 0 {
			continue
		}
		lines = append(lines, fmt.Sprintf("  %v\t%v\t%v\t%v",
			podInfo.Namespace,
			podInfo.Name,
			podInfo.RequestGPU,
			strings.Join(podInfo.Allocation, ","),
		))
	}
	if len(lines) == 4 {
		lines = []string{}
	}
	lines = append(lines, "", "GPUs:", "  INDEX\tSTATUS\tHEALTHY")
	lines = append(lines, "  ----\t------\t-------")
	unhealthyGPUs := 0
	usedGPUs := 0
	totalGPUs := len(nodeInfo.Devices)
	for _, dev := range nodeInfo.Devices {
		if !dev.Healthy {
			unhealthyGPUs++
		}
		if dev.Status == "using" {
			usedGPUs++
		}
		lines = append(lines, fmt.Sprintf("  GPU%v\t%v\t%v", dev.Index, dev.Status, dev.Healthy))
	}
	unhealthyGPUField := fmt.Sprintf("%v/%v", unhealthyGPUs, totalGPUs)
	if unhealthyGPUs != 0 {
		percent := float32(0)
		if totalGPUs != 0 {
			percent = float32(unhealthyGPUs) / float32(totalGPUs) * 100
		}
		if percent != float32(0) {
			unhealthyGPUField = fmt.Sprintf("%v/%v(%.1f%%)", unhealthyGPUs, totalGPUs, percent)
		}
	}
	usedGPUsField := fmt.Sprintf("%v/%v", usedGPUs, totalGPUs)
	if usedGPUs != 0 {
		percent := float32(0)
		if totalGPUs != 0 {
			percent = float32(usedGPUs) / float32(totalGPUs) * 100
		}
		if percent != float32(0) {
			usedGPUsField = fmt.Sprintf("%v/%v(%.1f%%)", usedGPUs, totalGPUs, percent)
		}
	}
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("  Total: %v,Allocated: %v,Unhealthy: %v", totalGPUs, usedGPUsField, unhealthyGPUField))
	if len(nodeInfo.GPUTopology.LinkMatrix) != 0 {
		header := []string{"  "}
		lines = append(lines, "", "LinkTypeMatrix:")
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
		lines = append(lines, "", "BandwidthMatrix:")
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
	return fmt.Sprintf(strings.TrimRight(gpuTopologyTemplate, "\n"),
		nodeInfo.Name,
		nodeInfo.Status,
		role,
		nodeInfo.Type,
		nodeInfo.IP,
		nodeInfo.TotalGPUs,
		usedGPUs,
		unhealthyGPUs,
		strings.Join(lines, "\n"),
	)
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
	PrintLine(w, "===================================== GPUTopologyNode ===================================")
	totalGPUs := 0
	unhealthyGPUs := 0
	usedGPUs := 0
	unhealthyPercent := float32(0)
	usedPercent := float32(0)
	for _, node := range nodes {
		nodeInfo := node.Convert2NodeInfo().(types.GPUTopologyNodeInfo)
		totalGPUs += nodeInfo.TotalGPUs
		usedGPUs += nodeInfo.UsedGPUs
		nodeUnhealthy := node.CapacityResourceCount() - node.AllocatableResourceCount()
		unhealthyGPUs += nodeUnhealthy
		PrintLine(w, node.WideFormat())
	}
	if totalGPUs != 0 {
		usedPercent = float32(usedGPUs) / float32(totalGPUs) * 100
	}
	PrintLine(w, fmt.Sprintf("Allocated/Total GPUs In Cluster: %v/%v(%.1f%%)", usedGPUs, totalGPUs, usedPercent))
	if unhealthyGPUs != 0 {
		if totalGPUs != 0 {
			unhealthyPercent = float32(unhealthyGPUs) / float32(totalGPUs) * 100
		}
		PrintLine(w, fmt.Sprintf("Unhealthy/Total GPUs In Cluster: %v/%v(%.1f%%)", unhealthyGPUs, totalGPUs, unhealthyPercent))
	}
	PrintLine(w, "")
}

func displayGPUTopologyNodeSummary(w *tabwriter.Writer, nodes []Node, isUnhealthy, showNodeType bool) {
	for _, node := range nodes {
		nodeInfo := node.Convert2NodeInfo().(types.GPUTopologyNodeInfo)
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
		items = append(items, fmt.Sprintf("%v", nodeInfo.UsedGPUs))
		if showNodeType {
			for _, typeInfo := range types.NodeTypeSlice {
				if typeInfo.Name == types.GPUTopologyNode {
					items = append(items, typeInfo.Alias)
				}
			}
		}
		if isUnhealthy {
			items = append(items, fmt.Sprintf("%v", nodeInfo.UnHealthyGPUs))
		}
		PrintLine(w, items...)
	}
}

func NewGPUTopologyNodeProcesser() NodeProcesser {
	return &nodeProcesser{
		nodeType:            types.GPUTopologyNode,
		key:                 "gpuTopologyNodes",
		builder:             NewGPUTopologyNode,
		canBuildNode:        IsGPUTopologyNode,
		displayNodesDetails: displayGPUTopologyNodeDetails,
		displayNodesSummary: displayGPUTopologyNodeSummary,
	}
}
