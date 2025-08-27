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

var QGPUNodeDescription = `
  1.This node is enabled qgpu mode.
  2.Pods can request resource 'tke.cloud.tencent.com/qgpu-core / tke.cloud.tencent.com/qgpu-memory' to use qgpu feature on this node 
`

var qgpuTemplate = `
Name:    %v
Status:  %v
Role:    %v
Type:    %v
Address: %v
Description:
%v
%v
`

var qgpuSummary = `
GPU Summary:
  Total GPUs:           %v
  Allocated GPUs:       %v
  Unhealthy GPUs:       %v
  Total GPU Memory:     %.1f GiB
  Allocated GPU Memory: %.1f GiB
  Used GPU Memory:      %.1f GiB
`

type qgpu struct {
	node       *v1.Node
	pods       []*v1.Pod
	gpuMetrics types.NodeGpuMetric
	baseNode
}

func NewQGPUNode(client *kubernetes.Clientset, node *v1.Node, index int, args buildNodeArgs) (Node, error) {
	pods := getNodePods(node, args.pods)
	return &qgpu{
		node:       node,
		pods:       pods,
		gpuMetrics: getGPUMetricsByNodeName(node.Name, args.nodeGPUMetrics),
		baseNode: baseNode{
			index:    index,
			node:     node,
			pods:     pods,
			nodeType: types.QGPUNode,
		},
	}, nil
}

func (g *qgpu) gpuMetricsIsEnabled() bool {
	return len(g.gpuMetrics) != 0
}

func (g *qgpu) getTotalGPUs() float64 {
	if len(g.gpuMetrics) != 0 {
		return float64(len(g.gpuMetrics))
	}
	val, ok := g.node.Status.Capacity[v1.ResourceName(types.QGPUCoreResourceName)]
	if !ok {
		return 0
	}
	return float64(val.Value()) / 100
}

