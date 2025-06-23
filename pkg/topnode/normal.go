// Copyright 2024 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package topnode

import (
	"fmt"
	"strings"
	"text/tabwriter"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/kubeflow/arena/pkg/apis/types"
)

var NormalNodeDescription = `
  This node has none gpu devices
`

var normalNodeTemplate = `
Name:    %v
Status:  %v
Role:    %v
Type:    %v
Address: %v
Description:
%v
`

type normalNode struct {
	node *corev1.Node
	pods []*corev1.Pod
	baseNode
}

func NewNormalNode(client *kubernetes.Clientset, node *corev1.Node, index int, args buildNodeArgs) (Node, error) {
	return &normalNode{
		node: node,
		pods: []*corev1.Pod{},
		baseNode: baseNode{
			index:    index,
			node:     node,
			pods:     []*corev1.Pod{},
			nodeType: types.NormalNode,
		},
	}, nil
}

func (n *normalNode) AllDevicesAreHealthy() bool {
	return true
}

func (n *normalNode) convert2NodeInfo() types.NormalNodeInfo {
	role := strings.Join(n.Role(), ",")
	return types.NormalNodeInfo{
		CommonNodeInfo: types.CommonNodeInfo{
			Name:        n.Name(),
			IP:          n.IP(),
			Status:      n.Status(),
			Role:        role,
			Type:        types.NormalNode,
			Description: NormalNodeDescription,
		},
	}
}

func (n *normalNode) Convert2NodeInfo() interface{} {
	return n.convert2NodeInfo()
}

func (n *normalNode) WideFormat() string {
	role := strings.Join(n.Role(), ",")
	if role == "" {
		role = "<none>"
	}
	nodeInfo := n.convert2NodeInfo()
	return fmt.Sprintf(normalNodeTemplate,
		n.Name(),
		n.Status(),
		role,
		n.Type(),
		n.IP(),
		strings.Trim(nodeInfo.Description, "\n"),
	)
}

/*
display like:

===================================== NormalNode ========================================

Name:    cn-shanghai.192.168.7.186
Status:  Ready
Role:    <none>
Type:    Normal
Address: 192.168.7.186
-----------------------------------------------------------------------------------------

===================================== End ===============================================
*/
func displayNormalNodeDetails(w *tabwriter.Writer, nodes []Node) {
	if len(nodes) == 0 {
		return
	}
	for _, node := range nodes {
		PrintLine(w, node.WideFormat())
	}
}

func displayNormalNodeSummary(w *tabwriter.Writer, nodes []Node, isUnhealthy, showMode bool) (float64, float64, float64) {
	for _, node := range nodes {
		nodeInfo := node.Convert2NodeInfo().(types.NormalNodeInfo)
		items := []string{}
		items = append(items, node.Name())
		items = append(items, node.IP())
		role := nodeInfo.Role
		if role == "" {
			role = "<none>"
		}
		items = append(items, role)
		items = append(items, node.Status())
		items = append(items, "0")
		items = append(items, "0")
		if showMode {
			for _, typeInfo := range types.NodeTypeSlice {
				if typeInfo.Name == types.NormalNode {
					items = append(items, typeInfo.Alias)
				}
			}
		}
		if isUnhealthy {
			items = append(items, "0")
		}
		PrintLine(w, items...)
	}
	return 0, 0, 0
}

func displayNormalNodeCustomSummary(w *tabwriter.Writer, nodes []Node) {
	if len(nodes) == 0 {
		return
	}
	header := []string{"NAME", "IPADDRESS", "ROLE", "STATUS", "GPU(Total)", "GPU(Allocated)"}
	PrintLine(w, header...)
	for _, node := range nodes {
		nodeInfo := node.Convert2NodeInfo().(types.NormalNodeInfo)
		items := []string{}
		items = append(items, node.Name())
		items = append(items, node.IP())
		role := nodeInfo.Role
		if role == "" {
			role = "<none>"
		}
		items = append(items, role)
		items = append(items, node.Status())
		items = append(items, "0")
		items = append(items, "0")
		PrintLine(w, items...)
	}
}

func NewNormalNodeProcesser() NodeProcesser {
	return &nodeProcesser{
		nodeType:                  types.NormalNode,
		key:                       "normalNodes",
		builder:                   NewNormalNode,
		canBuildNode:              func(node *corev1.Node) bool { return true },
		displayNodesDetails:       displayNormalNodeDetails,
		displayNodesSummary:       displayNormalNodeSummary,
		displayNodesCustomSummary: displayNormalNodeCustomSummary,
	}
}
