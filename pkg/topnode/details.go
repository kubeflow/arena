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

func ListNodeDetails(nodeNames []string, nodeType types.NodeType) (types.AllNodeInfo, error) {
	nodes, err := BuildNodes(nodeNames, nodeType)
	allNodeInfos := types.AllNodeInfo{}
	if err != nil {
		return allNodeInfos, err
	}
	for _, processer := range GetSupportedNodePorcessers() {
		allNodeInfos = processer.Convert2NodeInfos(nodes, allNodeInfos)
	}
	return allNodeInfos, nil
}

func DisplayNodeDetails(nodeNames []string, nodeType types.NodeType, format types.FormatStyle) error {
	nodes, err := BuildNodes(nodeNames, nodeType)
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
