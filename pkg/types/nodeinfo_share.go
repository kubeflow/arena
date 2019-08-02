package types

import (
	"fmt"
	"strconv"

	log "github.com/golang/glog"
	"k8s.io/api/core/v1"
)

const (
	resourceName  = "aliyun.com/gpu-mem"
	resourceCount = "aliyun.com/gpu-count"
	envNVGPUID    = "ALIYUN_COM_GPU_MEM_IDX"
)

type ShareNodeInfo struct {
	pods           []v1.Pod
	node           v1.Node
	Devs           map[int]*DeviceInfo
	GpuCount       int
	gpuTotalMemory int
}

type DeviceInfo struct {
	idx         int
	Pods        []v1.Pod
	UsedGPUMem  int
	TotalGPUMem int
	node        v1.Node
}

func (d *DeviceInfo) String() string {
	if d.idx == -1 {
		return fmt.Sprintf("%d", d.UsedGPUMem)
	}
	return fmt.Sprintf("%d/%d", d.UsedGPUMem, d.TotalGPUMem)
}

//For all GPUShare nodes,decide whether the memory of GPU is measured by MiB or GiB
func BuildAllShareNodeInfos(allPods []v1.Pod, nodes []v1.Node) ([]*ShareNodeInfo, error) {
	SharenodeInfos := buildShareNodeInfosWithPods(allPods, nodes)
	for _, SharenodeInfo := range SharenodeInfos {
		if SharenodeInfo.gpuTotalMemory > 0 {
			setUnit(SharenodeInfo.gpuTotalMemory, SharenodeInfo.GpuCount)
			err := SharenodeInfo.buildDeviceInfo()
			if err != nil {
				log.Warningf("Failed due to %v", err)
				continue
			}
		}
	}
	return SharenodeInfos, nil
}

//For one GPUShare node,decide whether the memory of GPU is measured by MiB or GiB
func BuildShareNodeInfo(allPods []v1.Pod, node v1.Node) (*ShareNodeInfo, error) {
	SharenodeInfo := buildShareNodeInfoWithPods(allPods, node)

	if SharenodeInfo.gpuTotalMemory > 0 {
		setUnit(SharenodeInfo.gpuTotalMemory, SharenodeInfo.GpuCount)
		err := SharenodeInfo.buildDeviceInfo()
		if err != nil {
			log.Warningf("Failed due to %v", err)
		}
	}

	return SharenodeInfo, nil
}

//Create  ShareNodeInfos for all gpushare nodes
func buildShareNodeInfosWithPods(pods []v1.Pod, nodes []v1.Node) []*ShareNodeInfo {
	nodeMap := map[string]*ShareNodeInfo{}
	nodeList := []*ShareNodeInfo{}

	for _, node := range nodes {
		var info *ShareNodeInfo = &ShareNodeInfo{}
		if value, ok := nodeMap[node.Name]; ok {
			info = value
		} else {
			nodeMap[node.Name] = info
			info.node = node
			info.pods = []v1.Pod{}
			info.GpuCount = getGPUCountInNode(node)
			info.gpuTotalMemory = getTotalGPUMemory(node)
			info.Devs = map[int]*DeviceInfo{}

			for i := 0; i < info.GpuCount; i++ {
				dev := &DeviceInfo{
					Pods:        []v1.Pod{},
					idx:         i,
					TotalGPUMem: info.gpuTotalMemory / info.GpuCount,
					node:        info.node,
				}
				info.Devs[i] = dev
			}

		}

		for _, pod := range pods {
			if pod.Spec.NodeName == node.Name {
				info.pods = append(info.pods, pod)
			}
		}
	}

	for _, v := range nodeMap {
		nodeList = append(nodeList, v)
	}
	return nodeList
}

