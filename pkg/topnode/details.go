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
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/kubeflow/arena/pkg/apis/types"
	"gopkg.in/yaml.v2"
)

func ListNodeDetails(nodeNames []string, nodeType types.NodeType, showMetric bool) (types.AllNodeInfo, error) {
	nodes, err := BuildNodes(nodeNames, nodeType, showMetric)
	allNodeInfos := types.AllNodeInfo{}
	if err != nil {
		return allNodeInfos, err
	}
	for _, processer := range GetSupportedNodePorcessers() {
		allNodeInfos = processer.Convert2NodeInfos(nodes, allNodeInfos)
	}
	return allNodeInfos, nil
}

func DisplayNodeDetails(nodeNames []string, nodeType types.NodeType, format types.FormatStyle, showMetric bool) error {
	nodes, err := BuildNodes(nodeNames, nodeType, showMetric)
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
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	processers := GetSupportedNodePorcessers()
	for i := len(processers) - 1; i >= 0; i-- {
		processer := processers[i]
		processer.DisplayNodesDetails(w, nodes)
	}
	_ = w.Flush()
	return nil
}

func PrintLine(w io.Writer, fields ...string) {
	buffer := strings.Join(fields, "\t")
	fmt.Fprintln(w, buffer)
}

func isNeededNodeType(nodeType, targetType types.NodeType) bool {
	if targetType == types.AllKnownNode {
		return true
	}
	return nodeType == targetType
}
