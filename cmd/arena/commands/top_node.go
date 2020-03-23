// Copyright 2018 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package commands

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/kubeflow/arena/pkg/util"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/kubernetes"
)

var (
	showDetails bool
)

type NodeInfo struct {
	node v1.Node
	pods []v1.Pod
}

func NewTopNodeCommand() *cobra.Command {

	var command = &cobra.Command{
		Use:   "node",
		Short: "Display Resource (GPU) usage of nodes.",
		Run: func(cmd *cobra.Command, args []string) {
			client, err := util.GetClientSet()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			allPods, err = acquireAllActivePods(client)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			nd := newNodeDescriber(client, allPods)
			nodeInfos, err := nd.getAllNodeInfos()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			displayTopNode(nodeInfos)
		},
	}

	command.Flags().BoolVarP(&showDetails, "details", "d", false, "Display details")
	return command
}

type NodeDescriber struct {
	client  *kubernetes.Clientset
	allPods []v1.Pod
}

func newNodeDescriber(client *kubernetes.Clientset, pods []v1.Pod) *NodeDescriber {
	return &NodeDescriber{
		client:  client,
		allPods: pods,
	}
}

func (d *NodeDescriber) getAllNodeInfos() ([]NodeInfo, error) {
	nodeInfoList := []NodeInfo{}

	nodeList, err := d.client.CoreV1().Nodes().List(metav1.ListOptions{})

	if err != nil {
		return nodeInfoList, err
	}

	for _, node := range nodeList.Items {

		pods := d.getPodsFromNode(node)
		nodeInfo := NodeInfo{
			node: node,
			pods: pods,
		}
		nodeInfoList = append(nodeInfoList, nodeInfo)
	}
	return nodeInfoList, nil
}

func (d *NodeDescriber) getPodsFromNode(node v1.Node) []v1.Pod {
	pods := []v1.Pod{}
	for _, pod := range d.allPods {
		if pod.Spec.NodeName == node.Name {
			pods = append(pods, pod)
		}
	}

	return pods
}

func displayTopNode(nodes []NodeInfo) {
	if showDetails {
		displayTopNodeDetails(nodes)
	} else {
		displayTopNodeSummary(nodes)
	}
}
func displayTopNodeSummary(nodeInfos []NodeInfo) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	var (
		totalGPUsInCluster            int64
		totalUnhealthyGPUsInCluster   int64
		allocatedGPUsInCluster        int64
		totalGPUsOnReadyNodeInCluster int64
		hasUnhealthyGPUNode           bool
	)

	for _, NodeInfo := range nodeInfos {
		if hasUnhealthyGPU(NodeInfo) {
			hasUnhealthyGPUNode = true
			break
		}
	}

	if hasUnhealthyGPUNode {
		fmt.Fprintf(w, "NAME\tIPADDRESS\tROLE\tSTATUS\tGPU(Total)\tGPU(Allocated)\tGPU(Unhealthy)\n")
	} else {
		fmt.Fprintf(w, "NAME\tIPADDRESS\tROLE\tSTATUS\tGPU(Total)\tGPU(Allocated)\n")
	}

	for _, nodeInfo := range nodeInfos {
		// Skip NotReady node
		//if ! isNodeReady(nodeInfo.node) {
		//	continue
		//}
		var totalGPU int64
		var allocatableGPU int64
		var allocatedGPU int64
		totalGPU, allocatableGPU, allocatedGPU = calculateNodeGPU(nodeInfo)

		totalGPUsInCluster += totalGPU
		allocatedGPUsInCluster += allocatedGPU
		unhealthGPU := totalGPU - allocatableGPU
		totalUnhealthyGPUsInCluster += unhealthGPU

		address := getNodeInternalAddress(nodeInfo.node)

		role := strings.Join(findNodeRoles(&nodeInfo.node), ",")
		if len(role) == 0 {
			role = "<none>"
		}

		status := "ready"
		if !isNodeReady(nodeInfo.node) {
			status = "notReady"
		} else {
			totalGPUsOnReadyNodeInCluster += totalGPU
		}

		if hasUnhealthyGPUNode {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n", nodeInfo.node.Name,
				address,
				role,
				status,
				strconv.FormatInt(totalGPU, 10),
				strconv.FormatInt(allocatedGPU, 10),
				strconv.FormatInt(unhealthGPU, 10))
		} else {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", nodeInfo.node.Name,
				address,
				role,
				status,
				strconv.FormatInt(totalGPU, 10),
				strconv.FormatInt(allocatedGPU, 10))

		}
	}

	if hasUnhealthyGPUNode {
		fmt.Fprintf(w, "---------------------------------------------------------------------------------------------------\n")

	} else {
		fmt.Fprintf(w, "-----------------------------------------------------------------------------------------\n")
	}
	fmt.Fprintf(w, "Allocated/Total GPUs In Cluster:\n")
	log.Debugf("gpu: %s, allocated GPUs %s", strconv.FormatInt(totalGPUsInCluster, 10),
		strconv.FormatInt(allocatedGPUsInCluster, 10))
	var gpuUsage float64 = 0
	if totalGPUsInCluster > 0 {
		gpuUsage = float64(allocatedGPUsInCluster) / float64(totalGPUsInCluster) * 100
	}
	fmt.Fprintf(w, "%s/%s (%d%%)\t\n",
		strconv.FormatInt(allocatedGPUsInCluster, 10),
		strconv.FormatInt(totalGPUsInCluster, 10),
		int64(gpuUsage))
	if totalGPUsInCluster != totalGPUsOnReadyNodeInCluster {
		if totalGPUsOnReadyNodeInCluster > 0 {
			gpuUsage = float64(allocatedGPUsInCluster) / float64(totalGPUsOnReadyNodeInCluster) * 100
		} else {
			gpuUsage = 0
		}
		fmt.Fprintf(w, "Allocated/Total GPUs(Active) In Cluster:\n")
		fmt.Fprintf(w, "%s/%s (%d%%)\t\n",
			strconv.FormatInt(allocatedGPUsInCluster, 10),
			strconv.FormatInt(totalGPUsOnReadyNodeInCluster, 10),
			int64(gpuUsage))
	}

	if hasUnhealthyGPUNode {
		fmt.Fprintf(w, "Unhealthy/Total GPUs In Cluster:\n")
		var gpuUnhealthyPercentage float64 = 0
		if totalGPUsInCluster > 0 {
			gpuUnhealthyPercentage = float64(totalUnhealthyGPUsInCluster) / float64(totalGPUsInCluster) * 100
		}
		fmt.Fprintf(w, "%s/%s (%d%%)\t\n",
			strconv.FormatInt(totalUnhealthyGPUsInCluster, 10),
			strconv.FormatInt(totalGPUsInCluster, 10),
			int64(gpuUnhealthyPercentage))
	}

	// fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", ...)

	_ = w.Flush()
}

