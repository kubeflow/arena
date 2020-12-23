package topnode

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/kubeflow/arena/pkg/apis/types"
	"gopkg.in/yaml.v2"
)

/*
format like:

NAME                       IPADDRESS      ROLE    STATUS  GPU(Total)  GPU(Allocated)  GPU_MODE
cn-shanghai.192.168.7.178  192.168.7.178  master  Ready   0           0               none
cn-shanghai.192.168.7.179  192.168.7.179  master  Ready   0           0               none
cn-shanghai.192.168.7.180  192.168.7.180  master  Ready   0           0               none
cn-shanghai.192.168.7.181  192.168.7.181  <none>  Ready   0           0               none
cn-shanghai.192.168.7.182  192.168.7.182  <none>  Ready   1           0               exclusive
cn-shanghai.192.168.7.186  192.168.7.186  <none>  Ready   4           0               topology
cn-shanghai.192.168.7.183  192.168.7.183  <none>  Ready   4           2.1             share
*/
func DisplayNodeSummary(nodeNames []string, targetNodeType types.NodeType, format types.FormatStyle) error {
	totalGPUs := 0
	allocatedGPUs := 0
	unhealthyGPUs := 0
	nodes, err := BuildNodes(nodeNames, targetNodeType)
	if err != nil {
		return err
	}
	allNodeInfos := types.AllNodeInfo{}
	for _, processer := range GetSupportedNodePorcessers() {
		allNodeInfos = processer.Convert2NodeInfos(nodes, allNodeInfos)
	}
	switch format {
	case types.JsonFormat:
		data, _ := json.MarshalIndent(allNodeInfos, "", "    ")
		fmt.Printf("%v", string(data))
		return nil
	case types.YamlFormat:
		data, _ := yaml.Marshal(allNodeInfos)
		fmt.Printf("%v", string(data))
		return nil
	}
	var showNodeType bool
	var isUnhealthy bool
	nodeTypes := map[types.NodeType]bool{}
	for _, node := range nodes {
		if node.Type() != types.NormalNode {
			nodeTypes[node.Type()] = true
		}
		if !node.AllDevicesAreHealthy() {
			isUnhealthy = true
		}
	}
	if len(nodeTypes) != 1 {
		showNodeType = true
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	header := []string{"NAME", "IPADDRESS", "ROLE", "STATUS", "GPU(Total)", "GPU(Allocated)"}
	if showNodeType {
		header = append(header, "GPU(Mode)")
	}
	if isUnhealthy {
		header = append(header, "UNHEALTHY")
	}
	PrintLine(w, header...)
	processers := GetSupportedNodePorcessers()
	for i := len(processers) - 1; i >= 0; i-- {
		processer := processers[i]
		t, a, u := processer.DisplayNodesSummary(w, nodes, showNodeType, isUnhealthy)
		totalGPUs += t
		allocatedGPUs += a
		unhealthyGPUs += u
	}
	PrintLine(w, "---------------------------------------------------------------------------------------------------")
	PrintLine(w, "Allocated/Total GPUs In Cluster:")
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
		PrintLine(w, "Unhealthy/Total GPUs In Cluster:")
		PrintLine(w, fmt.Sprintf("%v/%v (%.1f%%)", unhealthyGPUs, totalGPUs, unhealthyPercent))
	}
	_ = w.Flush()
	return nil
}
