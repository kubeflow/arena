package topnode

import (
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/kubeflow/arena/pkg/apis/config"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/apis/utils"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

// NodeProcesser process the node
type NodeProcesser interface {
	// BuildNode builds the nodes and return the skip nodes
	BuildNode(v1nodes *v1.Node, pods []*v1.Pod, nodes []Node, targetNodeType types.NodeType, index int, args ...interface{}) ([]Node, bool)
	// Convert2NodeInfos filters nodes
	Convert2NodeInfos(nodes []Node, allNodes types.AllNodeInfo) types.AllNodeInfo
	// DisplayNodesDetails display nodes which the processer knowns
	DisplayNodesDetails(w *tabwriter.Writer, nodes []Node)
	// DisplayNodesSummary display nodes summary
	DisplayNodesSummary(w *tabwriter.Writer, nodes []Node, showNodeType, isUnhealthy bool)
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
	GetV1Pods() []*v1.Pod
	// GetV1Node returns the v1.node
	GetV1Node() *v1.Node
	// Capacity returns the total resource
	CapacityResourceCount() int
	// Allocatable returns the allocatable resource
	AllocatableResourceCount() int
	// UsedResourceCount returns the used resource count
	UsedResourceCount() int
	// Convert2NodeInfo convert node to node info
	Convert2NodeInfo() interface{}
	// WideFormat is used to display node information with wide format
	WideFormat() string
	// IsHealthy return the node is healthy or not
	IsHealthy() bool
}

const (
	k8sNodeRoleLabelKeyPrefix = "node-role.kubernetes.io/"
	nodeLabelRole             = "kubernetes.io/role"
)

type baseNode struct {
	index    int
	pods     []*v1.Pod
	node     *v1.Node
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

func (b *baseNode) GetV1Pods() []*v1.Pod {
	return b.pods
}

func (b *baseNode) GetV1Node() *v1.Node {
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
	key                 string
	nodeType            types.NodeType
	builder             func(node *v1.Node, pods []*v1.Pod, index int, args ...interface{}) (Node, error)
	canBuildNode        func(node *v1.Node) bool
	displayNodesDetails func(w *tabwriter.Writer, nodes []Node)
	displayNodesSummary func(w *tabwriter.Writer, nodes []Node, isUnhealthy, showNodeType bool)
}

func (n *nodeProcesser) BuildNode(v1node *v1.Node, pods []*v1.Pod, nodes []Node, targetNodeType types.NodeType, index int, args ...interface{}) ([]Node, bool) {
	skip := true
	if !isNeededNodeType(n.nodeType, targetNodeType) || !n.canBuildNode(v1node) {
		return nodes, skip
	}
	myNode, err := n.builder(v1node, pods, index, args...)
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

func (n *nodeProcesser) DisplayNodesSummary(w *tabwriter.Writer, nodes []Node, showNodeType, isUnhealthy bool) {
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
	n.displayNodesSummary(w, myNodes, isUnhealthy, showNodeType)
}

func (n *nodeProcesser) SupportedNodeType() types.NodeType {
	return n.nodeType
}

func BuildNodes(nodeNames []string, targetNodeType types.NodeType) ([]Node, error) {
	client := config.GetArenaConfiger().GetClientSet()
	allPods, err := utils.AcquireAllActivePods(client)
	if err != nil {
		return nil, err
	}
	nodeList, err := client.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	configMapList, err := client.CoreV1().ConfigMaps("kube-system").List(metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ListOptions",
			APIVersion: "v1",
		}, LabelSelector: types.GPUTopologyNodeLabels,
	})
	if err != nil {
		return nil, err
	}
	configmaps := []*v1.ConfigMap{}
	for _, c := range configMapList.Items {
		configmaps = append(configmaps, c.DeepCopy())
	}
	names := map[string]bool{}
	for _, name := range nodeNames {
		names[name] = true
	}
	nodes := []Node{}
	for index, n := range nodeList.Items {
		node := n.DeepCopy()
		pods := []*v1.Pod{}
		if !filterNode(names, node.Name) {
			continue
		}
		for _, pod := range allPods {
			if pod.Spec.NodeName != node.Name {
				continue
			}
			pods = append(pods, pod)
		}
		for _, processer := range GetSupportedNodePorcessers() {
			var skip bool
			nodes, skip = processer.BuildNode(node, pods, nodes, targetNodeType, index, configmaps)
			if !skip {
				log.Debugf("the processer %v process the node %v", processer.SupportedNodeType(), node.Name)
				break
			}
			log.Debugf("the processer %v skips to process the node %v", processer.SupportedNodeType(), node.Name)
		}
	}
	return nodes, nil
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