func displayTopNodeDetails(nodeInfos []NodeInfo) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	var (
		totalGPUsInCluster          int64
		totalUnhealthyGPUsInCluster int64
		allocatedGPUsInCluster      int64
		hasUnhealthyGPUNode         bool
	)

	for _, NodeInfo := range nodeInfos {
		if hasUnhealthyGPU(NodeInfo) {
			hasUnhealthyGPUNode = true
			break
		}
	}

	fmt.Fprintf(w, "\n")
	for _, nodeInfo := range nodeInfos {
		// Skip NotReady node
		//if ! isNodeReady(nodeInfo.node) {
		//	continue
		//}
		var totalGPU int64
		var allocatableGPU int64
		var allocatedGPU int64
		totalGPU, allocatableGPU, allocatedGPU = calculateNodeGPU(nodeInfo)

		totalGPUsInCluster += totalGPU
		allocatedGPUsInCluster += allocatedGPU
		unhealthyGPUs := totalGPU - allocatableGPU
		totalUnhealthyGPUsInCluster += unhealthyGPUs

		address := getNodeInternalAddress(nodeInfo.node)

		role := strings.Join(findNodeRoles(&nodeInfo.node), ",")
		if len(role) == 0 {
			role = "<none>"
		}

		fmt.Fprintf(w, "\n")
		fmt.Fprintf(w, "NAME:\t%s\n", nodeInfo.node.Name)
		fmt.Fprintf(w, "IPADDRESS:\t%s\n", address)
		fmt.Fprintf(w, "ROLE:\t%s\n", role)

		pods := gpuPods(nodeInfo.pods)
		if len(pods) > 0 {
			fmt.Fprintf(w, "\n")
			fmt.Fprintf(w, "NAMESPACE\tNAME\tGPU REQUESTS\t \n")
			for _, pod := range pods {
				fmt.Fprintf(w, "%s\t%s\t%s\t\n", pod.Namespace,
					pod.Name,
					strconv.FormatInt(gpuInPod(pod), 10))
			}
			fmt.Fprintf(w, "\n")
		}

		var gpuUsageInNode float64 = 0
		if totalGPU > 0 {
			gpuUsageInNode = float64(allocatedGPU) / float64(totalGPU) * 100
		} else {
			fmt.Fprintf(w, "\n")
		}

		var gpuUnhealthyPercentageInNode float64 = 0
		if totalGPU > 0 {
			gpuUnhealthyPercentageInNode = float64(unhealthyGPUs) / float64(totalGPU) * 100
		}

		fmt.Fprintf(w, "Total GPUs In Node %s:\t%s \t\n", nodeInfo.node.Name, strconv.FormatInt(totalGPU, 10))
		fmt.Fprintf(w, "Allocated GPUs In Node %s:\t%s (%d%%)\t\n", nodeInfo.node.Name, strconv.FormatInt(allocatedGPU, 10), int64(gpuUsageInNode))
		if hasUnhealthyGPUNode {
			fmt.Fprintf(w, "Unhealthy GPUs In Node %s:\t%s (%d%%)\t\n", nodeInfo.node.Name, strconv.FormatInt(unhealthyGPUs, 10), int64(gpuUnhealthyPercentageInNode))

		}
		log.Debugf("gpu: %s, allocated GPUs %s", strconv.FormatInt(totalGPU, 10),
			strconv.FormatInt(allocatedGPU, 10))

		fmt.Fprintf(w, "-----------------------------------------------------------------------------------------\n")
	}
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "Allocated/Total GPUs In Cluster:\t")
	log.Debugf("gpu: %s, allocated GPUs %s", strconv.FormatInt(totalGPUsInCluster, 10),
		strconv.FormatInt(allocatedGPUsInCluster, 10))

	var gpuUsage float64 = 0
	if totalGPUsInCluster > 0 {
		gpuUsage = float64(allocatedGPUsInCluster) / float64(totalGPUsInCluster) * 100
	}
	fmt.Fprintf(w, "%s/%s (%d%%)\t\n",
		strconv.FormatInt(allocatedGPUsInCluster, 10),
		strconv.FormatInt(totalGPUsInCluster, 10),
		int64(gpuUsage))
	// fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", ...)
	if hasUnhealthyGPUNode {
		fmt.Fprintf(w, "Unhealthy/Total GPUs In Cluster:\t")
		var gpuUnhealthyPercentage float64 = 0
		if totalGPUsInCluster > 0 {
			gpuUnhealthyPercentage = float64(totalUnhealthyGPUsInCluster) / float64(totalGPUsInCluster) * 100
		}
		fmt.Fprintf(w, "%s/%s (%d%%)\t\n",
			strconv.FormatInt(totalUnhealthyGPUsInCluster, 10),
			strconv.FormatInt(totalGPUsInCluster, 10),
			int64(gpuUnhealthyPercentage))
	}

	_ = w.Flush()
}

