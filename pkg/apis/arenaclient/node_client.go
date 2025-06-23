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

// ListAndPrintNodes is used to display nodes informations
func (t *NodeClient) ListAndPrintNodes(nodeNames []string, nodeType types.NodeType, format types.FormatStyle, details bool, notStop bool, showMetric bool) error {
	if format == types.UnknownFormat {
		return fmt.Errorf("unknown output format,only support:[wide|json|yaml]")
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