func (g *qgpu) getAllocatedGPUs() float64 {
	total := float64(0)
	for _, pod := range g.pods {
		if utils.IsCompletedPod(pod) || !g.isQGPUPod(pod) {
			continue
		}
		allocation := g.getPodAllocation(pod)
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

func (g *qgpu) getTotalGPUMemory() float64 {
	totalGPUMemory := float64(0)
	for _, metric := range g.gpuMetrics {
		totalGPUMemory += metric.GpuMemoryTotal
	}
	// if gpu metric is enable,return the value given by prometheus
	if totalGPUMemory != 0 {
		return totalGPUMemory
	}
	val, ok := g.node.Status.Capacity[v1.ResourceName(types.QGPUMemoryResourceName)]
	if !ok {
		return float64(0)
	}
	return utils.DataUnitTransfer("GiB", "bytes", float64(val.Value()))
}

func (g *qgpu) getAllocatedGPUMemory() float64 {
	allocatedGPUMemory := float64(0)
	for _, pod := range g.pods {
		if utils.IsCompletedPod(pod) || !g.isQGPUPod(pod) {
			continue
		}
		allocation := g.getPodAllocation(pod)
		for _, gpuMem := range allocation {
			allocatedGPUMemory += float64(gpuMem)
		}
	}
	return utils.DataUnitTransfer("GiB", "bytes", allocatedGPUMemory)
}

func (g *qgpu) getUsedGPUMemory() float64 {
	if !g.gpuMetricsIsEnabled() {
		return float64(0)
	}
	usedGPUMemory := float64(0)
	for _, gpuMetric := range g.gpuMetrics {
		usedGPUMemory += gpuMetric.GpuMemoryUsed
	}
	return usedGPUMemory
}

func (g *qgpu) getDutyCycle() float64 {
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

func (g *qgpu) getUnhealthyGPUs() float64 {
	totalGPUs := g.getTotalGPUs()
	totalGPUMemory, ok := g.node.Status.Capacity[v1.ResourceName(types.QGPUMemoryResourceName)]
	if !ok {
		return 0
	}
	allocatableGPUMemory, ok := g.node.Status.Allocatable[v1.ResourceName(types.QGPUMemoryResourceName)]
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

func (g *qgpu) getTotalGPUMemoryOfDevice(id string) float64 {
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

func (g *qgpu) GPUMetricsIsEnabled() bool {
	return len(g.gpuMetrics) != 0
}

func (g *qgpu) convert2NodeInfo() types.QGPUNodeInfo {
	podInfos := []types.QGPUPodInfo{}
	metrics := []*types.AdvancedGpuMetric{}
	for _, metric := range g.gpuMetrics {
		metrics = append(metrics, metric)
	}
	qgpuInfo := types.QGPUNodeInfo{
		CommonNodeInfo: types.CommonNodeInfo{
			Name:        g.Name(),
			Description: QGPUNodeDescription,
			IP:          g.IP(),
			Status:      g.Status(),
			Type:        types.QGPUNode,
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
	deviceMap := map[string]types.QGPUNodeDevice{}
	for i := 0; i < int(g.getTotalGPUs()); i++ {
		gpuId := fmt.Sprintf("%v", i)
		deviceMap[gpuId] = types.QGPUNodeDevice{
			Id:             gpuId,
			TotalGPUMemory: g.getTotalGPUMemoryOfDevice(gpuId),
		}
	}
	for _, pod := range g.pods {
		if utils.IsCompletedPod(pod) || !g.isQGPUPod(pod) {
			continue
		}
		allocation := g.getPodAllocation(pod)
		if len(allocation) == 0 {
			continue
		}
		for gpuId, gpuMem := range allocation {
			_, ok := deviceMap[gpuId]
			if !ok {
				deviceMap[gpuId] = types.QGPUNodeDevice{
					Id:             gpuId,
					TotalGPUMemory: g.getTotalGPUMemoryOfDevice(gpuId),
				}
			}
			deviceInfo := deviceMap[gpuId]
			deviceInfo.AllocatedGPUMemory += utils.DataUnitTransfer("GiB", "bytes", float64(gpuMem))
			deviceMap[gpuId] = deviceInfo
		}
		status, _, _, _ := utils.DefinePodPhaseStatus(*pod)
		podInfos = append(podInfos, types.QGPUPodInfo{
			Name:          pod.Name,
			Namespace:     pod.Namespace,
			Status:        status,
			RequestMemory: g.qGPUMemoryInPod(pod),
			Allocation:    allocation,
		})
	}
	devices := []types.QGPUNodeDevice{}
	for _, dev := range deviceMap {
		devices = append(devices, dev)
	}
	qgpuInfo.Devices = devices
	qgpuInfo.PodInfos = podInfos
	qgpuInfo.GPUMetrics = metrics
	return qgpuInfo
}

func (g *qgpu) Convert2NodeInfo() interface{} {
	return g.convert2NodeInfo()
}

func (g *qgpu) AllDevicesAreHealthy() bool {
	return g.getUnhealthyGPUs() == 0
}

func (g *qgpu) WideFormat() string {
	role := strings.Join(g.Role(), ",")
	if role == "" {
		role = "<none>"
	}
	nodeInfo := g.convert2NodeInfo()
	lines := []string{}
	lines = g.displayPodInfos(lines, nodeInfo)
	lines = g.displayDeviceInfos(lines, nodeInfo)
	lines = append(lines, "")
	return fmt.Sprintf(qgpuTemplate,
		nodeInfo.Name,
		nodeInfo.Status,
		role,
		nodeInfo.Type,
		nodeInfo.IP,
		strings.Trim(nodeInfo.Description, "\n"),
		strings.Join(lines, "\n"),
	)
}

func (g *qgpu) displayPodInfos(lines []string, nodeInfo types.QGPUNodeInfo) []string {
	podLines := []string{"Pods:", "  NAMESPACE\tNAME\tGPU_MEM(Requested)\tGPU_MEM(Allocated)"}
	podLines = append(podLines, "  ---------\t----\t------------------\t------------------")
	deviceMap := map[string]types.QGPUNodeDevice{}
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

func (g *qgpu) displayDeviceInfos(lines []string, nodeInfo types.QGPUNodeInfo) []string {
	if !g.GPUMetricsIsEnabled() {
		return g.displayDeviceUnderNoGPUMetric(lines, nodeInfo)
	}
	return g.displayDeviceUnderGPUMetric(lines, nodeInfo)
}

func (g *qgpu) displayDeviceUnderNoGPUMetric(lines []string, nodeInfo types.QGPUNodeInfo) []string {
	deviceLines := []string{"GPUs:", "  INDEX\tMEMORY(Total)\tMEMORY(Allocated)\tPERCENT"}
	deviceLines = append(deviceLines, "  -----\t-------------\t-----------------\t-------")
	deviceMap := map[string]types.QGPUNodeDevice{}
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
		deviceLines = append(deviceLines, fmt.Sprintf("  %v\t%v GiB\t%v GiB\t%.1f%%",
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
	deviceLines = append(deviceLines, fmt.Sprintf("  Total GPUs: %v", nodeInfo.TotalGPUs))
	deviceLines = append(deviceLines, fmt.Sprintf("  Allocated GPUs: %v", nodeInfo.AllocatedGPUs))
	deviceLines = append(deviceLines, fmt.Sprintf("  Unhealthy GPUs: %v", g.getUnhealthyGPUs()))
	deviceLines = append(deviceLines, fmt.Sprintf("  Total GPU Memory: %.1f GiB", utils.DataUnitTransfer("bytes", "GiB", nodeInfo.TotalGPUMemory)))
	deviceLines = append(deviceLines, fmt.Sprintf("  Allocated GPU Memory: %.1f GiB", utils.DataUnitTransfer("bytes", "GiB", nodeInfo.AllocatedGPUMemory)))
	lines = append(lines, deviceLines...)
	return lines
}

func (g *qgpu) displayDeviceUnderGPUMetric(lines []string, nodeInfo types.QGPUNodeInfo) []string {
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
		fmt.Sprintf(strings.Trim(qgpuSummary, "\n"),
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
===================================== QGPUNode ======================================

Name:     cn-shanghai.192.168.7.183
Status:   Ready
Role:     <none>
Type:     QGPU
Address:  192.168.7.183

Instances:
  NAMESPACE  NAME                                                       GPU_MEM(Requested)  GPU_MEM(Allocated)
  ---------  ----                                                       ------------------  ------------------
  default    binpack-0                                                  3                   GPU3->3
  default    fast-style-transfer-alpha-custom-serving-754c5ff685-vzjmt  5                   GPU3->5
  default    multi-qgpu-f4rgv                                       8                   GPU0->2,GPU1->2,GPU2->2,GPU3->2
  default    multi-qgpu-qcsqq                                       8                   GPU0->2,GPU1->2,GPU2->2,GPU3->2
  default    multi-qgpu-vz6xc                                       8                   GPU0->2,GPU1->2,GPU2->2,GPU3->2

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
func displayQGPUNodeDetails(w *tabwriter.Writer, nodes []Node) {
	if len(nodes) == 0 {
		return
	}
	totalGPUMemory := float64(0)
	totalGPUs := float64(0)
	allocatedGPUs := float64(0)
	allocatedGPUMemory := float64(0)
	unhealthyGPUs := float64(0)
	for _, node := range nodes {
		nodeInfo := node.Convert2NodeInfo().(types.QGPUNodeInfo)
		totalGPUs += nodeInfo.TotalGPUs
		totalGPUMemory += nodeInfo.TotalGPUMemory
		allocatedGPUs += nodeInfo.AllocatedGPUs
		allocatedGPUMemory += nodeInfo.AllocatedGPUMemory
		unhealthyGPUs += nodeInfo.UnhealthyGPUs
		PrintLine(w, node.WideFormat())
	}
}

func displayQGPUNodeSummary(w *tabwriter.Writer, nodes []Node, isUnhealthy, showNodeType bool) (float64, float64, float64) {
	totalGPUs := float64(0)
	allocatedGPUs := float64(0)
	unhealthyGPUs := float64(0)
	for _, node := range nodes {
		nodeInfo := node.Convert2NodeInfo().(types.QGPUNodeInfo)
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
				if typeInfo.Name == types.QGPUNode {
					items = append(items, typeInfo.Alias)
				}
			}
		}
		if isUnhealthy && nodeInfo.TotalGPUs != 0 {
			items = append(items, fmt.Sprintf("%v", nodeInfo.UnhealthyGPUs))
		}
		PrintLine(w, items...)
	}
	return totalGPUs, allocatedGPUs, unhealthyGPUs
}

func displayQGPUNodesCustomSummary(w *tabwriter.Writer, nodes []Node) {
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
		nodeInfo := node.Convert2NodeInfo().(types.QGPUNodeInfo)
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
		items = append(items, fmt.Sprintf("%v/%v", nodeInfo.AllocatedGPUs, nodeInfo.TotalGPUs))
		items = append(items, fmt.Sprintf("%.1f/%.1f GiB",
			utils.DataUnitTransfer("bytes", "GiB", nodeInfo.AllocatedGPUMemory),
			utils.DataUnitTransfer("bytes", "GiB", nodeInfo.TotalGPUMemory)))
		if isUnhealthy {
			items = append(items, fmt.Sprintf("%v", nodeInfo.UnhealthyGPUs))
		}
		PrintLine(w, items...)
	}
	PrintLine(w, "-----------------------------------------------------------------------------------------------------------")
	// 1. print the utilization of gpus
	PrintLine(w, "Allocated/Total GPUs of nodes which own resource tke.cloud.tencent.com/qgpu-memory In Cluster:")
	allocatedPercent := float64(0)
	if totalGPUs != 0 {
		allocatedPercent = float64(allocatedGPUs) / float64(totalGPUs) * 100
	}
	PrintLine(w, fmt.Sprintf("%v/%v (%.1f%%)", allocatedGPUs, totalGPUs, allocatedPercent))
	// 2. print the utilization of gpu memory
	PrintLine(w, "Allocated/Total GPU Memory of nodes which own resource tke.cloud.tencent.com/qgpu-memory In Cluster:")
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
		PrintLine(w, "Unhealthy/Total GPUs of nodes which own resource tke.cloud.tencent.com/qgpu-memory In Cluster:")
		PrintLine(w, fmt.Sprintf("%v/%v (%.1f%%)", unhealthyGPUs, totalGPUs, unhealthyPercent))
	}
}

func IsQGPUNode(node *v1.Node) bool {
	labels := strings.Split(types.QGPUNodeLabels, ",")
	for _, label := range labels {
		qgpuLey := strings.Split(label, "=")[0]
		qgpuVal := strings.Split(label, "=")[1]
		if val, ok := node.Labels[qgpuLey]; ok && val == qgpuVal {
			return true
		}
	}
	return false
}

func NewQGPUNodeProcesser() NodeProcesser {
	return &nodeProcesser{
		nodeType:                  types.QGPUNode,
		key:                       "qGPUNodes",
		builder:                   NewQGPUNode,
		canBuildNode:              IsQGPUNode,
		displayNodesDetails:       displayQGPUNodeDetails,
		displayNodesSummary:       displayQGPUNodeSummary,
		displayNodesCustomSummary: displayQGPUNodesCustomSummary,
	}
}

func (g *qgpu) getPodAllocation(pod *v1.Pod) map[string]float64 {
	gpuContainers := map[string][]string{}
	for k, v := range pod.Annotations {
		c := strings.Replace(k, types.QGPUIndexPrefix, "", 1)
		for _, i := range strings.Split(v, ",") {
			gpuContainers[i] = append(gpuContainers[i], c)
		}
	}

	gpuMemContainers := g.resourceInContainers(pod, types.QGPUMemoryResourceName)
	gpuCoreContainers := g.resourceInContainers(pod, types.QGPUCoreResourceName)

	result := map[string]float64{}
	for i, cons := range gpuContainers {
		for _, c := range cons {
			if gpuCoreContainers[c] < 100 {
				result[i] += gpuMemContainers[c]
			} else {
				result[i] = utils.DataUnitTransfer("bytes", "GiB", g.getTotalGPUMemoryOfDevice(i))
			}
		}
	}

	return result
}

func (g *qgpu) resourceInContainers(pod *v1.Pod, resourceName string) map[string]float64 {
	total := make(map[string]float64)
	containers := pod.Spec.Containers
	for _, container := range containers {
		if val, ok := container.Resources.Limits[v1.ResourceName(resourceName)]; ok && int(val.Value()) != 0 {
			total[container.Name] = float64(val.Value())
		}
	}
	return total
}

func (g *qgpu) qGPUMemoryInPod(pod *v1.Pod) float64 {
	gpuMemoryContainers := g.resourceInContainers(pod, types.QGPUMemoryResourceName)

	total := float64(0)
	for k, v := range g.resourceInContainers(pod, types.QGPUCoreResourceName) {
		if v < 100 {
			total += gpuMemoryContainers[k]
		} else {
			ids := strings.Split(pod.Annotations[fmt.Sprintf("%s-%s", types.QGPUIndexPrefix, k)], ",")
			for _, i := range ids {
				total += utils.DataUnitTransfer("bytes", "GiB", g.getTotalGPUMemoryOfDevice(i))
			}
		}
	}

	return total
}

func (g *qgpu) isQGPUPod(pod *v1.Pod) bool {
	_, yes := pod.Annotations[types.QGPUAllocationLabel]
	return yes
}