// calculate the GPU count of each node
func calculateNodeGPU(nodeInfo NodeInfo) (totalGPU, allocatableGPU, allocatedGPU int64) {
	node := nodeInfo.node
	totalGPU = totalGpuInNode(node)
	allocatableGPU = allocatableGpuInNode(node)
	// allocatedGPU = gpuInPod()

	for _, pod := range nodeInfo.pods {
		allocatedGPU += gpuInPod(pod)
	}

	fractionalGPUsUsedInNode := sharedGPUsUsedInNode(nodeInfo)
	allocatedGPU += fractionalGPUsUsedInNode
	totalGPU += fractionalGPUsUsedInNode

	return totalGPU, allocatableGPU, allocatedGPU
}

// Does the node have unhealthy GPU
func hasUnhealthyGPU(nodeInfo NodeInfo) (unhealthy bool) {
	node := nodeInfo.node
	totalGPU := totalGpuInNode(node)
	allocatableGPU := allocatableGpuInNode(node)

	unhealthy = totalGPU > allocatableGPU

	if unhealthy {
		log.Debugf("node: %s, allocated GPUs %s, total GPUs %s is unhealthy", nodeInfo.node.Name, strconv.FormatInt(totalGPU, 10),
			strconv.FormatInt(allocatableGPU, 10))
	}

	return unhealthy
}

func isMasterNode(node v1.Node) bool {
	if _, ok := node.Labels[masterLabelRole]; ok {
		return true
	}

	return false
}

func (nodeInfo NodeInfo) isGPUExclusiveNode() bool {
	value, ok := nodeInfo.node.Status.Allocatable[NVIDIAGPUResourceName]

	if ok {
		ok = (int(value.Value()) > 0)
	}

	return ok
}

// findNodeRoles returns the roles of a given node.
// The roles are determined by looking for:
// * a node-role.kubernetes.io/<role>="" label
// * a kubernetes.io/role="<role>" label
func findNodeRoles(node *v1.Node) []string {
	roles := sets.NewString()
	for k, v := range node.Labels {
		switch {
		case strings.HasPrefix(k, labelNodeRolePrefix):
			if role := strings.TrimPrefix(k, labelNodeRolePrefix); len(role) > 0 {
				roles.Insert(role)
			}

		case k == nodeLabelRole && v != "":
			roles.Insert(v)
		}
	}
	return roles.List()
}

func isNodeReady(node v1.Node) bool {
	for _, condition := range node.Status.Conditions {
		if condition.Type == v1.NodeReady && condition.Status == v1.ConditionTrue {
			return true
		}
	}
	return false
}

func getNodeInternalAddress(node v1.Node) string {
	address := "unknown"
	if len(node.Status.Addresses) > 0 {
		//address = nodeInfo.node.Status.Addresses[0].Address
		for _, addr := range node.Status.Addresses {
			if addr.Type == v1.NodeInternalIP {
				address = addr.Address
			}
		}
	}
	return address
}
