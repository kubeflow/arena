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
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/kubeflow/arena/pkg/k8saccesser"

	"github.com/kubeflow/arena/pkg/apis/config"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/apis/utils"
	"github.com/kubeflow/arena/pkg/prometheus"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/kubernetes"
)

type buildNodeArgs struct {
	pods           []*corev1.Pod
	configmaps     []*corev1.ConfigMap
	nodeGPUMetrics map[string]types.NodeGpuMetric
}

// NodeProcesser process the node
type NodeProcesser interface {
	// BuildNode builds the nodes and return the skip nodes
	BuildNode(client *kubernetes.Clientset, v1nodes *corev1.Node, nodes []Node, targetNodeType types.NodeType, index int, args buildNodeArgs) ([]Node, bool)
	// Convert2NodeInfos filters nodes
	Convert2NodeInfos(nodes []Node, allNodes types.AllNodeInfo) types.AllNodeInfo
	// DisplayNodesDetails display nodes which the processer knowns
	DisplayNodesDetails(w *tabwriter.Writer, nodes []Node)
	// DisplayNodesSummary display nodes summary
	DisplayNodesSummary(w *tabwriter.Writer, nodes []Node, showNodeType, isUnhealthy bool) (float64, float64, float64)
	// DisplayNodesCustomSummary display custom format of target type nodes
	DisplayNodesCustomSummary(w *tabwriter.Writer, nodes []Node)
	// SupportedNodeType Type returns the supported node type
	SupportedNodeType() types.NodeType
}

type Node interface {
	// Index is used to sort the nodes
	Index() int
	// Name return the node name
	Name() string
	// Type return the node type
	Type() types.NodeType
	// Role returns the role of node
	Role() []string
	// IP returns the node ip
	IP() string
	// Status returns the node status
	Status() string
	// GetV1Pods returns the pods of node
	GetV1Pods() []*corev1.Pod
	// GetV1Node returns the v1.node
	GetV1Node() *corev1.Node
	// Convert2NodeInfo convert node to node info
	Convert2NodeInfo() interface{}
	// AllDevicesAreHealthy returns the all devices are healthy
	AllDevicesAreHealthy() bool
	// WideFormat is used to display node information with wide format
	WideFormat() string
}

const (
	k8sNodeRoleLabelKeyPrefix = "node-role.kubernetes.io/"
	nodeLabelRole             = "kubernetes.io/role"
)

type baseNode struct {
	index    int
	pods     []*corev1.Pod
	node     *corev1.Node
	nodeType types.NodeType
}

func (b *baseNode) Index() int {
	return b.index
}

func (b *baseNode) Name() string {
	return b.node.Name
}

// findNodeRoles returns the roles of a given node.
// The roles are determined by looking for:
// * a node-role.kubernetes.io/<role>="" label
// * a kubernetes.io/role="<role>" label
func (b *baseNode) Role() []string {
	roles := sets.NewString()
	for k, v := range b.node.Labels {
		switch {
		case strings.HasPrefix(k, k8sNodeRoleLabelKeyPrefix):
			if role := strings.TrimPrefix(k, k8sNodeRoleLabelKeyPrefix); len(role) > 0 {
				roles.Insert(role)
			}
		case k == nodeLabelRole && v != "":
			roles.Insert(v)
		}
	}
	return roles.List()
}

// IP returns the node ip
func (b *baseNode) IP() string {
	if len(b.node.Status.Addresses) == 0 {
		return "N/A"
	}
	address := b.node.Status.Addresses[0]
	return address.Address
}

// Status return the node status
func (b *baseNode) Status() string {
	return utils.DefineNodeStatus(b.node)
}

func (b *baseNode) GetV1Pods() []*corev1.Pod {
	return b.pods
}

func (b *baseNode) GetV1Node() *corev1.Node {
	return b.node
}

func (b *baseNode) Type() types.NodeType {
	return b.nodeType
}

// NewNormalNodeProcesser must be placed at last,it will match all unknown nodes
func GetSupportedNodePorcessers() []NodeProcesser {
	return []NodeProcesser{
		NewGPUShareNodeProcesser(),
		NewGPUTopologyNodeProcesser(),
		NewGPUExclusiveNodeProcesser(),
		NewNormalNodeProcesser(),
	}
}

type nodeProcesser struct {
	key                       string
	nodeType                  types.NodeType
	builder                   func(client *kubernetes.Clientset, node *corev1.Node, index int, args buildNodeArgs) (Node, error)
	canBuildNode              func(node *corev1.Node) bool
	displayNodesDetails       func(w *tabwriter.Writer, nodes []Node)
	displayNodesSummary       func(w *tabwriter.Writer, nodes []Node, isUnhealthy, showNodeType bool) (float64, float64, float64)
	displayNodesCustomSummary func(w *tabwriter.Writer, nodes []Node)
}

func (n *nodeProcesser) BuildNode(client *kubernetes.Clientset, v1node *corev1.Node, nodes []Node, targetNodeType types.NodeType, index int, args buildNodeArgs) ([]Node, bool) {
	skip := true
	if !isNeededNodeType(n.nodeType, targetNodeType) || !n.canBuildNode(v1node) {
		return nodes, skip
	}
	myNode, err := n.builder(client, v1node, index, args)
	if err != nil {
		log.Debugf("failed to build node: %v", err)
		return nodes, skip
	}
	nodes = append(nodes, myNode)
	return nodes, !skip
}

func (n *nodeProcesser) Convert2NodeInfos(nodes []Node, allNodes types.AllNodeInfo) types.AllNodeInfo {
	myNodes := []interface{}{}
	for _, node := range nodes {
		if node.Type() != n.nodeType {
			continue
		}
		myNodes = append(myNodes, node.Convert2NodeInfo())
	}
	allNodes[n.key] = myNodes
	return allNodes
}

