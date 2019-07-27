package commands

import (
	"fmt"
	"strconv"

	log "github.com/golang/glog"
	"k8s.io/api/core/v1"
)

type ShareNodeInfo struct {
	pods           []v1.Pod
	node           v1.Node
	devs           map[int]*DeviceInfo
	gpuCount       int
	gpuTotalMemory int
}

type DeviceInfo struct {
	idx         int
	pods        []v1.Pod
	usedGPUMem  int
	totalGPUMem int
	node        v1.Node
}

func (d *DeviceInfo) String() string {
	if d.idx == -1 {
		return fmt.Sprintf("%d", d.usedGPUMem)
	}
	return fmt.Sprintf("%d/%d", d.usedGPUMem, d.totalGPUMem)
}

//For all GPUShare nodes,decide whether the memory of GPU is measured by MiB or GiB
func buildAllShareNodeInfos(allPods []v1.Pod, nodes []v1.Node) ([]*ShareNodeInfo, error) {
	SharenodeInfos := buildShareNodeInfosWithPods(allPods, nodes)
	for _, SharenodeInfo := range SharenodeInfos {
		if SharenodeInfo.gpuTotalMemory > 0 {
			setUnit(SharenodeInfo.gpuTotalMemory, SharenodeInfo.gpuCount)
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
func buildShareNodeInfo(allPods []v1.Pod, node v1.Node) (*ShareNodeInfo, error) {
	SharenodeInfo := buildShareNodeInfoWithPods(allPods, node)

	if SharenodeInfo.gpuTotalMemory > 0 {
		setUnit(SharenodeInfo.gpuTotalMemory, SharenodeInfo.gpuCount)
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
			info.gpuCount = getGPUCountInNode(node)
			info.gpuTotalMemory = getTotalGPUMemory(node)
			info.devs = map[int]*DeviceInfo{}

			for i := 0; i < info.gpuCount; i++ {
				dev := &DeviceInfo{
					pods:        []v1.Pod{},
					idx:         i,
					totalGPUMem: info.gpuTotalMemory / info.gpuCount,
					node:        info.node,
				}
				info.devs[i] = dev
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
	info.gpuCount = getGPUCountInNode(node)
	info.gpuTotalMemory = getTotalGPUMemory(node)
	info.devs = map[int]*DeviceInfo{}

	for i := 0; i < info.gpuCount; i++ {
		dev := &DeviceInfo{
			pods:        []v1.Pod{},
			idx:         i,
			totalGPUMem: info.gpuTotalMemory / info.gpuCount,
			node:        info.node,
		}
		info.devs[i] = dev
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
		if dev, ok = n.devs[devID]; !ok {
			totalGPUMem := 0
			if n.gpuCount > 0 {
				totalGPUMem = n.gpuTotalMemory / n.gpuCount
			}

			dev = &DeviceInfo{
				pods:        []v1.Pod{},
				idx:         devID,
				totalGPUMem: totalGPUMem,
				node:        n.node,
			}
			n.devs[devID] = dev
		}

		dev.usedGPUMem = dev.usedGPUMem + usedGPUMem
		dev.pods = append(dev.pods, pod)
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
	_, found := n.devs[-1]
	return found
}

func isGPUSharingNode(node v1.Node) bool {
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
