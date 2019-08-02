package types

import (
	"bytes"
	"fmt"
	log "github.com/golang/glog"
	"k8s.io/api/core/v1"
	"os"
	"strconv"
	"text/tabwriter"
)

func DisplayShareDetails(nodeInfos []*ShareNodeInfo) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	var (
		totalGPUMemInCluster int64
		usedGPUMemInCluster  int64
		prtLineLen           int
	)

	for _, nodeInfo := range nodeInfos {
		address := "unknown"
		if len(nodeInfo.node.Status.Addresses) > 0 {
			//address = nodeInfo.node.Status.Addresses[0].Address
			for _, addr := range nodeInfo.node.Status.Addresses {
				if addr.Type == v1.NodeInternalIP {
					address = addr.Address
					break
				}
			}
		}

		totalGPUMemInNode := nodeInfo.gpuTotalMemory
		if totalGPUMemInNode <= 0 {
			continue
		}

		fmt.Fprintf(w, "\n")
		fmt.Fprintf(w, "NAME:\t%s\n", nodeInfo.node.Name)
		fmt.Fprintf(w, "IPADDRESS:\t%s\n", address)
		fmt.Fprintf(w, "\n")

		usedGPUMemInNode := 0
		var buf bytes.Buffer
		buf.WriteString("NAME\tNAMESPACE\t")
		for i := 0; i < nodeInfo.GpuCount; i++ {
			buf.WriteString(fmt.Sprintf("GPU%d(Allocated)\t", i))
		}

		buf.WriteString("\n")
		fmt.Fprintf(w, buf.String())

		var buffer bytes.Buffer
		for i, dev := range nodeInfo.Devs {
			usedGPUMemInNode += dev.UsedGPUMem
			for _, pod := range dev.Pods {

				buffer.WriteString(fmt.Sprintf("%s\t%s\t", pod.Name, pod.Namespace))
				count := nodeInfo.GpuCount

				for k := 0; k < count; k++ {
					if k == i || (i == -1 && k == nodeInfo.GpuCount) {
						buffer.WriteString(fmt.Sprintf("%d\t", GetGPUMemoryInPod(pod)))
					} else {
						buffer.WriteString("0\t")
					}
				}
				buffer.WriteString("\n")
			}
		}
		if prtLineLen == 0 {
			prtLineLen = buffer.Len() + 10
		}
		fmt.Fprintf(w, buffer.String())

		var gpuUsageInNode float64 = 0
		if totalGPUMemInNode > 0 {
			gpuUsageInNode = float64(usedGPUMemInNode) / float64(totalGPUMemInNode) * 100
		} else {
			fmt.Fprintf(w, "\n")
		}

		fmt.Fprintf(w, "Allocated :\t%d (%d%%)\t\n", usedGPUMemInNode, int64(gpuUsageInNode))
		fmt.Fprintf(w, "Total :\t%d \t\n", nodeInfo.gpuTotalMemory)
		// fmt.Fprintf(w, "-----------------------------------------------------------------------------------------\n")
		var prtLine bytes.Buffer
		for i := 0; i < prtLineLen; i++ {
			prtLine.WriteString("-")
		}
		prtLine.WriteString("\n")
		fmt.Fprintf(w, prtLine.String())
		totalGPUMemInCluster += int64(totalGPUMemInNode)
		usedGPUMemInCluster += int64(usedGPUMemInNode)
	}
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "Allocated/Total GPU Memory In GPUShare Node:\n")
	log.V(2).Infof("gpu: %s, allocated GPU Memory %s", strconv.FormatInt(totalGPUMemInCluster, 10),
		strconv.FormatInt(usedGPUMemInCluster, 10))

	var gpuUsage float64 = 0
	if totalGPUMemInCluster > 0 {
		gpuUsage = float64(usedGPUMemInCluster) / float64(totalGPUMemInCluster) * 100
	}
	fmt.Fprintf(w, "%s/%s (%s) (%d%%)\t\n",
		strconv.FormatInt(usedGPUMemInCluster, 10),
		strconv.FormatInt(totalGPUMemInCluster, 10),
		memoryUnit,
		int64(gpuUsage))
	// fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", ...)

	_ = w.Flush()
}