func (n *nodeProcesser) DisplayNodesDetails(w *tabwriter.Writer, nodes []Node) {
	myNodes := []Node{}
	for _, node := range nodes {
		if node.Type() != n.nodeType {
			continue
		}
		myNodes = append(myNodes, node)
	}
	sort.Slice(myNodes, func(i, j int) bool {
		return myNodes[i].Index() < myNodes[j].Index()
	})
	n.displayNodesDetails(w, myNodes)
}

func (n *nodeProcesser) DisplayNodesSummary(w *tabwriter.Writer, nodes []Node, showNodeType, isUnhealthy bool) (float64, float64, float64) {
	totalGPUs := float64(0)
	allocatedGPUs := float64(0)
	unhealthyGPUs := float64(0)
	myNodes := []Node{}
	for _, node := range nodes {
		if node.Type() != n.nodeType {
			continue
		}
		myNodes = append(myNodes, node)
	}
	sort.Slice(myNodes, func(i, j int) bool {
		return myNodes[i].Index() < myNodes[j].Index()
	})
	t, a, u := n.displayNodesSummary(w, myNodes, isUnhealthy, showNodeType)
	totalGPUs += t
	allocatedGPUs += a
	unhealthyGPUs += u
	return totalGPUs, allocatedGPUs, unhealthyGPUs
}

func (n *nodeProcesser) DisplayNodesCustomSummary(w *tabwriter.Writer, nodes []Node) {
	myNodes := []Node{}
	for _, node := range nodes {
		if node.Type() != n.nodeType {
			continue
		}
		myNodes = append(myNodes, node)
	}
	sort.Slice(myNodes, func(i, j int) bool {
		return myNodes[i].Index() < myNodes[j].Index()
	})
	n.displayNodesCustomSummary(w, myNodes)
}

func (n *nodeProcesser) SupportedNodeType() types.NodeType {
	return n.nodeType
}

func BuildNodes(nodeNames []string, targetNodeType types.NodeType, showMetric bool) ([]Node, error) {
	client := config.GetArenaConfiger().GetClientSet()
	allNodes, err := k8saccesser.GetK8sResourceAccesser().ListNodes("")
	if err != nil {
		return nil, err
	}
	configmaps, err := k8saccesser.GetK8sResourceAccesser().ListConfigMaps("kube-system", types.GPUTopologyNodeLabels)
	if err != nil {
		return nil, err
	}
	names := map[string]bool{}
	for _, name := range nodeNames {
		names[name] = true
	}
	nodeGPUMetrics := map[string]types.NodeGpuMetric{}
	if showMetric {
		nodeGPUMetrics, err = GetNodeGpuMetrics(client)
		if err != nil {
			log.Debugf("failed to get node metrics: %v", err)
			nodeGPUMetrics = map[string]types.NodeGpuMetric{}
		}
	}
	pods, err := listRunningPods()
	if err != nil {
		log.Errorf("failed to list active running pods,reason: %v", err)
		return nil, err
	}
	nodes := []Node{}
	args := buildNodeArgs{
		configmaps:     configmaps,
		nodeGPUMetrics: nodeGPUMetrics,
		pods:           pods,
	}
	for index, n := range allNodes {
		node := n.DeepCopy()
		if !filterNode(names, node.Name) {
			continue
		}
		for _, processer := range GetSupportedNodePorcessers() {
			var skip bool
			nodes, skip = processer.BuildNode(client, node, nodes, targetNodeType, index, args)
			if !skip {
				log.Debugf("the processer %v process the node %v", processer.SupportedNodeType(), node.Name)
				break
			}
			log.Debugf("the processer %v skips to process the node %v", processer.SupportedNodeType(), node.Name)
		}
	}
	if len(nodes) == 0 {
		return nil, fmt.Errorf("failed to display nodes's informations: not found nodes")
	}
	return nodes, nil
}

func listRunningPods() ([]*corev1.Pod, error) {
	labelSelector := fmt.Sprintf("status.phase!=%v,status.phase!=%v", corev1.PodFailed, corev1.PodSucceeded)
	var filterFunc = func(pod *corev1.Pod) bool {
		if pod.Status.Phase == corev1.PodFailed || pod.Status.Phase == corev1.PodSucceeded {
			return false
		}
		return true
	}
	pods, err := k8saccesser.GetK8sResourceAccesser().ListPods("", "", labelSelector, filterFunc)
	if err != nil {
		return nil, err
	}
	return pods, nil
}
func filterNode(names map[string]bool, nodeName string) bool {
	if len(names) == 0 {
		return true
	}
	if _, ok := names[nodeName]; ok {
		return true
	}
	return false
}

func GetNodeGpuMetrics(client *kubernetes.Clientset) (map[string]types.NodeGpuMetric, error) {
	return prometheus.GetNodeGPUMetrics(client, []string{".*"})
}

func getGPUMetricsByNodeName(nodeName string, metrics map[string]types.NodeGpuMetric) types.NodeGpuMetric {
	result := types.NodeGpuMetric{}
	if metrics == nil || metrics[nodeName] == nil {
		return result
	}
	return metrics[nodeName]
}

func getNodePods(node *corev1.Node, pods []*corev1.Pod) []*corev1.Pod {
	nodePods := []*corev1.Pod{}
	for _, pod := range pods {
		if pod.Spec.NodeName != node.Name {
			continue
		}
		nodePods = append(nodePods, pod)
	}
	return nodePods
}
