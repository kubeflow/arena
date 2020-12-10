package arenaclient

import (
	"fmt"
	"strings"

	"github.com/kubeflow/arena/pkg/apis/config"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/apis/utils"
	"github.com/kubeflow/arena/pkg/topnode"
)

type NodeClient struct {
	namespace string
	configer  *config.ArenaConfiger
}

// NewNodeClient creates a ServingJobClient
func NewNodeClient(namespace string, configer *config.ArenaConfiger) *NodeClient {
	return &NodeClient{
		namespace: namespace,
		configer:  configer,
	}
}

// Namespace sets the namespace,this operation does not change the default namespace
func (t *NodeClient) Namespace(namespace string) *NodeClient {
	copyNodeClient := &NodeClient{
		namespace: namespace,
		configer:  t.configer,
	}
	return copyNodeClient
}

// Details is used to serve api
func (t *NodeClient) Details(nodeNames []string, nodeType types.NodeType) (types.AllNodeInfo, error) {
	return topnode.ListNodeDetails(nodeNames, nodeType)
}

//  ListAndPrintNodes is used to display nodes informations
func (t *NodeClient) ListAndPrintNodes(nodeNames []string, nodeType types.NodeType, format types.FormatStyle, details bool) error {
	if format == types.UnknownFormat {
		return fmt.Errorf("Unknown output format,only support:[wide|json|yaml]")
	}
	if nodeType == types.UnknownNode {
		return fmt.Errorf("unknown node type,only supports:[%v]", strings.Join(utils.GetSupportedNodeTypes(), "|"))
	}
	if details {
		return topnode.DisplayNodeDetails(nodeNames, nodeType, format)
	}
	return topnode.DisplayNodeSummary(nodeNames, nodeType, format)
}