func DisplayShareSummary(nodeInfos []*ShareNodeInfo) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	var (
		maxGPUCount          int
		totalGPUMemInCluster int64
		usedGPUMemInCluster  int64
		prtLineLen           int
	)

	maxGPUCount = getMaxGPUCount(nodeInfos)

	var buffer bytes.Buffer
	buffer.WriteString("NAME\tIPADDRESS\t")
	for i := 0; i < maxGPUCount; i++ {
		buffer.WriteString(fmt.Sprintf("GPU%d(Allocated/Total)\t", i))
	}

	buffer.WriteString(fmt.Sprintf("\n"))

	fmt.Fprintf(w, buffer.String())
	for _, nodeInfo := range nodeInfos {
		address := "unknown"
		if len(nodeInfo.node.Status.Addresses) > 0 {
			// address = nodeInfo.node.Status.Addresses[0].Address
			for _, addr := range nodeInfo.node.Status.Addresses {
				if addr.Type == v1.NodeInternalIP {
					address = addr.Address
					break
				}
			}
		}

		gpuMemInfos := []string{}

		usedGPUMemInNode := 0
		totalGPUMemInNode := nodeInfo.gpuTotalMemory
		if totalGPUMemInNode <= 0 {
			continue
		}

		for i := 0; i < maxGPUCount; i++ {
			gpuMemInfo := "0/0"
			if dev, ok := nodeInfo.Devs[i]; ok {
				gpuMemInfo = dev.String()
				usedGPUMemInNode += dev.UsedGPUMem
			}
			gpuMemInfos = append(gpuMemInfos, gpuMemInfo)
		}

		var buf bytes.Buffer
		buf.WriteString(fmt.Sprintf("%s\t%s\t", nodeInfo.node.Name, address))
		for i := 0; i < maxGPUCount; i++ {
			buf.WriteString(fmt.Sprintf("%s\t", gpuMemInfos[i]))
		}

		buf.WriteString(fmt.Sprintf("\n"))
		fmt.Fprintf(w, buf.String())

		if prtLineLen == 0 {
			prtLineLen = buf.Len() + 20
		}

		usedGPUMemInCluster += int64(usedGPUMemInNode)
		totalGPUMemInCluster += int64(totalGPUMemInNode)
	}
	// fmt.Fprintf(w, "-----------------------------------------------------------------------------------------\n")
	var prtLine bytes.Buffer
	for i := 0; i < prtLineLen; i++ {
		prtLine.WriteString("-")
	}
	prtLine.WriteString("\n")
	fmt.Fprint(w, prtLine.String())

	fmt.Fprintf(w, "Allocated/Total GPU Memory In GPUShare Node:\n")
	log.V(2).Infof("gpu: %s, allocated GPU Memory %s", strconv.FormatInt(totalGPUMemInCluster, 10),
		strconv.FormatInt(usedGPUMemInCluster, 10))
	var gpuUsage float64 = 0
	if totalGPUMemInCluster > 0 {
		gpuUsage = float64(usedGPUMemInCluster) / float64(totalGPUMemInCluster) * 100
	}
	fmt.Fprintf(w, "%s/%s (%s) (%d%%)\t\n",
		strconv.FormatInt(usedGPUMemInCluster, 10),
		strconv.FormatInt(totalGPUMemInCluster, 10),
		memoryUnit,
		int64(gpuUsage))
	// fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", ...)

	_ = w.Flush()
}

func GetGPUMemoryInPod(pod v1.Pod) int {
	gpuMem := 0
	for _, container := range pod.Spec.Containers {
		if val, ok := container.Resources.Limits[resourceName]; ok {
			gpuMem += int(val.Value())
		}
	}
	return gpuMem
}

func getMaxGPUCount(nodeInfos []*ShareNodeInfo) (max int) {
	for _, node := range nodeInfos {
		if node.GpuCount > max {
			max = node.GpuCount
		}
	}

	return max
}
