package topnode

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/apis/utils"
	v1 "k8s.io/api/core/v1"
)

var gpuExclusiveTemplate = `
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

type gpuexclusive struct {
	node *v1.Node
	pods []*v1.Pod
	baseNode
}

func NewGPUExclusiveNode(node *v1.Node, pods []*v1.Pod, index int, args ...interface{}) (Node, error) {
	return &gpuexclusive{
		node: node,
		pods: pods,
		baseNode: baseNode{
			index:    index,
			node:     node,
			pods:     pods,
			nodeType: types.GPUExclusiveNode,
		},
	}, nil
}

func (g *gpuexclusive) CapacityResourceCount() int {
	val, ok := g.node.Status.Capacity[v1.ResourceName(types.NvidiaGPUResourceName)]
	if !ok {
		return 0
	}
	return int(val.Value())
}

func (g *gpuexclusive) AllocatableResourceCount() int {
	val, ok := g.node.Status.Allocatable[v1.ResourceName(types.NvidiaGPUResourceName)]
	if !ok {
		return 0
	}
	return int(val.Value())
}

func (g *gpuexclusive) UsedResourceCount() int {
	usedGPUMemory := 0
	for _, pod := range g.pods {
		if utils.IsCompletedPod(pod) {
			continue
		}
		usedGPUMemory += utils.GPUCountInPod(pod)
	}
	return usedGPUMemory
}

func (g *gpuexclusive) IsHealthy() bool {
	return g.AllocatableResourceCount() == g.CapacityResourceCount()
}

func (g *gpuexclusive) convert2NodeInfo() types.GPUExclusiveNodeInfo {
	podInfos := []types.GPUExclusivePodInfo{}
	gpuExclusiveInfo := types.GPUExclusiveNodeInfo{
		CommonNodeInfo: types.CommonNodeInfo{
			Name:   g.Name(),
			IP:     g.IP(),
			Status: g.Status(),
			Type:   types.GPUExclusiveNode,
		},
		UnHealthyGPUs: g.CapacityResourceCount() - g.AllocatableResourceCount(),
		TotalGPUs:     g.CapacityResourceCount(),
		UsedGPUs:      g.UsedResourceCount(),
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
	return gpuExclusiveInfo
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
	lines := []string{"", "Instances:", "  NAMESPACE\tNAME\tGPU(Requested)"}
	lines = append(lines, "  ---------\t----\t--------------")
	for _, podInfo := range nodeInfo.PodInfos {
		lines = append(lines, fmt.Sprintf("  %v\t%v\t%v", podInfo.Namespace, podInfo.Name, podInfo.RequestGPU))
	}
	if len(lines) == 4 {
		lines = []string{}
	}
	unhealthyGPUs := fmt.Sprintf("%v/%v", 0, g.CapacityResourceCount())
	if g.CapacityResourceCount()-g.AllocatableResourceCount() != 0 {
		percent := float32(0)
		gpus := g.CapacityResourceCount() - g.AllocatableResourceCount()
		if g.CapacityResourceCount() != 0 {
			percent = float32(gpus) / float32(g.CapacityResourceCount()) * 100
		}
		unhealthyGPUs = fmt.Sprintf("%v/%v", gpus, g.CapacityResourceCount())
		if percent != float32(0) {
			unhealthyGPUs = fmt.Sprintf("%v/%v(%.1f%%)", gpus, g.CapacityResourceCount(), percent)
		}
	}
	usedGPUs := fmt.Sprintf("%v/%v", nodeInfo.UsedGPUs, nodeInfo.TotalGPUs)
	if nodeInfo.UsedGPUs != 0 {
		percent := float32(0)
		if g.CapacityResourceCount() != 0 {
			percent = float32(nodeInfo.UsedGPUs) / float32(nodeInfo.TotalGPUs) * 100
		}
		if percent != float32(0) {
			usedGPUs = fmt.Sprintf("%v/%v(%.1f%%)", nodeInfo.UsedGPUs, nodeInfo.TotalGPUs, percent)
		}
	}
	return fmt.Sprintf(strings.TrimRight(gpuExclusiveTemplate, "\n"),
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
	PrintLine(w, "===================================== GPUExclusiveNode ==================================")
	totalGPUs := 0
	unhealthyGPUs := 0
	usedGPUs := 0
	unhealthyPercent := float32(0)
	usedPercent := float32(0)
	for _, node := range nodes {
		nodeInfo := node.Convert2NodeInfo().(types.GPUExclusiveNodeInfo)
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

func displayGPUExclusiveNodeSummary(w *tabwriter.Writer, nodes []Node, isUnhealthy, showNodeType bool) {
	for _, node := range nodes {
		nodeInfo := node.Convert2NodeInfo().(types.GPUExclusiveNodeInfo)
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
				if typeInfo.Name == types.GPUExclusiveNode {
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