//Create  ShareNodeInfo for one node
func buildShareNodeInfoWithPods(pods []v1.Pod, node v1.Node) *ShareNodeInfo {

	var info *ShareNodeInfo = &ShareNodeInfo{}
	info.node = node
	info.pods = []v1.Pod{}
	info.GpuCount = getGPUCountInNode(node)
	info.gpuTotalMemory = getTotalGPUMemory(node)
	info.Devs = map[int]*DeviceInfo{}

	for i := 0; i < info.GpuCount; i++ {
		dev := &DeviceInfo{
			Pods:        []v1.Pod{},
			idx:         i,
			TotalGPUMem: info.gpuTotalMemory / info.GpuCount,
			node:        info.node,
		}
		info.Devs[i] = dev
	}

	for _, pod := range pods {
		if pod.Spec.NodeName == node.Name {
			info.pods = append(info.pods, pod)
		}
	}

	return info
}

func getTotalGPUMemory(node v1.Node) int {
	val, ok := node.Status.Allocatable[resourceName]

	if !ok {
		return 0
	}

	return int(val.Value())
}

func getGPUCountInNode(node v1.Node) int {
	val, ok := node.Status.Allocatable[resourceCount]

	if !ok {
		return 0
	}

	return int(val.Value())
}

func gpuMemoryInPod(pod v1.Pod) int {
	var total int
	containers := pod.Spec.Containers
	for _, container := range containers {
		if val, ok := container.Resources.Limits[resourceName]; ok {
			total += int(val.Value())
		}
	}

	return total
}

// Get Deviceinfo of ShareNodeinfo
func (n *ShareNodeInfo) buildDeviceInfo() error {

GPUSearchLoop:
	for _, pod := range n.pods {
		if gpuMemoryInPod(pod) <= 0 {
			continue GPUSearchLoop
		}

		devID, usedGPUMem := n.getDeivceInfo(pod)

		var dev *DeviceInfo
		ok := false
		if dev, ok = n.Devs[devID]; !ok {
			totalGPUMem := 0
			if n.GpuCount > 0 {
				totalGPUMem = n.gpuTotalMemory / n.GpuCount
			}

			dev = &DeviceInfo{
				Pods:        []v1.Pod{},
				idx:         devID,
				TotalGPUMem: totalGPUMem,
				node:        n.node,
			}
			n.Devs[devID] = dev
		}

		dev.UsedGPUMem = dev.UsedGPUMem + usedGPUMem
		dev.Pods = append(dev.Pods, pod)
	}

	return nil
}

func (n *ShareNodeInfo) getDeivceInfo(pod v1.Pod) (devIdx int, gpuMemory int) {
	var err error
	id := -1

	if len(pod.ObjectMeta.Annotations) > 0 {
		value, found := pod.ObjectMeta.Annotations[envNVGPUID]
		if found {
			id, err = strconv.Atoi(value)
			if err != nil {
				log.Warningf("Failed to parse dev id %s due to %v for pod %s in ns %s",
					value,
					err,
					pod.Name,
					pod.Namespace)
				id = -1
			}
		} else {
			log.Warningf("Failed to get dev id %s for pod %s in ns %s",
				pod.Name,
				pod.Namespace)
		}
	}

	return id, gpuMemoryInPod(pod)
}

func hasPendingGPUMemory(nodeInfos []*ShareNodeInfo) (found bool) {
	for _, info := range nodeInfos {
		if info.hasPendingGPUMemory() {
			return true
		}
	}

	return false
}

func (n *ShareNodeInfo) hasPendingGPUMemory() bool {
	_, found := n.Devs[-1]
	return found
}

func IsGPUSharingNode(node v1.Node) bool {
	value, ok := node.Status.Allocatable[resourceName]

	if ok {
		ok = (int(value.Value()) > 0)
	}

	return ok
}

var (
	memoryUnit = ""
)

func setUnit(gpuMemory, gpuCount int) {
	if memoryUnit != "" {
		return
	}

	if gpuCount == 0 {
		return
	}

	gpuMemoryByDev := gpuMemory / gpuCount

	if gpuMemoryByDev > 100 {
		memoryUnit = "MiB"
	} else {
		memoryUnit = "GiB"
	}
}
