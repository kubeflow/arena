package arenaclient

import (
	"fmt"
	"strings"
	"time"

	"github.com/kubeflow/arena/pkg/apis/config"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/apis/utils"
	"github.com/kubeflow/arena/pkg/topnode"
	log "github.com/sirupsen/logrus"
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
func (t *NodeClient) Details(nodeNames []string, nodeType types.NodeType, showMetric bool) (types.AllNodeInfo, error) {
	return topnode.ListNodeDetails(nodeNames, nodeType, showMetric)
}

//  ListAndPrintNodes is used to display nodes informations
func (t *NodeClient) ListAndPrintNodes(nodeNames []string, nodeType types.NodeType, format types.FormatStyle, details bool, notStop bool, showMetric bool) error {
	if format == types.UnknownFormat {
		return fmt.Errorf("Unknown output format,only support:[wide|json|yaml]")
	}
	if nodeType == types.UnknownNode {
		return fmt.Errorf("unknown node type,only supports:[%v]", strings.Join(utils.GetSupportedNodeTypes(), "|"))
	}
	if details {
		if !notStop {
			return topnode.DisplayNodeDetails(nodeNames, nodeType, format, showMetric)
		}
		if len(nodeNames) != 1 {
			return fmt.Errorf("must specify only one node name when '-r' is enabled")
		}
		for {
			err := topnode.DisplayNodeDetails(nodeNames, nodeType, format, showMetric)
			if err != nil {
				log.Errorf("failed to display node details,reason: %v", err)
			}
			t := time.Now()
			line := "------------------------- %v -------------------------------------"
			fmt.Printf(line+"\n", t.Format("2006-01-02 15:04:05"))
			time.Sleep(2 * time.Second)
		}
	}
	return topnode.DisplayNodeSummary(nodeNames, nodeType, format, showMetric)
}
